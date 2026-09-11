package redisstream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/appleboy/com/bytesconv"
	"github.com/golang-queue/queue"
	"github.com/golang-queue/queue/core"
	"github.com/golang-queue/queue/job"
	"github.com/redis/go-redis/v9"
)

var _ core.Worker = (*Worker)(nil)

// Worker for Redis
type Worker struct {
	// redis config
	rdb       redis.Cmdable
	tasks     chan redis.XMessage
	stopFlag  int32
	stopOnce  sync.Once
	startOnce sync.Once
	stop      chan struct{}
	exit      chan struct{}
	opts      options
}

// NewWorker for struc
func NewWorker(opts ...Option) *Worker {
	var err error
	w := &Worker{
		opts:  newOptions(opts...),
		stop:  make(chan struct{}),
		exit:  make(chan struct{}),
		tasks: make(chan redis.XMessage),
	}

	if w.opts.connectionString != "" {
		options, err := redis.ParseURL(w.opts.connectionString)
		if err != nil {
			w.opts.logger.Fatal(err)
		}
		w.rdb = redis.NewClient(options)
	} else if w.opts.addr != "" {
		if w.opts.cluster {
			w.rdb = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:     strings.Split(w.opts.addr, ","),
				Username:  w.opts.username,
				Password:  w.opts.password,
				TLSConfig: w.opts.tls,
			})
		} else {
			options := &redis.Options{
				Addr:      w.opts.addr,
				Username:  w.opts.username,
				Password:  w.opts.password,
				DB:        w.opts.db,
				TLSConfig: w.opts.tls,
			}
			w.rdb = redis.NewClient(options)
		}
	}

	_, err = w.rdb.Ping(context.Background()).Result()
	if err != nil {
		w.opts.logger.Fatal(err)
	}

	return w
}

func (w *Worker) startConsumer() {
	w.startOnce.Do(func() {
		if err := w.ensureGroup(context.Background()); err != nil {
			w.opts.logger.Error(err)
		}

		go w.fetchTask()
	})
}

// ensureGroup creates the stream and its consumer group. BUSYGROUP just means
// another worker got there first, which is the normal case on restart.
func (w *Worker) ensureGroup(ctx context.Context) error {
	err := w.rdb.XGroupCreateMkStream(ctx, w.opts.streamName, w.opts.group, "$").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("create consumer group %s on stream %s: %w", w.opts.group, w.opts.streamName, err)
	}

	return nil
}

func (w *Worker) fetchTask() {
	for {
		select {
		case <-w.stop:
			return
		default:
		}

		ctx := context.Background()
		data, err := w.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    w.opts.group,
			Consumer: w.opts.consumer,
			Streams:  []string{w.opts.streamName, ">"},
			// count is number of entries we want to read from redis
			Count: 1,
			// we use the block command to make sure if no entry is found we wait
			// until an entry is found
			Block: w.opts.blockTime,
		}).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			// NOGROUP means the stream key is gone (eviction, FLUSHDB, trimmed
			// away) and took the group with it. XREADGROUP then returns
			// immediately, so without recreating the group this loop spins and
			// floods the log.
			if strings.Contains(err.Error(), "NOGROUP") {
				if errGroup := w.ensureGroup(ctx); errGroup == nil {
					continue
				} else {
					w.opts.logger.Error(errGroup)
				}
			} else {
				w.opts.logger.Errorf("error while reading from redis %v", err)
			}

			// Back off: the failing call did not block, so retrying straight
			// away is a busy loop.
			select {
			case <-w.stop:
				return
			case <-time.After(w.opts.blockTime):
			}
			continue
		}
		// we have received the data we should loop it and queue the messages
		// so that our tasks can start processing
		for _, result := range data {
			for _, message := range result.Messages {
				select {
				case w.tasks <- message:
					if err := w.rdb.XAck(ctx, w.opts.streamName, w.opts.group, message.ID).Err(); err != nil {
						w.opts.logger.Errorf("can't ack message: %s", message.ID)
					}
				case <-w.stop:
					// Todo: re-queue the task
					w.opts.logger.Info("re-queue the task: ", message.ID)
					if err := w.queue(message.Values); err != nil {
						w.opts.logger.Error("error to re-queue the task: ", message.ID)
					}
					close(w.exit)
					return
				}
			}
		}
	}
}

// Shutdown worker
func (w *Worker) Shutdown() error {
	if !atomic.CompareAndSwapInt32(&w.stopFlag, 0, 1) {
		return queue.ErrQueueShutdown
	}

	w.stopOnce.Do(func() {
		close(w.stop)

		// wait requeue
		select {
		case <-w.exit:
		case <-time.After(200 * time.Millisecond):
		}

		switch v := w.rdb.(type) {
		case *redis.Client:
			v.Close()
		case *redis.ClusterClient:
			v.Close()
		}
		close(w.tasks)
	})
	return nil
}

func (w *Worker) queue(data any) error {
	ctx := context.Background()

	// Publish a message.
	err := w.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: w.opts.streamName,
		MaxLen: w.opts.maxLength,
		Values: data,
	}).Err()

	return err
}

// Queue send notification to queue
func (w *Worker) Queue(task core.TaskMessage) error {
	if atomic.LoadInt32(&w.stopFlag) == 1 {
		return queue.ErrQueueShutdown
	}

	return w.queue(map[string]any{"body": bytesconv.BytesToStr(task.Bytes())})
}

// Run start the worker
func (w *Worker) Run(ctx context.Context, task core.TaskMessage) error {
	return w.opts.runFunc(ctx, task)
}

// Request a new task
func (w *Worker) Request() (core.TaskMessage, error) {
	clock := 0
	w.startConsumer()
loop:
	for {
		select {
		case task, ok := <-w.tasks:
			if !ok {
				return nil, queue.ErrQueueHasBeenClosed
			}
			body, ok := task.Values["body"].(string)
			if !ok {
				return nil, fmt.Errorf("redis stream message %q has no string body", task.ID)
			}

			var data job.Message
			if err := json.Unmarshal(bytesconv.StrToBytes(body), &data); err != nil {
				return nil, fmt.Errorf("unmarshal redis stream message %q: %w", task.ID, err)
			}
			return &data, nil
		case <-time.After(1 * time.Second):
			if clock == 5 {
				break loop
			}
			clock += 1
		}
	}

	return nil, queue.ErrNoTaskInQueue
}
