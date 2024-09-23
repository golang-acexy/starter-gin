package ginstarter

import (
	"net/http"
	"time"
)

type StatusCode int
type StatusMessage string
type BizErrorCode int
type BizErrorMessage string

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

// RestRespStatusStruct 框架默认的Rest请求状态结构
type RestRespStatusStruct struct {

	// 标识请求系统状态 200 标识网络请求层面的成功 见StatusCode
	StatusCode    StatusCode    `json:"statusCode"`
	StatusMessage StatusMessage `json:"statusMessage"`

	// 业务错误码 仅当StatusCode为200时进入业务错误判断
	BizErrorCode    *BizErrorCode    `json:"bizErrorCode"`
	BizErrorMessage *BizErrorMessage `json:"bizErrorMessage"`

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

// NewRestSuccess 响应标准成功Rest结构体
func NewRestSuccess(data ...interface{}) RestRespStruct {
	result := RestRespStruct{
		Status: &RestRespStatusStruct{
			StatusCode:    StatusCodeSuccess,
			StatusMessage: statusMessageSuccess,
			Timestamp:     time.Now().UnixMilli(),
		},
	}
	if len(data) > 0 {
		result.Data = data[0]
	}
	return result
}

// NewRestException 响应标准异常Rest结构体
func NewRestException(statusMessage ...string) RestRespStruct {
	status := &RestRespStatusStruct{
		StatusCode:    StatusCodeException,
		StatusMessage: statusMessageException,
		Timestamp:     time.Now().UnixMilli(),
	}
	if len(statusMessage) > 0 {
		status.StatusMessage = StatusMessage(statusMessage[0])
	}
	return RestRespStruct{
		Status: status,
	}
}

// NewRestBadBadParameters 响应标准参数错误Rest结构体
func NewRestBadBadParameters(statusMessage ...string) RestRespStruct {
	status := &RestRespStatusStruct{
		StatusCode:    StatusCodeBadRequestParameters,
		StatusMessage: statusMessageBadRequestParameters,
		Timestamp:     time.Now().UnixMilli(),
	}
	if len(statusMessage) > 0 {
		status.StatusMessage = StatusMessage(statusMessage[0])
	}
	return RestRespStruct{
		Status: status,
	}
}

// NewRestUnauthorized 响应标准未授权Rest结构体
func NewRestUnauthorized(statusMessage ...string) RestRespStruct {
	status := &RestRespStatusStruct{
		StatusCode:    StatusCodeUnauthorized,
		StatusMessage: statusMessageUnauthorized,
		Timestamp:     time.Now().UnixMilli(),
	}
	if len(statusMessage) > 0 {
		status.StatusMessage = StatusMessage(statusMessage[0])
	}
	return RestRespStruct{
		Status: status,
	}
}

// NewRestStatusError 响应标准错误Rest结构体
func NewRestStatusError(statusCode StatusCode, statusMessage ...StatusMessage) RestRespStruct {
	dataRest := RestRespStruct{
		Status: &RestRespStatusStruct{
			StatusCode: statusCode,
			Timestamp:  time.Now().UnixMilli(),
		},
	}
	if len(statusMessage) > 0 && statusMessage[0] != "" {
		dataRest.Status.StatusMessage = statusMessage[0]
	} else {
		dataRest.Status.StatusMessage = GetStatusMessage(statusCode)
	}
	return dataRest
}

// NewRestBizError 响应标准业务错误Rest结构体
func NewRestBizError(bizErrorCode BizErrorCode, bizErrorMessage BizErrorMessage) RestRespStruct {
	dataRest := RestRespStruct{
		Status: &RestRespStatusStruct{
			StatusCode:      StatusCodeSuccess,
			StatusMessage:   statusMessageSuccess,
			BizErrorCode:    &bizErrorCode,
			BizErrorMessage: &bizErrorMessage,
			Timestamp:       time.Now().UnixMilli(),
		},
	}
	return dataRest
}
