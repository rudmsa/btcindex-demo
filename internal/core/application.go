package core

import (
	"context"
	"sync"
)

type Runnable interface {
	Run(ctx context.Context) error
}

type Application struct {
	runnables   []Runnable
	muRunnables sync.Mutex
}

func NewApplication() *Application {
	return &Application{}
}

func (appl *Application) Register(r Runnable) {
	appl.muRunnables.Lock()
	defer appl.muRunnables.Unlock()
	appl.runnables = append(appl.runnables, r)
}

func (appl *Application) Run(ctx context.Context) {
	ctx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	wg := sync.WaitGroup{}

	// FIXME: handle signals

	for i := range appl.runnables {
		wg.Add(1)
		go startRunnable(ctx, &wg, appl.runnables[i])
	}
	wg.Wait()
}

func startRunnable(ctx context.Context, wg *sync.WaitGroup, r Runnable) error {
	defer wg.Done()
	return r.Run(ctx)
}
