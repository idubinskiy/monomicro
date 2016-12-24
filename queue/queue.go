package queue

import (
	"errors"
	"time"
)

type Interface interface {
	SendMessage(message []byte) error
	ReceiveMessage(timeout time.Duration) (id string, message []byte, err error)
	DeleteMessage(id string) error
}

var ErrNoMessages = errors.New("No messages in queue")
