package slot

import (
	"github.com/danclive/july/util"
)

type Slot struct {
	Id         string      `xorm:"pk 'id'" json:"id"`
	Name       string      `xorm:"'name'" json:"name"`
	Desc       string      `xorm:"'desc'" json:"desc"`
	Driver     string      `xorm:"'driver'" json:"driver"`           // 驱动 S7-TCP, MODBUS-TCP, MQTT
	Params     string      `xorm:"'params'" json:"params"`           // 参数
	Status     int32       `xorm:"'status'" json:"status"`           // 状态 1: ON，-1: OFF
	LinkStatus int32       `xorm:"'link_status'" json:"link_status"` // 设备连接状态(仅在mqtt等时有用)，1: ON，-1: OFF
	Order      int32       `xorm:"'order'" json:"order"`
	Version    int32       `xorm:"version" json:"version"`
	DeletedAt  util.MyTime `xorm:"deleted" json:"-"`
	CreatedAt  util.MyTime `xorm:"created" json:"created"`
	UpdatedAt  util.MyTime `xorm:"updated" json:"updated"`
}

func (*Slot) TableName() string {
	return "slots"
}
