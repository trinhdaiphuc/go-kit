package main

import (
	"fmt"
	"testing"

	"github.com/redis/go-redis/v9"
)

const batchSize = 100

// Benchmark Pipelining for Set
func Benchmark_Pipeline_Set_Small(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := JSONSmallStruct{initSmall()}

	b.ResetTimer()
	for i := 0; i < b.N; i += batchSize {
		pipe := rdb.Pipeline()
		count := 0
		for j := 0; j < batchSize && i+j < b.N; j++ {
			key := fmt.Sprintf("bench:pipe:small:%d", j) // Reuse keys to avoid explosion
			pipe.Set(ctx, key, &data, 0)
			count++
		}
		if count > 0 {
			_, err := pipe.Exec(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// Benchmark Pipelining for Get
func Benchmark_Pipeline_Get_Small(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := JSONSmallStruct{initSmall()}

	// Pre-populate keys 0-99
	pipe := rdb.Pipeline()
	for i := 0; i < batchSize; i++ {
		key := fmt.Sprintf("bench:pipe:small:%d", i)
		pipe.Set(ctx, key, &data, 0)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i += batchSize {
		pipe := rdb.Pipeline()
		count := 0
		for j := 0; j < batchSize && i+j < b.N; j++ {
			key := fmt.Sprintf("bench:pipe:small:%d", j)
			pipe.Get(ctx, key)
			count++
		}
		if count > 0 {
			cmds, err := pipe.Exec(ctx)
			if err != nil {
				b.Fatal(err)
			}
			// Simulate decoding
			for _, cmd := range cmds {
				var res JSONSmallStruct
				if err := cmd.(*redis.StringCmd).Scan(&res); err != nil {
					b.Fatal(err)
				}
			}
		}
	}
}

// Benchmark MSet (Multi-Set)
func Benchmark_MSet_Small(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := JSONSmallStruct{initSmall()}

	// Prepare a batch of arguments
	// MSet accepts pairs: key, value, key, value...
	args := make([]interface{}, 0, batchSize*2)
	for i := 0; i < batchSize; i++ {
		args = append(args, fmt.Sprintf("bench:mset:small:%d", i), &data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i += batchSize {
		// We execute the same batch repeatedly for the benchmark
		// In reality, you'd have different data, but this measures the mechanism overhead
		if err := rdb.MSet(ctx, args...).Err(); err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark MGet (Multi-Get)
func Benchmark_MGet_Small(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := JSONSmallStruct{initSmall()}

	// Pre-populate
	args := make([]interface{}, 0, batchSize*2)
	keys := make([]string, 0, batchSize)
	for i := 0; i < batchSize; i++ {
		key := fmt.Sprintf("bench:mget:small:%d", i)
		args = append(args, key, &data)
		keys = append(keys, key)
	}
	if err := rdb.MSet(ctx, args...).Err(); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i += batchSize {
		// MGet returns a slice of interface{}
		vals, err := rdb.MGet(ctx, keys...).Result()
		if err != nil {
			b.Fatal(err)
		}

		// Simulate decoding
		for _, val := range vals {
			if val == nil {
				continue
			}
			// MGet returns the raw string/bytes, we need to unmarshal manually
			// Since we used JSONSmallStruct (which implements BinaryMarshaler),
			// Redis stores the JSON string.
			var res JSONSmallStruct
			s, ok := val.(string)
			if !ok {
				b.Fatal("expected string")
			}
			if err := res.UnmarshalBinary([]byte(s)); err != nil {
				b.Fatal(err)
			}
		}
	}
}

// --- Large Struct Benchmarks ---

// Benchmark Pipelining for Set (Large)
func Benchmark_Pipeline_Set_Large(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := JSONLargeStruct{initLarge()}

	b.ResetTimer()
	for i := 0; i < b.N; i += batchSize {
		pipe := rdb.Pipeline()
		count := 0
		for j := 0; j < batchSize && i+j < b.N; j++ {
			key := fmt.Sprintf("bench:pipe:large:%d", j)
			pipe.Set(ctx, key, &data, 0)
			count++
		}
		if count > 0 {
			_, err := pipe.Exec(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// Benchmark Pipelining for Get (Large)
func Benchmark_Pipeline_Get_Large(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := JSONLargeStruct{initLarge()}

	// Pre-populate keys 0-99
	pipe := rdb.Pipeline()
	for i := 0; i < batchSize; i++ {
		key := fmt.Sprintf("bench:pipe:large:%d", i)
		pipe.Set(ctx, key, &data, 0)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i += batchSize {
		pipe := rdb.Pipeline()
		count := 0
		for j := 0; j < batchSize && i+j < b.N; j++ {
			key := fmt.Sprintf("bench:pipe:large:%d", j)
			pipe.Get(ctx, key)
			count++
		}
		if count > 0 {
			cmds, err := pipe.Exec(ctx)
			if err != nil {
				b.Fatal(err)
			}
			// Simulate decoding
			for _, cmd := range cmds {
				var res JSONLargeStruct
				if err := cmd.(*redis.StringCmd).Scan(&res); err != nil {
					b.Fatal(err)
				}
			}
		}
	}
}

// Benchmark MSet (Multi-Set) (Large)
func Benchmark_MSet_Large(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := JSONLargeStruct{initLarge()}

	// Prepare a batch of arguments
	args := make([]interface{}, 0, batchSize*2)
	for i := 0; i < batchSize; i++ {
		args = append(args, fmt.Sprintf("bench:mset:large:%d", i), &data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i += batchSize {
		if err := rdb.MSet(ctx, args...).Err(); err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark MGet (Multi-Get) (Large)
func Benchmark_MGet_Large(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := JSONLargeStruct{initLarge()}

	// Pre-populate
	args := make([]interface{}, 0, batchSize*2)
	keys := make([]string, 0, batchSize)
	for i := 0; i < batchSize; i++ {
		key := fmt.Sprintf("bench:mget:large:%d", i)
		args = append(args, key, &data)
		keys = append(keys, key)
	}
	if err := rdb.MSet(ctx, args...).Err(); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i += batchSize {
		vals, err := rdb.MGet(ctx, keys...).Result()
		if err != nil {
			b.Fatal(err)
		}

		for _, val := range vals {
			if val == nil {
				continue
			}
			var res JSONLargeStruct
			s, ok := val.(string)
			if !ok {
				b.Fatal("expected string")
			}
			if err := res.UnmarshalBinary([]byte(s)); err != nil {
				b.Fatal(err)
			}
		}
	}
}
