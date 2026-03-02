# Kafka Example

This example demonstrates how to use the go-kit Kafka client with both plain and SASL/PLAIN authentication.

## Prerequisites

- Docker and Docker Compose
- Go 1.23+

## Quick Start

### 1. Start Infrastructure

**Without authentication (plain):**

```bash
docker-compose up -d kafka jaeger
```

**With SASL/PLAIN authentication:**

```bash
docker-compose up -d kafka-sasl jaeger
```

**With SASL/SCRAM authentication (SHA-256/SHA-512):**

```bash
docker-compose up -d kafka-scram jaeger
```

Wait ~10 seconds for Kafka to be ready.

### 2. Run Consumer

**Without authentication:**

```bash
go run main.go consumer -b localhost:9092
```

**With SASL/PLAIN authentication:**

```bash
go run main.go consumer -b localhost:9093 -u admin -p admin-secret -a plain
```

### 3. Run Producer (in another terminal)

**Without authentication:**

```bash
go run main.go producer -b localhost:9092
```

**With SASL/PLAIN authentication:**

```bash
go run main.go producer -b localhost:9093 -u admin -p admin-secret -a plain
```

Type messages and press Enter to send them to Kafka. Press `Ctrl+C` to stop.

### 4. Run Batch Consumer

Buffers messages and flushes when the batch reaches `--batch-size` **or** `--batch-interval` elapses — whichever comes first.

**Without authentication:**

```bash
go run main.go batch-consumer -b localhost:9092 --batch-size 5 --batch-interval 500ms
```

**With SASL/PLAIN authentication:**

```bash
go run main.go batch-consumer -b localhost:9093 -u admin -p admin-secret -a plain \
  --batch-size 5 --batch-interval 500ms
```

### 5. Run Mock Producer (in another terminal)

Auto-publishes a pre-defined list of messages without requiring any stdin input.
Useful for testing the batch consumer end-to-end.

**Send all 10 pre-defined messages once (one every 100 ms):**

```bash
go run main.go mock-producer -b localhost:9092 --interval 100ms
```

**Send 50 messages by cycling the list, with a short delay:**

```bash
go run main.go mock-producer -b localhost:9092 --count 50 --interval 50ms
```

**Repeat indefinitely until `Ctrl+C`:**

```bash
go run main.go mock-producer -b localhost:9092 --interval 100ms --repeat
```

### End-to-end Batch Test

Open two terminals inside `examples/`:

```bash
# Terminal 1 — batch consumer (flush every 3 messages or 300 ms)
go run kafka/main.go batch-consumer -b localhost:9092 -t test-batch \
  -g test-group --batch-size 3 --batch-interval 300ms

# Terminal 2 — mock producer (10 messages, one every 100 ms)
go run kafka/main.go mock-producer -b localhost:9092 -t test-batch --interval 100ms
```

## Command Flags

### Global flags (all commands)

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--brokers` | `-b` | `localhost:9092` | Kafka brokers address |
| `--client-id` | `-c` | `client-1` | Client ID |
| `--group-id` | `-g` | `group-1` | Consumer group ID |
| `--topic` | `-t` | `topic-1` | Kafka topic |
| `--sasl-username` | `-u` | | SASL username |
| `--sasl-password` | `-p` | | SASL password |
| `--sasl-algorithm` | `-a` | `plain` | SASL algorithm (`plain`, `sha256`, `sha512`) |
| `--tracing` | `-r` | `true` | Enable OpenTelemetry tracing |
| `--metrics` | `-e` | `true` | Enable Prometheus metrics |

### `batch-consumer` flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--batch-size` | `-s` | `256` | Max messages per batch (must be ≤ Sarama `ChannelBufferSize`) |
| `--batch-interval` | `-n` | `100ms` | Max wait before flushing a partial batch |

### `mock-producer` flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--count` | `-N` | `10` | Total messages to send (cycles list if larger than list size) |
| `--interval` | `-d` | `50ms` | Delay between each message send |
| `--repeat` | `-R` | `false` | Loop the message list indefinitely until `Ctrl+C` |

## SASL Authentication

### SASL/PLAIN

The simplest authentication mechanism. Credentials are sent in plaintext (use with TLS in production).

```bash
go run main.go consumer -b localhost:9093 -u admin -p admin-secret -a plain
```

### SASL/SCRAM-SHA-256

More secure than PLAIN, uses challenge-response mechanism. Requires `kafka-scram` service.

```bash
docker-compose up -d kafka-scram jaeger
go run main.go consumer -b localhost:9094 -u admin -p admin-secret -a sha256
```

### SASL/SCRAM-SHA-512

Most secure SCRAM variant. Requires `kafka-scram` service.

```bash
docker-compose up -d kafka-scram jaeger
go run main.go consumer -b localhost:9094 -u admin -p admin-secret -a sha512
```

## Docker Services

| Service | Port | SASL Mechanism | Description |
|---------|------|----------------|-------------|
| `kafka` | 9092 | None | Kafka without authentication |
| `kafka-sasl` | 9093 | PLAIN | Kafka with SASL/PLAIN |
| `kafka-scram` | 9094 | SCRAM-SHA-256, SCRAM-SHA-512 | Kafka with SASL/SCRAM |
| `jaeger` | 16686 | - | Jaeger UI for tracing |

**Credentials for all SASL services:** `admin` / `admin-secret`

## Observability

### Tracing

View traces in Jaeger UI: http://localhost:16686

### Metrics

Prometheus metrics are exposed by the application when `--metrics` is enabled.

## Cleanup

```bash
docker-compose down -v
```
