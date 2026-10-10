package ginstarter

import (
	"time"
)

// RestRespStatusStruct describes a REST result independently of its transport HTTP status.
type RestRespStatusStruct struct {

	// StatusCode is the envelope status; 200 permits business-error evaluation.
	StatusCode StatusCode `json:"statusCode"`
	// StatusMessage is the client-visible description of StatusCode.
	StatusMessage StatusMessage `json:"statusMessage"`

	// BizErrorCode describes a business failure when StatusCode is 200; nil means none.
	BizErrorCode *BizErrorCode `json:"bizErrorCode"`
	// BizErrorMessage describes BizErrorCode; nil means no business message.
	BizErrorMessage *BizErrorMessage `json:"bizErrorMessage"`

	// Timestamp is the envelope creation time in Unix milliseconds.
	Timestamp int64 `json:"timestamp"`
}

// RestRespStruct is a mutable REST envelope; synchronize access when sharing it between goroutines.
type RestRespStruct struct {

	// Status describes the result; nil is treated as unsuccessful by success helpers.
	Status *RestRespStatusStruct `json:"status"`

	// Data carries success payloads by convention; constructors do not enforce payload ownership.
	Data any `json:"data"`
}

// IsSuccess reports status 200 with no business error code; a nil Status is unsuccessful.
func (r *RestRespStruct) IsSuccess() bool {
	if r.Status != nil {
		return r.Status.StatusCode == StatusCodeSuccess && r.Status.BizErrorCode == nil
	}
	return false
}

// IsSuccessWithData additionally requires a nonnil Data interface; a typed nil can satisfy this check.
func (r *RestRespStruct) IsSuccessWithData() bool {
	if r.Status != nil {
		return r.Status.StatusCode == StatusCodeSuccess && r.Status.BizErrorCode == nil && r.Data != nil
	}
	return false
}

// NewRestSuccess creates a successful envelope using only the first optional data value.
func NewRestSuccess(data ...any) *RestRespStruct {
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
	return &result
}

// NewRestException creates a system-error envelope using only the first optional message.
func NewRestException(statusMessage ...string) *RestRespStruct {
	status := &RestRespStatusStruct{
		StatusCode:    StatusCodeException,
		StatusMessage: statusMessageException,
		Timestamp:     time.Now().UnixMilli(),
	}
	if len(statusMessage) > 0 {
		status.StatusMessage = StatusMessage(statusMessage[0])
	}
	return &RestRespStruct{
		Status: status,
	}
}

// NewRestBadParameters creates a parameter-error envelope using only the first optional message.
func NewRestBadParameters(statusMessage ...string) *RestRespStruct {
	status := &RestRespStatusStruct{
		StatusCode:    StatusCodeBadRequestParameters,
		StatusMessage: statusMessageBadRequestParameters,
		Timestamp:     time.Now().UnixMilli(),
	}
	if len(statusMessage) > 0 {
		status.StatusMessage = StatusMessage(statusMessage[0])
	}
	return &RestRespStruct{
		Status: status,
	}
}

// NewRestUnauthorized creates an unauthorized envelope using only the first optional message.
func NewRestUnauthorized(statusMessage ...string) *RestRespStruct {
	status := &RestRespStatusStruct{
		StatusCode:    StatusCodeUnauthorized,
		StatusMessage: statusMessageUnauthorized,
		Timestamp:     time.Now().UnixMilli(),
	}
	if len(statusMessage) > 0 {
		status.StatusMessage = StatusMessage(statusMessage[0])
	}
	return &RestRespStruct{
		Status: status,
	}
}

// NewRestStatusError uses the first optional message when nonempty, otherwise the mapped default (possibly empty).
func NewRestStatusError(statusCode StatusCode, statusMessage ...StatusMessage) *RestRespStruct {
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
	return &dataRest
}

// NewRestBizError creates a business-error envelope with system status 200 and no success data.
func NewRestBizError(bizErrorCode BizErrorCode, bizErrorMessage BizErrorMessage) *RestRespStruct {
	dataRest := RestRespStruct{
		Status: &RestRespStatusStruct{
			StatusCode:      StatusCodeSuccess,
			StatusMessage:   statusMessageSuccess,
			BizErrorCode:    &bizErrorCode,
			BizErrorMessage: &bizErrorMessage,
			Timestamp:       time.Now().UnixMilli(),
		},
	}
	return &dataRest
}
