package queue

import (
	"sync"
	"time"

	"github.com/satori/go.uuid"
)

type Local struct {
	queue         [][]byte
	queueMutex    *sync.Mutex
	received      map[string]*time.Timer
	receivedMutex *sync.RWMutex
}

func NewLocal() *Local {
	return &Local{
		queue:         [][]byte{},
		queueMutex:    &sync.Mutex{},
		received:      map[string]*time.Timer{},
		receivedMutex: &sync.RWMutex{},
	}
}

func (l *Local) SendMessage(message []byte) error {
	l.queueMutex.Lock()
	l.queue = append(l.queue, message)
	l.queueMutex.Unlock()
	return nil
}

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
