package warehouse

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"swallow/wrapper"
	"sync"
	"time"

	"github.com/tidwall/gjson"
	"github.com/uopensail/swallow-idl/proto/swallowapi"
	"github.com/uopensail/ulib/prome"
	"github.com/uopensail/ulib/utils"
)

const success = "SUCCESS"
const interval int64 = 3600

type Shard struct {
	ins   *wrapper.Instance
	path  string
	start int64
}

type Warehouse struct {
	sync.RWMutex
	cur      *Shard
	ch       chan *Shard
	workdir  string
	ip       string
	pk       string
	interval int64
}

func NewWarehouse(workdir, pk string) *Warehouse {
	ip, _ := utils.GetLocalIp()
	dataDir := path.Join(workdir, ip)
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.Mkdir(dataDir, 0755); err != nil {
			panic(err)
		}
	}

	shards := listShards(dataDir)
	ch := make(chan *Shard, len(shards))
	var cur *Shard
	ts := time.Now().Unix()
	for i := 0; i < len(shards); i++ {
		if shards[i].start+interval > ts {
			cur = shards[i]
		} else {
			ch <- shards[i]
		}
	}
	if cur == nil {
		cur = &Shard{
			ins:   wrapper.NewInstance(path.Join(dataDir, fmt.Sprintf("%d", ts))),
			path:  dataDir,
			start: ts,
		}
	}
	w := &Warehouse{
		cur:      cur,
		ip:       ip,
		pk:       pk,
		ch:       ch,
		workdir:  workdir,
		interval: interval,
	}
	go w.run()
	go w.compact()
	return w
}

func (w *Warehouse) rotate() {
	stat := prome.NewStat("warehouse.rotate")
	defer stat.End()
	ts := time.Now().Unix()
	dataDir := path.Join(w.workdir, w.ip, fmt.Sprintf("%d", ts))
	shard := &Shard{
		ins:   wrapper.NewInstance(dataDir),
		path:  dataDir,
		start: ts,
	}
	w.Lock()
	pre := w.cur
	w.cur = shard
	w.Unlock()
	w.ch <- pre
}

func (w *Warehouse) compact() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			continue
		case shard := <-w.ch:
			stat := prome.NewStat("warehouse.compact")
			defer stat.End()
			shard.ins.Compact()
			shard.ins.Close()
			markSuccess(path.Join(shard.path, success))
		}
	}
}

func (w *Warehouse) run() {
	ticker := time.NewTicker(time.Second * time.Duration(w.interval))
	defer ticker.Stop()
	for {
		<-ticker.C
		w.rotate()
	}
}

func (w *Warehouse) Trace(kvs []*swallowapi.KV) {
	stat := prome.NewStat("warehouse.Put")
	defer stat.End()
	keys := make([][]byte, len(kvs))
	vs := make([][]byte, len(kvs))

	for i := 0; i < len(kvs); i++ {
		keys[i] = kvs[i].Key
		vs[i] = kvs[i].Value
	}
	stat.SetCounter(len(kvs))
	w.RLock()
	defer w.RUnlock()
	w.cur.ins.Put(keys, vs)
}

func (w *Warehouse) Log(kvs []*swallowapi.KV) {
	stat := prome.NewStat("warehouse.Put")
	defer stat.End()
	keys := make([][]byte, len(kvs))
	vs := make([][]byte, len(kvs))
	ts := time.Now().UnixNano()
	rd := rand.Int63n(1000000)

	for i := 0; i < len(kvs); i++ {
		vs[i] = kvs[i].Value
		keys[i] = []byte(fmt.Sprintf("%s|%d|%d-%d",
			gjson.GetBytes(vs[i], w.pk).String(), ts, rd, i))

	}
	stat.SetCounter(len(kvs))
	w.RLock()
	defer w.RUnlock()
	w.cur.ins.Put(keys, vs)
}

func (w *Warehouse) Close() {
	w.Lock()
	defer w.Unlock()
	w.cur.ins.Close()
	close(w.ch)
}
