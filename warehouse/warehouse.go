package warehouse

import (
	"swallow/wrapper"
	"sync"
)

type InstanceWrapper struct {
	ins   *wrapper.Instance
	start int64
	end   int64
	path  string
}

type Warehouse struct {
	cmtx *sync.RWMutex
	pmtx *sync.RWMutex
	cur  *InstanceWrapper
	ch   chan *InstanceWrapper
	prev []*InstanceWrapper
}

func (w *Warehouse) Rotate() {
	ptr := &InstanceWrapper{
		ins: wrapper.NewInstance()
	}
}
