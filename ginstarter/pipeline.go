package ginstarter

import "github.com/gin-gonic/gin"

// requestState 保存一次请求在框架内部流转的响应。
// Handler、拦截器和异常处理器只更新该状态，最终由响应管道统一写入客户端。
type requestState struct {
	response   Response
	stopped    bool
	formParsed bool
	formErr    error
}

func getRequestState(context *gin.Context) *requestState {
	if value, exists := context.Get(ginCtxKeyRequestState); exists {
		if state, ok := value.(*requestState); ok {
			return state
		}
	}
	state := &requestState{}
	context.Set(ginCtxKeyRequestState, state)
	return state
}

func setResponse(context *gin.Context, response Response) {
	if response != nil {
		getRequestState(context).response = response
	}
}

func currentResponse(context *gin.Context) Response {
	return getRequestState(context).response
}
