package cache

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLocalGet(t *testing.T) {
	Convey("Given a local cache with an entry", t, func() {
		c := NewLocal()
		key, value := "foo", []byte("bar")
		c.cache[key] = value

		Convey("When the that entry is accessed", func() {
			cVal, err := c.Get(key)

			Convey("Then there should not be an error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the returned value should be the one in the cache", func() {
				So(cVal, ShouldResemble, value)
			})
		})

		Convey("When a non-existent entry is accessed", func() {
			cVal, err := c.Get("baz")

			Convey("Then there should not be an error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the returned value should be nil", func() {
				So(cVal, ShouldBeNil)
			})
		})
	})
}

func TestLocalDelete(t *testing.T) {
	Convey("Given a local cache with an an entry with an expiry", t, func() {
		c := NewLocal()
		key := "foo"
		c.cache[key] = []byte("bar")
		c.timeouts[key] = time.NewTimer(1 * time.Hour)

		Convey("When that entry is deleted", func() {
			err := c.Delete(key)

			Convey("Then there should not be an error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the entry should no longer be in the cache", func() {
				So(c.cache[key], ShouldBeNil)
			})

			Convey("Then the entry's timeout should be cleared", func() {
				So(c.timeouts[key], ShouldBeNil)
			})
		})
	})
}

func TestLocalSet(t *testing.T) {
	Convey("Given an empty local cache", t, func() {
		c := NewLocal()

		Convey("When an entry is set with no timeout", func() {
			key, value := "foo", []byte("bar")
			err := c.Set(key, value, 0)

			Convey("Then there should not be an error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the entry should be in the cache", func() {
				So(c.cache[key], ShouldResemble, value)
			})

			Convey("Then there should not be a timeout stored for the entry", func() {
				So(c.timeouts[key], ShouldBeNil)
			})
		})

		Convey("When an entry is set with a timeout", func() {
			key, value := "foo", []byte("bar")
			timeout := 100 * time.Millisecond
			err := c.Set(key, value, timeout)

			Convey("Then there should not be an error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the entry should be in the cache", func() {
				c.cacheMutex.RLock()
				defer c.cacheMutex.RUnlock()
				So(c.cache[key], ShouldResemble, value)
			})

			Convey("Then the entry's timeout should be stored", func() {
				c.timeoutsMutex.Lock()
				defer c.timeoutsMutex.Unlock()
				So(c.timeouts[key], ShouldNotBeNil)
			})

			Convey("When the timeout expires", func() {
				time.Sleep(timeout)

				Convey("Then the entry should no longer be in the cache", func() {
					c.cacheMutex.RLock()
					defer c.cacheMutex.RUnlock()
					So(c.cache[key], ShouldBeNil)
				})

				Convey("Then the entry's timeout should be cleared", func() {
					c.timeoutsMutex.Lock()
					defer c.timeoutsMutex.Unlock()
					So(c.timeouts[key], ShouldBeNil)
				})
			})
		})
	})

	Convey("Given a local cache with an an entry with an expiry", t, func() {
		c := NewLocal()
		key := "foo"
		c.cache[key] = []byte("bar")
		timeout := 100 * time.Millisecond
		c.timeouts[key] = time.NewTimer(timeout)

		Convey("When that entry is overwritten with no expiry", func() {
			value := []byte("baz")
			err := c.Set(key, value, 0)

			Convey("Then there should not be an error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the entry's new value should be in the cache", func() {
				So(c.cache[key], ShouldResemble, value)
			})

			Convey("Then the entry's timeout should be cleared", func() {
				So(c.timeouts[key], ShouldBeNil)
			})

			Convey("When the timeout expires", func() {
				time.Sleep(timeout)

				Convey("Then the entry should still be in the cache", func() {
					So(c.cache[key], ShouldResemble, value)
				})
			})
		})

		Convey("When that entry is overwritten with a new expiry", func() {
			value := []byte("baz")
			newTimeout := 200 * time.Millisecond
			err := c.Set(key, value, newTimeout)

			Convey("Then there should not be an error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the entry's new value should be in the cache", func() {
				c.cacheMutex.RLock()
				defer c.cacheMutex.RUnlock()
				So(c.cache[key], ShouldResemble, value)
			})

			Convey("Then the entry's new timeout should be stored", func() {
				c.timeoutsMutex.Lock()
				defer c.timeoutsMutex.Unlock()
				So(c.timeouts[key], ShouldNotBeNil)
			})

			Convey("When the original timeout expires", func() {
				time.Sleep(timeout)

				Convey("Then the entry should still be in the cache", func() {
					c.cacheMutex.RLock()
					defer c.cacheMutex.RUnlock()
					So(c.cache[key], ShouldResemble, value)
				})

				Convey("When the new timeout expires", func() {
					time.Sleep(newTimeout - timeout)

					Convey("Then the entry should no longer be in the cache", func() {
						c.cacheMutex.RLock()
						defer c.cacheMutex.RUnlock()
						So(c.cache[key], ShouldBeNil)
					})

					Convey("Then the entry's timeout should be cleared", func() {
						c.timeoutsMutex.Lock()
						defer c.timeoutsMutex.Unlock()
						So(c.timeouts[key], ShouldBeNil)
					})
				})
			})
		})
	})
}
