package slot

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/danclive/july/bolt"
	"github.com/danclive/july/log"
	"github.com/danclive/nson-go"
	"go.etcd.io/bbolt"
)

var _cache *Cache

func initCache() {
	_cache = newCache()
}

func GetCache() *Cache {
	return _cache
}

type Cache struct {
	cache map[string]Tag
	lock  sync.Mutex
}

func newCache() *Cache {
	bus := &Cache{
		cache: make(map[string]Tag),
	}

	return bus
}

func (d *Cache) Clear() {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.cache = nil
	d.cache = make(map[string]Tag)
}

func (d *Cache) getTag(tagName string) (Tag, bool) {
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

func (d *Cache) Get(tagName string) (Tag, bool) {
	d.lock.Lock()
	defer d.lock.Unlock()

	tag, has := d.getTag(tagName)
	if !has {
		return tag, has
	}

	if tag.TagType == TagTypeCFG {
		var value nson.Value

		err := bolt.GetBoltDB().View(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket(bolt.CFG_BUCKET)
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
			log.Suger.Debugf("bolt.BoltDB.View: %s", err)
		}

		if err == nil && value != nil {
			if tag.Value.Tag() == value.Tag() {
				tag.Value = value
			} else {
				log.Suger.Errorf("tag.Value.Tag(%v) != value.Tag(%v)", tag.Value.Tag(), value.Tag())
			}
		}
	}

	return tag, has
}

func (d *Cache) Delete(tagName string) {
	d.lock.Lock()
	defer d.lock.Unlock()

	delete(d.cache, tagName)
}

func (d *Cache) GetValue(tagName string) (nson.Value, bool) {
	tag, has := d.Get(tagName)
	if has {
		return tag.Value, true
	}

	return nil, false
}

func (d *Cache) SetValue(tagName string, value nson.Value, ioWrite bool) error {
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

			return b.Put([]byte(tagName), buffer.Bytes())
		})

		if err != nil {
			log.Suger.Error("bolt.BoltDB.Update:", err)
			return err
		}
	}

	return nil
}

// 此方法用来在修改标签配置后更新配置，到保留之前的值
func (d *Cache) UpdateTagConfig(tagName string) {
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
func (d *Cache) getTagByName(tagName string) *Tag {
	tag, err := GetService().GetTagByName(tagName)
	if err != nil {
		log.Suger.Errorf("Service.GetTagByName: %s", err)
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

func (d *Cache) IoWrite(tag Tag) {
	err := GetCollect().Write([]Tag{tag})
	if err != nil {
		log.Suger.Warnf("Collect.Write([]Tag{tag}): %s", err)
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
