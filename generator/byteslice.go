package generator

import (
	"sync"

	"google.golang.org/protobuf/compiler/protogen"
)

type (
	ByteSlice struct {
		lock  sync.Mutex
		marks map[protogen.GoImportPath]map[string]string
	}
)

var (
	byteSlice     *ByteSlice
	byteSliceLock sync.Mutex
)

func GetByteSlice() *ByteSlice {
	byteSliceLock.Lock()
	defer byteSliceLock.Unlock()

	if byteSlice == nil {
		byteSlice = &ByteSlice{
			marks: make(map[protogen.GoImportPath]map[string]string),
		}
	}

	return byteSlice
}

func (bs *ByteSlice) GetBytesPoolName(goIdent protogen.GoIdent) string {
	bs.lock.Lock()
	defer bs.lock.Unlock()

	if _, ok := bs.marks[goIdent.GoImportPath]; !ok {
		bs.marks[goIdent.GoImportPath] = make(map[string]string)
	}

	name := `vtprotoPool_` + goIdent.GoName + `_bytes`
	bs.marks[goIdent.GoImportPath][goIdent.GoName] = name

	return name
}

func (bs *ByteSlice) GenerateCode(importPath protogen.GoImportPath, p *GeneratedFile) {
	bs.lock.Lock()
	defer bs.lock.Unlock()

	for goName, name := range bs.marks[importPath] {
		if name == "" {
			continue
		}
		bs.marks[importPath][goName] = ""

		p.P(`var `, name, ` = `, p.Ident("sync", "Pool"), `{`)
		p.P(`New: func() interface{} {`)
		p.P(`b := make([]byte, 0, 8)`)
		p.P(`return &b`)
		p.P(`},`)
		p.P(`}`)
		p.P()
	}
}
