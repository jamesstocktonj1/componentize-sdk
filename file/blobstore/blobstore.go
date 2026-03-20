package blobstore

import (
	"errors"
	"io"
	"os"

	blobstore "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_blobstore_blobstore"
)

func Open(name string) (Container, error) {
	contRes := blobstore.GetContainer(name)
	if contRes.IsErr() {
		return nil, errors.New(contRes.Err())
	}
	return &containerImpl{cont: contRes.Ok()}, nil
}

func Create(name string) (Container, error) {
	contRes := blobstore.CreateContainer(name)
	if contRes.IsErr() {
		return nil, errors.New(contRes.Err())
	}
	return &containerImpl{cont: contRes.Ok()}, nil
}

type Container interface {
	io.Closer
	Name() (string, error)
	Open(string) (Object, error)
}

type Object interface {
	io.ReadWriteCloser
	Stat() (os.FileInfo, error)
}
