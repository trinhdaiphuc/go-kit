package mysql

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm/schema"
)

type IntType interface {
	int | int8 | int16 | int32 | int64
}

type IntSliceSerializer[N IntType] struct{}

func (IntSliceSerializer[N]) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
	if dbValue == nil {
		return nil
	}

	var str string
	switch v := dbValue.(type) {
	case []byte:
		str = string(v)
	case string:
		str = v
	default:
		return fmt.Errorf("invalid type %T", dbValue)
	}

	if str == "" {
		return nil
	}

	intStrings := strings.Split(str, ",")
	intSlice := make([]N, len(intStrings))
	for i, intStr := range intStrings {
		intVal, err := strconv.ParseInt(intStr, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert string to int: %v", err)
		}
		intSlice[i] = N(intVal)
	}

	value := field.ReflectValueOf(ctx, dst)
	value.Set(reflect.ValueOf(intSlice))
	return nil
}

func (IntSliceSerializer[N]) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	intSlice, ok := fieldValue.([]N)
	if !ok {
		return "", fmt.Errorf("failed to convert value to []int: %v", fieldValue)
	}

	strValues := make([]string, len(intSlice))
	for i, intVal := range intSlice {
		strValues[i] = strconv.Itoa(int(intVal))
	}
	return strings.Join(strValues, ","), nil
}

func (IntSliceSerializer[N]) SerializerType() string {
	return string(schema.String)
}
