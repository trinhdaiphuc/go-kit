package mysql

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gorm.io/gorm/schema"
)

type FloatType interface {
	float32 | float64
}

type FloatSliceSerializer[F FloatType] struct{}

func (FloatSliceSerializer[F]) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
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

	floatStrings := strings.Split(str, ",")
	floatSlice := make([]F, len(floatStrings))
	for i, floatStr := range floatStrings {
		floatVal, err := strconv.ParseFloat(floatStr, 64)
		if err != nil {
			return fmt.Errorf("failed to convert string to float: %v", err)
		}
		floatSlice[i] = F(floatVal)
	}

	value := field.ReflectValueOf(ctx, dst)
	value.Set(reflect.ValueOf(floatSlice))
	return nil
}

func (FloatSliceSerializer[F]) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	floatSlice, ok := fieldValue.([]F)
	if !ok {
		return "", fmt.Errorf("failed to convert value to []float: %v", fieldValue)
	}

	strValues := make([]string, len(floatSlice))
	for i, floatVal := range floatSlice {
		strValues[i] = strconv.FormatFloat(float64(floatVal), 'E', -1, 64)
	}
	return strings.Join(strValues, ","), nil
}

func (FloatSliceSerializer[F]) SerializerType() string {
	return string(schema.String)
}
