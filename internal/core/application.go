package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Runnable interface {
	Run(ctx context.Context) error
}

var (
	ErrIsAlreadyStarted = errors.New("is already started")
)

type Application struct {
	runnables   []Runnable
	muRunnables sync.Mutex
	isStarted   bool
}

func NewApplication() *Application {
	return &Application{}
}

func (appl *Application) Register(r Runnable) {
	appl.muRunnables.Lock()
	defer appl.muRunnables.Unlock()
	appl.runnables = append(appl.runnables, r)
}

func (appl *Application) Run(ctx context.Context) error {
	if appl.isStarted {
		return ErrIsAlreadyStarted
	}

	ctx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	errCh := make(chan error, len(appl.runnables))
	defer close(errCh)

	wg := sync.WaitGroup{}
	wg.Add(len(appl.runnables))

	// FIXME: handle signals

	for i := range appl.runnables {
		go startRunnable(ctx, &wg, appl.runnables[i], errCh)
	}

	// #FIXME: looks ugly - why do you think 100 ms is enough?
	select {
	case err := <-errCh:
		return fmt.Errorf("failed to start application: %w", err)

	case <-time.After(100 * time.Millisecond):
	}

	appl.isStarted = true
	wg.Wait()

	return nil
}

func startRunnable(ctx context.Context, wg *sync.WaitGroup, r Runnable, errCh chan<- error) {
	defer wg.Done()
	if err := r.Run(ctx); err != nil {
		errCh <- err
	}
}
