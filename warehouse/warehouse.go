package warehouse

import (
	"swallow/wrapper"
	"sync"
)

type Shard struct {
	ins   *wrapper.Instance
	start int64
	path  string
}

type Warehouse struct {
	cmtx *sync.RWMutex
	pmtx *sync.RWMutex
	cur  *Shard
	shards []*Shard
	workdir string
	datadir string
	ip string
	ch   chan *Shard
}

func (w *Warehouse) Rotate() {
	ptr := &InstanceWrapper{
		ins: wrapper.NewInstance()
	}
}
