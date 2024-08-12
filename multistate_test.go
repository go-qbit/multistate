package multistate_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-qbit/multistate"
	. "github.com/go-qbit/multistate/expr"
)

type testEntity struct {
	state uint64
}

func (*testEntity) StartAction(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (m *testEntity) GetState(ctx context.Context) (uint64, error) {
	return m.state, nil
}

func (m *testEntity) SetState(ctx context.Context, newState uint64, params ...interface{}) error {
	m.state = newState
	return nil
}

func (*testEntity) EndAction(ctx context.Context, err error) error {
	return err
}

func (testEntity) GetId() interface{} {
	return 1
}

func (testEntity) GetTx() interface{} {
	return nil
}

func TestMultistate_DoAction(t *testing.T) {
	mst := multistate.New("New")

	signedA := mst.MustAddState(0, "signed_a", "Signed A")
	signedB := mst.MustAddState(1, "signed_b", "Signed B")
	signedC := mst.MustAddState(2, "signed_c", "Signed C")
	signedD := mst.MustAddState(3, "signed_d", "Signed D")
	signedE := mst.MustAddState(4, "signed_e", "Signed E")
	signedF := mst.MustAddState(5, "signed_f", "Signed F")

	mst.MustAddAction(
		"sign_a", "Sign A", Empty(),
		multistate.States{signedA}, nil,
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustAddAction(
		"sign_b", "Sign B", Empty(),
		multistate.States{signedB}, nil,
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustAddAction(
		"sign_c", "Sign C", And(Or(signedA, signedB), Not(signedC)),
		multistate.States{signedC}, nil,
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustAddAction(
		"sign_d", "Sign D", And(Or(signedC, signedE), Not(signedD)),
		multistate.States{signedD}, nil,
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustAddAction(
		"sign_e", "Sign E", And(Or(signedC, signedD), Not(signedE)),
		multistate.States{signedE}, nil,
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustAddAction(
		"sign_f", "Sign F", And(signedD, signedE, Not(signedF)),
		multistate.States{signedF}, nil,
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.AddCluster("Cluster 1", signedA)
	mst.AddCluster("Cluster 2", signedB)

	mst.MustCompile()

	m := &testEntity{}

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

	// ioutil.WriteFile("/tmp/graph.svg", []byte(mst.GetGraphSVG()), 0644)
}

func TestMultistate_DoAction2(t *testing.T) {
	mst := multistate.New("New")

	signedA := mst.MustAddState(0, "signed_a", "Signed A")
	signedB := mst.MustAddState(1, "signed_b", "Signed B")
	signedC := mst.MustAddState(2, "signed_c", "Signed C")
	signedD := mst.MustAddState(3, "signed_d", "Signed D")
	signedE := mst.MustAddState(4, "signed_e", "Signed E")
	signedF := mst.MustAddState(5, "signed_f", "Signed F")

	mst.MustAddAction(
		"sign_a", "Sign A", Empty(),
		multistate.States{signedA}, nil,
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustAddAction(
		"sign_b", "Sign B", Empty(),
		multistate.States{signedB}, nil,
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustAddAction(
		"sign_c", "Sign C", And(Or(signedA, signedB), Not(signedC)),
		multistate.States{signedC}, multistate.States{signedA, signedB},
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustAddAction(
		"sign_d", "Sign D", And(Or(signedC, signedE), Not(signedD)),
		multistate.States{signedD}, multistate.States{signedC},
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustAddAction(
		"sign_e", "Sign E", And(Or(signedC, signedD), Not(signedE)),
		multistate.States{signedE}, multistate.States{signedC},
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustAddAction(
		"sign_f", "Sign F", And(signedD, signedE, Not(signedF)),
		multistate.States{signedF}, multistate.States{signedD, signedE},
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustCompile()

	m := &testEntity{}

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

	// ioutil.WriteFile("/tmp/graph_wo.svg", []byte(mst.GetGraphSVG()), 0644)
}

func TestMultistate_GetMultistatesByRequiredAndForbiddenStateIds_NoError(t *testing.T) {
	mst := multistate.New("New")

	signedA := mst.MustAddState(0, "signed_a", "Signed A")
	signedB := mst.MustAddState(1, "signed_b", "Signed B")
	signedC := mst.MustAddState(2, "signed_c", "Signed C")
	signedD := mst.MustAddState(3, "signed_d", "Signed D")
	signedE := mst.MustAddState(4, "signed_e", "Signed E")

	mst.MustAddAction(
		"sign_a", "Sign A", Empty(),
		multistate.States{signedA}, nil,
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustAddAction(
		"sign_b", "Sign B", Empty(),
		multistate.States{signedB}, nil,
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustAddAction(
		"sign_c", "Sign C", And(Or(signedA, signedB), Not(signedC)),
		multistate.States{signedC}, nil,
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustAddAction(
		"sign_d", "Sign D", And(Or(signedC, signedE), Not(signedD)),
		multistate.States{signedD}, nil,
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustAddAction(
		"sign_e", "Sign E", And(Or(signedC, signedD), Not(signedE)),
		multistate.States{signedE}, multistate.States{signedC},
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustCompile()

	requiredIds := []string{"signed_a", "signed_c"}
	forbiddenIds := []string{"signed_d"}

	// Two options do satisfy the requirement:
	// 1. signed_a (1) + signed_c (4) = 5
	// 2. signed_a (1) + signed_c (4) + signed_e (16) = 21
	expectedResult := []uint64{5, 21}

	msts, err := mst.GetMultistatesByRequiredAndForbiddenStateIds(requiredIds, forbiddenIds)

	assert.Equal(t, expectedResult, msts)
	assert.NoError(t, err)
}

func TestMultistate_GetMultistatesByRequiredAndForbiddenStateIds_ErrorInvalidRequiredStateId(t *testing.T) {
	mst := multistate.New("New")

	signedA := mst.MustAddState(0, "signed_a", "Signed A")
	signedB := mst.MustAddState(1, "signed_b", "Signed B")
	signedC := mst.MustAddState(2, "signed_c", "Signed C")

	mst.MustAddAction(
		"sign_a", "Sign A", Empty(),
		multistate.States{signedA}, nil,
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustAddAction(
		"sign_b", "Sign B", Empty(),
		multistate.States{signedB}, nil,
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustAddAction(
		"sign_c", "Sign C", And(Or(signedA, signedB), Not(signedC)),
		multistate.States{signedC}, nil,
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustCompile()

	requiredIds := []string{"signed_a", "signed_e"}
	forbiddenIds := []string{"signed_d"}

	msts, err := mst.GetMultistatesByRequiredAndForbiddenStateIds(requiredIds, forbiddenIds)

	assert.Equal(t, []uint64(nil), msts)
	assert.ErrorIs(t, err, multistate.ErrInvalidState)
}

func TestMultistate_GetMultistatesByRequiredAndForbiddenStateIds_ErrorInvalidForbiddenStateId(t *testing.T) {
	mst := multistate.New("New")

	signedA := mst.MustAddState(0, "signed_a", "Signed A")
	signedB := mst.MustAddState(1, "signed_b", "Signed B")
	signedC := mst.MustAddState(2, "signed_c", "Signed C")

	mst.MustAddAction(
		"sign_a", "Sign A", Empty(),
		multistate.States{signedA}, nil,
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustAddAction(
		"sign_b", "Sign B", Empty(),
		multistate.States{signedB}, nil,
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustAddAction(
		"sign_c", "Sign C", And(Or(signedA, signedB), Not(signedC)),
		multistate.States{signedC}, nil,
		func(_ context.Context, _ multistate.Entity, _ ...interface{}) error {
			return nil
		},
		nil,
	)

	mst.MustCompile()

	requiredIds := []string{"signed_a", "signed_c"}
	forbiddenIds := []string{"signed_e"}

	msts, err := mst.GetMultistatesByRequiredAndForbiddenStateIds(requiredIds, forbiddenIds)

	assert.Equal(t, []uint64(nil), msts)
	assert.ErrorIs(t, err, multistate.ErrInvalidState)
}
