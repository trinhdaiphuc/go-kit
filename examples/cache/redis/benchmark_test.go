package main

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	goccy "github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
	pb "github.com/trinhdaiphuc/go-kit/examples/cache/redis/proto"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/proto"
)

// Struct definitions moved to models.go

// JSON Wrappers for Standard JSON Benchmarks
type JSONSmallStruct struct{ SmallStruct }

func (s *JSONSmallStruct) MarshalBinary() ([]byte, error) {
	return json.Marshal(s.SmallStruct)
}

func (s *JSONSmallStruct) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &s.SmallStruct)
}

type JSONMediumStruct struct{ MediumStruct }

func (m *JSONMediumStruct) MarshalBinary() ([]byte, error) {
	return json.Marshal(m.MediumStruct)
}

func (m *JSONMediumStruct) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &m.MediumStruct)
}

type JSONLargeStruct struct{ LargeStruct }

func (l *JSONLargeStruct) MarshalBinary() ([]byte, error) {
	return json.Marshal(l.LargeStruct)
}

func (l *JSONLargeStruct) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &l.LargeStruct)
}

type JSONLargeNestedStruct struct{ LargeNestedStruct }

func (l *JSONLargeNestedStruct) MarshalBinary() ([]byte, error) {
	return json.Marshal(l.LargeNestedStruct)
}

func (l *JSONLargeNestedStruct) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &l.LargeNestedStruct)
}

// --- Goccy JSON Wrappers ---

type GoccySmallStruct struct{ SmallStruct }

func (s *GoccySmallStruct) MarshalBinary() ([]byte, error) {
	return goccy.Marshal(s.SmallStruct)
}
func (s *GoccySmallStruct) UnmarshalBinary(data []byte) error {
	return goccy.Unmarshal(data, &s.SmallStruct)
}

type GoccyMediumStruct struct{ MediumStruct }

func (s *GoccyMediumStruct) MarshalBinary() ([]byte, error) {
	return goccy.Marshal(s.MediumStruct)
}
func (s *GoccyMediumStruct) UnmarshalBinary(data []byte) error {
	return goccy.Unmarshal(data, &s.MediumStruct)
}

type GoccyLargeStruct struct{ LargeStruct }

func (s *GoccyLargeStruct) MarshalBinary() ([]byte, error) {
	return goccy.Marshal(s.LargeStruct)
}
func (s *GoccyLargeStruct) UnmarshalBinary(data []byte) error {
	return goccy.Unmarshal(data, &s.LargeStruct)
}

type GoccyLargeNestedStruct struct{ LargeNestedStruct }

func (s *GoccyLargeNestedStruct) MarshalBinary() ([]byte, error) {
	return goccy.Marshal(s.LargeNestedStruct)
}
func (s *GoccyLargeNestedStruct) UnmarshalBinary(data []byte) error {
	return goccy.Unmarshal(data, &s.LargeNestedStruct)
}

// --- MsgPack Wrappers ---

type MsgPackSmallStruct struct{ SmallStruct }

func (s *MsgPackSmallStruct) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(&s.SmallStruct)
}
func (s *MsgPackSmallStruct) UnmarshalBinary(data []byte) error {
	return msgpack.Unmarshal(data, &s.SmallStruct)
}

type MsgPackMediumStruct struct{ MediumStruct }

func (s *MsgPackMediumStruct) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(&s.MediumStruct)
}
func (s *MsgPackMediumStruct) UnmarshalBinary(data []byte) error {
	return msgpack.Unmarshal(data, &s.MediumStruct)
}

type MsgPackLargeStruct struct{ LargeStruct }

func (s *MsgPackLargeStruct) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(&s.LargeStruct)
}
func (s *MsgPackLargeStruct) UnmarshalBinary(data []byte) error {
	return msgpack.Unmarshal(data, &s.LargeStruct)
}

type MsgPackLargeNestedStruct struct{ LargeNestedStruct }

func (s *MsgPackLargeNestedStruct) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(&s.LargeNestedStruct)
}
func (s *MsgPackLargeNestedStruct) UnmarshalBinary(data []byte) error {
	return msgpack.Unmarshal(data, &s.LargeNestedStruct)
}

// --- Tinylib Msgp Wrappers ---

type TinylibSmallStruct struct{ SmallStruct }

func (s *TinylibSmallStruct) MarshalBinary() ([]byte, error) {
	return s.SmallStruct.MarshalMsg(nil)
}
func (s *TinylibSmallStruct) UnmarshalBinary(data []byte) error {
	_, err := s.SmallStruct.UnmarshalMsg(data)
	return err
}

