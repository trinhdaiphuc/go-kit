# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

`github.com/trinhdaiphuc/go-kit` — a Go 1.25 **library** of reusable packages for microservices. It is
imported by services; it is not an application. There is no `main`, no binary, no runnable server
(the only `main` packages live under `examples/`). Consequences:

- Exported API is the product. A signature change is a breaking change for downstream services.
- Every top-level directory is one concern and is independently importable. Keep new code inside the
  package that owns the concern; do not create cross-package coupling that forces consumers to pull
  in dependencies (Kafka, Redis, MySQL, gRPC) they don't use.
- `README.md` has the authoritative package table and usage snippets. Update it when adding a package
  or changing a public constructor.

## Commands

```bash
make test        # go test -v -race -coverprofile=coverage.out ./... ; opens the HTML report
make coverage    # same without opening a browser (atomic covermode)
make lint        # golangci-lint run ./...
make fmt         # goimports -w on every .go file
make generate    # fmt, then go generate ./...  (regenerates all mocks)
make scan        # trivy image + fs scan (needs Docker)

go test -race -run TestName ./cache/...   # single test
```

`make test` calls `open coverage.out`, which is macOS-only and opens the raw profile — use
`make coverage` in CI or when you only need the numbers.

Lint note: `.golangci.yml` sets `issues-exit-code: 0` and `tests: false`. **`make lint` always exits 0**
— read its output, never trust the exit status. Enabled linters include `gosec`, `gocognit` (min 40),
`prealloc`, `errcheck` with `check-blank: true` (so `_ = f()` is a finding, not a workaround).

## Code generation

- **Mocks are generated, never hand-edited.** Interface files carry
  `//go:generate mockgen -destination=./mocks/$GOFILE -source=$GOFILE -package=<pkg>mock`.
  Package naming is inconsistent by design of history (`cachemock`, `kafkamock`, `redislock`, plain
  `mocks` for `uuid`/`clock`/`http/client`) — match the existing directive in the package you touch.
  Change the interface → run `make generate` → commit the regenerated file.
- Protos build with `buf` (`buf.gen.yaml`: `protoc-gen-go`, `paths=source_relative`, output in place).
  Only `examples/cache/redis/proto/models.proto` exists today.
- `examples/cache/redis/models.go` uses `//go:generate msgp` (msgpack codegen) for benchmarks.

## Architecture conventions

**Functional options everywhere.** Constructors take `(required args, ...Option)`. Two shapes coexist:
a plain `type Option func(*Options)` (most packages) and an `Option` *interface* with an
`optionFunc` adapter (`metrics`, `tracing`) when the option set must stay closed to outside packages.
Generic stores parameterize the option too: `type Option[K comparable, V any] func(*Options[K, V])`.
Follow whichever shape the package already uses.

**Generics for storage abstractions.** `cache.Store[K comparable, V any]` is the central interface;
`cache/redis`, `cache/local` implement it, and `cache/loader` layers cache-aside loading with
Redsync distributed locking + singleflight on top. New cache backends implement `cache.Store`, they
do not add methods to it — the interface is wide already and every implementer pays for additions.

**Three package-level singletons — know them before adding a second one:**
- `log.New(cfg)` builds the process logger behind a `sync.Once`; `log.Bg()` / `log.For(ctx)` lazily
  self-initialize with a JSON/info default if `New` was never called. Repeat `New` calls are no-ops.
- `metrics.NewServerMonitor(serviceName)` assigns a package-global `monitor` and calls
  `prom.MustRegister` — calling it twice **panics** on duplicate registration.
- `tracing.TracerProvider(...)` returns the provider plus a cleanup func that must be deferred.

**Logging:** `log.For(ctx)` is the entry point — it injects the OTel trace context into every record.
Use it, not `log.Bg()`, anywhere a `context.Context` is available. `log.Fn` lets callers attach
context-derived fields.

**Errors:** use `errorx`, not stdlib `errors`, for anything that crosses a service boundary.
`errorx.ErrorWrapper` models a Google AIP-193 error body (`code`/`status`/`message`/`details`) and
carries both an HTTP code and a gRPC `codes.Code`. Builder chain: `errorx.Wrap(err, msg).WithStatus(codes.NotFound).WithDetails(...)`.
Details must implement the sealed `IsErrorDetail` interface (`isErrorDetail()`), so new detail types
belong in `errorx` itself. `GetHTTPCode` / `GetGRPCCode` unwrap arbitrary errors for handlers;
`ParseValidateDetails` converts `validator/v10` errors into a `BadRequest` detail.

**Instrumentation is wired into the clients, not bolted on.** `http/client`, `grpc/client`,
`database/mysql`, `cache/redis`, and `kafka` already emit Prometheus metrics and OTel spans through
`metrics/` and `tracing/`. Adding a new outbound client means wiring those two packages in the same
way — don't leave a new integration unobserved.

## Testing

Coverage is thin (~15 `_test.go` files across ~30 packages), so new code is expected to bring its own
table-driven tests with `t.Run` subtests. Tests run with `-race`; anything with goroutines,
singletons, or shared maps must be race-clean. Redis-backed packages test against `redismock/v9`
rather than a live server — follow that, keep tests hermetic.
