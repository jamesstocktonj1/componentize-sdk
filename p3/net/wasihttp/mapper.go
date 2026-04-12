package wasihttp

import (
	"fmt"
	"net/url"

	httpTypes "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_http_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
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
		return mapErrorCodeDnsError(e.DnsError())
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
		return mapErrorCodeTlsAlertReceived(e.TlsAlertReceived())
	case httpTypes.ErrorCodeHttpRequestDenied:
		return ErrHttpRequestDenied
	case httpTypes.ErrorCodeHttpRequestLengthRequired:
		return ErrHttpRequestLengthRequired
	case httpTypes.ErrorCodeHttpRequestBodySize:
		return mapErrorCodeHttpRequestBodySize(e.HttpRequestBodySize())
	case httpTypes.ErrorCodeHttpRequestMethodInvalid:
		return ErrHttpRequestMethodInvalid
	case httpTypes.ErrorCodeHttpRequestUriInvalid:
		return ErrHttpRequestUriInvalid
	case httpTypes.ErrorCodeHttpRequestUriTooLong:
		return ErrHttpRequestUriTooLong
	case httpTypes.ErrorCodeHttpRequestHeaderSectionSize:
		return mapErrorCodeHttpRequestHeaderSectionSize(e.HttpRequestHeaderSectionSize())
	case httpTypes.ErrorCodeHttpRequestHeaderSize:
		return mapErrorCodeHttpRequestHeaderSize(e.HttpRequestHeaderSize())
	case httpTypes.ErrorCodeHttpRequestTrailerSectionSize:
		return mapErrorCodeHttpRequestTrailerSectionSize(e.HttpRequestTrailerSectionSize())
	case httpTypes.ErrorCodeHttpRequestTrailerSize:
		return mapErrorCodeHttpRequestTrailerSize(e.HttpRequestTrailerSize())
	case httpTypes.ErrorCodeHttpResponseIncomplete:
		return ErrHttpResponseIncomplete
	case httpTypes.ErrorCodeHttpResponseHeaderSectionSize:
		return mapErrorCodeHttpResponseHeaderSectionSize(e.HttpResponseHeaderSectionSize())
	case httpTypes.ErrorCodeHttpResponseHeaderSize:
		return mapErrorCodeHttpResponseHeaderSize(e.HttpResponseHeaderSize())
	case httpTypes.ErrorCodeHttpResponseBodySize:
		return mapErrorCodeHttpResponseBodySize(e.HttpResponseBodySize())
	case httpTypes.ErrorCodeHttpResponseTrailerSectionSize:
		return mapErrorCodeHttpResponseTrailerSectionSize(e.HttpResponseTrailerSectionSize())
	case httpTypes.ErrorCodeHttpResponseTrailerSize:
		return mapErrorCodeHttpResponseTrailerSize(e.HttpResponseTrailerSize())
	case httpTypes.ErrorCodeHttpResponseTransferCoding:
		return mapErrorCodeHttpResponseTransferCoding(e.HttpResponseTransferCoding())
	case httpTypes.ErrorCodeHttpResponseContentCoding:
		return mapErrorCodeHttpResponseContentCoding(e.HttpResponseContentCoding())
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
		return mapErrorCodeInternalError(e.InternalError())
	default:
		return fmt.Errorf("unknown HTTP error code: %d", e.Tag())
	}
}

func mapErrorCodeDnsError(p httpTypes.DnsErrorPayload) error {
	rcode, infoCode := "", uint16(0)
	if !p.Rcode.IsNone() {
		rcode = p.Rcode.Some()
	}
	if !p.InfoCode.IsNone() {
		infoCode = p.InfoCode.Some()
	}
	return fmt.Errorf("DNS error: rcode=%q, infoCode=%d", rcode, infoCode)
}

func mapErrorCodeTlsAlertReceived(p httpTypes.TlsAlertReceivedPayload) error {
	alertId, alertMsg := uint8(0), ""
	if !p.AlertId.IsNone() {
		alertId = p.AlertId.Some()
	}
	if !p.AlertMessage.IsNone() {
		alertMsg = p.AlertMessage.Some()
	}
	return fmt.Errorf("TLS alert received: alertId=%d, alertMessage=%q", alertId, alertMsg)
}

