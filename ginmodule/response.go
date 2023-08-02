package ginmodule

type Status struct {
	StatusCode      StatusCode      `json:"statusCode"`
	StatusMessage   StatusMessage   `json:"statusMessage"`
	BizErrorCode    BizErrorCode    `json:"bizErrorCode"`
	BizErrorMessage BizErrorMessage `json:"bizErrorMessage"`
}

type Response struct {
	Status *Status `json:"status"`
	Data   any     `json:"data"`
}

func NewSuccess(data ...any) *Response {
	response := Response{
		Status: &Status{
			StatusCode:      statusCodeSuccess,
			StatusMessage:   statusMessageSuccess,
			BizErrorCode:    bizErrorCodeSuccess,
			BizErrorMessage: bizErrorMessageSuccess,
		},
	}
	if len(data) > 0 {
		response.Data = data[0]
	}
	return &response
}

// NewException 系统异常响应
func NewException() *Response {
	return &Response{
		Status: &Status{
			StatusCode:    statusCodeException,
			StatusMessage: statusMessageException,
		},
	}
}

// NewError 其他StatusCode错误
func NewError(statusCode StatusCode) *Response {
	return &Response{
		Status: &Status{
			StatusCode:    statusCode,
			StatusMessage: GetStatusMessage(statusCode),
		},
	}
}

func NewBizError(bizErrorCode BizErrorCode, bizErrorMessage BizErrorMessage) *Response {
	return &Response{
		Status: &Status{
			StatusCode:      statusCodeSuccess,
			StatusMessage:   GetStatusMessage(statusCodeSuccess),
			BizErrorCode:    bizErrorCode,
			BizErrorMessage: bizErrorMessage,
		},
	}
}
