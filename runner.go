package gotransactions

type Scheduler string

const (
	SchedulerTransaction Scheduler = "sequential"
	SchedulerParallel   Scheduler = "parallel"
)

type Runner[S any, T Stager[S]] struct {
	transactions []*Transaction[S, T]
	Syncer       StateSyncer
}


func NewRunner[S any, T Stager[S]](syncer StateSyncer) *Runner[S, T] {
	return &Runner[S, T]{Syncer: syncer}
}

func (r *Runner[S, T]) Run(t *Transaction[S, T]) {
	r.transactions = append(r.transactions, t)
}

func (r *Runner[S, T]) Start() {

}

func (r *Runner[S, T]) Stop() {

}
