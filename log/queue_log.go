package log

import (
	"fmt"

	"github.com/golang-queue/queue"
)

type QueueLogger struct {
}

func NewQueueLogger() queue.Logger {
	return &QueueLogger{}
}

func (q *QueueLogger) Infof(format string, args ...any) {
	factory().ll.Info(fmt.Sprintf(format, args...))
}

func (q *QueueLogger) Errorf(format string, args ...any) {
	factory().ll.Error(fmt.Sprintf(format, args...))
}

func (q *QueueLogger) Fatalf(format string, args ...any) {
	factory().ll.Fatal(fmt.Sprintf(format, args...))
}

func (q *QueueLogger) Info(args ...any) {
	factory().ll.Info(fmt.Sprintf("%s", args...))
}

func (q *QueueLogger) Error(args ...any) {
	factory().ll.Error(fmt.Sprintf("%s", args...))
}

func (q *QueueLogger) Fatal(args ...any) {
	factory().ll.Fatal(fmt.Sprintf("%s", args...))
}
