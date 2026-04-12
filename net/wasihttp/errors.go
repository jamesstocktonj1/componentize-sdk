package wasihttp

import "errors"

// Sentinel errors for WASI HTTP error codes with no payload.
var (
	// ErrDnsTimeout is returned when a DNS lookup timed out.
	ErrDnsTimeout = errors.New("DNS timeout")
	// ErrDestinationNotFound is returned when the destination hostname could not be resolved.
	ErrDestinationNotFound = errors.New("destination not found")
	// ErrDestinationUnavailable is returned when the destination host is currently unreachable.
	ErrDestinationUnavailable = errors.New("destination unavailable")
	// ErrDestinationIpProhibited is returned when a network policy prohibits connecting to the destination IP.
	ErrDestinationIpProhibited = errors.New("destination IP prohibited")
	// ErrDestinationIpUnroutable is returned when the destination IP address has no route.
	ErrDestinationIpUnroutable = errors.New("destination IP unroutable")
	// ErrConnectionRefused is returned when the destination host actively refused the connection.
	ErrConnectionRefused = errors.New("connection refused")
	// ErrConnectionTerminated is returned when an established connection was closed before the request completed.
	ErrConnectionTerminated = errors.New("connection terminated")
	// ErrConnectionTimeout is returned when a connection could not be established within the allowed time.
	ErrConnectionTimeout = errors.New("connection timeout")
	// ErrConnectionReadTimeout is returned when waiting for data from the server exceeded the allowed time.
	ErrConnectionReadTimeout = errors.New("connection read timeout")
	// ErrConnectionWriteTimeout is returned when sending data to the server exceeded the allowed time.
	ErrConnectionWriteTimeout = errors.New("connection write timeout")
	// ErrConnectionLimitReached is returned when no new connections can be made because the connection pool is full.
	ErrConnectionLimitReached = errors.New("connection limit reached")
	// ErrTlsProtocolError is returned when a TLS handshake or protocol violation occurred.
	ErrTlsProtocolError = errors.New("TLS protocol error")
	// ErrTlsCertificateError is returned when the server's TLS certificate could not be validated.
	ErrTlsCertificateError = errors.New("TLS certificate error")
	// ErrHttpRequestDenied is returned when the HTTP gateway denied the outbound request.
	ErrHttpRequestDenied = errors.New("HTTP request denied")
	// ErrHttpRequestLengthRequired is returned when the server requires a Content-Length header but none was provided.
	ErrHttpRequestLengthRequired = errors.New("HTTP request length required")
	// ErrHttpRequestMethodInvalid is returned when the HTTP method is not valid or not permitted.
	ErrHttpRequestMethodInvalid = errors.New("HTTP request method invalid")
	// ErrHttpRequestUriInvalid is returned when the request URI is malformed.
	ErrHttpRequestUriInvalid = errors.New("HTTP request URI invalid")
	// ErrHttpRequestUriTooLong is returned when the request URI exceeds the maximum allowed length.
	ErrHttpRequestUriTooLong = errors.New("HTTP request URI too long")
	// ErrHttpResponseIncomplete is returned when the response was cut off before it was fully received.
	ErrHttpResponseIncomplete = errors.New("HTTP response incomplete")
	// ErrHttpResponseTimeout is returned when the server did not send a complete response within the allowed time.
	ErrHttpResponseTimeout = errors.New("HTTP response timeout")
	// ErrHttpUpgradeFailed is returned when an HTTP protocol upgrade (e.g. WebSocket) was rejected or failed.
	ErrHttpUpgradeFailed = errors.New("HTTP upgrade failed")
	// ErrHttpProtocolError is returned when the server sent a response that violates the HTTP protocol.
	ErrHttpProtocolError = errors.New("HTTP protocol error")
	// ErrLoopDetected is returned when a request cycle was detected (e.g. a proxy forwarding to itself).
	ErrLoopDetected = errors.New("loop detected")
	// ErrConfigurationError is returned when the HTTP handler or proxy is misconfigured.
	ErrConfigurationError = errors.New("configuration error")
)
