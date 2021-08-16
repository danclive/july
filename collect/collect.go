package collect

import (
	"errors"
	"sync"
	"time"

	"github.com/danclive/july/device"
	"github.com/danclive/july/log"
	"github.com/danclive/march/consts"
)

var _service *Service

func InitService(readInterval int, keepalive int, connectInterval int) {
	initCache()

	_service = &Service{
		keepalive:       time.Second * time.Duration(keepalive),
		connectInterval: time.Second * time.Duration(connectInterval),
		readInterval:    time.Second * time.Duration(readInterval),
		wires:           make(map[string]*Wire),
		close:           make(chan struct{}),
	}
}

func GetService() *Service {
	return _service
}

type Service struct {
	keepalive       time.Duration
	connectInterval time.Duration
	readInterval    time.Duration
	wires           map[string]*Wire
	lock            sync.RWMutex
	close           chan struct{}
}

func Run() {
	log.Logger.Info("starting collect")

	device.GetService().SlotReset("")

	go _service.connect()
}

func Stop() error {
	select {
	case <-_service.close:
		return nil
	default:
		close(_service.close)
	}

	_service.lock.RLock()
	defer _service.lock.RUnlock()

	for slotId, wire := range _service.wires {
		wire.Close()
		device.GetService().SlotOffline(slotId)
	}

	return nil
}

// 更新 slot/tag 后，需要调用此函数
func (c *Service) Reset(slotId string) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if wire, ok := c.wires[slotId]; ok {
		wire.Close()
		device.GetService().SlotOffline(slotId)
	}
}

// 注意，要写入的标签必须为同一个 slot
func (c *Service) Write(tags []device.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	slotID := tags[0].SlotID

	for _, tag := range tags {
		if tag.SlotID != slotID {
			return errors.New("the tag to write to must be the same slot")
		}

		if tag.Access != consts.ON {
			return errors.New("tag.Access != RW(consts.ON)")
		}

		if tag.Value == nil {
			return errors.New("tag.Value == nil")
		}

		tag.WriteConvert()
	}

	slot, err := device.GetService().GetSlot(slotID)
	if err != nil {
		return err
	}

	if slot.Status != consts.ON {
		return errors.New("slot don't enable")
	}

	c.lock.RLock()
	defer c.lock.RUnlock()

	if wire, ok := c.wires[slotID]; ok {
		wire.Send(tags)
	}

	return nil
}

func (c *Service) connect() {
	for {
		time.Sleep(c.connectInterval)

		select {
		case <-c.close:
			return
		default:
			slots, err := device.GetService().ListSlotStatusOn()
			if err != nil {
				log.Suger.Error(err)
				continue
			}

			for _, slot := range slots {
				func(slot device.Slot) {
					c.lock.Lock()

					if _, ok := c.wires[slot.ID]; ok {
						c.lock.Unlock()
						return
					}
					c.lock.Unlock()

					if driver, ok := _drivers[slot.Driver]; ok {
						conn, err := driver.Connect(slot)
						if err != nil {
							log.Suger.Error(err)
							return
						}

						device.GetService().SlotOnline(slot.ID)

						tags, err := device.GetService().ListTagStatusOnAndTypeIO(slot.ID)
						if err != nil {
							log.Suger.Error(err)
							return
						}

						// fmt.Printf("%#v", tags[0])

						wire := NewWire(conn, slot.ID, tags, c.keepalive, c.readInterval)

						c.lock.Lock()
						c.wires[slot.ID] = wire
						c.lock.Unlock()

						go func(wire *Wire) {
							wire.Run()

							c.lock.Lock()
							delete(c.wires, wire.slotID)
							c.lock.Unlock()

							log.Suger.Errorf("wire break: %v", wire.err)
						}(wire)
					}
				}(slot)
			}
		}
	}
}
