package gotransactions

import (
	"bytes"
	"encoding/gob"
)

type StateSerializer[S any] interface {
	SerializeState(*S) ([]byte, error)
	DeserializeState([]byte) (*S, error)
}

// Transactioner is a generic interface for a transaction
// Transactions contain a list of stages, a state (which can be anything that gob can serialize), and a name
type Transaction[S any, T Stager[S]] struct {
	State *S
	Stages []*T
	Name string
}

// NewTransaction creates a new transaction with the given name and state
func NewTransaction[S any, T Stager[S]](Name string, State *S, Stages []*T) *Transaction[S, T] {
	return &Transaction[S, T]{State: State, Name: Name, Stages: Stages}
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
