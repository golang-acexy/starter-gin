package ginmodule

type StatusCode int
type StatusMessage string
type BizErrorCode int
type BizErrorMessage string

const (
	statusCodeSuccess            = 200
	statusCodeServiceUnavailable = 10000 - iota
	statusCodeExceededLimit
	statusCodeTimeout
	statusCodeException
	statusCodeNotFound
	statusCodeForbidden
	statusCodeAuthorityUnavailable

	statusCodeMethodNotAllowed = 9007 - iota
	statusCodeMediaTypeNotAllowed
	statusCodeUploadLimitExceeded
	statusCodeUnauthorized
	statusCodeBadRequestParameters
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
	statusCodeServiceUnavailable:   statusMessageServiceUnavailable,
	statusCodeExceededLimit:        statusMessageExceededLimit,
	statusCodeTimeout:              statusMessageTimeout,
	statusCodeException:            statusMessageException,
	statusCodeNotFound:             statusMessageNotFound,
	statusCodeForbidden:            statusMessageForbidden,
	statusCodeAuthorityUnavailable: statusMessageAuthorityUnavailable,
	statusCodeMethodNotAllowed:     statusMessageMethodNotAllowed,
	statusCodeMediaTypeNotAllowed:  statusMessageMediaTypeNotAllowed,
	statusCodeUploadLimitExceeded:  statusMessageUploadLimitExceeded,
	statusCodeUnauthorized:         statusMessageUnauthorized,
	statusCodeBadRequestParameters: statusMessageBadRequestParameters,
}

func GetStatusMessage(statusCode StatusCode) StatusMessage {
	return statusCodeWithMessage[statusCode]
}
