package redisstream

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-queue/queue"
	"github.com/golang-queue/queue/core"
	"github.com/golang-queue/queue/job"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/goleak"

	"github.com/trinhdaiphuc/go-kit/tracing"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func setupRedisCluserContainer(ctx context.Context, t *testing.T) (testcontainers.Container, string) {
	req := testcontainers.ContainerRequest{
		Image: "vishnunair/docker-redis-cluster:latest",
		ExposedPorts: []string{
			"6379/tcp",
			"6380/tcp",
			"6381/tcp",
			"6382/tcp",
			"6383/tcp",
			"6384/tcp",
		},
		WaitingFor: wait.NewExecStrategy(
			[]string{"redis-cli", "-h", "localhost", "-p", "6379", "cluster", "info"},
		),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	endpoint, err := redisC.Endpoint(ctx, "")
	require.NoError(t, err)

	return redisC, endpoint
}

func setupRedisContainer(ctx context.Context, t *testing.T) (testcontainers.Container, string) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:6",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor: wait.NewExecStrategy(
			[]string{"redis-cli", "-h", "localhost", "-p", "6379", "ping"},
		),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	endpoint, err := redisC.Endpoint(ctx, "")
	require.NoError(t, err)

	return redisC, endpoint
}

func TestWithRedis(t *testing.T) {
	ctx := context.Background()
	redisC, _ := setupRedisContainer(ctx, t)
	testcontainers.CleanupContainer(t, redisC)
}

type mockMessage struct {
	Message string
}

func (m mockMessage) Bytes() []byte {
	return []byte(m.Message)
}

func TestRedisDefaultFlow(t *testing.T) {
	ctx := context.Background()
	redisC, endpoint := setupRedisContainer(ctx, t)
	defer testcontainers.CleanupContainer(t, redisC)

	m := &mockMessage{
		Message: "foo",
	}
	w := NewWorker(
		WithAddr(endpoint),
		WithStreamName("test"),
	)
	q, err := queue.NewQueue(
		queue.WithWorker(w),
		queue.WithWorkerCount(2),
	)
	assert.NoError(t, err)
	assert.NoError(t, q.Queue(m))
	q.Start()
	time.Sleep(100 * time.Millisecond)
	q.Release()
}

func TestRedisShutdown(t *testing.T) {
	ctx := context.Background()
	redisC, endpoint := setupRedisContainer(ctx, t)
	defer testcontainers.CleanupContainer(t, redisC)
	w := NewWorker(
		WithAddr(endpoint),
		WithStreamName("test2"),
	)
	q, err := queue.NewQueue(
		queue.WithWorker(w),
		queue.WithWorkerCount(2),
	)
	assert.NoError(t, err)
	q.Start()
	time.Sleep(1 * time.Second)
	q.Shutdown()
	// check shutdown once
	assert.Error(t, w.Shutdown())
	assert.Equal(t, queue.ErrQueueShutdown, w.Shutdown())
	q.Wait()
}

func TestCustomFuncAndWait(t *testing.T) {
	ctx := context.Background()
	redisC, endpoint := setupRedisContainer(ctx, t)
	defer testcontainers.CleanupContainer(t, redisC)
	m := &mockMessage{
		Message: "foo",
	}
	w := NewWorker(
		WithAddr(endpoint),
		WithStreamName("test3"),
		WithRunFunc(func(ctx context.Context, m core.TaskMessage) error {
			time.Sleep(500 * time.Millisecond)
			return nil
		}),
	)
	q := queue.NewPool(
		5,
		queue.WithWorker(w),
	)
	time.Sleep(100 * time.Millisecond)
	assert.NoError(t, q.Queue(m))
	assert.NoError(t, q.Queue(m))
	assert.NoError(t, q.Queue(m))
	assert.NoError(t, q.Queue(m))
	time.Sleep(1000 * time.Millisecond)
	q.Release()
	// you will see the execute time > 1000ms
}

