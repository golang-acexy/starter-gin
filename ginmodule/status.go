package ginmodule

type StatusCode int
type StatusMessage string
type BizErrorCode int
type BizErrorMessage string

const (
	StatusCodeSuccess            = 200
	StatusCodeServiceUnavailable = 10000 - iota
	StatusCodeExceededLimit
	StatusCodeTimeout
	StatusCodeException
	StatusCodeNotFound
	StatusCodeForbidden
	StatusCodeAuthorityUnavailable

	StatusCodeMethodNotAllowed = 9007 - iota
	StatusCodeMediaTypeNotAllowed
	StatusCodeUploadLimitExceeded
	StatusCodeUnauthorized
	StatusCodeBadRequestParameters
)

const (
	statusMessageSuccess              = "Request Success"
	statusMessageServiceUnavailable   = "Service Unavailable"
	statusMessageExceededLimit        = "Request Exceeded Limit"
	statusMessageTimeout              = "Service Timeout"
	statusMessageException            = "System Error"
	statusMessageNotFound             = "Request Not Found"
	statusMessageForbidden            = "Request Forbidden"
	statusMessageAuthorityUnavailable = "Authority Service Unavailable"

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
	StatusCodeAuthorityUnavailable: statusMessageAuthorityUnavailable,
	StatusCodeMethodNotAllowed:     statusMessageMethodNotAllowed,
	StatusCodeMediaTypeNotAllowed:  statusMessageMediaTypeNotAllowed,
	StatusCodeUploadLimitExceeded:  statusMessageUploadLimitExceeded,
	StatusCodeUnauthorized:         statusMessageUnauthorized,
	StatusCodeBadRequestParameters: statusMessageBadRequestParameters,
}

var (
	bizErrorCodeSuccess    = BizErrorCode(200)
	bizErrorMessageSuccess = BizErrorMessage("Success")
)

func GetStatusMessage(statusCode StatusCode) StatusMessage {
	return statusCodeWithMessage[statusCode]
}
