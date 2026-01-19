package mysql

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm/schema"
)

type UintType interface {
	uint | uint8 | uint16 | uint32 | uint64
}

type UintSliceSerializer[U UintType] struct{}

func (UintSliceSerializer[U]) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
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

	uintStrings := strings.Split(str, ",")
	uintSlice := make([]U, len(uintStrings))
	for i, intStr := range uintStrings {
		uintVal, err := strconv.ParseUint(intStr, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert string to uint: %v", err)
		}
		uintSlice[i] = U(uintVal)
	}

	value := field.ReflectValueOf(ctx, dst)
	value.Set(reflect.ValueOf(uintSlice))
	return nil
}

func (UintSliceSerializer[U]) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	uintSlice, ok := fieldValue.([]U)
	if !ok {
		return "", fmt.Errorf("failed to convert value to []uint: %v", fieldValue)
	}

	strValues := make([]string, len(uintSlice))
	for i, uintVal := range uintSlice {
		strValues[i] = strconv.FormatUint(uint64(uintVal), 64)
	}
	return strings.Join(strValues, ","), nil
}

func (UintSliceSerializer[U]) SerializerType() string {
	return string(schema.String)
}
