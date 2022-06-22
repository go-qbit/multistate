package multistate_test

import (
	"testing"

	"github.com/go-qbit/multistate"
)

type impl struct {
	St1 multistate.State `bit:"0" caption:"State 1"`
	St2 multistate.State `bit:"1" caption:"State 2"`

	S string
}

func (i *impl) Action1() multistate.Action {
	return multistate.Action{
		Caption: i.S + " action",
		Set:     multistate.States{i.St2},
	}
}

func TestNew(t *testing.T) {
	multistate.NewFromStruct(&impl{S: "ttt"})
}
