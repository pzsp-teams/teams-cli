package messaging

import "errors"

// ErrMessageSkipped marks a message as skipped in case ignoreError is not set when sending messages
var ErrMessageSkipped = errors.New("skipped sending message, earlier message failed")
