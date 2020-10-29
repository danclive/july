package device

import (
	"errors"

	"github.com/danclive/july/consts"
	"github.com/danclive/july/log"
	"github.com/danclive/july/util"
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

func (s *Service) CreateSlot(params *Slot) (bool, error) {
	if params.Id == "" {
		params.Id = util.RandomID()
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

	_, err := s.InsertOne(params)

	return true, err
}

func (s *Service) UpdateSlot(params *Slot) (bool, error) {
	_, err := s.ID(params.Id).Update(params)

	// GetCollect().Reset(params.Id)

	return true, err
}

func (s *Service) DeleteSlot(params *Slot) error {
	var err error

	session := s.NewSession()
	defer session.Close()

	_, err = session.ID(params.Id).Delete(params)
	if err != nil {
		session.Rollback()
		return err
	}

	_, err = session.Delete(&Tag{SlotId: params.Id})
	if err != nil {
		session.Rollback()
		return err
	}

	err = session.Commit()

	// GetCollect().Reset(params.Id)

	return err
}

func (s *Service) CreateTag(params *Tag) (bool, error) {
	if params.Id == "" {
		params.Id = util.RandomID()
	}

	if params.SlotId == "" {
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

	if params.AccessMode == 0 {
		params.AccessMode = consts.RO
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

	// GetCollect().Reset(params.SlotId)

	return true, err
}

func (s *Service) UpdateTag(params *Tag) (bool, error) {
	_, err := s.ID(params.Id).Update(params)

	// GetCache().Delete(params.Name)
	// GetCollect().Reset(params.SlotId)

	return true, err
}

func (s *Service) DeleteTag(params *Tag) error {
	_, err := s.ID(params.Id).Delete(params)

	// GetCache().Delete(params.Name)
	// GetCollect().Reset(params.SlotId)

	return err
}

// list

func (s *Service) ListSlotSimple() ([]Slot, error) {
	items := make([]Slot, 0)

	err := s.Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *Service) ListSlotSimpleStatusOn() ([]Slot, error) {
	items := make([]Slot, 0)

	err := s.Where("status = ?", consts.ON).Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *Service) ListSlot(params util.ListParams) ([]Slot, int64, error) {
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

func (s *Service) ListTagSimple(slotId string) ([]Tag, error) {
	items := make([]Tag, 0)

	err := s.Where("slot_id = ?", slotId).Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *Service) ListTagSimpleStatusOnAndIO(slotId string) ([]Tag, error) {
	items := make([]Tag, 0)

	err := s.Where("slot_id = ?", slotId).And("type = ?", TypeIO).And("status = ?", consts.ON).Find(&items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *Service) ListTagForHisData() ([]Tag, error) {
	slots, err := s.ListSlotSimpleStatusOn()
	if err != nil {
		return nil, err
	}

	items := make([]Tag, 0)

	for i := 0; i < len(slots); i++ {
		err = s.Where("slot_id = ?", slots[i].Id).And("status = ?", consts.ON).And("save = ?", consts.ON).Find(&items)
		if err != nil {
			return nil, err
		}
	}

	return items, nil
}

func (s *Service) ListTagForUpload() ([]Tag, error) {
	slots, err := s.ListSlotSimpleStatusOn()
	if err != nil {
		return nil, err
	}

	items := make([]Tag, 0)

	for i := 0; i < len(slots); i++ {
		err = s.Where("slot_id = ?", slots[i].Id).And("status = ?", consts.ON).And("upload = ?", consts.ON).Find(&items)
		if err != nil {
			return nil, err
		}
	}

	return items, nil
}

type ListTagParams struct {
	util.ListParams
	SlotId string `form:"slot_id" json:"slot_id"`
}

func (s *Service) ListTag(params ListTagParams) ([]Tag, int64, error) {
	items := make([]Tag, 0)

	err := s.Where("slot_id = ?", params.SlotId).Limit(params.Limit, params.Start).Desc("order").Find(&items)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.Where("slot_id = ?", params.SlotId).Count(&Tag{})
	if err != nil {
		return nil, 0, err
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

func (s *Service) GetTagBySlotIdAndName(slotId, name string) (*Tag, error) {
	var item Tag
	has, err := s.Where("slot_id = ?", slotId).And("name = ?", name).Get(&item)
	if err != nil {
		return nil, err
	}

	if has {
		return &item, nil
	}

	return nil, nil
}

func (s *Service) GetTagBySlotIdAndAddress(slotId, address string) (*Tag, error) {
	var item Tag
	has, err := s.Where("slot_id = ?", slotId).And("address = ?", address).Get(&item)
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

	_, err = s.ID(slot.Id).Update(slot)

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

	_, err = s.ID(slot.Id).Update(slot)

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

		_, err := s.Update(item.Id, item)
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
