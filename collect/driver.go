package collect

import (
	"sync"
	"time"

	"github.com/danclive/july/device"
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

type driverWrap struct {
	driver  Driver
	lastUse time.Time
	tags    []device.Tag
	lock    sync.Mutex
}

func (d *driverWrap) read(tags []device.Tag) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.lastUse = time.Now()

	return d.driver.Read(tags)
}

func (d *driverWrap) write(tags []device.Tag) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.lastUse = time.Now()

	return d.driver.Write(tags)
}
