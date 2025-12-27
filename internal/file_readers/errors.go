package file_readers

import "errors"

var (
	errJSONDecodeFailed = errors.New("failed to decode JSON data")
	errYAMLDecodeFailed = errors.New("failed to decode YAML data")
	errTOMLDecodeFailed = errors.New("failed to decode TOML data")
)