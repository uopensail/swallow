package wrapper

type Instance struct {
}

func NewInstance(dir string) *Instance {
	return nil
}

func (ins *Instance) Put(keys, values []string) {

}

func (ins *Instance) Scan(start, end string) []string {
	return nil
}

func (ins *Instance) Compact() {

}

func (ins *Instance) Close() {}
