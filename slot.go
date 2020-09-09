package july

import (
	"encoding/json"
	"time"

	"github.com/danclive/july/util"
)

type Slot struct {
	Id         string    `xorm:"pk 'id'" json:"id"`
	DeletedAt  time.Time `xorm:"deleted" json:"-"`
	CreatedAt  time.Time `xorm:"created" json:"created"`
	UpdatedAt  time.Time `xorm:"updated" json:"updated"`
	Name       string    `xorm:"'name'" json:"name"`
	Desc       string    `xorm:"'desc'" json:"desc"`
	Driver     string    `xorm:"'driver'" json:"driver"`         // 驱动 S7-TCP, MODBUS-TCP, MQTT
	Params     string    `xorm:"'params'" json:"params"`         // 参数
	Status     int       `xorm:"'status'" json:"status"`         // 状态 1: ON，-1: OFF
	LinkStatus int       `xorm:"link_status" json:"link_status"` // 设备连接状态(仅在mqtt等时有用)，1: ON，-1: OFF
	Order      int       `xorm:"'order'" json:"order"`
	Version    int       `xorm:"version" json:"version"`
}

func (*Slot) TableName() string {
	return "slots"
}

func (s *Slot) MarshalJSON() ([]byte, error) {
	createdAt := ""
	if !s.CreatedAt.IsZero() {
		createdAt = util.TimeFormat(s.CreatedAt)
	}

	updatedAt := ""
	if !s.UpdatedAt.IsZero() {
		updatedAt = util.TimeFormat(s.UpdatedAt)
	}

	tmpItem := struct {
		Slot
		CreatedAt string `json:"created"`
		UpdatedAt string `json:"updated"`
	}{
		Slot:      *s,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	return json.Marshal(tmpItem)
}
