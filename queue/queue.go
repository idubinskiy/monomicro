package queue

import (
	"errors"
	"time"
)

// Interface represents the minimum implementation of an asynchronous queue.
//
// All methods should be safe for concurrent access.
type Interface interface {
	SendMessage(message []byte) error
	ReceiveMessage(timeout time.Duration) (id string, message []byte, err error)
	DeleteMessage(id string) error
}

// ErrNoMessages should be returned by the ReceiveMessage method of types implementing
// queue.Interface when there are no messages in the queue waiting to be received.
var ErrNoMessages = errors.New("No messages in queue")
