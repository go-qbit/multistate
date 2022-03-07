package multistate_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-qbit/multistate"
	. "github.com/go-qbit/multistate/expr"
)

type testModel struct {
	state uint64
}

func (*testModel) StartAction(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (m *testModel) GetState(ctx context.Context) (uint64, error) {
	return m.state, nil
}

func (m *testModel) SetState(ctx context.Context, newState uint64, params ...interface{}) error {
	m.state = newState
	return nil
}

func (*testModel) EndAction(ctx context.Context, err error) error {
	return err
}

func TestMultistate_DoAction(t *testing.T) {
	mst := multistate.New("New")

	signedA := mst.AddState(0, "signed_a", "Signed A")
	signedB := mst.AddState(1, "signed_b", "Signed B")
	signedC := mst.AddState(2, "signed_c", "Signed C")
	signedD := mst.AddState(3, "signed_d", "Signed D")
	signedE := mst.AddState(4, "signed_e", "Signed E")
	signedF := mst.AddState(5, "signed_f", "Signed F")

	mst.AddAction(
		"sign_a", "Sign A", Empty(),
		[]multistate.IState{signedA}, nil,
		func(ctx context.Context, model multistate.IModel, opts ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.AddAction(
		"sign_b", "Sign B", Empty(),
		[]multistate.IState{signedB}, nil,
		func(ctx context.Context, model multistate.IModel, opts ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.AddAction(
		"sign_c", "Sign C", And(Or(signedA, signedB), Not(signedC)),
		[]multistate.IState{signedC}, nil,
		func(ctx context.Context, model multistate.IModel, opts ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.AddAction(
		"sign_d", "Sign D", And(Or(signedC, signedE), Not(signedD)),
		[]multistate.IState{signedD}, nil,
		func(ctx context.Context, model multistate.IModel, opts ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.AddAction(
		"sign_e", "Sign E", And(Or(signedC, signedD), Not(signedE)),
		[]multistate.IState{signedE}, nil,
		func(ctx context.Context, model multistate.IModel, opts ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.AddAction(
		"sign_f", "Sign F", And(signedD, signedE, Not(signedF)),
		[]multistate.IState{signedF}, nil,
		func(ctx context.Context, model multistate.IModel, opts ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.Compile()

	m := &testModel{}

	_, err := mst.DoAction(context.Background(), m, "sign_a")
	assert.NoError(t, err)

	_, err = mst.DoAction(context.Background(), m, "sign_c")
	assert.NoError(t, err)

	_, err = mst.DoAction(context.Background(), m, "sign_d")
	assert.NoError(t, err)

	_, err = mst.DoAction(context.Background(), m, "sign_e")
	assert.NoError(t, err)

	_, err = mst.DoAction(context.Background(), m, "sign_f")
	assert.NoError(t, err)

	assert.Equal(t, uint64(61), m.state)

	//ioutil.WriteFile("/tmp/graph.svg", []byte(mst.GetGraphSVG()), 0644)
}

func TestMultistate_DoAction2(t *testing.T) {
	mst := multistate.New("New")

	signedA := mst.AddState(0, "signed_a", "Signed A")
	signedB := mst.AddState(1, "signed_b", "Signed B")
	signedC := mst.AddState(2, "signed_c", "Signed C")
	signedD := mst.AddState(3, "signed_d", "Signed D")
	signedE := mst.AddState(4, "signed_e", "Signed E")
	signedF := mst.AddState(5, "signed_f", "Signed F")

	mst.AddAction(
		"sign_a", "Sign A", Empty(),
		[]multistate.IState{signedA}, nil,
		func(ctx context.Context, model multistate.IModel, opts ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.AddAction(
		"sign_b", "Sign B", Empty(),
		[]multistate.IState{signedB}, nil,
		func(ctx context.Context, model multistate.IModel, opts ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.AddAction(
		"sign_c", "Sign C", And(Or(signedA, signedB), Not(signedC)),
		[]multistate.IState{signedC}, []multistate.IState{signedA, signedB},
		func(ctx context.Context, model multistate.IModel, opts ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.AddAction(
		"sign_d", "Sign D", And(Or(signedC, signedE), Not(signedD)),
		[]multistate.IState{signedD}, []multistate.IState{signedC},
		func(ctx context.Context, model multistate.IModel, opts ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.AddAction(
		"sign_e", "Sign E", And(Or(signedC, signedD), Not(signedE)),
		[]multistate.IState{signedE}, []multistate.IState{signedC},
		func(ctx context.Context, model multistate.IModel, opts ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.AddAction(
		"sign_f", "Sign F", And(signedD, signedE, Not(signedF)),
		[]multistate.IState{signedF}, []multistate.IState{signedD, signedE},
		func(ctx context.Context, model multistate.IModel, opts ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.Compile()

	m := &testModel{}

	_, err := mst.DoAction(context.Background(), m, "sign_a")
	assert.NoError(t, err)

	_, err = mst.DoAction(context.Background(), m, "sign_c")
	assert.NoError(t, err)

	_, err = mst.DoAction(context.Background(), m, "sign_d")
	assert.NoError(t, err)

	_, err = mst.DoAction(context.Background(), m, "sign_e")
	assert.NoError(t, err)

	_, err = mst.DoAction(context.Background(), m, "sign_f")
	assert.NoError(t, err)

	assert.Equal(t, uint64(32), m.state)

	//ioutil.WriteFile("/tmp/graph_wo.svg", []byte(mst.GetGraphSVG()), 0644)
}
