package multistate

import (
	"context"

	"github.com/go-qbit/multistate/expr"
)

type action struct {
	id         string
	caption    string
	from       expr.Expression
	set        []uint64
	reset      []uint64
	do         ActionDoFunc
	availabler Availabler
}

type Availabler interface {
	String() string
	IsAvailable(ctx context.Context) bool
}

type ActionDoFunc func(ctx context.Context, id interface{}, opts ...interface{}) error

type Entity interface {
	StartAction(ctx context.Context) (context.Context, error)
	GetState(ctx context.Context) (uint64, error)
	SetState(ctx context.Context, newState uint64, params ...interface{}) error
	EndAction(ctx context.Context, err error) error
	GetId() interface{}
}
