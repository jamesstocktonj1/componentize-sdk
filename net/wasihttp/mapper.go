package wasihttp

import (
	"fmt"
	"net/http"

	types "github.com/jamesstocktonj1/componentize-sdk/gen/wasi_http_types"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

func mapMethod(m string) types.Method {
	switch m {
	case http.MethodGet:
		return types.MakeMethodGet()
	case http.MethodHead:
		return types.MakeMethodHead()
	case http.MethodPost:
		return types.MakeMethodPost()
	case http.MethodPut:
		return types.MakeMethodPut()
	case http.MethodDelete:
		return types.MakeMethodDelete()
	case http.MethodConnect:
		return types.MakeMethodConnect()
	case http.MethodOptions:
		return types.MakeMethodOptions()
	case http.MethodTrace:
		return types.MakeMethodTrace()
	case http.MethodPatch:
		return types.MakeMethodPatch()
	default:
		return types.MakeMethodOther(m)
	}
}

func mapRequestOptions() witTypes.Option[*types.RequestOptions] {
	return witTypes.None[*types.RequestOptions]()
}

func mapErrorCode(e types.ErrorCode) error {
	switch e.Tag() {
	case types.ErrorCodeDnsTimeout:
		return ErrDnsTimeout

	case types.ErrorCodeDnsError:
		p := e.DnsError()
		rcode, infoCode := "", uint16(0)
		if !p.Rcode.IsNone() {
			rcode = p.Rcode.Some()
		}
		if !p.InfoCode.IsNone() {
			infoCode = p.InfoCode.Some()
		}
		return fmt.Errorf("DNS error: rcode=%q, infoCode=%d", rcode, infoCode)

	case types.ErrorCodeDestinationNotFound:
		return ErrDestinationNotFound

	case types.ErrorCodeDestinationUnavailable:
		return ErrDestinationUnavailable

	case types.ErrorCodeDestinationIpProhibited:
		return ErrDestinationIpProhibited

	case types.ErrorCodeDestinationIpUnroutable:
		return ErrDestinationIpUnroutable

	case types.ErrorCodeConnectionRefused:
		return ErrConnectionRefused

	case types.ErrorCodeConnectionTerminated:
		return ErrConnectionTerminated

	case types.ErrorCodeConnectionTimeout:
		return ErrConnectionTimeout

	case types.ErrorCodeConnectionReadTimeout:
		return ErrConnectionReadTimeout

	case types.ErrorCodeConnectionWriteTimeout:
		return ErrConnectionWriteTimeout

	case types.ErrorCodeConnectionLimitReached:
		return ErrConnectionLimitReached

	case types.ErrorCodeTlsProtocolError:
		return ErrTlsProtocolError

	case types.ErrorCodeTlsCertificateError:
		return ErrTlsCertificateError

	case types.ErrorCodeTlsAlertReceived:
		p := e.TlsAlertReceived()
		alertId, alertMsg := uint8(0), ""
		if !p.AlertId.IsNone() {
			alertId = p.AlertId.Some()
		}
		if !p.AlertMessage.IsNone() {
			alertMsg = p.AlertMessage.Some()
		}
		return fmt.Errorf("TLS alert received: alertId=%d, alertMessage=%q", alertId, alertMsg)

	case types.ErrorCodeHttpRequestDenied:
		return ErrHttpRequestDenied

	case types.ErrorCodeHttpRequestLengthRequired:
		return ErrHttpRequestLengthRequired

	case types.ErrorCodeHttpRequestBodySize:
		if size := e.HttpRequestBodySize(); !size.IsNone() {
			return fmt.Errorf("HTTP request body size error: limit=%d", size.Some())
		}
		return fmt.Errorf("HTTP request body size error")

	case types.ErrorCodeHttpRequestMethodInvalid:
		return ErrHttpRequestMethodInvalid

	case types.ErrorCodeHttpRequestUriInvalid:
		return ErrHttpRequestUriInvalid

	case types.ErrorCodeHttpRequestUriTooLong:
		return ErrHttpRequestUriTooLong

	case types.ErrorCodeHttpRequestHeaderSectionSize:
		if size := e.HttpRequestHeaderSectionSize(); !size.IsNone() {
			return fmt.Errorf("HTTP request header section size error: limit=%d", size.Some())
		}
		return fmt.Errorf("HTTP request header section size error")

	case types.ErrorCodeHttpRequestHeaderSize:
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

	case types.ErrorCodeHttpRequestTrailerSectionSize:
		if size := e.HttpRequestTrailerSectionSize(); !size.IsNone() {
			return fmt.Errorf("HTTP request trailer section size error: limit=%d", size.Some())
		}
		return fmt.Errorf("HTTP request trailer section size error")

	case types.ErrorCodeHttpRequestTrailerSize:
		p := e.HttpRequestTrailerSize()
		fieldName, fieldSize := "", uint32(0)
		if !p.FieldName.IsNone() {
			fieldName = p.FieldName.Some()
		}
		if !p.FieldSize.IsNone() {
			fieldSize = p.FieldSize.Some()
		}
		return fmt.Errorf("HTTP request trailer size error: field=%q, limit=%d", fieldName, fieldSize)

	case types.ErrorCodeHttpResponseIncomplete:
		return ErrHttpResponseIncomplete

	case types.ErrorCodeHttpResponseHeaderSectionSize:
		if size := e.HttpResponseHeaderSectionSize(); !size.IsNone() {
			return fmt.Errorf("HTTP response header section size error: limit=%d", size.Some())
		}
		return fmt.Errorf("HTTP response header section size error")

	case types.ErrorCodeHttpResponseHeaderSize:
		p := e.HttpResponseHeaderSize()
		fieldName, fieldSize := "", uint32(0)
		if !p.FieldName.IsNone() {
			fieldName = p.FieldName.Some()
		}
		if !p.FieldSize.IsNone() {
			fieldSize = p.FieldSize.Some()
		}
		return fmt.Errorf("HTTP response header size error: field=%q, limit=%d", fieldName, fieldSize)

	case types.ErrorCodeHttpResponseBodySize:
		if size := e.HttpResponseBodySize(); !size.IsNone() {
			return fmt.Errorf("HTTP response body size error: limit=%d", size.Some())
		}
		return fmt.Errorf("HTTP response body size error")

	case types.ErrorCodeHttpResponseTrailerSectionSize:
		if size := e.HttpResponseTrailerSectionSize(); !size.IsNone() {
			return fmt.Errorf("HTTP response trailer section size error: limit=%d", size.Some())
		}
		return fmt.Errorf("HTTP response trailer section size error")

	case types.ErrorCodeHttpResponseTrailerSize:
		p := e.HttpResponseTrailerSize()
		fieldName, fieldSize := "", uint32(0)
		if !p.FieldName.IsNone() {
			fieldName = p.FieldName.Some()
		}
		if !p.FieldSize.IsNone() {
			fieldSize = p.FieldSize.Some()
		}
		return fmt.Errorf("HTTP response trailer size error: field=%q, limit=%d", fieldName, fieldSize)

	case types.ErrorCodeHttpResponseTransferCoding:
		if coding := e.HttpResponseTransferCoding(); !coding.IsNone() {
			return fmt.Errorf("HTTP response transfer coding error: coding=%q", coding.Some())
		}
		return fmt.Errorf("HTTP response transfer coding error")

	case types.ErrorCodeHttpResponseContentCoding:
		if coding := e.HttpResponseContentCoding(); !coding.IsNone() {
			return fmt.Errorf("HTTP response content coding error: coding=%q", coding.Some())
		}
		return fmt.Errorf("HTTP response content coding error")

	case types.ErrorCodeHttpResponseTimeout:
		return ErrHttpResponseTimeout

	case types.ErrorCodeHttpUpgradeFailed:
		return ErrHttpUpgradeFailed

	case types.ErrorCodeHttpProtocolError:
		return ErrHttpProtocolError

	case types.ErrorCodeLoopDetected:
		return ErrLoopDetected

	case types.ErrorCodeConfigurationError:
		return ErrConfigurationError

	case types.ErrorCodeInternalError:
		if msg := e.InternalError(); !msg.IsNone() {
			return fmt.Errorf("internal error: %s", msg.Some())
		}
		return fmt.Errorf("internal error")

	default:
		return fmt.Errorf("unknown HTTP error code: %d", e.Tag())
	}
}
