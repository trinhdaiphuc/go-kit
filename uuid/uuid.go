package uuid

import "github.com/google/uuid"

//go:generate mockgen -destination=./mocks/mock_$GOFILE -source=$GOFILE -package=mocks
type Generator interface {
	New() string
}

type UUID struct{}

func (u UUID) New() string {
	return uuid.New().String()
}

func NewUUID() Generator {
	return UUID{}
}