func mapErrorCodeHttpRequestBodySize(size witTypes.Option[uint64]) error {
	if !size.IsNone() {
		return fmt.Errorf("HTTP request body size error: limit=%d", size.Some())
	}
	return fmt.Errorf("HTTP request body size error")
}

func mapErrorCodeHttpRequestHeaderSectionSize(size witTypes.Option[uint32]) error {
	if !size.IsNone() {
		return fmt.Errorf("HTTP request header section size error: limit=%d", size.Some())
	}
	return fmt.Errorf("HTTP request header section size error")
}

func mapErrorCodeHttpRequestHeaderSize(opt witTypes.Option[httpTypes.FieldSizePayload]) error {
	if !opt.IsNone() {
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
}

func mapErrorCodeHttpRequestTrailerSectionSize(size witTypes.Option[uint32]) error {
	if !size.IsNone() {
		return fmt.Errorf("HTTP request trailer section size error: limit=%d", size.Some())
	}
	return fmt.Errorf("HTTP request trailer section size error")
}

func mapErrorCodeHttpRequestTrailerSize(p httpTypes.FieldSizePayload) error {
	fieldName, fieldSize := "", uint32(0)
	if !p.FieldName.IsNone() {
		fieldName = p.FieldName.Some()
	}
	if !p.FieldSize.IsNone() {
		fieldSize = p.FieldSize.Some()
	}
	return fmt.Errorf("HTTP request trailer size error: field=%q, limit=%d", fieldName, fieldSize)
}

func mapErrorCodeHttpResponseHeaderSectionSize(size witTypes.Option[uint32]) error {
	if !size.IsNone() {
		return fmt.Errorf("HTTP response header section size error: limit=%d", size.Some())
	}
	return fmt.Errorf("HTTP response header section size error")
}

func mapErrorCodeHttpResponseHeaderSize(p httpTypes.FieldSizePayload) error {
	fieldName, fieldSize := "", uint32(0)
	if !p.FieldName.IsNone() {
		fieldName = p.FieldName.Some()
	}
	if !p.FieldSize.IsNone() {
		fieldSize = p.FieldSize.Some()
	}
	return fmt.Errorf("HTTP response header size error: field=%q, limit=%d", fieldName, fieldSize)
}

func mapErrorCodeHttpResponseBodySize(size witTypes.Option[uint64]) error {
	if !size.IsNone() {
		return fmt.Errorf("HTTP response body size error: limit=%d", size.Some())
	}
	return fmt.Errorf("HTTP response body size error")
}

func mapErrorCodeHttpResponseTrailerSectionSize(size witTypes.Option[uint32]) error {
	if !size.IsNone() {
		return fmt.Errorf("HTTP response trailer section size error: limit=%d", size.Some())
	}
	return fmt.Errorf("HTTP response trailer section size error")
}

func mapErrorCodeHttpResponseTrailerSize(p httpTypes.FieldSizePayload) error {
	fieldName, fieldSize := "", uint32(0)
	if !p.FieldName.IsNone() {
		fieldName = p.FieldName.Some()
	}
	if !p.FieldSize.IsNone() {
		fieldSize = p.FieldSize.Some()
	}
	return fmt.Errorf("HTTP response trailer size error: field=%q, limit=%d", fieldName, fieldSize)
}

func mapErrorCodeHttpResponseTransferCoding(coding witTypes.Option[string]) error {
	if !coding.IsNone() {
		return fmt.Errorf("HTTP response transfer coding error: coding=%q", coding.Some())
	}
	return fmt.Errorf("HTTP response transfer coding error")
}

func mapErrorCodeHttpResponseContentCoding(coding witTypes.Option[string]) error {
	if !coding.IsNone() {
		return fmt.Errorf("HTTP response content coding error: coding=%q", coding.Some())
	}
	return fmt.Errorf("HTTP response content coding error")
}

func mapErrorCodeInternalError(msg witTypes.Option[string]) error {
	if !msg.IsNone() {
		return fmt.Errorf("internal error: %s", msg.Some())
	}
	return fmt.Errorf("internal error")
}
