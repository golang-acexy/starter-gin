package ginstarter

import "errors"

var (
	// ErrParamNotSet reports param not set.
	ErrParamNotSet = errors.New("param not set")
	// ErrUnsupportedContent reports unsupported content-type.
	ErrUnsupportedContent = errors.New("unsupported content-type")
	// ErrMediaTypeNotAllowed reports a disallowed request media type.
	ErrMediaTypeNotAllowed = errors.New(statusMessageMediaTypeNotAllowed)
	// ErrBadJSONPayload reports bad json payload.
	ErrBadJSONPayload = errors.New("bad json payload")
	// ErrJSONTypeMismatch reports json type mismatch.
	ErrJSONTypeMismatch = errors.New("json type mismatch")
	// ErrGinStarterAlreadyStarted reports gin starter already started.
	ErrGinStarterAlreadyStarted = errors.New("gin starter already started")
	// ErrGinServerNotStarted reports gin server not started.
	ErrGinServerNotStarted = errors.New("gin server not started")
	// ErrRouterInfoNil reports router info is nil.
	ErrRouterInfoNil = errors.New("router info is nil")
	// ErrResponseEntityBuilderNil reports response entity builder is nil.
	ErrResponseEntityBuilderNil = errors.New("response entity builder is nil")
	// ErrResponseEntityNil reports response entity is nil.
	ErrResponseEntityNil = errors.New("response entity is nil")
	// ErrSaveUploadedFile reports save uploaded file failed.
	ErrSaveUploadedFile = errors.New("save uploaded file failed")
)
