package blobstore

import (
	"errors"
	"fmt"
	"io"
	"os"

	container "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_blobstore_container"
	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_blobstore_types"
	"github.com/jamesstocktonj1/componentize-sdk/internal/pollable"
)

type objectImpl struct {
	name     string
	cont     *container.Container
	cursor   uint64
	outgoing *types.OutgoingValue
	stream   *types.OutputStream
}

var _ Object = (*objectImpl)(nil)

func newObject(name string, cont *container.Container) (*objectImpl, error) {
	outgoing := types.OutgoingValueNewOutgoingValue()
	streamRes := outgoing.OutgoingValueWriteBody()
	if streamRes.IsErr() {
		outgoing.Drop()
		return nil, errors.New("failed to open outgoing write body")
	}
	return &objectImpl{
		name:     name,
		cont:     cont,
		outgoing: outgoing,
		stream:   streamRes.Ok(),
	}, nil
}

func (o *objectImpl) Close() error {
	flushRes := o.stream.Flush()
	if flushRes.IsErr() {
		o.stream.Drop()
		return fmt.Errorf("failed to flush outgoing stream: %v", flushRes.Err())
	}
	pollable.AwaitAndDrop(o.stream.Subscribe())
	o.stream.Drop()
	contWriteRes := o.cont.WriteData(o.name, o.outgoing)
	if contWriteRes.IsErr() {
		return errors.New(contWriteRes.Err())
	}
	return nil
}

func (o *objectImpl) Read(p []byte) (int, error) {
	getRes := o.cont.GetData(o.name, o.cursor, o.cursor+uint64(len(p)))
	if getRes.IsErr() {
		return 0, errors.New(getRes.Err())
	}
	val := getRes.Ok()
	defer val.Drop()

	dataRes := types.IncomingValueIncomingValueConsumeSync(val)
	if dataRes.IsErr() {
		return 0, errors.New(dataRes.Err())
	}

	data := dataRes.Ok()
	copy(p, data)

	dataLen := len(data)
	o.cursor += uint64(dataLen)
	if dataLen < len(p) {
		return dataLen, io.EOF
	}
	return dataLen, nil
}

func (o *objectImpl) Write(p []byte) (int, error) {
	writeRes := o.stream.Write(p)
	if writeRes.IsErr() {
		return 0, fmt.Errorf("failed to write to outgoing stream: %v", writeRes.Err())
	}
	return len(p), nil
}

func (o *objectImpl) Stat() (os.FileInfo, error) {
	metaRes := o.cont.ObjectInfo(o.name)
	if metaRes.IsErr() {
		return nil, errors.New(metaRes.Err())
	}
	return &objectMeta{meta: metaRes.Ok()}, nil
}
