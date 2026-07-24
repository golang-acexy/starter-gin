package ginstarter

import "errors"

var (
	ErrParamNotSet               = errors.New("param not set")
	ErrUnsupportedContent        = errors.New("unsupported content-type")
	ErrMediaTypeNotAllowed       = errors.New(statusMessageMediaTypeNotAllowed)
	ErrBadJSONPayload            = errors.New("bad json payload")
	ErrJSONTypeMismatch          = errors.New("json type mismatch")
	ErrGinStarterAlreadyStarted = errors.New("gin starter already started")
	ErrGinServerNotStarted      = errors.New("gin server not started")
	ErrRouterInfoNil            = errors.New("router info is nil")
	ErrResponseEntityBuilderNil = errors.New("response entity builder is nil")
	ErrResponseEntityNil        = errors.New("response entity is nil")
	ErrSaveUploadedFile         = errors.New("save uploaded file failed")
)
