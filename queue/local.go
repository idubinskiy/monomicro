package queue

import (
	"sync"
	"time"

	"github.com/satori/go.uuid"
)

// Local implements Interface as an in-memory queue.
//
// No guarantees are made about performance or efficiency.
type Local struct {
	queue         [][]byte
	queueMutex    *sync.Mutex
	received      map[string]*time.Timer
	receivedMutex *sync.RWMutex
}

// NewLocal initializes and returns a Local in-memory queue.
func NewLocal() *Local {
	return &Local{
		queue:         [][]byte{},
		queueMutex:    &sync.Mutex{},
		received:      map[string]*time.Timer{},
		receivedMutex: &sync.RWMutex{},
	}
}

// SendMessage adds a message to the end of the queue.
func (l *Local) SendMessage(message []byte) error {
	l.queueMutex.Lock()
	l.queue = append(l.queue, message)
	l.queueMutex.Unlock()
	return nil
}

// ReceiveMessage receives a message from the head of the queue.
//
// A timeout of 0 (or less) will cause the message to be removed from the queue;
// id will be an empty string in this case.
//
// A positive timeout will cause the message to appear back at the head of the
// queue once the timeout expires. The returned id should be passed to DeleteMessage
// once the message is processed to remove it from the queue.
func (l *Local) ReceiveMessage(timeout time.Duration) (id string, message []byte, err error) {
	l.queueMutex.Lock()
	if len(l.queue) == 0 {
		err = ErrNoMessages
		l.queueMutex.Unlock()
		return
	}
	message, l.queue = l.queue[0], l.queue[1:]
	l.queueMutex.Unlock()

	if timeout < 1 {
		return
	}

	l.receivedMutex.Lock()
	for {
		id = uuid.NewV4().String()
		_, ok := l.received[id]
		if !ok {
			break
		}
	}
	l.received[id] = time.AfterFunc(timeout, func() {
		l.queueMutex.Lock()
		l.queue = append([][]byte{message}, l.queue...)
		l.queueMutex.Unlock()

		l.receivedMutex.Lock()
		delete(l.received, id)
		l.receivedMutex.Unlock()
	})
	l.receivedMutex.Unlock()

	return
}

// DeleteMessage removes the message with the passed id from the queue. An invalid
// id (or one for a message that has already been deleted) results in a no-op.
func (l *Local) DeleteMessage(id string) error {
	l.receivedMutex.Lock()
	timer, ok := l.received[id]
	if ok {
		timer.Stop()
		delete(l.received, id)
	}
	l.receivedMutex.Unlock()
	return nil
}