func TestRedisCluster(t *testing.T) {
	t.Skip()

	ctx := context.Background()
	redisC, _ := setupRedisCluserContainer(ctx, t)
	defer testcontainers.CleanupContainer(t, redisC)

	masterPort, err := redisC.MappedPort(ctx, "6379")
	assert.NoError(t, err)

	slavePort, err := redisC.MappedPort(ctx, "6382")
	assert.NoError(t, err)

	hostIP, err := redisC.Host(ctx)
	assert.NoError(t, err)

	m := &mockMessage{
		Message: "foo",
	}

	masterName := fmt.Sprintf("%s:%s", hostIP, masterPort.Port())
	slaveName := fmt.Sprintf("%s:%s", hostIP, slavePort.Port())

	hosts := []string{masterName, slaveName}

	w := NewWorker(
		WithAddr(strings.Join(hosts, ",")),
		WithStreamName("testCluster"),
		WithCluster(),
		WithRunFunc(func(ctx context.Context, m core.TaskMessage) error {
			time.Sleep(500 * time.Millisecond)
			return nil
		}),
	)
	q := queue.NewPool(
		5,
		queue.WithWorker(w),
	)
	time.Sleep(100 * time.Millisecond)
	assert.NoError(t, q.Queue(m))
	assert.NoError(t, q.Queue(m))
	assert.NoError(t, q.Queue(m))
	assert.NoError(t, q.Queue(m))
	time.Sleep(1000 * time.Millisecond)
	q.Release()
	// you will see the execute time > 1000ms
}

func TestEnqueueJobAfterShutdown(t *testing.T) {
	ctx := context.Background()
	redisC, endpoint := setupRedisContainer(ctx, t)
	defer testcontainers.CleanupContainer(t, redisC)
	m := mockMessage{
		Message: "foo",
	}
	w := NewWorker(
		WithAddr(endpoint),
	)
	q, err := queue.NewQueue(
		queue.WithWorker(w),
		queue.WithWorkerCount(2),
	)
	assert.NoError(t, err)
	q.Start()
	time.Sleep(50 * time.Millisecond)
	q.Shutdown()
	// can't queue task after shutdown
	err = q.Queue(m)
	assert.Error(t, err)
	assert.Equal(t, queue.ErrQueueShutdown, err)
	q.Wait()
}

func TestJobReachTimeout(t *testing.T) {
	ctx := context.Background()
	redisC, endpoint := setupRedisContainer(ctx, t)
	defer testcontainers.CleanupContainer(t, redisC)
	m := mockMessage{
		Message: "foo",
	}
	w := NewWorker(
		WithAddr(endpoint),
		WithStreamName("timeout"),
		WithRunFunc(func(ctx context.Context, m core.TaskMessage) error {
			for {
				select {
				case <-ctx.Done():
					log.Println("get data:", string(m.Payload()))
					if errors.Is(ctx.Err(), context.Canceled) {
						log.Println("queue has been shutdown and cancel the job")
					} else if errors.Is(ctx.Err(), context.DeadlineExceeded) {
						log.Println("job deadline exceeded")
					}
					return nil
				default:
				}
				time.Sleep(50 * time.Millisecond)
			}
		}),
	)
	q, err := queue.NewQueue(
		queue.WithWorker(w),
		queue.WithWorkerCount(2),
	)
	assert.NoError(t, err)
	q.Start()
	time.Sleep(50 * time.Millisecond)
	assert.NoError(t, q.Queue(m, job.AllowOption{
		Timeout: job.Time(20 * time.Millisecond),
	}))
	time.Sleep(2 * time.Second)
	q.Shutdown()
	q.Wait()
}

