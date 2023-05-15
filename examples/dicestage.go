package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/ehrlich-b/gotransactions"
)

// Dice stage rolls a dice and prints the result. If the result is anything but 6, it yields for 1 second.
type DiceStage struct {
}

func (s *DiceStage) Name() string {
	return "DiceStage"
}

func (s *DiceStage) Concurrency() int {
	return 1
}

func (s *DiceStage) Execute(t *gotransactions.Transaction) gotransactions.StageResult {
	result := rand.Intn(6) + 1
	Print(fmt.Sprintf("Rolled a %v", result))
	if result != 6 {
		return gotransactions.StageResult{
			Status: gotransactions.StageStatusYield,
			Yield:  1 * time.Second,
		}
	}
	return gotransactions.StageResult{
		Status: gotransactions.StageStatusSuccess,
	}
}

func (s *DiceStage) Rollback(t *gotransactions.Transaction) error {
	return nil
}
