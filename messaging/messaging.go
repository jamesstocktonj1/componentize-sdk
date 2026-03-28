package messaging

import (
	"errors"

	gen "github.com/jamesstocktonj1/componentize-sdk/gen/wasmcloud_messaging_handler"
	gentypes "github.com/jamesstocktonj1/componentize-sdk/gen/wasmcloud_messaging_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// Message represents a message exchanged with a message broker.
type Message struct {
	// Subject is the topic or routing key for the message.
	Subject string
	// Body holds the raw message payload.
	Body []byte
	// ReplyTo is the optional subject to reply to. nil means no reply-to is set.
	ReplyTo *string
}

// HandleMessage forwards msg to the imported wasmcloud:messaging/handler handle-message function.
func HandleMessage(msg Message) error {
	var replyTo witTypes.Option[string]
	if msg.ReplyTo != nil {
		replyTo = witTypes.Some(*msg.ReplyTo)
	} else {
		replyTo = witTypes.None[string]()
	}

	result := gen.HandleMessage(gentypes.BrokerMessage{
		Subject: msg.Subject,
		Body:    msg.Body,
		ReplyTo: replyTo,
	})
	if result.IsErr() {
		return errors.New(result.Err())
	}
	return nil
}
