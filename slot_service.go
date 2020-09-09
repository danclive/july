package july

import (
	"errors"
	"fmt"

	"github.com/danclive/july/util"
	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
	"xorm.io/xorm/caches"
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
	TagTypeIO  = "IO"
	TagTypeMEM = "MEM"
	TagTypeCFG = "CFG"
)

const (
	ON  = 1
	OFF = -1
)

const (
	RO = 1
	RW = 2
)

const (
	// DriverS7_TCP     = "S7-TCP"
	// DriverMODBUS_TCP = "MODBUS-TCP"
	DriverMQTT = "MQTT"
)

type ListParams struct {
	Start  int    `form:"start" json:"start"`
	Limit  int    `form:"limit" json:"limit"`
	Search string `form:"search" json:"search"`
	SortBy string `form:"sort_by" json:"sort_by"`
	Asc    bool   `form:"asc" json:"asc"`
}

type SlotService struct {
	crate Crate
	XormHelper
}

func newSlotService(crate Crate, dbPath string) (*SlotService, error) {
	engine, err := xorm.NewEngine("sqlite3", fmt.Sprintf("%s?cache=shared&_journal_mode=WAL", dbPath))
	if err != nil {
		return nil, err
	}

	engine.Ping()

	cacher := caches.NewLRUCacher(caches.NewMemoryStore(), 1500)
	engine.SetDefaultCacher(cacher)

	{
		has, err := engine.IsTableExist(Slot{})
		if err != nil {
			return nil, err
		}

		if !has {
			err = engine.Sync2(Slot{})
			if err != nil {
				return nil, err
			}
		}

		has, err = engine.IsTableExist(Tag{})
		if err != nil {
			return nil, err
		}

		if !has {
			err = engine.Sync2(Tag{})
			if err != nil {
				return nil, err
			}
		}
	}

	return &SlotService{crate: crate, XormHelper: NewXormHelper(engine)}, nil
}

func (s *SlotService) CreateSlot(params *Slot) (bool, error) {
	if params.Id == "" {
		params.Id = util.RandomID()
	}

	if params.Name == "" {
		return false, errors.New("插槽名称不能为空")
	}

	if params.Status == 0 {
		params.Status = OFF
	}

	if params.LinkStatus == 0 {
		params.LinkStatus = OFF
	}

	return true, s.Insert(params)
}

func (s *SlotService) UpdateSlot(params *Slot) (bool, error) {
	if params.Name == "" {
		return false, errors.New("插槽名称不能为空")
	}

	err := s.Update(params.Id, params)

	s.crate.Collect().Reset(params.Id)

	return true, err
}

func (s *SlotService) DeleteSlot(params *Slot) error {
	tags, err := s.ListTagSimple(params.Id)
	if err != nil {
		return err
	}

	session := s.engine.NewSession()
	defer session.Close()

	for i := 0; i < len(tags); i++ {

		_, err = session.ID(tags[i].Id).Delete(&tags[i])
		if err != nil {
			session.Rollback()
			return err
		}

		s.crate.DBus().Delete(tags[i].Name)
	}

	_, err = session.ID(params.Id).Delete(params)
	if err != nil {
		session.Rollback()
		return err
	}

	err = session.Commit()

	s.crate.Collect().Reset(params.Id)

	return err
}