type TinylibMediumStruct struct{ MediumStruct }

func (s *TinylibMediumStruct) MarshalBinary() ([]byte, error) {
	return s.MediumStruct.MarshalMsg(nil)
}
func (s *TinylibMediumStruct) UnmarshalBinary(data []byte) error {
	_, err := s.MediumStruct.UnmarshalMsg(data)
	return err
}

type TinylibLargeStruct struct{ LargeStruct }

func (s *TinylibLargeStruct) MarshalBinary() ([]byte, error) {
	return s.LargeStruct.MarshalMsg(nil)
}
func (s *TinylibLargeStruct) UnmarshalBinary(data []byte) error {
	_, err := s.LargeStruct.UnmarshalMsg(data)
	return err
}

type TinylibLargeNestedStruct struct{ LargeNestedStruct }

func (s *TinylibLargeNestedStruct) MarshalBinary() ([]byte, error) {
	return s.LargeNestedStruct.MarshalMsg(nil)
}
func (s *TinylibLargeNestedStruct) UnmarshalBinary(data []byte) error {
	_, err := s.LargeNestedStruct.UnmarshalMsg(data)
	return err
}

// --- Tinylib Msgp Benchmarks ---

func Benchmark_Tinylib_Set_Small(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := TinylibSmallStruct{initSmall()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.Set(ctx, "bench:tinylib:small", &data, 0)
	}
}

func Benchmark_Tinylib_Get_Small(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := TinylibSmallStruct{initSmall()}
	rdb.Set(ctx, "bench:tinylib:small", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res TinylibSmallStruct
		rdb.Get(ctx, "bench:tinylib:small").Scan(&res)
	}
}

func Benchmark_Tinylib_Set_Medium(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := TinylibMediumStruct{initMedium()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.Set(ctx, "bench:tinylib:medium", &data, 0)
	}
}

func Benchmark_Tinylib_Get_Medium(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := TinylibMediumStruct{initMedium()}
	rdb.Set(ctx, "bench:tinylib:medium", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res TinylibMediumStruct
		rdb.Get(ctx, "bench:tinylib:medium").Scan(&res)
	}
}

func Benchmark_Tinylib_Set_Large(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := TinylibLargeStruct{initLarge()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.Set(ctx, "bench:tinylib:large", &data, 0)
	}
}

func Benchmark_Tinylib_Get_Large(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := TinylibLargeStruct{initLarge()}
	rdb.Set(ctx, "bench:tinylib:large", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res TinylibLargeStruct
		rdb.Get(ctx, "bench:tinylib:large").Scan(&res)
	}
}

func Benchmark_Tinylib_Set_LargeNested(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := TinylibLargeNestedStruct{initLargeNested()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.Set(ctx, "bench:tinylib:largenested", &data, 0)
	}
}

func Benchmark_Tinylib_Get_LargeNested(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := TinylibLargeNestedStruct{initLargeNested()}
	rdb.Set(ctx, "bench:tinylib:largenested", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res TinylibLargeNestedStruct
		rdb.Get(ctx, "bench:tinylib:largenested").Scan(&res)
	}
}

func getRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

var ctx = context.Background()

// --- Initialization ---

func initSmall() SmallStruct {
	return SmallStruct{
		Field1: "value1", Field2: 2, Field3: true, Field4: "value4", Field5: 5,
	}
}

func initMedium() MediumStruct {
	return MediumStruct{
		SmallStruct: initSmall(),
		Field6:      "value6", Field7: 7, Field8: false, Field9: 9.9, Field10: "value10",
	}
}

func initLarge() LargeStruct {
	return LargeStruct{
		MediumStruct: initMedium(),
		Field11:      "val11", Field12: 12, Field13: true, Field14: "val14", Field15: 15,
		Field16: "val16", Field17: 17, Field18: false, Field19: "val19", Field20: 20,
		Field21: "val21", Field22: 22, Field23: true, Field24: "val24", Field25: 25,
	}
}

func initLargeNested() LargeNestedStruct {
	sub := SubStruct{SubField1: "sub1", SubField2: 100}
	return LargeNestedStruct{
		Field1: "val1", Field2: 2, Field3: true, Field4: "val4", Field5: 5,
		Nested1: sub, Nested2: sub, Nested3: sub, Nested4: sub, Nested5: sub,
		Nested6: sub, Nested7: sub, Nested8: sub, Nested9: sub, Nested10: sub,
	}
}

