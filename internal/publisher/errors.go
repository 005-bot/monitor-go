package publisher

import "errors"

var ErrNoSubscribers = errors.New("no subscribers for redis channel")
