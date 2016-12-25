package queue

import (
	"sync"
	"time"

	"github.com/satori/go.uuid"
)

// Local implements queue.Interface as an in-memory queue.
//
// No guarantees are made about performance or efficiency.
type Local struct {
	queue         []string
	queueMutex    *sync.Mutex
	messages      map[string][]byte
	messagesMutex *sync.RWMutex
	received      map[string]*time.Timer
	receivedMutex *sync.RWMutex
}

// NewLocal initializes and returns a Local in-memory queue.
func NewLocal() *Local {
	return &Local{
		queue:         []string{},
		queueMutex:    &sync.Mutex{},
		messages:      map[string][]byte{},
		messagesMutex: &sync.RWMutex{},
		received:      map[string]*time.Timer{},
		receivedMutex: &sync.RWMutex{},
	}
}

// SendMessage adds a message to the end of the queue.
func (l *Local) SendMessage(message []byte) error {
	var id string
	l.messagesMutex.Lock()
	// make sure id is truly unique; usually this loop should always finish
	// after the first iteration, but better to be safe than sorry
	for {
		id = uuid.NewV4().String()
		if _, ok := l.messages[id]; !ok {
			break
		}
	}
	l.messages[id] = message
	l.messagesMutex.Unlock()
	l.queueMutex.Lock()
	l.queue = append(l.queue, id)
	l.queueMutex.Unlock()
	return nil
}

// ReceiveMessage receives a message from the head of the queue.
//
// A timeout of 0 (or less) will cause the message to be removed from the queue
// immediately.
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
	l.messagesMutex.RLock()
	// skip/remove any messages that ended up back in the queue after a timeout
	// but were subsequently deleted
	for {
		id, l.queue = l.queue[0], l.queue[1:]
		var ok bool
		if message, ok = l.messages[id]; ok {
			break
		}
	}
	l.messagesMutex.RUnlock()
	l.queueMutex.Unlock()

	if timeout < 1 {
		// delete the message immediately; no need to set timer
		l.messagesMutex.Lock()
		delete(l.messages, id)
		l.messagesMutex.Unlock()
		return
	}

	l.receivedMutex.Lock()
	l.received[id] = time.AfterFunc(timeout, func() {
		l.queueMutex.Lock()
		l.queue = append([]string{id}, l.queue...)
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
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		delete(l.received, id)
	}
	l.receivedMutex.Unlock()

	l.messagesMutex.Lock()
	delete(l.messages, id)
	l.messagesMutex.Unlock()
	return nil
}