// --- Flattening Helper for Nested Structs in Hash ---
// This is a simplified flattener.
func flattenStruct(v interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	flatten(reflect.ValueOf(v), "", out)
	return out
}

func flatten(v reflect.Value, prefix string, out map[string]interface{}) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		out[prefix] = v.Interface()
		return
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		val := v.Field(i)
		key := field.Name
		if tag := field.Tag.Get("json"); tag != "" {
			key = tag
		}
		if prefix != "" {
			key = prefix + "." + key
		}
		if val.Kind() == reflect.Struct {
			flatten(val, key, out)
		} else {
			out[key] = val.Interface()
		}
	}
}

// --- Benchmarks ---

// 1. Small Struct

func Benchmark_JSON_Set_Small(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := JSONSmallStruct{initSmall()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.Set(ctx, "bench:json:small", &data, 0)
	}
}

func Benchmark_JSON_Get_Small(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := JSONSmallStruct{initSmall()}
	rdb.Set(ctx, "bench:json:small", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res JSONSmallStruct
		rdb.Get(ctx, "bench:json:small").Scan(&res)
	}
}

func Benchmark_Hash_Set_Small(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := initSmall()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.HSet(ctx, "bench:hash:small", data)
	}
}

func Benchmark_Hash_Get_Small(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := initSmall()
	rdb.HSet(ctx, "bench:hash:small", data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res SmallStruct
		rdb.HGetAll(ctx, "bench:hash:small").Scan(&res)
	}
}

// 2. Medium Struct

func Benchmark_JSON_Set_Medium(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := JSONMediumStruct{initMedium()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.Set(ctx, "bench:json:medium", &data, 0)
	}
}

func Benchmark_JSON_Get_Medium(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := JSONMediumStruct{initMedium()}
	rdb.Set(ctx, "bench:json:medium", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res JSONMediumStruct
		rdb.Get(ctx, "bench:json:medium").Scan(&res)
	}
}

func Benchmark_Hash_Set_Medium(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := initMedium()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.HSet(ctx, "bench:hash:medium", data)
	}
}

func Benchmark_Hash_Get_Medium(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := initMedium()
	rdb.HSet(ctx, "bench:hash:medium", data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res MediumStruct
		rdb.HGetAll(ctx, "bench:hash:medium").Scan(&res)
	}
}

// 3. Large Struct (Flat)

func Benchmark_JSON_Set_Large(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := JSONLargeStruct{initLarge()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.Set(ctx, "bench:json:large", &data, 0)
	}
}

func Benchmark_JSON_Get_Large(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := JSONLargeStruct{initLarge()}
	rdb.Set(ctx, "bench:json:large", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res JSONLargeStruct
		rdb.Get(ctx, "bench:json:large").Scan(&res)
	}
}

func Benchmark_Hash_Set_Large(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := initLarge()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.HSet(ctx, "bench:hash:large", data)
	}
}

func Benchmark_Hash_Get_Large(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := initLarge()
	rdb.HSet(ctx, "bench:hash:large", data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res LargeStruct
		rdb.HGetAll(ctx, "bench:hash:large").Scan(&res)
	}
}

// 4. Large Nested Struct (Custom Flattening for Hash)

func Benchmark_JSON_Set_LargeNested(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := JSONLargeNestedStruct{initLargeNested()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.Set(ctx, "bench:json:largenested", &data, 0)
	}
}

func Benchmark_JSON_Get_LargeNested(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := JSONLargeNestedStruct{initLargeNested()}
	rdb.Set(ctx, "bench:json:largenested", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res JSONLargeNestedStruct
		rdb.Get(ctx, "bench:json:largenested").Scan(&res)
	}
}

func Benchmark_Hash_Set_LargeNested(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := initLargeNested()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Must flatten first
		flat := flattenStruct(data)
		rdb.HSet(ctx, "bench:hash:largenested", flat)
	}
}

// Note: HGetAll + Scan does NOT support unflattening automatically in go-redis without custom logic.
// So this benchmark is "unfair" if we don't include unflattening cost.
// However, writing a generic unflatten is complex.
// For the sake of the benchmark "save a struct", we can measure the Write cost.
// For Read, we can just HGetAll and see the raw map, or try to Scan (which will fail to populate nested fields).
// I will implement a simple Read that just fetches the map, as that's the "Redis" part.
// Or I can skip the Read benchmark for Nested Hash if it's too complex to implement correctly in a short script.
// Let's just benchmark the SET for nested, as that's where the flattening overhead is.
// And for GET, we'll just HGetAll.

