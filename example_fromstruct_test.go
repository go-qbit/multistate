package multistate_test

import (
	"context"
	"github.com/go-qbit/multistate"
	. "github.com/go-qbit/multistate/expr"
	"log"
)

type ExampleImpl struct {
	// Multistate flags
	St1 multistate.State `bit:"0" caption:"State 1"`
	St2 multistate.State `bit:"1" caption:"State 2"`

	// Here can be other data
}

// Global callback, will be called on each DoAction
func (*ExampleImpl) OnDoAction(ctx context.Context, prevState, newState uint64, action string, opts ...interface{}) error {
	log.Printf("%d -> %s -> %d", prevState, action, newState)
	return nil
}

// All actions must start from Action prefix
func (i *ExampleImpl) ActionTest() multistate.Action {
	return multistate.Action{
		Caption: "Test action",
		From:    Any(),
		Set:     multistate.States{i.St2},
		Reset:   nil,
		OnDo: func(ctx context.Context, id interface{}, opts ...interface{}) error {
			return nil
		},
	}
}

func ExampleNewFromStruct() {
	mst := multistate.NewFromStruct(&ExampleImpl{})

	e := &testEntity{}
	_, err := mst.DoAction(context.Background(), e, "test")
	if err != nil {
		panic(err)
	}
}
