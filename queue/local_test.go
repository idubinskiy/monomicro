package queue

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func getQueueContents(q *Local) [][]byte {
	qMessages := make([][]byte, 0, len(q.queue))
	for _, id := range q.queue {
		qMessages = append(qMessages, q.messages[id])
	}
	return qMessages
}

func TestLocalSendMessage(t *testing.T) {
	Convey("Given a local queue", t, func() {
		q := NewLocal()

		Convey("When a message is sent", func() {
			message := []byte("foo")
			err := q.SendMessage(message)

			Convey("Then there should not be an error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then it should be in the queue", func() {
				So(getQueueContents(q), ShouldContain, message)
			})
		})

		Convey("When multiple messages are sent", func() {
			messages := [][]byte{[]byte("foo"), []byte("bar"), []byte("baz")}
			for _, message := range messages {
				q.SendMessage(message)
			}

			Convey("Then they should all be in the queue", func() {
				So(getQueueContents(q), ShouldResemble, messages)
			})
		})
	})
}

func TestLocalReceiveMessage(t *testing.T) {
	Convey("Given a local queue with two messages in it", t, func() {
		q := NewLocal()
		sentMessage1 := []byte("foo")
		sentMessage2 := []byte("bar")
		q.SendMessage(sentMessage1)
		q.SendMessage(sentMessage2)

		Convey("When a message is received with no timeout", func() {
			_, message, err := q.ReceiveMessage(0)

			Convey("Then there should not be an error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then it should match the original", func() {
				So(message, ShouldResemble, sentMessage1)
			})

			Convey("Then the queue should no longer contain the message", func() {
				So(q.queue, ShouldNotContain, sentMessage1)
			})

			Convey("Then the queue should have no receive timers", func() {
				So(q.received, ShouldBeEmpty)
			})
		})

		Convey("When a message is received with a timeout", func() {
			timeout := 100 * time.Millisecond
			id, message, err := q.ReceiveMessage(timeout)

			Convey("Then there should not be an error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then it should match the original", func() {
				So(message, ShouldResemble, sentMessage1)
			})

			Convey("Then `id` should be not be an empty string", func() {
				So(id, ShouldNotEqual, "")
			})

			Convey("Then the queue should no longer contain the message", func() {
				So(q.queue, ShouldNotContain, sentMessage1)
			})

			Convey("Then the queue should have a receive timer for the message", func() {
				So(q.received, ShouldContainKey, id)
			})

			Convey("When the timeout expires", func() {
				time.Sleep(timeout)

				Convey("Then the queue should contain the message", func() {
					So(getQueueContents(q), ShouldContain, sentMessage1)
				})

				Convey("Then the message should be the next one received", func() {
					_, nextMessage, _ := q.ReceiveMessage(0)
					So(nextMessage, ShouldResemble, sentMessage1)
				})

				Convey("Then the queue should not have a receive timer for the message", func() {
					So(q.received, ShouldNotContainKey, id)
				})
			})
		})
	})

	Convey("Given a local queue with no messages", t, func() {
		q := NewLocal()

		Convey("When a message is received", func() {
			_, _, err := q.ReceiveMessage(0)

			Convey("Then the error should specify that there are no messages", func() {
				So(err, ShouldEqual, ErrNoMessages)
			})
		})
	})
}

func TestLocalDeleteMessage(t *testing.T) {
	Convey("Given a local queue with a receive timer for a message", t, func() {
		q := NewLocal()
		id := "foo"
		timeout := 100 * time.Millisecond
		timer := time.NewTimer(timeout)
		q.received[id] = timer

		Convey("When the message is deleted", func() {
			q.DeleteMessage(id)

			Convey("Then the queue should not have a receive timer for the message", func() {
				So(q.received, ShouldNotContainKey, id)
			})

			Convey("Then the timer should not fire", func() {
				time.Sleep(timeout)

				So(timer.C, ShouldBeEmpty)
			})
		})
	})
}