func (s *SlotService) ListSlot(params ListParams) ([]Slot, int64, error) {
	items := make([]Slot, 0)

	err := s.engine.Limit(params.Limit, params.Start).Desc("order").Find(&items)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.engine.Count(&Slot{})
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (s *SlotService) ListSlotSimple() ([]Slot, error) {
	items := make([]Slot, 0)

	err := s.engine.Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *SlotService) ListSlotByStatusOn() ([]Slot, error) {
	items := make([]Slot, 0)

	err := s.engine.Where("status = ?", ON).Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *SlotService) GetSlot(id string) (*Slot, error) {
	var item Slot
	has, err := s.GetById(id, &item)
	if err != nil {
		return nil, err
	}

	if has {
		return &item, nil
	}

	return nil, nil
}

func (s *SlotService) SlotOnline(id string) error {
	slot, err := s.GetSlot(id)
	if err != nil {
		return err
	}

	slot.LinkStatus = ON

	return s.Update(slot.Id, slot)
}

func (s *SlotService) SlotOffline(id string) error {
	slot, err := s.GetSlot(id)
	if err != nil {
		return err
	}

	slot.LinkStatus = OFF

	return s.Update(slot.Id, slot)
}

func (s *SlotService) SlotReset(driver string) error {
	items := make([]Slot, 0)
	var err error

	if driver != "" {
		err = s.engine.Where("driver = ?", driver).Find(&items)
	} else {
		err = s.engine.Find(&items)
	}

	if err != nil {
		return err
	}

	for _, item := range items {
		item.LinkStatus = OFF

		err := s.Update(item.Id, item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SlotService) CreateTag(params *Tag) (bool, error) {
	if params.Id == "" {
		params.Id = util.RandomID()
	}

	if params.SlotId == "" {
		return false, errors.New("插槽ID不能为空")
	}

	if params.Name == "" {
		return false, errors.New("标签名称不能为空")
	}

	if params.TagType == "" {
		return false, errors.New("标签类型不能为空")
	}

	if params.DataType == "" {
		return false, errors.New("标签数据类型不能为空")
	}

	if params.Status == 0 {
		params.Status = OFF
	}

	if params.Upload == 0 {
		params.Upload = OFF
	}

	if params.Save == 0 {
		params.Save = OFF
	}

	if params.Visible == 0 {
		params.Visible = OFF
	}

	err := s.Insert(params)

	s.crate.Collect().Reset(params.SlotId)

	return true, err
}

func (s *SlotService) UpdateTag(params *Tag) (bool, error) {
	if params.SlotId == "" {
		return false, errors.New("插槽ID不能为空")
	}

	if params.Name == "" {
		return false, errors.New("标签名称不能为空")
	}

	if params.TagType == "" {
		return false, errors.New("标签类型不能为空")
	}

	if params.DataType == "" {
		return false, errors.New("标签数据类型不能为空")
	}

	err := s.Update(params.Id, params)

	// s.crate.DBus().UpdateTagConfig(params.Name)
	s.crate.DBus().Delete(params.Name)
	s.crate.Collect().Reset(params.SlotId)

	return true, err
}

func (s *SlotService) DeleteTag(params *Tag) error {
	err := s.Delete(params.Id, params)

	s.crate.DBus().Delete(params.Name)
	s.crate.Collect().Reset(params.SlotId)

	return err
}

type ListTagParams struct {
	ListParams
	SlotId string `form:"slot_id" json:"slot_id"`
}

func (s *SlotService) ListTag(params ListTagParams) ([]Tag, int64, error) {
	items := make([]Tag, 0)

	err := s.engine.Where("slot_id = ?", params.SlotId).Limit(params.Limit, params.Start).Desc("order").Find(&items)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.engine.Where("slot_id = ?", params.SlotId).Count(&Tag{})
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (s *SlotService) ListTagByVisableOn(params ListParams) ([]Tag, int64, error) {
	items := make([]Tag, 0)

	err := s.engine.Where("status = ?", ON).Limit(params.Limit, params.Start).Desc("order").Find(&items)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.engine.Where("status = ?", ON).Count(&Tag{})
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (s *SlotService) ListTagSimple(slotId string) ([]Tag, error) {
	items := make([]Tag, 0)

	err := s.engine.Where("slot_id = ?", slotId).Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *SlotService) ListTagByStatusOnAndIO(slotId string) ([]Tag, error) {
	items := make([]Tag, 0)

	err := s.engine.Where("slot_id = ?", slotId).And("tag_type = ?", TagTypeIO).And("status = ?", ON).Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *SlotService) ListTagForHisData() ([]Tag, error) {
	slots, err := s.ListSlotByStatusOn()
	if err != nil {
		return nil, err
	}

	items := make([]Tag, 0)

	for i := 0; i < len(slots); i++ {
		err = s.engine.Where("slot_id = ?", slots[i].Id).And("status = ?", ON).And("save = ?", ON).Find(&items)
		if err != nil {
			return nil, err
		}
	}

	return items, nil
}

func (s *SlotService) ListTagForUpload() ([]Tag, error) {
	slots, err := s.ListSlotByStatusOn()
	if err != nil {
		return nil, err
	}

	items := make([]Tag, 0)

	for i := 0; i < len(slots); i++ {
		err = s.engine.Where("slot_id = ?", slots[i].Id).Where("status = ?", ON).And("upload = ?", ON).Find(&items)
		if err != nil {
			return nil, err
		}
	}

	return items, nil
}

func (s *SlotService) GetTag(id string) (*Tag, error) {
	var item Tag
	has, err := s.GetById(id, &item)
	if err != nil {
		return nil, err
	}

	if has {
		return &item, nil
	}

	return nil, nil
}

func (s *SlotService) GetTagByName(tagName string) (*Tag, error) {
	var item Tag
	has, err := s.GetByName(tagName, &item)
	if err != nil {
		return nil, err
	}

	if has {
		return &item, nil
	}

	return nil, nil
}

func (s *SlotService) GetTagBySlotIdAndAddress(slotId, address string) (*Tag, error) {
	var item Tag
	has, err := s.engine.Where("slot_id = ?", slotId).And("address = ?", address).Get(&item)
	if err != nil {
		return nil, err
	}

	if has {
		return &item, nil
	}

	return nil, nil
}
