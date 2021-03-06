package device

import (
	"fmt"
	"strconv"

	"github.com/danclive/july/util"
	"github.com/danclive/march/consts"
	"github.com/danclive/nson-go"
)

type Slot struct {
	ID         string      `xorm:"pk 'id'" json:"id"`
	Name       string      `xorm:"'name'" json:"name"`
	Desc       string      `xorm:"'desc'" json:"desc"`
	Model      string      `xorm:"'model'" json:"model"`       // 型号
	Driver     string      `xorm:"'driver'" json:"driver"`     // 驱动 S7-TCP, MODBUS-TCP, MQTT
	Params     string      `xorm:"'params'" json:"params"`     // 参数
	Config     string      `xorm:"'cfg'" json:"cfg"`           // 配置
	ConfigFile string      `xorm:"'cfg_file'" json:"cfg_file"` // 配置文件
	LinkStatus int32       `xorm:"'link'" json:"link"`         // 连接状态，1: ON，-1: OFF
	Fault      int32       `xorm:"'fault'" json:"fault"`       // 故障状态，1: 故障，-1：正常
	Update     int32       `xorm:"'update'" json:"update"`     // 更新状态，1: 已更新，-1：未更新
	Status     int32       `xorm:"'status'" json:"status"`     // 状态 1: ON，-1: OFF
	Order      int32       `xorm:"'order'" json:"order"`
	Version    int32       `xorm:"version" json:"version"`
	DeletedAt  util.MyTime `xorm:"deleted" json:"-"`
	CreatedAt  util.MyTime `xorm:"created" json:"created"`
	UpdatedAt  util.MyTime `xorm:"updated" json:"updated"`
}

func (*Slot) TableName() string {
	return "dev_slots"
}

type Tag struct {
	ID              string      `xorm:"pk 'id'" json:"id"`
	SlotID          string      `xorm:"slot_id" json:"slot_id"`
	Name            string      `xorm:"'name'" json:"name"`
	Desc            string      `xorm:"'desc'" json:"desc"`
	Unit            string      `xorm:"'unit'" json:"unit"`       // 数据单位
	Type            string      `xorm:"'type'" json:"type"`       // 标签类型 MEM，IO，CFG
	DataType        string      `xorm:"'dtype'" json:"dtype"`     // 数据类型
	Format          string      `xorm:"'format'" json:"format"`   // 数据格式化
	Address         string      `xorm:"'address'" json:"address"` // 寄存器
	Config          string      `xorm:"'cfg'" json:"cfg"`         // 配置
	Access          int32       `xorm:"'access'" json:"access"`   // 读写数据模式， 1: RW，-1: RO
	Upload          int32       `xorm:"'upload'" json:"upload"`   // 上传数据，1: ON，-1: OFF
	Save            int32       `xorm:"'save'" json:"save"`       // 保存数据，1: ON，-1: OFF
	Visible         int32       `xorm:"'visible'" json:"visible"` // 可见性，1: ON，-1: OFF
	Status          int32       `xorm:"'status'" json:"status"`   // 状态 1: ON，-1: OFF
	Order           int32       `xorm:"'order'" json:"order"`     // 排序
	Version         int32       `xorm:"version" json:"version"`
	Convert         int32       `xorm:"'convert'" json:"convert"` // 量程转换，1: ON，-1: OFF
	ConvertDataType string      `xorm:"'cdtype'" json:"cdtype"`   // 转换后数据类型
	HLimit          float64     `xorm:"'hlimit'" json:"hlimit"`   // 工程量上限
	LLimit          float64     `xorm:"'llimit'" json:"llimit"`   // 工程量下限
	HValue          float64     `xorm:"'hvalue'" json:"hvalue"`   // 量程上限
	LValue          float64     `xorm:"'lvalue'" json:"lvalue"`   // 量程下限
	Value           nson.Value  `xorm:"-" json:"value"`
	DeletedAt       util.MyTime `xorm:"deleted" json:"-"`
	CreatedAt       util.MyTime `xorm:"created" json:"created"`
	UpdatedAt       util.MyTime `xorm:"updated" json:"updated"`
}

func (*Tag) TableName() string {
	return "dev_tags"
}

const DriverMQTT = "MQTT"

const (
	TypeIO  = "IO"
	TypeMEM = "MEM"
	TypeCFG = "CFG"
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

const (
	DefaultSlotName = "default"
)

func (t *Tag) DType() string {
	if t.Convert == consts.ON && IsNumber(t.DataType) {
		return t.ConvertDataType
	}

	return t.DataType
}

func (t *Tag) DefaultValue() nson.Value {
	switch t.DType() {
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

func (t *Tag) ReadConvert() {
	if !IsNumber(t.DataType) {
		return
	}

	if t.Convert == consts.ON {
		// OUT = [(IN - K1)/(K2 - K1)) * (HI_LIM - LO_LIM)] + LO_LIM
		if (t.HLimit-t.LLimit == 0) || (t.HValue-t.LValue == 0) {
			return
		}

		value, ok := util.NsonValueToFloat64(t.Value)
		if !ok {
			return
		}

		out := (value-t.LLimit)/(t.HLimit-t.LLimit)*(t.HValue-t.LValue) + t.LValue

		switch t.ConvertDataType {
		case TypeI8, TypeI16, TypeI32:
			t.Value = nson.I32(out)
		case TypeU8, TypeU16, TypeU32:
			t.Value = nson.U32(out)
		case TypeI64:
			t.Value = nson.I64(out)
		case TypeU64:
			t.Value = nson.U64(out)
		case TypeF32:
			t.Value = nson.F32(out)
		case TypeF64:
			t.Value = nson.F64(out)
		}
	}
}

func (t *Tag) WriteConvert() {
	if !IsNumber(t.DataType) {
		return
	}

	if t.Convert == consts.ON {
		if (t.HLimit-t.LLimit == 0) || (t.HValue-t.LValue == 0) {
			return
		}

		value, ok := util.NsonValueToFloat64(t.Value)
		if !ok {
			return
		}

		out := (value-t.LValue)/(t.HValue-t.LValue)*(t.HLimit-t.LLimit) + t.LLimit

		switch t.DataType {
		case TypeI8, TypeI16, TypeI32:
			t.Value = nson.I32(out)
		case TypeU8, TypeU16, TypeU32:
			t.Value = nson.U32(out)
		case TypeI64:
			t.Value = nson.I64(out)
		case TypeU64:
			t.Value = nson.U64(out)
		case TypeF32:
			t.Value = nson.F32(out)
		case TypeF64:
			t.Value = nson.F64(out)
		}
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

func TypeSize(t string) int {
	switch t {
	case TypeBool, TypeI8, TypeU8:
		return 1
	case TypeI16, TypeU16:
		return 2
	case TypeI32, TypeU32, TypeF32:
		return 4
	case TypeI64, TypeU64, TypeF64:
		return 8
	}

	return 0
}

func IsNumber(t string) bool {
	switch t {
	case TypeI8, TypeI16, TypeI32,
		TypeU8, TypeU16, TypeU32,
		TypeI64, TypeU64,
		TypeF32, TypeF64:
		return true
	}

	return false
}
