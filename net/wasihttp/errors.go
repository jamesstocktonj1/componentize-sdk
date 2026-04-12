package wasihttp

import "errors"

// Sentinel errors for WASI HTTP error codes with no payload.
var (
	ErrDnsTimeout                = errors.New("DNS timeout")
	ErrDestinationNotFound       = errors.New("destination not found")
	ErrDestinationUnavailable    = errors.New("destination unavailable")
	ErrDestinationIpProhibited   = errors.New("destination IP prohibited")
	ErrDestinationIpUnroutable   = errors.New("destination IP unroutable")
	ErrConnectionRefused         = errors.New("connection refused")
	ErrConnectionTerminated      = errors.New("connection terminated")
	ErrConnectionTimeout         = errors.New("connection timeout")
	ErrConnectionReadTimeout     = errors.New("connection read timeout")
	ErrConnectionWriteTimeout    = errors.New("connection write timeout")
	ErrConnectionLimitReached    = errors.New("connection limit reached")
	ErrTlsProtocolError          = errors.New("TLS protocol error")
	ErrTlsCertificateError       = errors.New("TLS certificate error")
	ErrHttpRequestDenied         = errors.New("HTTP request denied")
	ErrHttpRequestLengthRequired = errors.New("HTTP request length required")
	ErrHttpRequestMethodInvalid  = errors.New("HTTP request method invalid")
	ErrHttpRequestUriInvalid     = errors.New("HTTP request URI invalid")
	ErrHttpRequestUriTooLong     = errors.New("HTTP request URI too long")
	ErrHttpResponseIncomplete    = errors.New("HTTP response incomplete")
	ErrHttpResponseTimeout       = errors.New("HTTP response timeout")
	ErrHttpUpgradeFailed         = errors.New("HTTP upgrade failed")
	ErrHttpProtocolError         = errors.New("HTTP protocol error")
	ErrLoopDetected              = errors.New("loop detected")
	ErrConfigurationError        = errors.New("configuration error")
)
