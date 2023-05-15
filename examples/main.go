package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ehrlich-b/gotransactions"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	stages := []gotransactions.Stager{
		&PrintStage{},
		&SleepStage{},
		&DiceStage{},
		&PrintStage{},
	}
	saver := &gotransactions.FileStateSaver{}
	t := gotransactions.NewTransaction("HelloWorld", stages, saver)
	t.SetState("", "message", "My roll is gonna be great!")
	t2 := gotransactions.NewTransaction("HelloWorld", stages, saver)
	t2.SetState("", "message", "Mine even better!")
	t3 := gotransactions.NewTransaction("HelloWorld", stages, saver)
	t3.SetState("", "message", "Here I go")
	r := gotransactions.NewRunner()
	ctx := context.Background()
	r.Start(ctx)
	r.ExecuteAll([]*gotransactions.Transaction{t, t2, t3})
	time.Sleep(100 * time.Second)
}

var start = time.Now()

func Print(message any) {
	fmt.Printf("%v: %v\n", time.Since(start), message)
}

var _ gotransactions.Stager = (*PrintStage)(nil)
