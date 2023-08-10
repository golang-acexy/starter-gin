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
			StatusCode:    StatusCodeSuccess,
			StatusMessage: statusMessageSuccess,
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
			StatusCode:    StatusCodeException,
			StatusMessage: statusMessageException,
		},
	}
}

// ResponseError 其他StatusCode错误
func ResponseError(statusCode StatusCode, statusMessage ...StatusMessage) *Response {
	response := &Response{
		Status: &Status{
			StatusCode: statusCode,
		},
	}
	if len(statusMessage) > 0 {
		response.Status.StatusMessage = statusMessage[0]
	} else {
		response.Status.StatusMessage = GetStatusMessage(statusCode)
	}
	return response
}

func ResponseBizError(bizErrorCode *BizErrorCode, bizErrorMessage *BizErrorMessage) *Response {
	return &Response{
		Status: &Status{
			StatusCode:      StatusCodeSuccess,
			StatusMessage:   GetStatusMessage(StatusCodeSuccess),
			BizErrorCode:    bizErrorCode,
			BizErrorMessage: bizErrorMessage,
		},
	}
}
