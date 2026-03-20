package blobstore

import (
	"os"
	"time"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_blobstore_types"
)

type objectMeta struct {
	meta types.ObjectMetadata
}

var _ os.FileInfo = (*objectMeta)(nil)

func (m *objectMeta) Name() string {
	return m.meta.Name
}

func (m *objectMeta) Size() int64 {
	return int64(m.meta.Size)
}

func (m *objectMeta) Mode() os.FileMode {
	return 0
}

func (m *objectMeta) ModTime() time.Time {
	return time.Unix(int64(m.meta.CreatedAt), 0)
}

func (m *objectMeta) IsDir() bool {
	return false
}

func (m *objectMeta) Sys() any {
	return m.meta
}
