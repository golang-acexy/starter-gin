package ginstarter

import "github.com/gin-gonic/gin"

// requestState stores the response flowing through one request.
// Handlers, interceptors, and recovery update state; the outer pipeline writes the result.
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
