package gotransactions

import "time"

type StageStatus string

type StageResult struct {
	Status StageStatus
	Error  error
	Yield  time.Duration
}

type Stager[S any] interface {
	Execute(*S) StageResult
	Rollback(*S) error
}
