# Go kit

Common packages for golang microservices.

## Installation

```shell
go get -u github.com/trinhdaiphuc/go-kit
```

## Package Overview

| Package | Purpose |
|---------|---------|
| `breaker/` | Circuit breaker pattern implementation using Sony's gobreaker |
| `cache/` | Caching abstraction with Redis and local implementations |
| `cache/loader/` | Cache loaders with distributed locking (Redsync) and singleflight |
| `cache/redis/` | Redis cache store with distributed locking support |
| `cache/local/` | Local in-memory cache |
| `clock/` | Clock abstraction for time utilities |
| `collection/` | Generic collection utilities (array/slice and map helpers) |
| `database/mysql/` | MySQL/GORM database connection with tracing and Prometheus metrics |
| `errorx/` | Structured error handling following Google AIP-193 |
| `grpc/client/` | gRPC client with retry and OpenTelemetry tracing |
| `grpc/interceptor/` | gRPC interceptors for logging and circuit breaker |
| `grpc/request/` | gRPC request metadata parsing |
| `grpc/server/` | gRPC server utilities (health checks) |
| `header/` | HTTP header parsing and models |
| `http/client/` | HTTP client with retry, tracing, and Prometheus metrics |
| `http/middleware/` | HTTP middleware utilities (Gin logger, high latency detection) |
| `http/tripperware/` | HTTP RoundTripper middleware (retry with backoff) |
| `kafka/` | Kafka producer/consumer using IBM Sarama with SASL/TLS support |
| `log/` | Structured logging using Zap with OpenTelemetry trace context |
| `mailbox/` | Microsoft Outlook mailbox client via Microsoft Graph API (ROPC OAuth2) |
| `metrics/` | Prometheus metrics for HTTP, gRPC, Kafka, Redis, and circuit breaker |
| `network/` | Network utilities (IP address) |
| `queue/redis-stream/` | Redis Stream queue worker |
| `repository/` | Base repository patterns |
| `thread/` | Thread/goroutine utilities |
| `tracing/` | OpenTelemetry tracing setup (Jaeger/OTLP exporters) |
| `url/` | URL manipulation utilities |
| `uuid/` | UUID generator with interface |
| `validate/` | Vietnamese text validation utilities |
| `version/` | Semantic version comparison utilities |

## Usage

### Cache

```go
package main

import (
	"context"
	"time"

	"github.com/trinhdaiphuc/go-kit/cache"
	cacheredis "github.com/trinhdaiphuc/go-kit/cache/redis"
	"github.com/trinhdaiphuc/go-kit/cache/local"
)

func main() {
	ctx := context.Background()

	// Redis cache
	redisStore := cacheredis.NewStore[string, MyData](
		cacheredis.WithAddr("localhost:6379"),
		cacheredis.WithPassword(""),
		cacheredis.WithDB(0),
	)
	redisStore.Set(ctx, "key", MyData{}, cacheredis.WithTTL(time.Hour))
	data, err := redisStore.Get(ctx, "key")

	// Local cache
	localStore := local.NewStore[string, MyData](
		local.WithMaxSize(1000),
		local.WithTTL(time.Minute),
	)
	localStore.Set(ctx, "key", MyData{})
	data, err = localStore.Get(ctx, "key")
}
```

### MySQL

```go
package main

import (
	"github.com/trinhdaiphuc/go-kit/database/mysql"
)

func main() {
	db, cleanup, err := mysql.ConnectMySQL(&mysql.Config{
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password",
		Database: "mydb",
	}, "service-name")
	defer cleanup()
}
```

### HTTP Client

```go
package main

import (
	"context"
	"time"

	httpclient "github.com/trinhdaiphuc/go-kit/http/client"
)

func main() {
	client := httpclient.New("service-name",
		httpclient.WithRequestTimeout(10*time.Second),
	)

	ctx := context.Background()
	resp, err := client.Get(ctx, "https://api.example.com/data")
}
```

### gRPC Client

```go
package main

import (
	"context"

	grpcclient "github.com/trinhdaiphuc/go-kit/grpc/client"
)

func main() {
	conn, err := grpcclient.NewConnection("localhost:9090",
		grpcclient.WithInsecure(),
	)
	defer conn.Close()
}
```

### Kafka

```go
package main

import (
	"context"

	"github.com/trinhdaiphuc/go-kit/kafka"
)

func main() {
	// Producer
	producer, err := kafka.NewProducer(
		kafka.WithBrokers([]string{"localhost:9092"}),
		kafka.WithTopic("my-topic"),
	)
	defer producer.Close()

	producer.Publish(context.Background(), []byte("key"), []byte("message"))

	// Single-message consumer
	consumer, err := kafka.NewConsumer(
		kafka.WithBrokers([]string{"localhost:9092"}),
		kafka.WithTopic("my-topic"),
		kafka.WithGroupID("my-group"),
	)
	defer consumer.Close()
}
```

### Kafka Batch Consumer

The batch consumer buffers messages and flushes the batch when either
`batchSize` messages accumulate **or** `delayInterval` elapses — whichever
comes first.

