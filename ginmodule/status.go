package ginmodule

type StatusCode int
type StatusMessage string
type BizErrorCode int
type BizErrorMessage string

const (
	// StatusCodeSuccess 请求成功
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

func GetStatusMessage(statusCode StatusCode) StatusMessage {
	return statusCodeWithMessage[statusCode]
}

// RestRespStatusStruct 框架默认的Rest请求状态结构
type RestRespStatusStruct struct {

	// 标识请求系统状态 200 标识网络请求层面的成功 见StatusCode
	StatusCode    StatusCode    `json:"statusCode"`
	StatusMessage StatusMessage `json:"statusMessage"`

	// 业务错误码 仅当StatusCode为200时进入业务错误判断
	BizErrorCode    BizErrorCode    `json:"bizErrorCode"`
	BizErrorMessage BizErrorMessage `json:"bizErrorMessage"`

	// 系统响应时间戳
	Timestamp int64 `json:"timestamp"`
}

// RestRespStruct 框架默认的Rest请求结构
type RestRespStruct struct {

	// 请求状态描述
	Status *RestRespStatusStruct `json:"status"`

	// 仅当StatusCode为200 无业务错误码BizErrorCode 响应成功数据
	Data any `json:"data"`
}
