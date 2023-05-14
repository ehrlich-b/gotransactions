package gotransactions

import "time"

type StageStatus string

const (
	StageStatusSuccess     StageStatus = "success"
	StageStatusFailure     StageStatus = "failure"
	StageStatusYield       StageStatus = "yield"
	StageStatusJobComplete StageStatus = "job_complete"
	StageStatusJobFailed   StageStatus = "job_failed"
)

type StageResult struct {
	Status StageStatus
	Error  error
	Yield  time.Duration
}

type Stager interface {
	Name() string
	Concurrency() int
	Execute(*Transaction) StageResult
	Rollback(*Transaction) error
}