// 5. Goccy JSON Benchmarks

func Benchmark_GoccyJSON_Set_Small(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := GoccySmallStruct{initSmall()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.Set(ctx, "bench:goccy:small", &data, 0)
	}
}

func Benchmark_GoccyJSON_Get_Small(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := GoccySmallStruct{initSmall()}
	rdb.Set(ctx, "bench:goccy:small", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res GoccySmallStruct
		rdb.Get(ctx, "bench:goccy:small").Scan(&res)
	}
}

func Benchmark_GoccyJSON_Set_Medium(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := GoccyMediumStruct{initMedium()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.Set(ctx, "bench:goccy:medium", &data, 0)
	}
}

func Benchmark_GoccyJSON_Get_Medium(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := GoccyMediumStruct{initMedium()}
	rdb.Set(ctx, "bench:goccy:medium", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res GoccyMediumStruct
		rdb.Get(ctx, "bench:goccy:medium").Scan(&res)
	}
}

func Benchmark_GoccyJSON_Set_Large(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := GoccyLargeStruct{initLarge()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.Set(ctx, "bench:goccy:large", &data, 0)
	}
}

func Benchmark_GoccyJSON_Get_Large(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := GoccyLargeStruct{initLarge()}
	rdb.Set(ctx, "bench:goccy:large", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res GoccyLargeStruct
		rdb.Get(ctx, "bench:goccy:large").Scan(&res)
	}
}

func Benchmark_GoccyJSON_Set_LargeNested(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := GoccyLargeNestedStruct{initLargeNested()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.Set(ctx, "bench:goccy:largenested", &data, 0)
	}
}

func Benchmark_GoccyJSON_Get_LargeNested(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := GoccyLargeNestedStruct{initLargeNested()}
	rdb.Set(ctx, "bench:goccy:largenested", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res GoccyLargeNestedStruct
		rdb.Get(ctx, "bench:goccy:largenested").Scan(&res)
	}
}

// 6. MsgPack Benchmarks

