package gotransactions

import (
	"github.com/google/uuid"
)

type Transaction struct {
	AutoRollback bool
	state        map[string]any
	stages       []Stager
	saver        StateSaver
	currentStage int
	name         string
	guid         string
}

func NewTransaction(name string, stages []Stager, saver StateSaver) *Transaction {
	return &Transaction{
		state:        make(map[string]any),
		stages:       stages,
		saver:        saver,
		currentStage: 0,
		name:         name,
		guid:         uuid.New().String(),
	}
}

func (t *Transaction) Run() StageResult {
	if t.currentStage >= len(t.stages) {
		return StageResult{
			Status: StageStatusJobComplete,
		}
	}

	result := t.stages[t.currentStage].Execute(t)
	if result.Status == StageStatusSuccess {
		t.currentStage++
	} else if result.Status == StageStatusYield {
		return result
	} else if result.Status == StageStatusFailure {
		if t.AutoRollback {
			t.Rollback()
		}
		return result
	}

	if t.currentStage >= len(t.stages) {
		return StageResult{
			Status: StageStatusJobComplete,
		}
	}

	return result
}

func (t *Transaction) RunAll() StageResult {
	var result StageResult
	for t.currentStage < len(t.stages) {
		result = t.Run()
		t.SaveState()
		if result.Status != StageStatusSuccess {
			break
		}
	}

	return result
}

// Return [transaction_name].[stage_name]
func (t *Transaction) CurrentStageName() string {
	if t.currentStage < 0 || t.currentStage >= len(t.stages) {
		return ""
	}

	return t.name + "." + t.stages[t.currentStage].Name()
}

func (t *Transaction) CurrentStage() Stager {
	if t.currentStage < 0 || t.currentStage >= len(t.stages) {
		return nil
	}

	return t.stages[t.currentStage]
}

func (t *Transaction) SaveState() error {
	return t.saver.SaveState(t)
}

func (t *Transaction) GetState(stageName, key string) any {
	if v, ok := t.state[getConfigKey(stageName, key)]; ok {
		return v
	}
	return nil
}

func (t *Transaction) SetState(stageName, key string, value any) {
	t.state[getConfigKey(stageName, key)] = value
}

func getConfigKey(stageName, key string) string {
	if stageName == "" {
		return key
	}
	return stageName + "." + key
}

func (t *Transaction) Rollback() error {
	if t.currentStage <= 0 {
		return nil
	}

	for i := t.currentStage; i >= 0; i-- {
		err := t.stages[i].Rollback(t)
		if err != nil {
			return err
		}
		t.currentStage--
	}

	return nil
}
