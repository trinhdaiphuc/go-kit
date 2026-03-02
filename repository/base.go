package repository

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

type QueryParams struct {
	Limit  int32
	Offset int32
	Sort   Sort
}

type Clause func(tx *gorm.DB)

type Model any

type Base[M Model, ID comparable] interface {
	Create(ctx context.Context, i *M) (*M, error)
	List(ctx context.Context, params QueryParams, clauses ...Clause) ([]*M, error)
	Get(ctx context.Context, i *M) (*M, error)
	GetByID(ctx context.Context, id ID) (*M, error)
	UpdateByID(ctx context.Context, id ID, i *M, clauses ...Clause) (rowsAffected int64, err error)
	Updates(ctx context.Context, i *M, clauses ...Clause) (rowsAffected int64, err error)
	UpdateColumns(ctx context.Context, id ID, columns map[string]any, clauses ...Clause) (rowsAffected int64, err error)
	Count(ctx context.Context, clauses ...Clause) (rowsAffected int64, err error)
	DeleteByID(ctx context.Context, id ID) (success bool, err error)
	Delete(ctx context.Context, i *M) (success bool, err error)
}

type baseRepository[M Model, ID comparable] struct {
	model *M
	db    *gorm.DB
}

func NewBase[M Model, ID comparable](db *gorm.DB) Base[M, ID] {
	return &baseRepository[M, ID]{
		model: new(M),
		db:    db,
	}
}

func (b *baseRepository[M, ID]) Create(ctx context.Context, i *M) (*M, error) {
	err := b.db.WithContext(ctx).Model(b.model).Select("*").Create(i).Error
	return i, err
}

func (b *baseRepository[M, ID]) UpdateByID(ctx context.Context, id ID, o *M, clauses ...Clause) (rowsAffected int64, err error) {
	tx := b.db.WithContext(ctx).Model(b.model)
	for _, f := range clauses {
		f(tx)
	}
	result := tx.Omit("CreatedAt").Where("id = ?", id).Updates(o)
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (b *baseRepository[M, ID]) Updates(ctx context.Context, o *M, clauses ...Clause) (rowsAffected int64, err error) {
	tx := b.db.WithContext(ctx).Model(o)
	for _, f := range clauses {
		f(tx)
	}
	result := tx.Omit("CreatedAt").Updates(o)
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (b *baseRepository[M, ID]) UpdateColumns(ctx context.Context, id ID, columns map[string]any, clauses ...Clause) (rowsAffected int64, err error) {
	var o *M
	tx := b.db.WithContext(ctx).Model(&o).Where("id = ?", id)
	for _, f := range clauses {
		f(tx)
	}
	result := tx.Updates(columns)
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

func (b *baseRepository[M, ID]) Count(ctx context.Context, clauses ...Clause) (rowsAffected int64, err error) {
	var count int64
	tx := b.db.WithContext(ctx).Model(b.model).WithContext(ctx)
	for _, f := range clauses {
		f(tx)
	}
	tx.Count(&count)
	return count, tx.Error
}

func (b *baseRepository[M, ID]) List(ctx context.Context, params QueryParams, clauses ...Clause) ([]*M, error) {
	var oList []*M
	tx := b.db.WithContext(ctx).
		Model(b.model).
		Limit(int(params.Limit)).
		Offset(int(params.Offset)).
		Order(params.Sort.Parse())
	for _, f := range clauses {
		f(tx)
	}
	tx.Find(&oList)
	return oList, tx.Error
}

func (b *baseRepository[M, ID]) GetByID(ctx context.Context, id ID) (*M, error) {
	var o *M

	result := b.db.WithContext(ctx).First(&o, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}

	return o, nil
}

func (b *baseRepository[M, ID]) Get(ctx context.Context, o *M) (*M, error) {
	result := b.db.WithContext(ctx).First(&o)
	if result.Error != nil {
		return nil, result.Error
	}

	return o, nil
}

func (b *baseRepository[M, ID]) Delete(ctx context.Context, i *M) (success bool, err error) {
	result := b.db.WithContext(ctx).Delete(i)
	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}

func (b *baseRepository[M, ID]) DeleteByID(ctx context.Context, id ID) (success bool, err error) {
	result := b.db.WithContext(ctx).Delete(b.model, id)
	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}

type Sort struct {
	Origin string
}

// Parse the query string to order string (Ex: http://example.com/messages?sort=created_at.asc,updated_at.acs
// => order string: created_at asc,updated_at acs)
func (s Sort) Parse() string {
	return strings.ReplaceAll(s.Origin, ".", " ")
}