func TestCancelJobAfterShutdown(t *testing.T) {
	ctx := context.Background()
	redisC, endpoint := setupRedisContainer(ctx, t)
	defer testcontainers.CleanupContainer(t, redisC)
	m := mockMessage{
		Message: "test",
	}
	w := NewWorker(
		WithAddr(endpoint),
		WithStreamName("cancel"),
		WithLogger(queue.NewLogger()),
		WithRunFunc(func(ctx context.Context, m core.TaskMessage) error {
			for {
				select {
				case <-ctx.Done():
					log.Println("get data:", string(m.Payload()))
					if errors.Is(ctx.Err(), context.Canceled) {
						log.Println("queue has been shutdown and cancel the job")
					} else if errors.Is(ctx.Err(), context.DeadlineExceeded) {
						log.Println("job deadline exceeded")
					}
					return nil
				default:
				}
				time.Sleep(50 * time.Millisecond)
			}
		}),
	)
	q, err := queue.NewQueue(
		queue.WithWorker(w),
		queue.WithWorkerCount(2),
	)
	assert.NoError(t, err)
	q.Start()
	time.Sleep(50 * time.Millisecond)
	assert.NoError(t, q.Queue(m, job.AllowOption{
		Timeout: job.Time(3 * time.Second),
	}))
	time.Sleep(2 * time.Second)
	q.Shutdown()
	q.Wait()
}

func TestGoroutineLeak(t *testing.T) {
	ctx := context.Background()
	redisC, endpoint := setupRedisContainer(ctx, t)
	defer testcontainers.CleanupContainer(t, redisC)
	m := mockMessage{
		Message: "foo",
	}
	w := NewWorker(
		WithAddr(endpoint),
		WithStreamName("GoroutineLeak"),
		WithLogger(queue.NewEmptyLogger()),
		WithRunFunc(func(ctx context.Context, m core.TaskMessage) error {
			for {
				select {
				case <-ctx.Done():
					log.Println("get data:", string(m.Payload()))
					if errors.Is(ctx.Err(), context.Canceled) {
						log.Println("queue has been shutdown and cancel the job")
					} else if errors.Is(ctx.Err(), context.DeadlineExceeded) {
						log.Println("job deadline exceeded")
					}
					return nil
				default:
					log.Println("get data:", string(m.Payload()))
					time.Sleep(50 * time.Millisecond)
					return nil
				}
			}
		}),
	)
	q, err := queue.NewQueue(
		queue.WithLogger(queue.NewEmptyLogger()),
		queue.WithWorker(w),
		queue.WithWorkerCount(10),
	)
	assert.NoError(t, err)
	q.Start()
	time.Sleep(50 * time.Millisecond)
	for i := range 50 {
		m.Message = fmt.Sprintf("foobar: %d", i+1)
		assert.NoError(t, q.Queue(m))
	}
	time.Sleep(1 * time.Second)
	q.Release()
	time.Sleep(1 * time.Second)
	fmt.Println("number of goroutines:", runtime.NumGoroutine())
}

func TestGoroutinePanic(t *testing.T) {
	ctx := context.Background()
	redisC, endpoint := setupRedisContainer(ctx, t)
	defer testcontainers.CleanupContainer(t, redisC)
	m := mockMessage{
		Message: "foo",
	}
	w := NewWorker(
		WithAddr(endpoint),
		WithStreamName("GoroutinePanic"),
		WithRunFunc(func(ctx context.Context, m core.TaskMessage) error {
			panic("missing something")
		}),
	)
	q, err := queue.NewQueue(
		queue.WithWorker(w),
		queue.WithWorkerCount(2),
	)
	assert.NoError(t, err)
	q.Start()
	time.Sleep(50 * time.Millisecond)
	assert.NoError(t, q.Queue(m))
	assert.NoError(t, q.Queue(m))
	time.Sleep(200 * time.Millisecond)
	q.Shutdown()
	assert.Error(t, q.Queue(m))
	q.Wait()
}

// A stream entry written by something other than this worker has no "body"
// field; Request must report that instead of panicking on the type assertion.
func TestRequestRejectsMalformedMessage(t *testing.T) {
	ctx := context.Background()
	redisC, endpoint := setupRedisContainer(ctx, t)
	defer testcontainers.CleanupContainer(t, redisC)

	w := NewWorker(
		WithAddr(endpoint),
		WithStreamName("malformed"),
		WithBlockTime(100*time.Millisecond),
	)
	defer func() {
		assert.NoError(t, w.Shutdown())
	}()

	// Create the group first so the worker sees entries added after it.
	w.startConsumer()
	require.NoError(t, w.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "malformed",
		Values: map[string]any{"not_body": 1},
	}).Err())

	task, err := w.Request()
	assert.Nil(t, task)
	assert.ErrorContains(t, err, "no string body")
}

