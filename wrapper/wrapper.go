package wrapper

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -L/usr/local/lib -L../lib -lrocksdb -lswallow -Wl,-rpath,./../lib
#include "../cpp/swallow.h"
#include <stdlib.h>
*/
import "C"

import "unsafe"

type Instance struct {
	ptr unsafe.Pointer
}

func NewInstance(dir string) *Instance {
	ptr := C.swallow_open(unsafe.Pointer(&dir), C.ulonglong(len(dir)))
	return &Instance{
		ptr: ptr,
	}
}

func (ins *Instance) Put(keys, values []string) {
	req := C.swallow_new_request()
	defer C.swallow_del_request(req)
	for i := 0; i < len(keys); i++ {
		C.swallow_request_append(req, unsafe.Pointer(&keys[i]), C.ulonglong(len(keys[i])),
			unsafe.Pointer(&values[i]), C.ulonglong(len(values[i])))
	}
	C.swallow_put(ins.ptr, req)
}

func (ins *Instance) Compact() {
	C.swallow_compact(ins.ptr)
}

func (ins *Instance) Close() {
	C.swallow_close(ins.ptr)
}
