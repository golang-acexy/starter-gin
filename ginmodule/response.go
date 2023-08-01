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
			StatusCode:    statusCodeSuccess,
			StatusMessage: statusMessageSuccess,
		},
	}
	if len(data) > 0 {
		response.Data = data[0]
	}
	return &response
}

func NewException() *Response {
	return &Response{
		Status: &Status{
			StatusCode:    statusCodeException,
			StatusMessage: statusMessageException,
		},
	}
}

func NewError(statusCode StatusCode) *Response {
	return &Response{
		Status: &Status{
			StatusCode:    statusCode,
			StatusMessage: GetStatusMessage(statusCode),
		},
	}
}
