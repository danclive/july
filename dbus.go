package july

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/danclive/nson-go"
	"go.etcd.io/bbolt"
)

var DBUS_BUCKET = []byte("dbus")

type DBus interface {
	Get(tagName string) (Tag, bool)
	Delete(tagName string)
	SetValue(tagName string, value nson.Value, ioWrite bool) error
	GetValue(tagName string) (nson.Value, bool)
	UpdateTagConfig(tagName string)
}

var _ DBus = &dBus{}

type dBus struct {
	crate Crate
	cache map[string]Tag
	lock  sync.Mutex
}

func newDBus(crate Crate) DBus {
	bus := &dBus{
		crate: crate,
		cache: make(map[string]Tag),
	}

	return bus
}

// func (d *dBus) Clear() {
// 	d.lock.RLock()
// 	defer d.lock.RUnlock()

// 	d.cache = nil
// 	d.cache = make(map[string]Tag)
// }

func (d *dBus) getTag(tagName string) (Tag, bool) {
	tag, ok := d.cache[tagName]
	if ok {
		return tag, true
	}

	tag2 := d.getTagByName(tagName)
	if tag2 != nil {
		d.cache[tagName] = *tag2
		return *tag2, true
	}

	return Tag{}, false
}

func (d *dBus) Get(tagName string) (Tag, bool) {
	d.lock.Lock()
	defer d.lock.Unlock()

	tag, has := d.getTag(tagName)
	if !has {
		return tag, has
	}

	if tag.TagType == TagTypeCFG {
		var value nson.Value

		err := d.crate.BboltDB().View(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket(DBUS_BUCKET)
			v := bucket.Get([]byte(tagName))
			if v == nil {
				return errors.New("not exist")
			}

			buffer := bytes.NewBuffer(v)

			tag, err := buffer.ReadByte()
			if err != nil {
				return err
			}

			v2, err := decode_value(buffer, tag)
			if err != nil {
				return err
			}

			value = v2
			return nil

		})

		if err != nil {
			suger.Debugf("crate.BboltDB().View: %s", err)
		}

		if err == nil && value != nil {
			if tag.Value.Tag() == value.Tag() {
				tag.Value = value
			} else {
				suger.Errorf("tag.Value.Tag(%v) != value.Tag(%v)", tag.Value.Tag(), value.Tag())
			}
		}
	}

	return tag, has
}

func (d *dBus) Delete(tagName string) {
	d.lock.Lock()
	defer d.lock.Unlock()

	delete(d.cache, tagName)
}

func (d *dBus) GetValue(tagName string) (nson.Value, bool) {
	tag, has := d.Get(tagName)
	if has {
		return tag.Value, true
	}

	return nil, false
}

func (d *dBus) SetValue(tagName string, value nson.Value, ioWrite bool) error {
	if value == nil {
		return nil
	}

	d.lock.Lock()
	defer d.lock.Unlock()

	tag, has := d.getTag(tagName)
	if !has {
		return errors.New("not exist")
	}

	switch tag.DataType {
	case TypeI8, TypeI16, TypeI32:
		if value.Tag() != nson.TAG_I32 {
			return errors.New("value.Tag() != nson.TAG_I32")
		}
	case TypeU8, TypeU16, TypeU32:
		if value.Tag() != nson.TAG_U32 {
			return errors.New("value.Tag() != nson.TAG_U32")
		}
	case TypeI64:
		if value.Tag() != nson.TAG_I64 {
			return errors.New("value.Tag() != nson.TAG_I64")
		}
	case TypeU64:
		if value.Tag() != nson.TAG_U64 {
			return errors.New("value.Tag() != nson.TAG_U64")
		}
	case TypeF32:
		if value.Tag() != nson.TAG_F32 {
			return errors.New("value.Tag() != nson.TAG_F32")
		}
	case TypeF64:
		if value.Tag() != nson.TAG_F64 {
			return errors.New("value.Tag() != nson.TAG_F64")
		}
	case TypeBool:
		if value.Tag() != nson.TAG_BOOL {
			return errors.New("value.Tag() != nson.TAG_BOOL")
		}
	case TypeString:
		if value.Tag() != nson.TAG_STRING {
			return errors.New("value.Tag() != nson.TAG_STRING")
		}
	default:
		return errors.New("value.Tag() 不支持、")
	}

	if tag.TagType != TagTypeCFG {
		tag.Value = value
	}

	d.cache[tagName] = tag

	if tag.TagType == TagTypeIO {
		// 如果 ioWrite 为 true，写入到设备
		if ioWrite {
			go d.IoWrite(tag)
		}
	} else if tag.TagType == TagTypeCFG {
		err := d.crate.BboltDB().Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket(DBUS_BUCKET)

			buffer := new(bytes.Buffer)

			err := buffer.WriteByte(value.Tag())
			if err != nil {
				return err
			}

			err = value.Encode(buffer)
			if err != nil {
				return err
			}

			return b.Put([]byte(tagName), buffer.Bytes())
		})

		if err != nil {
			suger.Error("db.BboltDB.Update:", err)
			return err
		}
	}

	return nil
}

// 此方法用来在修改标签配置后更新配置，到保留之前的值
func (d *dBus) UpdateTagConfig(tagName string) {
	d.lock.Lock()
	defer d.lock.Unlock()

	var value nson.Value

	if tag, ok := d.cache[tagName]; ok {
		value = tag.Value
		delete(d.cache, tagName)
	}

	tag := d.getTagByName(tagName)
	if tag == nil {
		return
	}

	tag.Value = value
	d.cache[tagName] = *tag
}

// 从数据库查询标签
func (d *dBus) getTagByName(tagName string) *Tag {
	tag, err := d.crate.SlotService().GetTagByName(tagName)
	if err != nil {
		suger.Errorf("crate.SlotService().GetTagByName: %s", err)
		return nil
	}

	if tag == nil {
		return nil
	}

	// if tag.Status != ON {
	// 	return nil
	// }

	tag.Value = tag.DefaultValue()

	return tag
}

func (d *dBus) IoWrite(tag Tag) {
	if coll := d.crate.Collect(); coll != nil {
		err := coll.Write([]Tag{tag})
		if err != nil {
			suger.Warnf("coll.Write([]Tag{tag}): %s", err)
		}
	}
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
