package loc

import (
	"reflect"
	"sync"
	"unsafe"
)

//nolint
type (
	funcID uint8

	funcInfo struct {
		entry *uintptr
		datap unsafe.Pointer
	}

	inlinedCall struct {
		parent   int16  // index of parent in the inltree, or < 0
		funcID   funcID // type of the called function
		_        byte
		file     int32 // fileno index into filetab
		line     int32 // line number of the call site
		func_    int32 // offset into pclntab for name of called function
		parentPc int32 // position of an instruction whose source position is the call site (offset from entry)
	}

	nfl struct {
		name string
		file string
		line int
	}
)

var (
	locmu sync.Mutex
	locc  = map[PC]nfl{}
)

//go:noescape
//go:linkname callers runtime.callers
func callers(skip int, pc []PC) int

//go:noescape
//go:linkname caller1 runtime.callers
func caller1(skip int, pc *PC, len, cap int) int //nolint:predeclared

// NameFileLine returns function name, file and line number for location.
//
// This works only in the same binary where location was captured.
//
// This functions is a little bit modified version of runtime.(*Frames).Next().
func (l PC) NameFileLine() (name, file string, line int) {
	if l == 0 {
		return
	}

	locmu.Lock()
	c, ok := locc[l]
	locmu.Unlock()
	if ok {
		return c.name, c.file, c.line
	}

	name, file, line = l.nameFileLine()

	if file != "" {
		file = cropFilename(file, name)
	}

	locmu.Lock()
	locc[l] = nfl{
		name: name,
		file: file,
		line: line,
	}
	locmu.Unlock()

	return
}

// FuncEntry is functions entry point.
func (l PC) FuncEntry() PC {
	funcInfo := findfunc(l)
	if funcInfo.entry == nil {
		return 0
	}
	return PC(*funcInfo.entry)
}

// SetCache sets name, file and line for the PC.
// It allows to work with PC in another binary the same as in original.
func SetCache(l PC, name, file string, line int) {
	locmu.Lock()
	if name == "" && file == "" {
		delete(locc, l)
	} else {
		locc[l] = nfl{
			name: name,
			file: file,
			line: line,
		}
	}
	locmu.Unlock()
}

func Cached(l PC) (ok bool) {
	locmu.Lock()
	_, ok = locc[l]
	locmu.Unlock()
	return
}

func noescapeSlize(b *[128]byte) []byte {
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&b[0])),
		Cap:  128,
	}))
}
