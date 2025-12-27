package templates

import "errors"

var (
	// Template errors
	errTemplateReadFailed   = errors.New("failed to read template content")
	errTemplateParseFailed  = errors.New("failed to parse template syntax")
	errTemplateRenderFailed = errors.New("failed to render template")

	// Data parsing errors
	errDataParseFailed = errors.New("failed to parse message data")

	// Registry errors
	errNoParserRegistered = errors.New("no parser registered for extension")
)
