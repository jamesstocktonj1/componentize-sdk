package blobstore

import container "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_blobstore_container"

type containerImpl struct {
	cont *container.Container
}

var _ Container = (*containerImpl)(nil)

func (c *containerImpl) Close() error {
	c.cont.Drop()
	return nil
}

func (c *containerImpl) Open(name string) (Object, error) {
	return newObject(name, c.cont)
}
