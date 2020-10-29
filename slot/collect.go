package slot

import (
	"errors"
	"sync"
	"time"

	"github.com/danclive/july/consts"
	"github.com/danclive/july/log"
)

var _collect *Collect

func initCollect(readInterval int, keepalive int) {
	_collect = newCollect(readInterval, keepalive)
}

func GetCollect() *Collect {
	return _collect
}

type Collect struct {
	keepalive    time.Duration
	readInterval time.Duration
	drivers      map[string]*driverWrap
	lock         sync.Mutex
	close        chan struct{}
}

func newCollect(readInterval int, keepalive int) *Collect {
	coll := &Collect{
		keepalive:    time.Second * time.Duration(keepalive),
		readInterval: time.Second * time.Duration(readInterval),
		drivers:      make(map[string]*driverWrap),
		close:        make(chan struct{}),
	}

	return coll
}

func (c *Collect) free() {
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-c.close:
			ticker.Stop()
			return
		case <-ticker.C:
			c.lock.Lock()
			for k, v := range c.drivers {
				// 如果一定时间内未使用，释放
				v.lock.Lock()
				if time.Now().After(v.lastUse.Add(c.keepalive)) {
					delete(c.drivers, k)
					go v.driver.Close()
				}
				v.lock.Unlock()
			}
			c.lock.Unlock()
		}
	}
}

func (c *Collect) read() {
	ticker := time.NewTicker(c.readInterval)
	for {
		select {
		case <-c.close:
			ticker.Stop()
			return
		case <-ticker.C:
			if err := c.connect(); err != nil {
				log.Suger.Errorf("c.connect(): %s", err)
				continue
			}

			c.lock.Lock()

			for slotId, driver := range c.drivers {
				go func(slotId string, driver *driverWrap) {
					err := driver.read(driver.tags)
					if err != nil {
						c.Reset(slotId)
					}
					//fmt.Println(driver.tags)

					for i := 0; i < len(driver.tags); i++ {
						GetCache().SetValue(driver.tags[i].Name, driver.tags[i].Value, false)
					}
				}(slotId, driver)
			}

			c.lock.Unlock()
		}
	}
}

func (c *Collect) connect() error {
	slots, err := GetService().ListSlotByStatusOn()
	if err != nil {
		return err
	}
	// fmt.Println(slots)

	c.lock.Lock()
	defer c.lock.Unlock()

	for _, slot := range slots {
		if _, ok := c.drivers[slot.Id]; ok {
			continue
		}

		if d, ok := _drivers[slot.Driver]; ok {
			driver, err := d.Connect(slot.Params)
			if err != nil {
				return err
			}

			tags, err := GetService().ListTagByStatusOnAndIO(slot.Id)
			if err != nil {
				return err
			}

			dw := &driverWrap{
				driver:  driver,
				lastUse: time.Now(),
				tags:    tags,
			}

			c.drivers[slot.Id] = dw
		}
	}

	return nil
}

func (c *Collect) Run() {
	log.Logger.Info("starting collect")
	go c.free()
	go c.read()
}

func (c *Collect) Stop() error {
	select {
	case <-c.close:
		return nil
	default:
		close(c.close)
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	for slotId, driver := range c.drivers {
		driver.lock.Lock()
		delete(c.drivers, slotId)
		go driver.driver.Close()
		driver.lock.Unlock()
	}

	return nil
}

// 更新 slot/tag 后，需要调用此函数
func (c *Collect) Reset(slotId string) {
	c.lock.Lock()
	if v, ok := c.drivers[slotId]; ok {
		v.lock.Lock()
		delete(c.drivers, slotId)
		go v.driver.Close()
		v.lock.Unlock()
	}
	c.lock.Unlock()
}

// 注意，要写入的标签必须为同一个 slot
func (c *Collect) Write(tags []Tag) error {
	if len(tags) == 0 {
		return nil
	}

	slotId := tags[0].SlotId

	for _, tag := range tags {
		if tag.SlotId != slotId {
			return errors.New("要写入的标签必须为同一个 slot")
		}

		if tag.AccessMode != consts.RW {
			return errors.New("tag.AccessMode != RW")
		}

		if tag.Value == nil {
			return errors.New("tag.Value == nil")
		}
	}

	slot, err := GetService().GetSlot(slotId)
	if err != nil {
		return err
	}

	if slot.Status != consts.ON {
		return nil
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	if dw, ok := c.drivers[slotId]; ok {
		go func() {
			err := dw.write(tags)
			if err != nil {
				c.Reset(slotId)
			}
		}()

		return nil
	}

	if d, ok := _drivers[slot.Driver]; ok {
		driver, err := d.Connect(slot.Params)
		if err != nil {
			return err
		}

		tags, err := GetService().ListTagByStatusOnAndIO(slot.Id)
		if err != nil {
			return err
		}

		dw := &driverWrap{
			driver:  driver,
			lastUse: time.Now(),
			tags:    tags,
		}

		go func() {
			err := dw.write(tags)
			if err != nil {
				c.Reset(slotId)
			}
		}()

		c.drivers[slotId] = dw
	}

	return nil
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
