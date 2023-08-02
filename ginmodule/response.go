package ginmodule

type Status struct {
	StatusCode      StatusCode       `json:"statusCode"`
	StatusMessage   StatusMessage    `json:"statusMessage"`
	BizErrorCode    *BizErrorCode    `json:"bizErrorCode"`
	BizErrorMessage *BizErrorMessage `json:"bizErrorMessage"`
}

type Response struct {
	Status *Status `json:"status"`
	Data   any     `json:"data"`
}

func ResponseSuccess(data ...any) *Response {
	response := Response{
		Status: &Status{
			StatusCode:      statusCodeSuccess,
			StatusMessage:   statusMessageSuccess,
			BizErrorCode:    &bizErrorCodeSuccess,
			BizErrorMessage: &bizErrorMessageSuccess,
		},
	}
	if len(data) > 0 {
		response.Data = data[0]
	}
	return &response
}

// ResponseException 系统异常响应
func ResponseException() *Response {
	return &Response{
		Status: &Status{
			StatusCode:    statusCodeException,
			StatusMessage: statusMessageException,
		},
	}
}

// ResponseError 其他StatusCode错误
func ResponseError(statusCode StatusCode) *Response {
	return &Response{
		Status: &Status{
			StatusCode:    statusCode,
			StatusMessage: GetStatusMessage(statusCode),
		},
	}
}

func ResponseBizError(bizErrorCode *BizErrorCode, bizErrorMessage *BizErrorMessage) *Response {
	return &Response{
		Status: &Status{
			StatusCode:      statusCodeSuccess,
			StatusMessage:   GetStatusMessage(statusCodeSuccess),
			BizErrorCode:    bizErrorCode,
			BizErrorMessage: bizErrorMessage,
		},
	}
}
