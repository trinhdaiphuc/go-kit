package mysql

import (
	"fmt"
	"strings"
)

// StringArray represents a one-dimensional array of the PostgreSQL character types.
type StringArray []string

// Scan implements the sql.Scanner interface.
func (a *StringArray) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("cannot convert %T to StringArray", src)
	}

	str := string(bytes)
	if str == "" {
		*a = nil
		return nil
	}

	*a = strings.Split(str, ",")

	return nil
}
