package main

import (
	"context"
	"fmt"

	"github.com/MetalBlueberry/go-minions/pkg/minions"
)

type PrintHelloWorld struct {
}

func (PrintHelloWorld) Work(context.Context) {
	fmt.Println("hello world")
}

func main() {
	lord := minions.NewLord()
	lord.StartQuest(context.Background(), 1, minions.NewQuest([]minions.Worker{
		PrintHelloWorld{},
	}))
	lord.Wait()
	fmt.Println("done")
}
