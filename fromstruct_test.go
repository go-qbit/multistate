package multistate_test

import (
	"context"
	"log"
	"testing"

	"github.com/go-qbit/multistate"
)

type impl struct {
	St1 multistate.State `bit:"0" caption:"State 1"`
	St2 multistate.State `bit:"1" caption:"State 2"`

	S string
}

func (i *impl) OnDoAction(ctx context.Context, prevState, newState uint64, action string, opts ...interface{}) error {
	log.Printf("%d -> %s -> %d", prevState, action, newState)
	return nil
}

func (i *impl) ActionTest() multistate.Action {
	return multistate.Action{
		Caption: i.S + " action",
		Set:     multistate.States{i.St2},
	}
}

func TestNew(t *testing.T) {
	m := multistate.NewFromStruct(&impl{S: "ttt"})

	e := &testEntity{}
	_, err := m.DoAction(context.Background(), e, "test")
	if err != nil {
		t.Error(err)
	}
}
