package multistate_test

import (
	"context"
	"fmt"

	"github.com/go-qbit/multistate"
	. "github.com/go-qbit/multistate/expr"
)

// The implementation structure
type ExampleImpl struct {
	// Multistate flags
	St1 multistate.State `bit:"0" caption:"State 1"`
	St2 multistate.State `bit:"1" caption:"State 2"`

	// Here can be other data
}

// Define clusters
func (i *ExampleImpl) Clusters() []multistate.Cluster {
	return []multistate.Cluster{
		{"test", i.St2},
	}
}

// Global callback, will be called on each DoAction
func (i *ExampleImpl) OnDoAction(ctx context.Context, entity multistate.Entity, prevState, newState uint64, action string, opts ...interface{}) error {
	fmt.Printf("[%d]: %d -> %s -> %d", entity.GetId(), prevState, action, newState)
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

// The example entity
type ExampleEntity struct {
	Id uint32
}

func (*ExampleEntity) StartAction(ctx context.Context) (context.Context, error) { return ctx, nil }
func (*ExampleEntity) GetState(context.Context) (uint64, error)                 { return 0, nil }
func (*ExampleEntity) SetState(context.Context, uint64, ...interface{}) error   { return nil }
func (*ExampleEntity) EndAction(context.Context, error) error                   { return nil }
func (e *ExampleEntity) GetId() interface{}                                     { return e.Id }

func ExampleNewFromStruct() {
	mst := multistate.NewFromStruct(&ExampleImpl{})

	e := &ExampleEntity{Id: 100}
	_, err := mst.DoAction(context.Background(), e, "test")
	if err != nil {
		panic(err)
	}
	// Output: [100]: 0 -> test -> 2
}
