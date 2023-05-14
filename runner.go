package gotransactions

import (
	"context"
	"sync"
)

type Runner struct {
	activeTransactions sync.Map
	runQ               chan *Transaction
	stageQs            map[string]chan *Transaction
	yieldQ             ConcurrentQueue[*Transaction]
	Syncer             StateSaver
}

func NewRunner() *Runner {
	return &Runner{
		activeTransactions: sync.Map{},
		Syncer:             NewFileStateSaver(),
	}
}

func (r *Runner) Run(t *Transaction) StageResult {
	r.activeTransactions.Store(t.guid, t)
	result := t.Run()
	r.activeTransactions.Delete(t.guid)
	return result
}

func (r *Runner) Start(ctx context.Context) {
	r.runQ = make(chan *Transaction)
	go r.runQWorker(ctx)
}

func (r *Runner) runQWorker(ctx context.Context) {
	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			return
		case t := <-r.runQ:
			stage := t.CurrentStageName()
			if _, ok := r.stageQs[stage]; !ok {
				r.stageQs[stage] = make(chan *Transaction)
				go r.stageQDistributor(ctx, stage)
			}
			r.stageQs[stage] <- t
		}
	}
}

func (r *Runner) stageQDistributor(ctx context.Context, stage string) {
	var initialized bool
	queue := NewConcurrentQueue[*Transaction]()
	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			return
		case t := <-r.stageQs[stage]:
			queue.Enqueue(t)
			if !initialized {
				workersToInitialize := t.CurrentStage().Concurrency()
				for i := 0; i < workersToInitialize; i++ {
					go r.stageQWorker(ctx, stage, queue)
				}
				initialized = true
			}
		}
	}
}

func (r *Runner) stageQWorker(ctx context.Context, stage string, queue *ConcurrentQueue[*Transaction]) {
	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			return
			case t := <-queue.Dequeue():
			// Run the transaction
			result := t.Run()
			// Sync the state of the transaction
			r.Syncer.SaveState(t)

			if result.Status == StageStatusYield {
				queue.Enqueue(t)
			}
			// If success, back to the runQ to be added to the next stage's queue
			if result.Status == StageStatusSuccess {
				r.runQ <- t
			}
			// All other states the transaction drops out of the runner
		}
	}
}
