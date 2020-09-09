package july

import (
	"bytes"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/danclive/july/dict"
	"github.com/danclive/july/util"
	"github.com/danclive/mqtt"
	"github.com/danclive/nson-go"
)

type Collect interface {
	Reset(slotId string)
	Write(tags []Tag) error
	Run()
	Stop() error
}

var _ Collect = &collect{}

type collect struct {
	crate        Crate
	keepalive    time.Duration
	readInterval time.Duration
	drivers      map[string]*driverWrap
	lock         sync.Mutex
	close        chan struct{}
}

func newCollect(crate Crate, readInterval int, keepalive int) Collect {
	coll := &collect{
		crate:        crate,
		keepalive:    time.Second * time.Duration(keepalive),
		readInterval: time.Second * time.Duration(readInterval),
		drivers:      make(map[string]*driverWrap),
		close:        make(chan struct{}),
	}

	return coll
}

func (c *collect) free() {
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

func (c *collect) read() {
	ticker := time.NewTicker(c.readInterval)
	for {
		select {
		case <-c.close:
			ticker.Stop()
			return
		case <-ticker.C:
			if err := c.connect(); err != nil {
				suger.Errorf("c.connect(): %s", err)
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
						c.crate.DBus().SetValue(driver.tags[i].Name, driver.tags[i].Value, false)
					}
				}(slotId, driver)
			}

			c.lock.Unlock()
		}
	}
}

func (c *collect) connect() error {
	slots, err := c.crate.SlotService().ListSlotByStatusOn()
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

		// if slot.Driver != DriverS7_TCP && slot.Driver != DriverMODBUS_TCP {
		// 	continue
		// }

		if d, ok := registerDriver[slot.Driver]; ok {
			driver, err := d.Connect(c.crate, slot.Params)
			if err != nil {
				return err
			}

			tags, err := c.crate.SlotService().ListTagByStatusOnAndIO(slot.Id)
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

func (c *collect) Run() {
	zaplog.Info("starting collect")
	go c.free()
	go c.read()
}

func (c *collect) Stop() error {
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
func (c *collect) Reset(slotId string) {
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
func (c *collect) Write(tags []Tag) error {
	if len(tags) == 0 {
		return nil
	}

	slotId := tags[0].SlotId

	for _, tag := range tags {
		if tag.SlotId != slotId {
			return errors.New("要写入的标签必须为同一个 slot")
		}

		if tag.AccessMode != RW {
			return errors.New("tag.AccessMode != RW")
		}

		if tag.Value == nil {
			return errors.New("tag.Value == nil")
		}
	}

	slot, err := c.crate.SlotService().GetSlot(slotId)
	if err != nil {
		return err
	}

	if slot.Status != ON {
		return nil
	}

	if slot.Driver == DriverMQTT {
		return c.writeToMqtt(tags)
	}

	// if slot.Driver != DriverS7_TCP && slot.Driver != DriverMODBUS_TCP {
	// 	if slot.Driver == DriverMQTT {
	// 		return c.writeToMqtt(tags)
	// 	}

	// 	return nil
	// }

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

	if d, ok := registerDriver[slot.Driver]; ok {
		driver, err := d.Connect(c.crate, slot.Params)
		if err != nil {
			return err
		}

		tags, err := c.crate.SlotService().ListTagByStatusOnAndIO(slot.Id)
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

func (c *collect) writeToMqtt(tags []Tag) error {
	slotId := tags[0].SlotId

	mqttServer := c.crate.MqttService().MqttServer()

	client, has := mqttServer.Client(slotId)

	flags := uint16(0)
	if has {
		f, has := client.Get("FLAGS")
		if has {
			flags = uint16(f.(nson.U32))
		}
	}

	if flags == 0 {
		data := make(map[string]interface{})

		for i := 0; i < len(tags); i++ {
			data[tags[i].Address] = tags[i].Value
		}

		msg := map[string]interface{}{
			dict.DATA: data,
		}

		bts, err := json.Marshal(msg)
		if err != nil {
			return err
		}

		buffer := new(bytes.Buffer)
		util.WriteUint16(flags, buffer)

		buffer.Write(bts)

		mqttServer.PublishService().PublishToClient(slotId, mqtt.NewMessage(dict.DEV_DATA_SET, buffer.Bytes(), 0), false)
	} else if flags == 1 {
		buffer := new(bytes.Buffer)

		util.WriteUint16(flags, buffer)

		data := nson.Message{}
		for i := 0; i < len(tags); i++ {
			data[tags[i].Address] = tags[i].Value
		}

		msg := nson.Message{
			"data": data,
		}

		err := msg.Encode(buffer)
		if err != nil {
			return err
		}

		mqttServer.PublishService().PublishToClient(slotId, mqtt.NewMessage(dict.DEV_DATA_SET, buffer.Bytes(), 0), false)
	}

	return nil
}
