package july

import "xorm.io/xorm"

type XormHelper struct {
	engine *xorm.Engine
}

func NewXormHelper(engine *xorm.Engine) XormHelper {
	return XormHelper{engine: engine}
}

func (s *XormHelper) Engine() *xorm.Engine {
	return s.engine
}

func (s *XormHelper) Insert(res interface{}) (err error) {
	_, err = s.engine.InsertOne(res)
	return
}

func (s *XormHelper) Update(id interface{}, res interface{}) (err error) {
	_, err = s.engine.ID(id).Update(res)
	return
}

func (s *XormHelper) Delete(id interface{}, res interface{}) (err error) {
	_, err = s.engine.ID(id).Delete(res)
	return
}

func (s *XormHelper) DeleteForce(id interface{}, res interface{}) (err error) {
	_, err = s.engine.ID(id).Unscoped().Delete(res)
	return
}

func (s *XormHelper) GetById(id interface{}, res interface{}) (has bool, err error) {
	has, err = s.engine.Where("id = ?", id).Get(res)
	return
}

func (s *XormHelper) GetByName(name string, res interface{}) (has bool, err error) {
	has, err = s.engine.Where("name = ?", name).Get(res)
	return
}
