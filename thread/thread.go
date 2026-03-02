package thread

import "sync"

type routineGroup struct {
	waitGroup sync.WaitGroup
}

type RoutineGroup interface {
	Run(fn func())
	Wait()
}

func NewRoutineGroup() RoutineGroup {
	return new(routineGroup)
}

func (g *routineGroup) Run(fn func()) {

	g.waitGroup.Go(func() {
		fn()
	})
}

func (g *routineGroup) Wait() {
	g.waitGroup.Wait()
}
