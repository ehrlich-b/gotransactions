# Go Transactions

This is an experimental project implementing something like a super simple version of temporal.io workers.

A unit of work is called a `Transaction`. `Transaction`s have `Stages` and a `State`. `Stages` are executed in order. `Stages` can return three different `State`s: `Success`, `Failure` and `Yield`. `Success` moves the `Transaction` to the next `Stage`. `Failure` stops the `Transaction` and `Yield` pauses the `Transaction` and puts it back into the queue after a specified delay.

Example `Stage`:

```go
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
```

`Transactions` wrap multiple `Stages` and specify a state saver. A `FileStateSaver` is provided.

Example `Transaction`:

```go
	stages := []gotransactions.Stager{
		&PrintStage{},
		&SleepStage{},
		&DiceStage{},
		&PrintStage{},
	}
	saver := &gotransactions.FileStateSaver{}
	t := gotransactions.NewTransaction("HelloWorld", stages, saver)
	t.SetState("", "message", "My roll is gonna be great!")
```

This transaction can be run as-is with t.RunAll(). But a "kanban" style runner is also provided. The runner looks at the `Concurrency()` of each `Stage` and runs that many workers to service that type of `Stage` across `Transaction` types.

Example runner:

```go
    r := gotransactions.NewRunner()
	ctx := context.Background()
	r.Start(ctx)
	r.ExecuteAll([]*gotransactions.Transaction{t, t2, t3})
```

The runner is cancellable via the context. `Start()` starts the goroutines necessary to run `Transaction`s in the background. `ExecuteAll()` is a non-blocking call that queues up `Transaction`s to be run, according to their `Concurrency()`.
