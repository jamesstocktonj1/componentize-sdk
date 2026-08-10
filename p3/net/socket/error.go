package socket

import (
	"errors"
	"fmt"

	lookup "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_sockets_ip_name_lookup"
	sockets "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_sockets_types"
)

var errorCodeMessages = map[uint8]error{
	sockets.ErrorCodeAccessDenied:       errors.New("socket error: access-denied"),
	sockets.ErrorCodeNotSupported:       errors.New("socket error: not-supported"),
	sockets.ErrorCodeInvalidArgument:    errors.New("socket error: invalid-argument"),
	sockets.ErrorCodeOutOfMemory:        errors.New("socket error: out-of-memory"),
	sockets.ErrorCodeTimeout:            errors.New("socket error: timeout"),
	sockets.ErrorCodeInvalidState:       errors.New("socket error: invalid-state"),
	sockets.ErrorCodeAddressNotBindable: errors.New("socket error: address-not-bindable"),
	sockets.ErrorCodeAddressInUse:       errors.New("socket error: address-in-use"),
	sockets.ErrorCodeRemoteUnreachable:  errors.New("socket error: remote-unreachable"),
	sockets.ErrorCodeConnectionRefused:  errors.New("socket error: connection-refused"),
	sockets.ErrorCodeConnectionBroken:   errors.New("socket error: connection-broken"),
	sockets.ErrorCodeConnectionReset:    errors.New("socket error: connection-reset"),
	sockets.ErrorCodeConnectionAborted:  errors.New("socket error: connection-aborted"),
	sockets.ErrorCodeDatagramTooLarge:   errors.New("socket error: datagram-too-large"),
	sockets.ErrorCodeOther:              errors.New("socket error: other"),
}

func mapErrorCode(err sockets.ErrorCode) error {
	if msg, ok := errorCodeMessages[err.Tag()]; ok {
		return msg
	}
	return fmt.Errorf("socket error: unknown(%d)", err.Tag())
}

var lookupErrorCodeMessages = map[uint8]error{
	lookup.ErrorCodeAccessDenied:             errors.New("lookup error: access-denied"),
	lookup.ErrorCodeInvalidArgument:          errors.New("lookup error: invalid-argument"),
	lookup.ErrorCodeNameUnresolvable:         errors.New("lookup error: name-unresolvable"),
	lookup.ErrorCodeTemporaryResolverFailure: errors.New("lookup error: temporary-resolver-failure"),
	lookup.ErrorCodePermanentResolverFailure: errors.New("lookup error: permanent-resolver-failure"),
	lookup.ErrorCodeOther:                    errors.New("lookup error: other"),
}

func mapLookupErrorCode(err lookup.ErrorCode) error {
	if msg, ok := lookupErrorCodeMessages[err.Tag()]; ok {
		return msg
	}
	return fmt.Errorf("lookup error: unknown(%d)", err.Tag())
}
