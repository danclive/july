package july

import (
	"sync"
	"time"
)

type Driver interface {
	Connect(crate Crate, params string) (Driver, error)
	Close() error
	Name() string
	Read([]Tag) error
	Write([]Tag) error
}

var registerDriver = make(map[string]Driver)

func RegisterDriver(name string, driver Driver) {
	registerDriver[name] = driver
}

type driverWrap struct {
	driver  Driver
	lastUse time.Time
	tags    []Tag
	lock    sync.Mutex
}

func (d *driverWrap) read(tags []Tag) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.lastUse = time.Now()

	return d.driver.Read(tags)
}

func (d *driverWrap) write(tags []Tag) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.lastUse = time.Now()

	return d.driver.Write(tags)
}
