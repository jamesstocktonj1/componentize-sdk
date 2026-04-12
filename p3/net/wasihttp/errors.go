package wasihttp

import "errors"

// ErrDnsTimeout is returned when a DNS lookup timed out.
var ErrDnsTimeout = errors.New("DNS timeout")

// ErrDestinationNotFound is returned when the destination hostname could not be resolved.
var ErrDestinationNotFound = errors.New("destination not found")

// ErrDestinationUnavailable is returned when the destination host is currently unreachable.
var ErrDestinationUnavailable = errors.New("destination unavailable")

// ErrDestinationIpProhibited is returned when a network policy prohibits connecting to the destination IP.
var ErrDestinationIpProhibited = errors.New("destination IP prohibited")

// ErrDestinationIpUnroutable is returned when the destination IP address has no route.
var ErrDestinationIpUnroutable = errors.New("destination IP unroutable")

// ErrConnectionRefused is returned when the destination host actively refused the connection.
var ErrConnectionRefused = errors.New("connection refused")

// ErrConnectionTerminated is returned when an established connection was closed before the request completed.
var ErrConnectionTerminated = errors.New("connection terminated")

// ErrConnectionTimeout is returned when a connection could not be established within the allowed time.
var ErrConnectionTimeout = errors.New("connection timeout")

// ErrConnectionReadTimeout is returned when waiting for data from the server exceeded the allowed time.
var ErrConnectionReadTimeout = errors.New("connection read timeout")

// ErrConnectionWriteTimeout is returned when sending data to the server exceeded the allowed time.
var ErrConnectionWriteTimeout = errors.New("connection write timeout")

// ErrConnectionLimitReached is returned when no new connections can be made because the connection pool is full.
var ErrConnectionLimitReached = errors.New("connection limit reached")

// ErrTlsProtocolError is returned when a TLS handshake or protocol violation occurred.
var ErrTlsProtocolError = errors.New("TLS protocol error")

// ErrTlsCertificateError is returned when the server's TLS certificate could not be validated.
var ErrTlsCertificateError = errors.New("TLS certificate error")

// ErrHttpRequestDenied is returned when the HTTP gateway denied the outbound request.
var ErrHttpRequestDenied = errors.New("HTTP request denied")

// ErrHttpRequestLengthRequired is returned when the server requires a Content-Length header but none was provided.
var ErrHttpRequestLengthRequired = errors.New("HTTP request length required")

// ErrHttpRequestMethodInvalid is returned when the HTTP method is not valid or not permitted.
var ErrHttpRequestMethodInvalid = errors.New("HTTP request method invalid")

// ErrHttpRequestUriInvalid is returned when the request URI is malformed.
var ErrHttpRequestUriInvalid = errors.New("HTTP request URI invalid")

// ErrHttpRequestUriTooLong is returned when the request URI exceeds the maximum allowed length.
var ErrHttpRequestUriTooLong = errors.New("HTTP request URI too long")

// ErrHttpResponseIncomplete is returned when the response was cut off before it was fully received.
var ErrHttpResponseIncomplete = errors.New("HTTP response incomplete")

// ErrHttpResponseTimeout is returned when the server did not send a complete response within the allowed time.
var ErrHttpResponseTimeout = errors.New("HTTP response timeout")

// ErrHttpUpgradeFailed is returned when an HTTP protocol upgrade (e.g. WebSocket) was rejected or failed.
var ErrHttpUpgradeFailed = errors.New("HTTP upgrade failed")

// ErrHttpProtocolError is returned when the server sent a response that violates the HTTP protocol.
var ErrHttpProtocolError = errors.New("HTTP protocol error")

// ErrLoopDetected is returned when a request cycle was detected (e.g. a proxy forwarding to itself).
var ErrLoopDetected = errors.New("loop detected")

// ErrConfigurationError is returned when the HTTP handler or proxy is misconfigured.
var ErrConfigurationError = errors.New("configuration error")