func Benchmark_MsgPack_Set_Small(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := MsgPackSmallStruct{initSmall()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := rdb.Set(ctx, "bench:msgpack:small", &data, 0).Err(); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_MsgPack_Get_Small(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := MsgPackSmallStruct{initSmall()}
	rdb.Set(ctx, "bench:msgpack:small", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res MsgPackSmallStruct
		rdb.Get(ctx, "bench:msgpack:small").Scan(&res)
	}
}

func Benchmark_MsgPack_Set_Medium(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := MsgPackMediumStruct{initMedium()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := rdb.Set(ctx, "bench:msgpack:medium", &data, 0).Err(); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_MsgPack_Get_Medium(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := MsgPackMediumStruct{initMedium()}
	rdb.Set(ctx, "bench:msgpack:medium", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res MsgPackMediumStruct
		rdb.Get(ctx, "bench:msgpack:medium").Scan(&res)
	}
}

func Benchmark_MsgPack_Set_Large(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := MsgPackLargeStruct{initLarge()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := rdb.Set(ctx, "bench:msgpack:large", &data, 0).Err(); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_MsgPack_Get_Large(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := MsgPackLargeStruct{initLarge()}
	rdb.Set(ctx, "bench:msgpack:large", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res MsgPackLargeStruct
		rdb.Get(ctx, "bench:msgpack:large").Scan(&res)
	}
}

func Benchmark_MsgPack_Set_LargeNested(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := MsgPackLargeNestedStruct{initLargeNested()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.Set(ctx, "bench:msgpack:largenested", &data, 0)
	}
}

func Benchmark_MsgPack_Get_LargeNested(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := MsgPackLargeNestedStruct{initLargeNested()}
	rdb.Set(ctx, "bench:msgpack:largenested", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res MsgPackLargeNestedStruct
		rdb.Get(ctx, "bench:msgpack:largenested").Scan(&res)
	}
}

// --- Proto Wrappers ---

type ProtoSmallStruct struct{ pb.SmallStruct }

func (s *ProtoSmallStruct) MarshalBinary() ([]byte, error) {
	return proto.Marshal(&s.SmallStruct)
}
func (s *ProtoSmallStruct) UnmarshalBinary(data []byte) error {
	return proto.Unmarshal(data, &s.SmallStruct)
}

type ProtoMediumStruct struct{ pb.MediumStruct }

func (s *ProtoMediumStruct) MarshalBinary() ([]byte, error) {
	return proto.Marshal(&s.MediumStruct)
}
func (s *ProtoMediumStruct) UnmarshalBinary(data []byte) error {
	return proto.Unmarshal(data, &s.MediumStruct)
}

type ProtoLargeStruct struct{ pb.LargeStruct }

func (s *ProtoLargeStruct) MarshalBinary() ([]byte, error) {
	return proto.Marshal(&s.LargeStruct)
}
func (s *ProtoLargeStruct) UnmarshalBinary(data []byte) error {
	return proto.Unmarshal(data, &s.LargeStruct)
}

type ProtoLargeNestedStruct struct{ pb.LargeNestedStruct }

func (s *ProtoLargeNestedStruct) MarshalBinary() ([]byte, error) {
	return proto.Marshal(&s.LargeNestedStruct)
}
func (s *ProtoLargeNestedStruct) UnmarshalBinary(data []byte) error {
	return proto.Unmarshal(data, &s.LargeNestedStruct)
}

// --- Proto Initialization ---

func initProtoSmall() ProtoSmallStruct {
	return ProtoSmallStruct{pb.SmallStruct{
		Field1: "value1", Field2: 2, Field3: true, Field4: "value4", Field5: 5,
	}}
}

func initProtoMedium() ProtoMediumStruct {
	return ProtoMediumStruct{pb.MediumStruct{
		Field1: "value1", Field2: 2, Field3: true, Field4: "value4", Field5: 5,
		Field6: "value6", Field7: 7, Field8: true, Field9: 9.9, Field10: "value10",
	}}
}

func initProtoLarge() ProtoLargeStruct {
	return ProtoLargeStruct{pb.LargeStruct{
		Field1: "value1", Field2: 2, Field3: true, Field4: "value4", Field5: 5,
		Field6: "value6", Field7: 7, Field8: true, Field9: 9.9, Field10: "value10",
		Field11: "value11", Field12: 12, Field13: true, Field14: "value14", Field15: 15,
		Field16: "value16", Field17: 17, Field18: true, Field19: "value19", Field20: 20,
		Field21: "value21", Field22: 22, Field23: true, Field24: "value24", Field25: 25,
	}}
}

func initProtoLargeNested() ProtoLargeNestedStruct {
	s := ProtoLargeNestedStruct{pb.LargeNestedStruct{
		Field1: "value1", Field2: 2, Field3: true, Field4: "value4", Field5: 5,
	}}
	// Populate nested
	sub := &pb.SubStruct{SubField1: "sub1", SubField2: 100}
	s.Nested1 = sub
	s.Nested2 = sub
	s.Nested3 = sub
	s.Nested4 = sub
	s.Nested5 = sub
	s.Nested6 = sub
	s.Nested7 = sub
	s.Nested8 = sub
	s.Nested9 = sub
	s.Nested10 = sub
	return s
}

// --- Proto Benchmarks ---

func Benchmark_Proto_Set_Small(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := initProtoSmall()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := rdb.Set(ctx, "bench:proto:small", &data, 0).Err(); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Proto_Get_Small(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := initProtoSmall()
	rdb.Set(ctx, "bench:proto:small", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res ProtoSmallStruct
		rdb.Get(ctx, "bench:proto:small").Scan(&res)
	}
}

func Benchmark_Proto_Set_Medium(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := initProtoMedium()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.Set(ctx, "bench:proto:medium", &data, 0)
	}
}

func Benchmark_Proto_Get_Medium(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := initProtoMedium()
	rdb.Set(ctx, "bench:proto:medium", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res ProtoMediumStruct
		rdb.Get(ctx, "bench:proto:medium").Scan(&res)
	}
}

func Benchmark_Proto_Set_Large(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := initProtoLarge()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.Set(ctx, "bench:proto:large", &data, 0)
	}
}

func Benchmark_Proto_Get_Large(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := initProtoLarge()
	rdb.Set(ctx, "bench:proto:large", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res ProtoLargeStruct
		rdb.Get(ctx, "bench:proto:large").Scan(&res)
	}
}

func Benchmark_Proto_Set_LargeNested(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := initProtoLargeNested()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rdb.Set(ctx, "bench:proto:largenested", &data, 0)
	}
}

func Benchmark_Proto_Get_LargeNested(b *testing.B) {
	rdb := getRedisClient()
	defer rdb.Close()
	data := initProtoLargeNested()
	rdb.Set(ctx, "bench:proto:largenested", &data, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res ProtoLargeNestedStruct
		rdb.Get(ctx, "bench:proto:largenested").Scan(&res)
	}
}
