package ginstarter

import "net/http"

// StatusCode is a REST envelope status, which may differ from the transport HTTP status.
type StatusCode int

// StatusMessage is a client-visible system-result description.
type StatusMessage string

// BizErrorCode identifies an application-level failure within a successful system envelope.
type BizErrorCode int

// BizErrorMessage is a client-visible application-level error description.
type BizErrorMessage string

const (
	ginCtxKeyRequestState = "_internal_request_state"
	ginCtxKeyRequestBody  = "_internal_request_body"
)
const (
	// StatusCodeSuccess represents HTTP StatusOK in the REST envelope.
	StatusCodeSuccess = http.StatusOK
	// StatusCodeServiceUnavailable represents HTTP StatusServiceUnavailable in the REST envelope.
	StatusCodeServiceUnavailable = http.StatusServiceUnavailable
	// StatusCodeExceededLimit represents HTTP StatusTooManyRequests in the REST envelope.
	StatusCodeExceededLimit = http.StatusTooManyRequests
	// StatusCodeTimeout represents HTTP StatusGatewayTimeout in the REST envelope.
	StatusCodeTimeout = http.StatusGatewayTimeout
	// StatusCodeException represents HTTP StatusInternalServerError in the REST envelope.
	StatusCodeException = http.StatusInternalServerError
	// StatusCodeNotFound represents HTTP StatusNotFound in the REST envelope.
	StatusCodeNotFound = http.StatusNotFound
	// StatusCodeForbidden represents HTTP StatusForbidden in the REST envelope.
	StatusCodeForbidden = http.StatusForbidden

	// StatusCodeMethodNotAllowed represents HTTP StatusMethodNotAllowed in the REST envelope.
	StatusCodeMethodNotAllowed = http.StatusMethodNotAllowed
	// StatusCodeMediaTypeNotAllowed represents HTTP StatusUnsupportedMediaType in the REST envelope.
	StatusCodeMediaTypeNotAllowed = http.StatusUnsupportedMediaType
	// StatusCodeUploadLimitExceeded represents HTTP StatusRequestEntityTooLarge in the REST envelope.
	StatusCodeUploadLimitExceeded = http.StatusRequestEntityTooLarge
	// StatusCodeUnauthorized represents HTTP StatusUnauthorized in the REST envelope.
	StatusCodeUnauthorized = http.StatusUnauthorized
	// StatusCodeBadRequestParameters represents HTTP StatusBadRequest in the REST envelope.
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

// GetStatusMessage returns the mapped error description, or an empty value for unmapped codes (including 200).
func GetStatusMessage(statusCode StatusCode) StatusMessage {
	return statusCodeWithMessage[statusCode]
}
