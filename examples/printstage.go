package main

import "github.com/ehrlich-b/gotransactions"

type PrintStage struct {
}

func (s *PrintStage) Name() string {
	return "PrintStage"
}

func (s *PrintStage) Concurrency() int {
	return 1
}

func (s *PrintStage) Execute(t *gotransactions.Transaction) gotransactions.StageResult {
	message := t.GetState("", "message")
	if message == nil {
		message = "Hello, World!"
	}
	Print(message)
	return gotransactions.StageResult{
		Status: gotransactions.StageStatusSuccess,
	}
}

func (s *PrintStage) Rollback(t *gotransactions.Transaction) error {
	return nil
}

var _ gotransactions.Stager = (*PrintStage)(nil)