```go
package main

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/trinhdaiphuc/go-kit/kafka"
)

func main() {
	cfg := &kafka.Config{
		Brokers: []string{"localhost:9092"},
		Topics:  []string{"my-topic"},
		GroupID: "my-group",
	}

	consumer, err := kafka.NewBatchConsumer(
		cfg,
		func(ctx context.Context, msgs []*sarama.ConsumerMessage) []kafka.MessageResult {
			results := make([]kafka.MessageResult, len(msgs))
			for i, m := range msgs {
				// process m …
				results[i] = kafka.MessageResult{Offset: m.Offset}
			}
			return results
		},
		kafka.DefaultBatchSize,    // flush at 256 messages
		100*time.Millisecond,      // or every 100 ms
	)
	if err != nil {
		panic(err)
	}

	go consumer.Start()
	defer consumer.Close()
}

### Mailbox

Client for reading Microsoft Outlook mailboxes via the [Microsoft Graph API](https://learn.microsoft.com/en-us/graph/api/resources/mail-api-overview).
Authentication uses the OAuth2 Resource Owner Password Credentials (ROPC) flow.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/trinhdaiphuc/go-kit/mailbox"
)

func main() {
	cfg := &mailbox.Config{
		Username:     "user@example.com",
		Password:     "s3cr3t",
		ClientID:     "azure-client-id",
		ClientSecret: "azure-client-secret", // optional for public clients
		TenantID:     "azure-tenant-id",
	}

	// Create client — all options are optional.
	client, err := mailbox.NewMailboxClient(cfg,
		mailbox.WithTimeout(15*time.Second),         // custom timeout
		mailbox.WithHTTPClient(&http.Client{}),      // or bring your own client
		mailbox.WithContext(context.Background()),    // base context stored on client
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// List recent messages (supports OData $filter).
	messages, err := client.GetMessages(ctx, 10, "")

	// Only unread messages.
	unread, err := client.GetUnreadMessages(ctx, 10)

	// Full message by ID.
	msg, err := client.GetMessage(ctx, messages[0].ID)
	fmt.Println(msg.Subject, msg.Body.Content)

	// Attachments for a message.
	attachments, err := client.GetAttachments(ctx, messages[0].ID)
	for _, att := range attachments {
		_ = mailbox.SaveAttachment(att, "/tmp/"+att.Name)
	}

	// Mark as read.
	_ = client.MarkAsRead(ctx, messages[0].ID)

	// Poll for new messages.
	_ = client.WatchMailbox(ctx, 30*time.Second, func(m mailbox.Message) {
		fmt.Println("new message:", m.Subject)
	})

	_ = unread
	_ = err
}
```

### Logging

```go
package main

import (
	"context"

	"github.com/trinhdaiphuc/go-kit/log"
)

func main() {
	ctx := context.Background()

	// Context-aware logging with trace correlation
	log.For(ctx).Info("operation completed", log.String("key", "value"))

	// Background logging (no context)
	log.Bg().Debug("debug message")
}
```

### Tracing

```go
package main

import (
	"context"

	"github.com/trinhdaiphuc/go-kit/tracing"
)

func main() {
	// Initialize tracer provider
	tp, cleanup, err := tracing.TracerProvider("service-name", "v1.0.0", &tracing.OtelExporter{
		OTLPEndpoint: "localhost:4317",
	})
	defer cleanup()

	ctx := context.Background()

	// Create spans
	ctx, span := tracing.CreateSpan(ctx, "operation-name")
	defer span.End()
}
```

### Metrics

```go
package main

import (
	"github.com/trinhdaiphuc/go-kit/metrics"
)

func main() {
	// Initialize server monitor
	monitor := metrics.NewServerMonitor("service-name")

	// Metrics are automatically collected for:
	// - HTTP requests (via tripperware)
	// - gRPC calls (via interceptors)
	// - Kafka operations
	// - Circuit breaker state
}
```

### Circuit Breaker

```go
package main

import (
	"github.com/trinhdaiphuc/go-kit/breaker"
)

func main() {
	cb := breaker.NewCircuitBreaker[MyResponse](
		breaker.WithCircuitBreakerName("my-service"),
		breaker.WithCircuitBreakerMaxRequests(5),
		breaker.WithCircuitBreakerInterval(10),
		breaker.WithCircuitBreakerTimeout(30),
	)

	result, err := cb.Execute(func() (MyResponse, error) {
		// Your logic here
		return MyResponse{}, nil
	})
}
```

### Error Handling

```go
package main

import (
	"net/http"

	"github.com/trinhdaiphuc/go-kit/errorx"
	"google.golang.org/grpc/codes"
)

func main() {
	err := errorx.New("operation failed").
		WithCode(http.StatusBadRequest).
		WithStatus(codes.InvalidArgument).
		WithDetails(&errorx.BadRequest{})
}
```

### Collection Utilities

```go
package main

import (
	"github.com/trinhdaiphuc/go-kit/collection"
)

func main() {
	// Array utilities
	arr := []int{1, 2, 3, 4, 5}
	contains := collection.Contains(arr, 3)
	filtered := collection.Filter(arr, func(v int) bool { return v > 2 })
}
```

### Clock

```go
package main

import (
	"github.com/trinhdaiphuc/go-kit/clock"
)

func main() {
	// Real clock for production
	c := clock.NewRealClock()
	now := c.Now()
}
```

### Version Utilities

```go
package main

import (
	"github.com/trinhdaiphuc/go-kit/version"
)

func main() {
	// Compare semantic versions
	result := version.Compare("1.2.3", "1.2.4")
	// result < 0 means first version is less than second
}
```

## Examples

Working examples are located in the `examples/` directory:

- `examples/cache/` - Local and Redis cache usage
- `examples/grpc/` - gRPC server
- `examples/kafka/` - Kafka producer, single-message consumer, and batch consumer
- `examples/main.go` - Microsoft Outlook mailbox client (reads messages and downloads attachments)

## Development

### Running Tests

```bash
make test
```

### Development Commands

```bash
# Format code
make fmt

# Run linter
make lint
```

## License

MIT License
