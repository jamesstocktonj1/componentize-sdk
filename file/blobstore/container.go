package blobstore

import (
	"errors"

	container "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_blobstore_container"
)

type containerImpl struct {
	cont *container.Container
}

var _ Container = (*containerImpl)(nil)

func (c *containerImpl) Close() error {
	c.cont.Drop()
	return nil
}

func (c *containerImpl) Name() (string, error) {
	nameRes := c.cont.Name()
	if nameRes.IsErr() {
		return "", errors.New(nameRes.Err())
	}
	return nameRes.Ok(), nil
}

func (c *containerImpl) Open(name string) (Object, error) {
	return newObject(name, c.cont)
}
