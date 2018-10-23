package multistate

import (
	"context"

	"github.com/go-qbit/multistate/expr"
	"github.com/go-qbit/rbac"
)

type action struct {
	id         string
	caption    string
	from       expr.IExpression
	set        []uint64
	reset      []uint64
	do         ActionDoFunc
	permission *rbac.Permission
}

type ActionDoFunc func(ctx context.Context, model IModel, opts ...interface{}) error

type IModel interface {
	StartAction(ctx context.Context) (context.Context, error)
	GetState(ctx context.Context) (uint64, error)
	SetState(ctx context.Context, newState uint64, params ...interface{}) error
	EndAction(ctx context.Context, err error) error
}
