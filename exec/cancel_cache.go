package exec

import (
	"context"
	"sync"
)

type ScriptCanceller struct {
	cancelFuncs []context.CancelFunc
	sync.Mutex
}

var scriptCanceller ScriptCanceller

func init() {
	scriptCanceller = ScriptCanceller{
		cancelFuncs: make([]context.CancelFunc, 0),
		Mutex:       sync.Mutex{},
	}
}

func CacheCancelFunc(f context.CancelFunc) {
	scriptCanceller.Lock()
	defer scriptCanceller.Unlock()
	scriptCanceller.cancelFuncs = append(scriptCanceller.cancelFuncs, f)
}

func CancelScripts() {
	scriptCanceller.Lock()
	defer scriptCanceller.Unlock()
	for _, f := range scriptCanceller.cancelFuncs {
		f()
	}
}
