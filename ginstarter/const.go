package ginstarter

import "net/http"

type StatusCode int
type StatusMessage string
type BizErrorCode int
type BizErrorMessage string

const (
	GinCtxKeyResponse = "_internal_response"
)
const (
	StatusCodeSuccess            = http.StatusOK
	StatusCodeServiceUnavailable = http.StatusServiceUnavailable
	StatusCodeExceededLimit      = http.StatusTooManyRequests
	StatusCodeTimeout            = http.StatusGatewayTimeout
	StatusCodeException          = http.StatusInternalServerError
	StatusCodeNotFound           = http.StatusNotFound
	StatusCodeForbidden          = http.StatusForbidden

	StatusCodeMethodNotAllowed     = http.StatusMethodNotAllowed
	StatusCodeMediaTypeNotAllowed  = http.StatusUnsupportedMediaType
	StatusCodeUploadLimitExceeded  = http.StatusRequestEntityTooLarge
	StatusCodeUnauthorized         = http.StatusUnauthorized
	StatusCodeBadRequestParameters = http.StatusBadRequest
)

const (
	statusMessageSuccess            = "Request Success"
	statusMessageServiceUnavailable = "Service Unavailable"
	statusMessageExceededLimit      = "Request Exceeded Limit"
	statusMessageTimeout            = "Service Timeout"
	statusMessageException          = "System Error"
	statusMessageNotFound           = "Request Not Found"
	statusMessageForbidden          = "Request Forbidden"

	statusMessageMethodNotAllowed     = "Request Method Not Allowed"
	statusMessageMediaTypeNotAllowed  = "Request Media Type Not Allowed"
	statusMessageUploadLimitExceeded  = "Upload File Size Limit Exceeded"
	statusMessageUnauthorized         = "Unauthorized Request"
	statusMessageBadRequestParameters = "Bad Request Parameters"
)

var statusCodeWithMessage = map[StatusCode]StatusMessage{
	StatusCodeServiceUnavailable:   statusMessageServiceUnavailable,
	StatusCodeExceededLimit:        statusMessageExceededLimit,
	StatusCodeTimeout:              statusMessageTimeout,
	StatusCodeException:            statusMessageException,
	StatusCodeNotFound:             statusMessageNotFound,
	StatusCodeForbidden:            statusMessageForbidden,
	StatusCodeMethodNotAllowed:     statusMessageMethodNotAllowed,
	StatusCodeMediaTypeNotAllowed:  statusMessageMediaTypeNotAllowed,
	StatusCodeUploadLimitExceeded:  statusMessageUploadLimitExceeded,
	StatusCodeUnauthorized:         statusMessageUnauthorized,
	StatusCodeBadRequestParameters: statusMessageBadRequestParameters,
}

func GetStatusMessage(statusCode StatusCode) StatusMessage {
	return statusCodeWithMessage[statusCode]
}
