package blobstore

import (
	"errors"
	"fmt"
	"io"

	container "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_blobstore_container"
	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_blobstore_types"
)

type objectImpl struct {
	name string
	cont *container.Container

	// read
	cursor uint64

	// write
}

var _ Object = (*objectImpl)(nil)

func (o *objectImpl) Close() error {
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
	if dataLen < len(p) {
		return dataLen, io.EOF
	}
	return dataLen, nil
}

func (o *objectImpl) Write(p []byte) (int, error) {
	// move to object open?
	outgoing := types.OutgoingValueNewOutgoingValue()
	defer outgoing.Drop()

	// move to object open?
	streamRes := outgoing.OutgoingValueWriteBody()
	if streamRes.IsErr() {
		return 0, errors.New("failed to open outgoing write body")
	}
	stream := streamRes.Ok()
	defer stream.Drop()

	// change to stream.Write
	writeRes := stream.BlockingWriteAndFlush(p)
	if writeRes.IsErr() {
		return 0, fmt.Errorf("failed to write to outgoing stream - %+v", writeRes.Err())
	}

	// move this to Close()
	contWriteRes := o.cont.WriteData(o.name, outgoing)
	if contWriteRes.IsErr() {
		return 0, errors.New(contWriteRes.Err())
	}
	return len(p), nil
}
