package wrapper

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -L/usr/local/lib -L../lib -lrocksdb -lswallow -Wl,-rpath,./../lib
#include "../cpp/swallow.h"
#include <stdlib.h>
*/
import "C"

import (
	"reflect"
	"unsafe"
)

type Instance struct {
	ptr unsafe.Pointer
}

func Str2bytes(s string) (b []byte) {
	/* #nosec G103 */
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	/* #nosec G103 */
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return b
}

func NewInstance(dir string) *Instance {
	ptr := C.swallow_open(unsafe.Pointer(&Str2bytes(dir)[0]), C.ulonglong(len(dir)))
	if ptr == nil {
		return nil
	}
	return &Instance{
		ptr: ptr,
	}
}

func (ins *Instance) Put(keys, values []string) {
	req := C.swallow_new_request()
	defer C.swallow_del_request(req)
	for i := 0; i < len(keys); i++ {
		C.swallow_request_append(req, unsafe.Pointer(&Str2bytes(keys[i])[0]), C.ulonglong(len(keys[i])),
			unsafe.Pointer(&Str2bytes(values[i])[0]), C.ulonglong(len(values[i])))
	}
	C.swallow_put(ins.ptr, req)
}

func (ins *Instance) Compact() {
	C.swallow_compact(ins.ptr)
}

func (ins *Instance) Close() {
	C.swallow_close(ins.ptr)
}
