package socket

import "errors"

var (
	ErrUnknown                  = errors.New("unknown error")
	ErrAccessDenied             = errors.New("access denied")
	ErrNotSupported             = errors.New("operation not supported")
	ErrInvalidArgument          = errors.New("invalid argument")
	ErrOutOfMemory              = errors.New("out of memory")
	ErrTimeout                  = errors.New("operation timed out")
	ErrConcurrencyConflict      = errors.New("concurrency conflict")
	ErrNotInProgress            = errors.New("operation not in progress")
	ErrWouldBlock               = errors.New("operation would block")
	ErrInvalidState             = errors.New("invalid state")
	ErrNewSocketLimit           = errors.New("new socket limit reached")
	ErrAddressNotBindable       = errors.New("address not bindable")
	ErrAddressInUse             = errors.New("address already in use")
	ErrRemoteUnreachable        = errors.New("remote unreachable")
	ErrConnectionRefused        = errors.New("connection refused")
	ErrConnectionReset          = errors.New("connection reset")
	ErrConnectionAborted        = errors.New("connection aborted")
	ErrDatagramTooLarge         = errors.New("datagram too large")
	ErrNameUnresolvable         = errors.New("name unresolvable")
	ErrTemporaryResolverFailure = errors.New("temporary resolver failure")
	ErrPermanentResolverFailure = errors.New("permanent resolver failure")
)
