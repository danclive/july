package slot

import (
	"fmt"
	"strconv"

	"github.com/danclive/july/util"
	"github.com/danclive/nson-go"
)

const (
	TagTypeIO  = "IO"
	TagTypeMEM = "MEM"
	TagTypeCFG = "CFG"
)

const (
	TypeBool   = "BOOL"
	TypeF32    = "F32"
	TypeF64    = "F64"
	TypeI8     = "I8"
	TypeU8     = "U8"
	TypeI16    = "I16"
	TypeU16    = "U16"
	TypeI32    = "I32"
	TypeI64    = "I64"
	TypeU32    = "U32"
	TypeU64    = "U64"
	TypeString = "STRING"
)

type Tag struct {
	Id         string      `xorm:"pk 'id'" json:"id"`
	SlotId     string      `xorm:"slot_id" json:"slot_id"`
	Name       string      `xorm:"'name'" json:"name"`
	Desc       string      `xorm:"'desc'" json:"desc"`
	Unit       string      `xorm:"'unit'" json:"unit"`               // 数据单位
	TagType    string      `xorm:"'tag_type'" json:"tag_type"`       // 标签类型 MEM，IO，CFG
	DataType   string      `xorm:"'data_type'" json:"data_type"`     // 数据类型
	Format     string      `xorm:"'format'" json:"format"`           // 数据格式化
	Address    string      `xorm:"'address'" json:"address"`         // 寄存器
	AccessMode int32       `xorm:"'access_mode'" json:"access_mode"` // 读写数据模式， 1: RO，2: RW
	Upload     int32       `xorm:"'upload'" json:"upload"`           // 上传数据，1: ON，-1: OFF
	Save       int32       `xorm:"'save'" json:"save"`               // 保存数据，1: ON，-1: OFF
	Visible    int32       `xorm:"'visible'" json:"visible"`         // 可见性，1: ON，-1: OFF
	Status     int32       `xorm:"'status'" json:"status"`           // 状态 1: ON，-1: OFF
	Order      int32       `xorm:"'order'" json:"order"`             // 排序
	Version    int32       `xorm:"version" json:"version"`
	Value      nson.Value  `xorm:"-" json:"value"`
	DeletedAt  util.MyTime `xorm:"deleted" json:"-"`
	CreatedAt  util.MyTime `xorm:"created" json:"created"`
	UpdatedAt  util.MyTime `xorm:"updated" json:"updated"`
}

// DefaultValue
// LValue
// LLVAlue
// HValue
// HHValue

func (*Tag) TableName() string {
	return "tags"
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
