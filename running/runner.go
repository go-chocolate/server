package running

import (
	"context"
	"fmt"
	"sync"
)

type Runner interface {
	Run(ctx context.Context) error
	Shutdown()
}

type Group struct {
	runner []Runner

	ctx    context.Context
	cancel context.CancelFunc
	err    chan error

	done       context.Context
	doneCancel context.CancelFunc
}

func NewGroup(runner ...Runner) *Group {
	return &Group{runner: runner}
}

func (g *Group) Add(runner ...Runner) *Group {
	g.runner = append(g.runner, runner...)
	return g
}

func (g *Group) Run() {
	wg := new(sync.WaitGroup)
	wg.Add(len(g.runner))
	g.ctx, g.cancel = context.WithCancel(context.Background())
	g.done, g.doneCancel = context.WithCancel(context.Background())
	g.err = make(chan error, 1)
	for _, v := range g.runner {
		go g.run(v, wg)
	}
	go g.wait(wg)
}

func (g *Group) wait(wg *sync.WaitGroup) {
	defer g.doneCancel()
	wg.Wait()
}

func (g *Group) run(runner Runner, wg *sync.WaitGroup) {
	err := g.safeRun(runner)
	wg.Done()
	if err != nil {
		g.err <- err
	}
}

func (g *Group) safeRun(runner Runner) error {
	var err error
	defer func() {
		if recovered := recover(); recovered != nil {
			if innerError, ok := recovered.(error); ok {
				err = innerError
			} else {
				err = fmt.Errorf("%v", recovered)
			}
		}
	}()
	err = runner.Run(g.ctx)
	return err
}

func (g *Group) Error() <-chan error {
	return g.err
}

func (g *Group) Done() <-chan struct{} {
	return g.done.Done()
}

func (g *Group) Shutdown() {
	g.cancel()
}
