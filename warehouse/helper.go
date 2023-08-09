package warehouse

import (
	"os"
	"path"
	"strconv"
	"swallow/wrapper"

	"github.com/uopensail/ulib/prome"
	"github.com/uopensail/ulib/zlog"
	"go.uber.org/zap"
)

func markSuccess(file string) {
	f, err := os.Create(file)
	if err != nil {
		return
	}
	defer f.Close()
}

func listShards(dir string) []*Shard {
	stat := prome.NewStat("warehouse.listShards")
	defer stat.End()
	files, err := os.ReadDir(dir)
	if err != nil {
		stat.MarkErr()
		zlog.LOG.Error("list shards error", zap.String("dir", dir))
		return nil
	}
	ret := make([]*Shard, 0, len(files))
	for i := 0; i < len(files); i++ {
		if !files[i].IsDir() {
			continue
		}
		ts, err := strconv.ParseInt(files[i].Name(), 10, 64)
		if err != nil {
			continue
		}
		if _, err := os.Stat(path.Join(dir, files[i].Name(), success)); os.IsNotExist(err) {
			dataDir := path.Join(dir, files[i].Name())
			ret = append(ret, &Shard{
				ins:   wrapper.NewInstance(dataDir),
				path:  dataDir,
				start: ts,
			})
		}
	}
	stat.SetCounter(len(ret))
	return ret
}
