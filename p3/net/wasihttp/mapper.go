package wasihttp

import (
	"errors"
	"fmt"
	"net/url"

	httpTypes "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

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

func mapUrlScheme(u *url.URL) witTypes.Option[httpTypes.Scheme] {
	switch u.Scheme {
	case "http":
		return witTypes.Some(httpTypes.MakeSchemeHttp())
	case "https":
		return witTypes.Some(httpTypes.MakeSchemeHttps())
	default:
		return witTypes.Some(httpTypes.MakeSchemeOther(u.Scheme))
	}
}

func mapErrorCode(e httpTypes.ErrorCode) error {
	switch e.Tag() {
	case httpTypes.ErrorCodeDnsTimeout:
		return ErrDnsTimeout

	case httpTypes.ErrorCodeDnsError:
		p := e.DnsError()
		rcode, infoCode := "", uint16(0)
		if !p.Rcode.IsNone() {
			rcode = p.Rcode.Some()
		}
		if !p.InfoCode.IsNone() {
			infoCode = p.InfoCode.Some()
		}
		return fmt.Errorf("DNS error: rcode=%q, infoCode=%d", rcode, infoCode)

	case httpTypes.ErrorCodeDestinationNotFound:
		return ErrDestinationNotFound

	case httpTypes.ErrorCodeDestinationUnavailable:
		return ErrDestinationUnavailable

	case httpTypes.ErrorCodeDestinationIpProhibited:
		return ErrDestinationIpProhibited

	case httpTypes.ErrorCodeDestinationIpUnroutable:
		return ErrDestinationIpUnroutable

	case httpTypes.ErrorCodeConnectionRefused:
		return ErrConnectionRefused

	case httpTypes.ErrorCodeConnectionTerminated:
		return ErrConnectionTerminated

	case httpTypes.ErrorCodeConnectionTimeout:
		return ErrConnectionTimeout

	case httpTypes.ErrorCodeConnectionReadTimeout:
		return ErrConnectionReadTimeout

	case httpTypes.ErrorCodeConnectionWriteTimeout:
		return ErrConnectionWriteTimeout

	case httpTypes.ErrorCodeConnectionLimitReached:
		return ErrConnectionLimitReached

	case httpTypes.ErrorCodeTlsProtocolError:
		return ErrTlsProtocolError

	case httpTypes.ErrorCodeTlsCertificateError:
		return ErrTlsCertificateError

	case httpTypes.ErrorCodeTlsAlertReceived:
		p := e.TlsAlertReceived()
		alertId, alertMsg := uint8(0), ""
		if !p.AlertId.IsNone() {
			alertId = p.AlertId.Some()
		}
		if !p.AlertMessage.IsNone() {
			alertMsg = p.AlertMessage.Some()
		}
		return fmt.Errorf("TLS alert received: alertId=%d, alertMessage=%q", alertId, alertMsg)

	case httpTypes.ErrorCodeHttpRequestDenied:
		return ErrHttpRequestDenied

	case httpTypes.ErrorCodeHttpRequestLengthRequired:
		return ErrHttpRequestLengthRequired

	case httpTypes.ErrorCodeHttpRequestBodySize:
		if size := e.HttpRequestBodySize(); !size.IsNone() {
			return fmt.Errorf("HTTP request body size error: limit=%d", size.Some())
		}
		return fmt.Errorf("HTTP request body size error")

	case httpTypes.ErrorCodeHttpRequestMethodInvalid:
		return ErrHttpRequestMethodInvalid

	case httpTypes.ErrorCodeHttpRequestUriInvalid:
		return ErrHttpRequestUriInvalid

	case httpTypes.ErrorCodeHttpRequestUriTooLong:
		return ErrHttpRequestUriTooLong

	case httpTypes.ErrorCodeHttpRequestHeaderSectionSize:
		if size := e.HttpRequestHeaderSectionSize(); !size.IsNone() {
			return fmt.Errorf("HTTP request header section size error: limit=%d", size.Some())
		}
		return fmt.Errorf("HTTP request header section size error")

	case httpTypes.ErrorCodeHttpRequestHeaderSize:
		if opt := e.HttpRequestHeaderSize(); !opt.IsNone() {
			p := opt.Some()
			fieldName, fieldSize := "", uint32(0)
			if !p.FieldName.IsNone() {
				fieldName = p.FieldName.Some()
			}
			if !p.FieldSize.IsNone() {
				fieldSize = p.FieldSize.Some()
			}
			return fmt.Errorf("HTTP request header size error: field=%q, limit=%d", fieldName, fieldSize)
		}
		return fmt.Errorf("HTTP request header size error")

	case httpTypes.ErrorCodeHttpRequestTrailerSectionSize:
		if size := e.HttpRequestTrailerSectionSize(); !size.IsNone() {
			return fmt.Errorf("HTTP request trailer section size error: limit=%d", size.Some())
		}
		return fmt.Errorf("HTTP request trailer section size error")

	case httpTypes.ErrorCodeHttpRequestTrailerSize:
		p := e.HttpRequestTrailerSize()
		fieldName, fieldSize := "", uint32(0)
		if !p.FieldName.IsNone() {
			fieldName = p.FieldName.Some()
		}
		if !p.FieldSize.IsNone() {
			fieldSize = p.FieldSize.Some()
		}
		return fmt.Errorf("HTTP request trailer size error: field=%q, limit=%d", fieldName, fieldSize)

	case httpTypes.ErrorCodeHttpResponseIncomplete:
		return ErrHttpResponseIncomplete

	case httpTypes.ErrorCodeHttpResponseHeaderSectionSize:
		if size := e.HttpResponseHeaderSectionSize(); !size.IsNone() {
			return fmt.Errorf("HTTP response header section size error: limit=%d", size.Some())
		}
		return fmt.Errorf("HTTP response header section size error")

	case httpTypes.ErrorCodeHttpResponseHeaderSize:
		p := e.HttpResponseHeaderSize()
		fieldName, fieldSize := "", uint32(0)
		if !p.FieldName.IsNone() {
			fieldName = p.FieldName.Some()
		}
		if !p.FieldSize.IsNone() {
			fieldSize = p.FieldSize.Some()
		}
		return fmt.Errorf("HTTP response header size error: field=%q, limit=%d", fieldName, fieldSize)

	case httpTypes.ErrorCodeHttpResponseBodySize:
		if size := e.HttpResponseBodySize(); !size.IsNone() {
			return fmt.Errorf("HTTP response body size error: limit=%d", size.Some())
		}
		return fmt.Errorf("HTTP response body size error")

	case httpTypes.ErrorCodeHttpResponseTrailerSectionSize:
		if size := e.HttpResponseTrailerSectionSize(); !size.IsNone() {
			return fmt.Errorf("HTTP response trailer section size error: limit=%d", size.Some())
		}
		return fmt.Errorf("HTTP response trailer section size error")

	case httpTypes.ErrorCodeHttpResponseTrailerSize:
		p := e.HttpResponseTrailerSize()
		fieldName, fieldSize := "", uint32(0)
		if !p.FieldName.IsNone() {
			fieldName = p.FieldName.Some()
		}
		if !p.FieldSize.IsNone() {
			fieldSize = p.FieldSize.Some()
		}
		return fmt.Errorf("HTTP response trailer size error: field=%q, limit=%d", fieldName, fieldSize)

	case httpTypes.ErrorCodeHttpResponseTransferCoding:
		if coding := e.HttpResponseTransferCoding(); !coding.IsNone() {
			return fmt.Errorf("HTTP response transfer coding error: coding=%q", coding.Some())
		}
		return fmt.Errorf("HTTP response transfer coding error")

	case httpTypes.ErrorCodeHttpResponseContentCoding:
		if coding := e.HttpResponseContentCoding(); !coding.IsNone() {
			return fmt.Errorf("HTTP response content coding error: coding=%q", coding.Some())
		}
		return fmt.Errorf("HTTP response content coding error")

	case httpTypes.ErrorCodeHttpResponseTimeout:
		return ErrHttpResponseTimeout

	case httpTypes.ErrorCodeHttpUpgradeFailed:
		return ErrHttpUpgradeFailed

	case httpTypes.ErrorCodeHttpProtocolError:
		return ErrHttpProtocolError

	case httpTypes.ErrorCodeLoopDetected:
		return ErrLoopDetected

	case httpTypes.ErrorCodeConfigurationError:
		return ErrConfigurationError

	case httpTypes.ErrorCodeInternalError:
		if msg := e.InternalError(); !msg.IsNone() {
			return fmt.Errorf("internal error: %s", msg.Some())
		}
		return fmt.Errorf("internal error")

	default:
		return fmt.Errorf("unknown HTTP error code: %d", e.Tag())
	}
}
