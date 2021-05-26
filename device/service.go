package device

import (
	"errors"
	"fmt"

	"github.com/danclive/july/log"
	"github.com/danclive/july/util"
	"github.com/danclive/march/consts"
	"xorm.io/xorm"
)

var _service *Service

func InitService(engine *xorm.Engine) {
	s := &Service{Engine: engine}

	if err := s.Sync(true); err != nil {
		log.Suger.Fatal(err)
	}

	_service = s
}

func GetService() *Service {
	return _service
}

type Service struct {
	*xorm.Engine
	collect Collect
}

type Collect interface {
	Reset(slotID string)
}

func (s *Service) Sync(force bool) error {
	var has bool
	var err error

	has, err = s.IsTableExist(Slot{})
	if err != nil {
		return err
	}

	if !has || force {
		err = s.Sync2(Slot{})
		if err != nil {
			return err
		}
	}

	has, err = s.IsTableExist(Tag{})
	if err != nil {
		return err
	}

	if !has || force {
		err = s.Sync2(Tag{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) SetCollect(c Collect) {
	s.collect = c
}

func (s *Service) CreateSlot(params *Slot) (bool, error) {
	if params.ID == "" {
		params.ID = util.RandomID()
	}

	if params.Name == "" {
		return false, errors.New("插槽名称不能为空")
	}

	if params.Status == 0 {
		params.Status = consts.OFF
	}

	if params.LinkStatus == 0 {
		params.LinkStatus = consts.OFF
	}

	if params.Fault == 0 {
		params.Fault = consts.OFF
	}

	if params.Update == 0 {
		params.Update = consts.OFF
	}

	_, err := s.InsertOne(params)

	return true, err
}

func (s *Service) UpdateSlot(params *Slot) (bool, error) {
	_, err := s.ID(params.ID).Update(params)

	if s.collect != nil {
		s.collect.Reset(params.ID)
	}

	return true, err
}

func (s *Service) DeleteSlot(params *Slot) error {
	var err error

	session := s.NewSession()
	defer session.Close()

	_, err = session.ID(params.ID).Delete(&Slot{})
	if err != nil {
		session.Rollback()
		return err
	}

	_, err = session.Delete(&Tag{SlotID: params.ID})
	if err != nil {
		session.Rollback()
		return err
	}

	err = session.Commit()

	if s.collect != nil {
		s.collect.Reset(params.ID)
	}

	return err
}

func (s *Service) DestorySlots() error {
	item := Slot{}
	_, err := s.Engine.Exec(fmt.Sprintf("DELETE FROM %v", item.TableName()))
	s.Engine.ClearCache(&item)
	return err
}

func (s *Service) CreateTag(params *Tag) (bool, error) {
	if params.ID == "" {
		params.ID = util.RandomID()
	}

	if params.SlotID == "" {
		return false, errors.New("插槽 ID 不能为空")
	}

	if params.Name == "" {
		return false, errors.New("标签名称不能为空")
	}

	if params.Type == "" {
		params.Type = TypeMEM
	}

	if params.DataType == "" {
		params.DataType = TypeI32
	}

	if params.Access == 0 {
		params.Access = consts.OFF
	}

	if params.Status == 0 {
		params.Status = consts.OFF
	}

	if params.Upload == 0 {
		params.Upload = consts.OFF
	}

	if params.Save == 0 {
		params.Save = consts.OFF
	}

	if params.Visible == 0 {
		params.Visible = consts.OFF
	}

	_, err := s.InsertOne(params)

	if s.collect != nil {
		s.collect.Reset(params.SlotID)
	}

	return true, err
}

func (s *Service) UpdateTag(params *Tag) (bool, error) {
	_, err := s.ID(params.ID).Update(params)

	if s.collect != nil {
		s.collect.Reset(params.SlotID)
	}

	return true, err
}

func (s *Service) DeleteTag(params *Tag) error {
	_, err := s.ID(params.ID).Delete(&Tag{})

	if s.collect != nil {
		s.collect.Reset(params.SlotID)
	}

	return err
}

func (s *Service) DestoryTags(slotID string) error {
	item := Tag{}
	var err error
	if slotID == "" {
		_, err = s.Engine.Exec(fmt.Sprintf("DELETE FROM %v", item.TableName()))
	} else {
		_, err = s.Engine.Exec(fmt.Sprintf("DELETE FROM %v WHERE slot_id = '%v'", item.TableName(), slotID))
	}

	s.Engine.ClearCache(&item)
	return err
}

// list

func (s *Service) ListSlot() ([]Slot, error) {
	items := make([]Slot, 0)

	err := s.Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *Service) ListSlotStatusOn() ([]Slot, error) {
	items := make([]Slot, 0)

	err := s.Where("status = ?", consts.ON).Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *Service) ListSlot2(params util.ListParams) ([]Slot, int64, error) {
	items := make([]Slot, 0)

	err := s.Limit(params.Limit, params.Start).Desc("order").Find(&items)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.Count(&Slot{})
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (s *Service) ListTag(slotID string) ([]Tag, error) {
	items := make([]Tag, 0)

	err := s.Where("slot_id = ?", slotID).Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *Service) ListTagStatusOnAndTypeIO(slotID string) ([]Tag, error) {
	items := make([]Tag, 0)

	err := s.Where("slot_id = ?", slotID).And("type = ?", TypeIO).And("status = ?", consts.ON).Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *Service) ListTagAndSave() ([]Tag, error) {
	slots, err := s.ListSlotStatusOn()
	if err != nil {
		return nil, err
	}

	items := make([]Tag, 0)

	for i := 0; i < len(slots); i++ {
		err = s.Where("slot_id = ?", slots[i].ID).And("status = ?", consts.ON).And("save = ?", consts.ON).Find(&items)
		if err != nil {
			return nil, err
		}
	}

	return items, nil
}

func (s *Service) ListTagAndUploadBySlot(slotID string) ([]Tag, error) {
	items := make([]Tag, 0)

	err := s.Where("slot_id = ?", slotID).And("status = ?", consts.ON).And("upload = ?", consts.ON).Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *Service) ListTagAndUpload() ([]Tag, error) {
	slots, err := s.ListSlotStatusOn()
	if err != nil {
		return nil, err
	}

	items := make([]Tag, 0)

	for i := 0; i < len(slots); i++ {
		err = s.Where("slot_id = ?", slots[i].ID).And("status = ?", consts.ON).And("upload = ?", consts.ON).Find(&items)
		if err != nil {
			return nil, err
		}
	}

	return items, nil
}

type ListTagParams struct {
	util.ListParams
	SlotID string `form:"slot_id" json:"slot_id"`
}

func (s *Service) ListTag2(params ListTagParams) ([]Tag, int64, error) {
	items := make([]Tag, 0)
	total := int64(0)
	if params.Search != "" {
		err := s.Where("slot_id = ?", params.SlotID).And(fmt.Sprintf("name LIKE '%%%s%%'", params.Search)).Limit(params.Limit, params.Start).Desc("order").Find(&items)
		if err != nil {
			return nil, 0, err
		}

		total, err = s.Where("slot_id = ?", params.SlotID).And(fmt.Sprintf("name LIKE '%%%s%%'", params.Search)).Count(&Tag{})
		if err != nil {
			return nil, 0, err
		}
	} else {
		err := s.Where("slot_id = ?", params.SlotID).Limit(params.Limit, params.Start).Desc("order").Find(&items)
		if err != nil {
			return nil, 0, err
		}

		total, err = s.Where("slot_id = ?", params.SlotID).Count(&Tag{})
		if err != nil {
			return nil, 0, err
		}
	}

	return items, total, nil
}

func (s *Service) ListTagByVisableOn(params util.ListParams) ([]Tag, int64, error) {
	items := make([]Tag, 0)

	err := s.Where("status = ?", consts.ON).Limit(params.Limit, params.Start).Desc("order").Find(&items)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.Where("status = ?", consts.ON).Count(&Tag{})
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// get

func (s *Service) GetSlot(id string) (*Slot, error) {
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

func (s *Service) GetSlotByName(name string) (*Slot, error) {
	var item Slot
	has, err := s.GetByName(name, &item)
	if err != nil {
		return nil, err
	}

	if has {
		return &item, nil
	}

	return nil, nil
}

func (s *Service) GetTag(id string) (*Tag, error) {
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

func (s *Service) GetTagByName(tagName string) (*Tag, error) {
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

func (s *Service) GetTagBySlotIDAndName(slotID, name string) (*Tag, error) {
	var item Tag
	has, err := s.Where("slot_id = ?", slotID).And("name = ?", name).Get(&item)
	if err != nil {
		return nil, err
	}

	if has {
		return &item, nil
	}

	return nil, nil
}

func (s *Service) GetTagBySlotIDAndAddress(slotID, address string) (*Tag, error) {
	var item Tag
	has, err := s.Where("slot_id = ?", slotID).And("address = ?", address).Get(&item)
	if err != nil {
		return nil, err
	}

	if has {
		return &item, nil
	}

	return nil, nil
}

// function

func (s *Service) SlotOnline(id string) error {
	slot, err := s.GetSlot(id)
	if err != nil {
		return err
	}

	if slot == nil {
		return errors.New("插槽不存在")
	}

	slot.LinkStatus = consts.ON

	_, err = s.ID(slot.ID).Update(slot)

	return err
}

func (s *Service) SlotOffline(id string) error {
	slot, err := s.GetSlot(id)
	if err != nil {
		return err
	}

	if slot == nil {
		return errors.New("插槽不存在")
	}

	slot.LinkStatus = consts.OFF

	_, err = s.ID(slot.ID).Update(slot)

	return err
}

func (s *Service) SlotReset(driver string) error {
	items := make([]Slot, 0)
	var err error

	if driver != "" {
		err = s.Where("driver = ?", driver).And("link = ?", consts.ON).Find(&items)
	} else {
		err = s.Where("link = ?", consts.ON).Find(&items)
	}

	if err != nil {
		return err
	}

	for _, item := range items {
		item.LinkStatus = consts.OFF

		_, err := s.Update(item.ID, item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) TagFn(fn string, ids []string) error {
	for _, id := range ids {
		tag, err := s.GetTag(id)
		if err != nil {
			return err
		}

		if tag == nil {
			continue
		}

		switch fn {
		case "on":
			tag.Status = consts.ON
		case "off":
			tag.Status = consts.OFF
		case "on_upload":
			tag.Upload = consts.ON
		case "off_upload":
			tag.Upload = consts.OFF
		case "on_save":
			tag.Save = consts.ON
		case "off_save":
			tag.Save = consts.OFF
		case "ro":
			tag.Access = consts.OFF
		case "rw":
			tag.Access = consts.ON
		case "on_visible":
			tag.Visible = consts.ON
		case "off_visible":
			tag.Visible = consts.OFF
		default:
			return errors.New("not support")
		}

		_, err = s.UpdateTag(tag)
		if err != nil {
			return err
		}
	}

	return nil
}

// helper

func (s *Service) GetById(id interface{}, res interface{}) (has bool, err error) {
	has, err = s.Where("id = ?", id).Get(res)
	return
}

func (s *Service) GetByName(name string, res interface{}) (has bool, err error) {
	has, err = s.Where("name = ?", name).Get(res)
	return
}

func (s *Service) DeleteForce(id interface{}, res interface{}) (err error) {
	_, err = s.ID(id).Unscoped().Delete(res)
	return
}
