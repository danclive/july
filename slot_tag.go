package july

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/danclive/july/util"
	"github.com/danclive/nson-go"
)

type Tag struct {
	Id         string     `xorm:"pk 'id'" json:"id"`
	DeletedAt  time.Time  `xorm:"deleted" json:"-"`
	CreatedAt  time.Time  `xorm:"created" json:"created"`
	UpdatedAt  time.Time  `xorm:"updated" json:"updated"`
	SlotId     string     `xorm:"slot_id" json:"slot_id"`
	Name       string     `xorm:"'name'" json:"name"`
	Desc       string     `xorm:"'desc'" json:"desc"`
	Unit       string     `xorm:"'unit'" json:"unit"`               // 数据单位
	TagType    string     `xorm:"'tag_type'" json:"tag_type"`       // 标签类型 MEM，IO，CFG
	DataType   string     `xorm:"'data_type'" json:"data_type"`     // 数据类型
	Format     string     `xorm:"'format'" json:"format"`           // 数据格式化
	Address    string     `xorm:"'address'" json:"address"`         // 寄存器
	AccessMode int        `xorm:"'access_mode'" json:"access_mode"` // 读写数据模式， 1: RO，2: RW
	Upload     int        `xorm:"'upload'" json:"upload"`           // 上传数据，1: ON，-1: OFF
	Save       int        `xorm:"'save'" json:"save"`               // 保存数据，1: ON，-1: OFF
	Visible    int        `xorm:"'visible'" json:"visible"`         // 可见性，1: ON，-1: OFF
	Status     int        `xorm:"'status'" json:"status"`           // 状态 1: ON，-1: OFF
	Order      int        `xorm:"'order'" json:"order"`             // 排序
	Version    int        `xorm:"version" json:"version"`
	Value      nson.Value `xorm:"-" json:"value"`
}

// DefaultValue
// LValue
// LLVAlue
// HValue
// HHValue

func (*Tag) TableName() string {
	return "tags"
}

func (t *Tag) MarshalJSON() ([]byte, error) {
	createdAt := ""
	if !t.CreatedAt.IsZero() {
		createdAt = util.TimeFormat(t.CreatedAt)
	}

	updatedAt := ""
	if !t.UpdatedAt.IsZero() {
		updatedAt = util.TimeFormat(t.UpdatedAt)
	}

	tmpItem := struct {
		Tag
		CreatedAt string `json:"created"`
		UpdatedAt string `json:"updated"`
	}{
		Tag:       *t,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	return json.Marshal(tmpItem)
}

func (t *Tag) DefaultValue() nson.Value {
	switch t.DataType {
	case TypeI8, TypeI16, TypeI32:
		return nson.I32(0)
	case TypeU8, TypeU16, TypeU32:
		return nson.U32(0)
	case TypeI64:
		return nson.I64(0)
	case TypeU64:
		return nson.U64(0)
	case TypeF32:
		return nson.F32(0)
	case TypeF64:
		return nson.F64(0)
	case TypeBool:
		return nson.Bool(false)
	case TypeString:
		return nson.String("")
	default:
		return nson.Null{}
	}
}

func (t *Tag) FormatValue() {
	if t.Value != nil && t.Format != "" {
		switch t.Value.Tag() {
		case nson.TAG_F32:
			s := fmt.Sprintf("%"+t.Format, t.Value.(nson.F32))
			if n, err := strconv.ParseFloat(s, 32); err == nil {
				t.Value = nson.F32(n)
			}
		case nson.TAG_F64:
			s := fmt.Sprintf("%"+t.Format, t.Value.(nson.F64))
			if n, err := strconv.ParseFloat(s, 64); err == nil {
				t.Value = nson.F64(n)
			}
		}
	}
}
