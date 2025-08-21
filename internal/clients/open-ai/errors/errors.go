package errors

import "errors"

var ErrNoChoicesFound = errors.New("openai api response contains no choices")
