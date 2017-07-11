package edit

import (
	"sync"

	"github.com/jumpscale/elvish/daemon/api"
	"github.com/jumpscale/elvish/eval"
)

// History implements the $le:history variable. It is list-like.
type History struct {
	mutex *sync.RWMutex
	st    *api.Client
}

var _ eval.ListLike = History{}

func (hv History) Kind() string {
	return "list"
}

func (hv History) Repr(int) string {
	return "$le:history"
}

func (hv History) Len() int {
	hv.mutex.RLock()
	defer hv.mutex.RUnlock()

	nextseq, err := hv.st.NextCmdSeq()
	maybeThrow(err)
	return nextseq - 1
}

func (hv History) Iterate(f func(eval.Value) bool) {
	hv.mutex.RLock()
	defer hv.mutex.RUnlock()

	n := hv.Len()
	cmds, err := hv.st.Cmds(1, n+1)
	maybeThrow(err)

	for _, cmd := range cmds {
		if !f(eval.String(cmd)) {
			break
		}
	}
}

func (hv History) IndexOne(idx eval.Value) eval.Value {
	hv.mutex.RLock()
	defer hv.mutex.RUnlock()

	slice, i, j := eval.ParseAndFixListIndex(eval.ToString(idx), hv.Len())
	if slice {
		cmds, err := hv.st.Cmds(i+1, j+1)
		maybeThrow(err)
		vs := make([]eval.Value, len(cmds))
		for i := range cmds {
			vs[i] = eval.String(cmds[i])
		}
		return eval.NewList(vs...)
	}
	s, err := hv.st.Cmd(i + 1)
	maybeThrow(err)
	return eval.String(s)
}
