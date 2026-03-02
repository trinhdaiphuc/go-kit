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
	instance.ll.Info(fmt.Sprintf(format, args...))
}

func (q *QueueLogger) Errorf(format string, args ...any) {
	instance.ll.Error(fmt.Sprintf(format, args...))
}

func (q *QueueLogger) Fatalf(format string, args ...any) {
	instance.ll.Fatal(fmt.Sprintf(format, args...))
}

func (q *QueueLogger) Info(args ...any) {
	instance.ll.Info(fmt.Sprintf("%s", args...))
}

func (q *QueueLogger) Error(args ...any) {
	instance.ll.Error(fmt.Sprintf("%s", args...))
}

func (q *QueueLogger) Fatal(args ...any) {
	instance.ll.Fatal(fmt.Sprintf("%s", args...))
}
