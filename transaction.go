package gotransactions

import (
	"bytes"
	"encoding/gob"

	"github.com/google/uuid"
)

type StateSerializer[S any] interface {
	SerializeState(*S) ([]byte, error)
	DeserializeState([]byte) (*S, error)
}

// Transactioner is a generic interface for a transaction
// Transactions contain a list of stages, a state (which can be anything that gob can serialize), and a name
type Transaction[S any, T Stager[S]] struct {
	State S
	Stages []T
	CurrentStage int
	Name string
	guid string
	rollbacks []func(S) error
}

// NewTransaction creates a new transaction with the given name and state
func NewTransaction[S any, T Stager[S]](Name string, State S, Stages []T) *Transaction[S, T] {
	t := &Transaction[S, T]{State: State, Name: Name, Stages: Stages}
	t.guid = uuid.New().String()

	return t
}

func (t *Transaction[S, T]) Run(autoRollback bool) StageResult {
	if t.CurrentStage >= len(t.Stages) {
		return StageResult{Status: StageStatusJobComplete}
	}
	if t.CurrentStage < 0 {
		return StageResult{Status: StageStatusJobFailed}
	}

	var result StageResult
	stage := t.Stages[t.CurrentStage]

	result = stage.Execute(t.State)
	t.rollbacks = append(t.rollbacks, stage.Rollback)
	if result.Status == StageStatusSuccess {
		t.CurrentStage++
	}

	if t.CurrentStage == len(t.Stages) {
		result.Status = StageStatusJobComplete
	}

	if result.Status == StageStatusFailure && autoRollback {
		err := t.Rollback()
		if err != nil {
			result.Error = err
		}
	}

	return result
}

func (t *Transaction[S, T]) RunAll(autoRollback bool) StageResult {
	var result StageResult
	for t.CurrentStage < len(t.Stages) {
		result = t.Run(autoRollback)
		if result.Status != StageStatusSuccess {
			break
		}
	}

	return result
}

func (t *Transaction[S, T]) Rollback() error {
	if t.CurrentStage < 0 {
		return nil
	}
	t.CurrentStage = -1
	for i := len(t.rollbacks) - 1; i >= 0; i-- {
		err := t.rollbacks[i](t.State)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Transaction[S, T]) SerializeState() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(t.State)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (t *Transaction[S, T]) DeserializeState(StateBytes []byte) (*S, error) {
	var state *S
	dec := gob.NewDecoder(bytes.NewReader(StateBytes))
	err := dec.Decode(&state)
	if err != nil {
		return nil, err
	}

	return state, nil
}