type countingLogger struct {
	mu     sync.Mutex
	errors []string
}

func (l *countingLogger) record(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errors = append(l.errors, msg)
}

func (l *countingLogger) errorCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.errors)
}

func (l *countingLogger) Infof(format string, args ...any)  {}
func (l *countingLogger) Info(args ...any)                  {}
func (l *countingLogger) Fatalf(format string, args ...any) {}
func (l *countingLogger) Fatal(args ...any)                 {}
func (l *countingLogger) Errorf(format string, args ...any) { l.record(fmt.Sprintf(format, args...)) }
func (l *countingLogger) Error(args ...any)                 { l.record(fmt.Sprint(args...)) }

// Losing the stream key (eviction, FLUSHDB, trim) drops the consumer group with
// it. XREADGROUP then fails instantly with NOGROUP, so the fetch loop used to
// spin and flood the log instead of recreating the group.
func TestFetchTaskRecoversFromNoGroup(t *testing.T) {
	ctx := context.Background()
	redisC, endpoint := setupRedisContainer(ctx, t)
	defer testcontainers.CleanupContainer(t, redisC)

	logger := &countingLogger{}
	w := NewWorker(
		WithAddr(endpoint),
		WithStreamName("nogroup"),
		WithBlockTime(100*time.Millisecond),
		WithLogger(logger),
	)
	defer func() {
		assert.NoError(t, w.Shutdown())
	}()

	w.startConsumer()

	// Drop the stream key: the group goes with it.
	require.NoError(t, w.rdb.Del(ctx, "nogroup").Err())
	time.Sleep(500 * time.Millisecond)

	assert.LessOrEqual(t, logger.errorCount(), 1, "NOGROUP must not be logged on every loop: %v", logger.errors)

	// The worker must have recreated the group and keep consuming.
	require.NoError(t, w.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "nogroup",
		Values: map[string]any{"body": `{"Body":"cGluZw=="}`},
	}).Err())
	task, err := w.Request()
	require.NoError(t, err)
	assert.NotNil(t, task)
}

// The header field must reach the run func through the context: queue.run()
// only accepts *job.Message, so it cannot ride on the task itself.
func TestRequestCarriesHeaderIntoRunContext(t *testing.T) {
	ctx := context.Background()
	redisC, endpoint := setupRedisContainer(ctx, t)
	defer testcontainers.CleanupContainer(t, redisC)

	got := make(chan map[string]string, 1)
	w := NewWorker(
		WithAddr(endpoint),
		WithStreamName("headered"),
		WithBlockTime(100*time.Millisecond),
		WithRunFunc(func(ctx context.Context, _ core.TaskMessage) error {
			got <- tracing.HeaderFromContext(ctx)
			return nil
		}),
	)
	defer func() {
		assert.NoError(t, w.Shutdown())
	}()

	w.startConsumer()
	require.NoError(t, w.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "headered",
		Values: map[string]any{
			fieldBody:   `{"body":"aGk=","timeout":60000000000}`,
			fieldHeader: `{"traceparent":"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"}`,
		},
	}).Err())

	task, err := w.Request()
	require.NoError(t, err)
	require.NoError(t, w.Run(context.Background(), task))

	select {
	case header := <-got:
		assert.Equal(t, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", header["traceparent"])
	case <-time.After(time.Second):
		t.Fatal("run func was not called")
	}

	// The entry is dropped once consumed, so the worker cannot accumulate headers.
	_, still := w.headers.Load(task)
	assert.False(t, still)
}

func TestParseHeaderDegradesOnGarbage(t *testing.T) {
	assert.Nil(t, parseHeader(nil))
	assert.Nil(t, parseHeader(""))
	assert.Nil(t, parseHeader("{not json"))
	assert.Equal(t, map[string]string{"b3": "x"}, parseHeader(`{"b3":"x"}`))
}
