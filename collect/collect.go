package collect

import (
	"errors"
	"sync"
	"time"

	"github.com/danclive/july/consts"
	"github.com/danclive/july/device"
	"github.com/danclive/july/log"
)

var _service *Service

func InitService(readInterval int, keepalive int) {
	initCache()

	_service = &Service{
		keepalive:    time.Second * time.Duration(keepalive),
		readInterval: time.Second * time.Duration(readInterval),
		drivers:      make(map[string]*driverWrap),
		close:        make(chan struct{}),
	}
}

func GetService() *Service {
	return _service
}

type Service struct {
	keepalive    time.Duration
	readInterval time.Duration
	drivers      map[string]*driverWrap
	lock         sync.Mutex
	close        chan struct{}
}

func Run() {
	log.Logger.Info("starting collect")

	device.GetService().SlotReset("")

	go _service.free()
	go _service.read()
}

func Stop() error {
	select {
	case <-_service.close:
		return nil
	default:
		close(_service.close)
	}

	_service.lock.Lock()
	defer _service.lock.Unlock()

	for slotId, driver := range _service.drivers {
		driver.lock.Lock()
		delete(_service.drivers, slotId)
		go driver.driver.Close()
		driver.lock.Unlock()

		device.GetService().SlotOffline(slotId)
	}

	return nil
}

// 更新 slot/tag 后，需要调用此函数
func (c *Service) Reset(slotId string) {
	c.lock.Lock()
	if v, ok := c.drivers[slotId]; ok {
		v.lock.Lock()
		delete(c.drivers, slotId)
		go v.driver.Close()
		v.lock.Unlock()

		device.GetService().SlotOffline(slotId)
	}
	c.lock.Unlock()
}

// 注意，要写入的标签必须为同一个 slot
func (c *Service) Write(tags []device.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	slotId := tags[0].SlotId

	for _, tag := range tags {
		if tag.SlotId != slotId {
			return errors.New("the tag to write to must be the same slot")
		}

		if tag.AccessMode != consts.RW {
			return errors.New("tag.AccessMode != RW")
		}

		if tag.Value == nil {
			return errors.New("tag.Value == nil")
		}
	}

	slot, err := device.GetService().GetSlot(slotId)
	if err != nil {
		return err
	}

	if slot.Status != consts.ON {
		return errors.New("slot don't enable")
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	if dw, ok := c.drivers[slotId]; ok {
		go func() {
			err := dw.write(tags)
			if err != nil {
				log.Suger.Error(err)
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

		tags, err := device.GetService().ListTagStatusOnAndTypeIO(slot.ID)
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
				log.Suger.Error(err)
				c.Reset(slotId)
			}
		}()

		c.drivers[slotId] = dw
	}

	return nil
}

func (c *Service) free() {
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

func (c *Service) read() {
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
						log.Suger.Error(err)
						c.Reset(slotId)
					}
					//fmt.Println(driver.tags)

					for i := 0; i < len(driver.tags); i++ {
						CacheSet(driver.tags[i].ID, driver.tags[i].Value)
					}
				}(slotId, driver)
			}

			c.lock.Unlock()
		}
	}
}

func (c *Service) connect() error {
	slots, err := device.GetService().ListSlotStatusOn()
	if err != nil {
		return err
	}
	// fmt.Println(slots)

	c.lock.Lock()
	defer c.lock.Unlock()

	for _, slot := range slots {
		if _, ok := c.drivers[slot.ID]; ok {
			continue
		}

		if d, ok := _drivers[slot.Driver]; ok {
			driver, err := d.Connect(slot.Params)
			if err != nil {
				return err
			}

			device.GetService().SlotOnline(slot.ID)

			tags, err := device.GetService().ListTagStatusOnAndTypeIO(slot.ID)
			if err != nil {
				return err
			}

			dw := &driverWrap{
				driver:  driver,
				lastUse: time.Now(),
				tags:    tags,
			}

			c.drivers[slot.ID] = dw
		}
	}

	return nil
}
