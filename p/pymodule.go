package p

/*
#include "Python.h"
*/
import "C"
import (
	"fmt"
	"pfi/sensorbee/sensorbee/data"
	"runtime"
	"unsafe"
)

// ObjectModule is a bind of `PyObject`, used as `PyModule`
type ObjectModule struct {
	Object
}

// LoadModule loads `name` module. The module needs to be placed at `sys.path`.
// User can set optional `sys.path` using `ImportSysAndAppendPath`
func LoadModule(name string) (ObjectModule, error) {
	cModule := C.CString(name)
	defer C.free(unsafe.Pointer(cModule))

	type Result struct {
		val ObjectModule
		err error
	}
	ch := make(chan *Result, 1)
	go func() {
		runtime.LockOSThread()
		state := GILState_Ensure()
		defer GILState_Release(state)

		pyMdl, err := C.PyImport_ImportModule(cModule)
		if pyMdl == nil {
			ch <- &Result{ObjectModule{}, fmt.Errorf("cannot load '%v' module: %v",
				name, err)}
			return
		}

		ch <- &Result{ObjectModule{Object{p: pyMdl}}, nil}
	}()
	res := <-ch

	return res.val, res.err
}

// NewInstance returns `name` constructor.
func (m *ObjectModule) NewInstance(name string, args ...data.Value) (
	ObjectInstance, error) {
	return newInstance(m, name, args...)
}

// NewInstanceWithKwd returns 'name' constructor with named arguments.
func (m *ObjectModule) NewInstanceWithKwd(name string, kwdArgs data.Map) (
	ObjectInstance, error) {
	return newInstance(m, name, kwdArgs)
}

// GetClass returns `name` class instance.
// User needs to call DecRef when finished using instance.
func (m *ObjectModule) GetClass(name string) (ObjectInstance, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	type Result struct {
		val ObjectInstance
		err error
	}
	ch := make(chan *Result, 1)
	go func() {
		runtime.LockOSThread()
		state := GILState_Ensure()
		defer GILState_Release(state)

		pyInstance := C.PyObject_GetAttrString(m.p, cName)
		if pyInstance == nil {
			ch <- &Result{ObjectInstance{}, fmt.Errorf("cannot get '%v' instance",
				name)}
			return
		}
		ch <- &Result{ObjectInstance{Object{p: pyInstance}}, nil}
	}()
	res := <-ch

	return res.val, res.err
}

// Call calls `name` function.
//  argument type: ...data.Value
//  return type:   data.Value
func (m *ObjectModule) Call(name string, args ...data.Value) (data.Value, error) {
	return invoke(m.p, name, args...)
}
