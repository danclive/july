package cache

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/danclive/july/bolt"
	"github.com/danclive/july/collect"
	"github.com/danclive/july/device"
	"github.com/danclive/july/log"
	"github.com/danclive/nson-go"
	"go.etcd.io/bbolt"
)

var _service *Service

func InitService() {
	_service = &Service{
		cache: make(map[string]nson.Value),
	}
}

func GetService() *Service {
	return _service
}

type Service struct {
	cache map[string]nson.Value
	lock  sync.RWMutex
}

func (s *Service) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.cache = nil
	s.cache = make(map[string]nson.Value)
}

func (s *Service) GetTagById(id string) (*device.Tag, error) {
	tag, err := device.GetService().GetTag(id)
	if err != nil {
		return nil, err
	}

	if tag == nil {
		return nil, errors.New("not found")
	}

	err = s.GetValue(tag)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *Service) GetTagByName(name string) (*device.Tag, error) {
	if strings.Contains(name, ".") {
		split := strings.Split(name, ".")
		if len(split) != 2 {
			return nil, errors.New("not found")
		}

		slot, err := device.GetService().GetSlotByName(split[0])
		if err != nil {
			return nil, err
		}

		tag, err := device.GetService().GetTagBySlotIdAndName(slot.Id, split[1])
		if err != nil {
			return nil, err
		}

		err = s.GetValue(tag)
		if err != nil {
			return nil, err
		}

		return tag, nil
	}

	tag, err := device.GetService().GetTagByName(name)

	if err != nil {
		return nil, err
	}

	if tag == nil {
		return nil, errors.New("not found")
	}

	err = s.GetValue(tag)
	if err != nil {
		log.Suger.Error(err)
	}

	return tag, nil
}

func (s *Service) GetValue(tag *device.Tag) error {
	switch tag.Type {
	case device.TypeIO:
		tag.Value = collect.CacheGet(tag.Id)
	case device.TypeCFG:
		err := bolt.GetBoltDB().View(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket(bolt.CFG_BUCKET)
			v := bucket.Get([]byte(tag.Id))
			if v == nil {
				return fmt.Errorf("tag: %v(%v) not exit", tag.Name, tag.Id)
			}

			buffer := bytes.NewBuffer(v)

			data_tag, err := buffer.ReadByte()
			if err != nil {
				return err
			}

			value, err := decode_value(buffer, data_tag)
			if err != nil {
				return err
			}

			if tag.DefaultValue().Tag() != value.Tag() {
				return fmt.Errorf("data type not match, expect: %v, provide: %v", tag.DefaultValue().Tag(), value.Tag())
			}

			tag.Value = value
			return nil
		})

		if err != nil {
			log.Suger.Warn("bolt.BoltDB.View:", err)
		}
	default:
		s.lock.RLock()
		defer s.lock.RUnlock()
		tag.Value = s.cache[tag.Id]
	}

	if tag.Value == nil {
		tag.Value = tag.DefaultValue()
	}

	return nil
}

func (s *Service) SetValueById(id string, value nson.Value) error {
	tag, err := device.GetService().GetTag(id)
	if err != nil {
		return err
	}

	if tag == nil {
		return errors.New("not found")
	}

	return s.SetValue(tag, value)
}

func (s *Service) SetValueByName(name string, value nson.Value) error {
	if strings.Contains(name, ".") {
		split := strings.Split(name, ".")
		if len(split) != 2 {
			return errors.New("not found")
		}

		slot, err := device.GetService().GetSlotByName(split[0])
		if err != nil {
			return err
		}

		tag, err := device.GetService().GetTagBySlotIdAndName(slot.Id, split[1])
		if err != nil {
			return err
		}

		if tag == nil {
			return errors.New("not found")
		}

		return s.SetValue(tag, value)
	}

	tag, err := device.GetService().GetTagByName(name)

	if err != nil {
		return err
	}

	if tag == nil {
		return errors.New("not found")
	}

	return s.SetValue(tag, value)
}

func (s *Service) SetValue(tag *device.Tag, value nson.Value) error {
	if tag == nil {
		return errors.New("tag in nil")
	}

	if tag.DefaultValue().Tag() != value.Tag() {
		return fmt.Errorf("data type not match, expect: %v, provide: %v", tag.DefaultValue().Tag(), value.Tag())
	}

	switch tag.Type {
	case device.TypeIO:
		tag.Value = value
		return collect.GetService().Write([]device.Tag{*tag})
	case device.TypeCFG:
		err := bolt.GetBoltDB().Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket(bolt.CFG_BUCKET)

			buffer := new(bytes.Buffer)

			err := buffer.WriteByte(value.Tag())
			if err != nil {
				return err
			}

			err = value.Encode(buffer)
			if err != nil {
				return err
			}

			return b.Put([]byte(tag.Id), buffer.Bytes())
		})

		if err != nil {
			log.Suger.Error("bolt.BoltDB.Update:", err)
			return err
		}
	default:
		s.lock.Lock()
		defer s.lock.Unlock()
		s.cache[tag.Id] = value
	}

	return nil
}

func decode_value(buf *bytes.Buffer, tag uint8) (nson.Value, error) {
	switch tag {
	case nson.TAG_F32:
		value, err := nson.F32(0).Decode(buf)
		if err != nil {
			return nil, err
		}

		return value, nil
	case nson.TAG_F64:
		value, err := nson.F64(0).Decode(buf)
		if err != nil {
			return nil, err
		}

		return value, nil
	case nson.TAG_I32:
		value, err := nson.I32(0).Decode(buf)
		if err != nil {
			return nil, err
		}

		return value, nil
	case nson.TAG_I64:
		value, err := nson.I64(0).Decode(buf)
		if err != nil {
			return nil, err
		}

		return value, nil
	case nson.TAG_U32:
		value, err := nson.U32(0).Decode(buf)
		if err != nil {
			return nil, err
		}

		return value, nil
	case nson.TAG_U64:
		value, err := nson.U64(0).Decode(buf)
		if err != nil {
			return nil, err
		}

		return value, nil
	case nson.TAG_STRING:
		value, err := nson.String("").Decode(buf)
		if err != nil {
			return nil, err
		}

		return value, nil
	case nson.TAG_ARRAY:
		value, err := nson.Array{}.Decode(buf)
		if err != nil {
			return nil, err
		}

		return value, nil
	case nson.TAG_BOOL:
		value, err := nson.Bool(false).Decode(buf)
		if err != nil {
			return nil, err
		}

		return value, nil
	case nson.TAG_NULL:
		value, err := nson.Null{}.Decode(buf)
		if err != nil {
			return nil, err
		}

		return value, nil
	case nson.TAG_BINARY:
		value, err := nson.Binary{}.Decode(buf)
		if err != nil {
			return nil, err
		}

		return value, nil
	case nson.TAG_TIMESTAMP:
		value, err := nson.Timestamp(0).Decode(buf)
		if err != nil {
			return nil, err
		}

		return value, nil
	case nson.TAG_MESSAGE_ID:
		value, err := nson.MessageId{}.Decode(buf)
		if err != nil {
			return nil, err
		}

		return value, nil
	case nson.TAG_MESSAGE:
		value, err := nson.Message{}.Decode(buf)
		if err != nil {
			return nil, err
		}

		return value, nil
	default:
		return nil, fmt.Errorf("Unsupported type '%X'", tag)
	}
}
