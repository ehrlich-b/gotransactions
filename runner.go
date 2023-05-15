package gotransactions

import (
	"context"
	"sync"
	"time"
)

type Runner struct {
	activeTransactions sync.Map
	runQ               chan *Transaction
	stageQs            map[string]chan *Transaction
	yieldQ             futureHeap // A queue of transactions that have yielded, ordered by the time they should be resumed
	Syncer             StateSaver
	newYield           chan *future
}

func NewRunner() *Runner {
	return &Runner{
		activeTransactions: sync.Map{},
		Syncer:             NewFileStateSaver(),
	}
}

type future struct {
	transaction *Transaction
	initTime    time.Time
	futureTime  time.Time
}

func (r *Runner) Execute(t *Transaction) {
	r.runQ <- t
}

func (r *Runner) ExecuteAll(t []*Transaction) {
	for _, t := range t {
		r.runQ <- t
	}
}

func (r *Runner) ExecuteAfter(t *Transaction, d time.Duration) {
	r.yield(t, d)
}

func (r *Runner) Start(ctx context.Context) {
	r.runQ = make(chan *Transaction)
	r.stageQs = make(map[string]chan *Transaction)
	r.yieldQ = make([]*future, 0)
	r.newYield = make(chan *future)
	go r.runQWorker(ctx)
	go r.yieldRunner(ctx)
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
				r.yield(t, result.Yield)
			}
			// If success, back to the runQ to be added to the next stage's queue
			if result.Status == StageStatusSuccess {
				r.Execute(t)
			}
			// All other states the transaction drops out of the runner
		}
	}
}

func (r *Runner) yield(t *Transaction, duration time.Duration) {
	f := &future{
		transaction: t,
		initTime:    time.Now(),
		futureTime:  time.Now().Add(duration),
	}

	r.newYield <- f
}

func (r *Runner) yieldRunner(ctx context.Context) {
	timer := time.NewTimer(time.Hour * 24 * 365 * 100) // 100 years
	if !timer.Stop() {
		<-timer.C
	}
	for ctx.Err() == nil {
		if r.yieldQ.Len() == 0 {
			timer.Reset(time.Hour * 24 * 365 * 100) // 100 years
		} else {
			nextRunAt := r.yieldQ[0].futureTime
			if time.Now().After(nextRunAt) {
				// Just run it right away
				f := r.yieldQ.Pop().(*future)
				r.Execute(f.transaction)
				continue
			}
			timer.Reset(time.Until(nextRunAt))
		}

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case f := <-r.newYield:
			r.yieldQ.Push(f)
			if !timer.Stop() {
				<-timer.C
			}
		case <-timer.C:
			f := r.yieldQ.Pop().(*future)
			r.Execute(f.transaction)
		}
	}
}

// implements heap.Interface
type futureHeap []*future

func (h futureHeap) Len() int {
	return len(h)
}

func (h futureHeap) Less(i, j int) bool {
	return h[i].futureTime.Before(h[j].futureTime)
}

func (h futureHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *futureHeap) Push(x interface{}) {
	*h = append(*h, x.(*future))
}

func (h *futureHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
