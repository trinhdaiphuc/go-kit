package mysql

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

const (
	ErrMySQLDuplicateEntry         uint16 = 1062
	ErrMySQLUpdateDeleteForeignKey uint16 = 1451
)

func IsDuplicateEntryErr(err error) bool {
	var myslErr *mysql.MySQLError
	ok := errors.As(err, &myslErr)
	return ok && myslErr.Number == ErrMySQLDuplicateEntry
}

func IsForeignKeyErr(err error) bool {
	var myslErr *mysql.MySQLError
	ok := errors.As(err, &myslErr)
	return ok && myslErr.Number == ErrMySQLUpdateDeleteForeignKey
}
