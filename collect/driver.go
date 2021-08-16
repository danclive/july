package collect

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/danclive/july/device"
	"github.com/danclive/july/log"
	"go.uber.org/zap"
)

type Driver interface {
	Connect(slot device.Slot) (Driver, error)
	Close() error
	Name() string
	Read([]device.Tag) error
	Write([]device.Tag) error
}

var _drivers = make(map[string]Driver)

func RegisterDriver(name string, driver Driver) {
	_drivers[name] = driver
}

type Wire struct {
	wg            sync.WaitGroup
	conn          Driver
	out           chan []device.Tag
	close         chan struct{}
	closeComplete chan struct{}
	error         chan error
	err           error

	lastUseTime time.Time
	opLock      sync.RWMutex
	lock        sync.Mutex

	keepalive    time.Duration
	readInterval time.Duration

	slotID string
	tags   []device.Tag
}

func NewWire(conn Driver, slotID string, tags []device.Tag, keepalive time.Duration, readInterval time.Duration) *Wire {
	w := &Wire{
		conn:          conn,
		out:           make(chan []device.Tag, 10),
		close:         make(chan struct{}),
		closeComplete: make(chan struct{}),
		error:         make(chan error, 1),
		lastUseTime:   time.Now(),
		keepalive:     keepalive,
		readInterval:  readInterval,
		slotID:        slotID,
		tags:          tags,
	}

	return w
}

func (w *Wire) Close() <-chan struct{} {
	w.setError(nil)
	return w.closeComplete
}

func (w *Wire) setError(err error) {
	select {
	case w.error <- err:
		if err != nil && err != io.EOF {
			log.Logger.Error("connection lost", zap.String("error_msg", err.Error()))
		}
	default:
	}
}

func (w *Wire) write(tags []device.Tag) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	err := w.conn.Write(tags)
	if err != nil {
		return err
	}

	w.setLastUseTime()
	return nil
}

func (w *Wire) read() error {
	w.lock.Lock()
	defer w.lock.Unlock()

	err := w.conn.Read(w.tags)
	if err != nil {
		log.Suger.Debugf("read: %v", err)
		return err
	}

	for i := 0; i < len(w.tags); i++ {
		if w.tags[i].Value != nil {
			w.tags[i].ReadConvert()
			CacheSet(w.tags[i].ID, w.tags[i].Value)
		}
	}

	w.setLastUseTime()
	return nil
}

func (w *Wire) setLastUseTime() {
	w.opLock.Lock()
	w.lastUseTime = time.Now()
	w.opLock.Unlock()
}

func (w *Wire) LastUseTime() time.Time {
	w.opLock.RLock()
	defer w.opLock.RUnlock()
	return w.lastUseTime
}

func (w *Wire) writeLoop() {
	var err error

	defer func() {
		if re := recover(); re != nil {
			err = errors.New(fmt.Sprint(re))
		}
		w.setError(err)
		w.wg.Done()
		log.Logger.Debug("write loop thread exit")
	}()

	for {
		select {
		case <-w.close:
			return
		case msg := <-w.out:

			err = w.write(msg)
			if err != nil {
				return
			}
		}
	}
}

func (w *Wire) readLoop() {
	var err error

	defer func() {
		if re := recover(); re != nil {
			err = errors.New(fmt.Sprint(re))
		}
		w.setError(err)
		w.wg.Done()
		log.Logger.Debug("read loop thread exit")
	}()

	for {
		time.Sleep(w.readInterval)
		log.Suger.Debugf("read, id: %s", w.slotID)
		err = w.read()
		if err != nil {
			return
		}
	}
}

func (w *Wire) errorWatch() {
	defer func() {
		w.wg.Done()
		log.Logger.Debug("error watch thread exit")
	}()

	select {
	case <-w.close:
		return
	case err := <-w.error: //有错误关闭
		w.err = err
		w.lock.Lock()
		w.conn.Close()
		w.lock.Unlock()
		close(w.close) //退出chanel
		return
	}
}

func (w *Wire) free() {
	defer func() {
		w.setError(errors.New("free"))
		w.wg.Done()
		log.Logger.Debug("free thread exit")
	}()

	ticker := time.NewTicker(w.keepalive)

	for {
		select {
		case <-w.close:
			return
		case t := <-ticker.C:
			if t.Sub(w.lastUseTime) > w.keepalive {
				return
			}
		}
	}
}

func (w *Wire) Send(tags []device.Tag) {
	select {
	case <-w.close:
		return
	case w.out <- tags:
		log.Suger.Debugf("send tags: %v", tags)
	}
}

func (w *Wire) Run() {
	log.Suger.Debugf("wire run, id: %v", w.slotID)
	defer close(w.closeComplete)

	w.wg.Add(4)

	go w.errorWatch()
	go w.readLoop()
	go w.writeLoop()
	go w.free()

	w.wg.Wait()
}
