package main

import (
	"time"

	"github.com/ehrlich-b/gotransactions"
)

type SleepStage struct {
}

func (s *SleepStage) Name() string {
	return "SleepStage"
}

func (s *SleepStage) Concurrency() int {
	return 5
}

func (s *SleepStage) Execute(t *gotransactions.Transaction) gotransactions.StageResult {
	Print("Sleeping...")
	time.Sleep(5 * time.Second)
	return gotransactions.StageResult{
		Status: gotransactions.StageStatusSuccess,
	}
}

func (s *SleepStage) Rollback(t *gotransactions.Transaction) error {
	return nil
}
