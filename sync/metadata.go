package sync

import (
	"encoding/json"
	"strings"

	"github.com/danclive/july/device"
	"github.com/danclive/july/log"
	"github.com/danclive/march/consts"
	"github.com/danclive/nson-go"
	"github.com/danclive/queen-go/client"
)

func SyncMetadata(_ *client.Client, recv *client.RecvMessage, back *client.SendMessage) {
	//fmt.Println(recv)

	data, err := recv.Body.GetMessage(consts.DATA)
	if err != nil {
		return
	}

	back2 := syncMetadata(data)

	if back != nil {
		back.Body().Insert(consts.DATA, back2)
	}
}

func syncMetadata(recv nson.Message) nson.Message {
	method, err := recv.GetString(consts.METHOD)
	if err != nil {
		log.Suger.Errorf("recv.GetString(consts.METHOD): %s", err)
		return nson.Message{
			consts.CODE:  nson.I32(400),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	params, err := recv.GetMessage(consts.PARAMS)
	if err != nil {
		log.Suger.Errorf("recv.GetMessage(consts.PARAMS): %s", err)
		return nson.Message{
			consts.CODE:  nson.I32(400),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	switch method {
	case "PullSlots":
		return pullSlots(params, recv)
	case "PullSlot":
		return pullSlot(params, recv)
	case "PullTags":
		return pullTags(params, recv)
	case "PullTag":
		return pullTag(params, recv)
	case "PushSlots":
		return pushSlots(params, recv)
	case "PushSlot":
		return pushSlot(params, recv)
	case "DeleteSlot":
		return deleteSlot(params, recv)
	case "PushTags":
		return pushTags(params, recv)
	case "PushTag":
		return pushTag(params, recv)
	case "DeleteTag":
		return deleteTag(params, recv)
	default:
		return nson.Message{
			consts.CODE:  nson.I32(404),
			consts.ERROR: nson.String("Not Found"),
		}
	}
}

func pullSlots(params, recv nson.Message) nson.Message {
	slots, err := device.GetService().ListSlot()
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(500),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	bytes, err := json.Marshal(slots)
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(500),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	return nson.Message{
		consts.CODE: nson.I32(0),
		consts.DATA: nson.Binary(bytes),
	}
}

func pullSlot(params, recv nson.Message) nson.Message {
	slotID, err := params.GetString("slotID")
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(400),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	slot, err := device.GetService().GetSlot(slotID)
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(500),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	if slot == nil {
		return nson.Message{
			consts.CODE:  nson.I32(404),
			consts.ERROR: nson.String("Not Found"),
		}
	}

	bytes, err := json.Marshal(&slot)
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(500),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	return nson.Message{
		consts.CODE: nson.I32(0),
		consts.DATA: nson.Binary(bytes),
	}
}

func pullTags(params, recv nson.Message) nson.Message {
	slotID, err := params.GetString("slotID")
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(400),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	tags, err := device.GetService().ListTag(slotID)
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(500),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	bytes, err := json.Marshal(tags)
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(500),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	return nson.Message{
		consts.CODE: nson.I32(0),
		consts.DATA: nson.Binary(bytes),
	}
}

func pullTag(params, recv nson.Message) nson.Message {
	tagID, err := params.GetString("tagID")
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(400),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	tag, err := device.GetService().GetTag(tagID)
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(500),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	if tag == nil {
		return nson.Message{
			consts.CODE:  nson.I32(404),
			consts.ERROR: nson.String("Not Found"),
		}
	}

	bytes, err := json.Marshal(&tag)
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(500),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	return nson.Message{
		consts.CODE: nson.I32(0),
		consts.DATA: nson.Binary(bytes),
	}
}

func pushSlots(params, recv nson.Message) nson.Message {
	data, err := params.GetBinary("slots")
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(400),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	slots := make([]device.Slot, 0)

	err = json.Unmarshal(data, &slots)
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(500),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	for i := 0; i < len(slots); i++ {
		slot, err := device.GetService().GetSlot(slots[i].ID)
		if err != nil {
			return nson.Message{
				consts.CODE:  nson.I32(500),
				consts.ERROR: nson.String(err.Error()),
			}
		}

		if slot != nil {
			slot.Name = slots[i].Name
			slot.Desc = slots[i].Desc
			slot.Model = slots[i].Model
			slot.Driver = slots[i].Driver
			slot.Params = slots[i].Params
			slot.Config = slots[i].Config
			slot.ConfigFile = slots[i].ConfigFile
			slot.Status = slots[i].Status
			slot.Order = slots[i].Order

			_, err := device.GetService().UpdateSlot(slot)
			if err != nil {
				return nson.Message{
					consts.CODE:  nson.I32(500),
					consts.ERROR: nson.String(err.Error()),
				}
			}

			continue
		}

	NEW:
		slot2 := device.Slot{
			ID:         slots[i].ID,
			Name:       slots[i].Name,
			Desc:       slots[i].Desc,
			Model:      slots[i].Model,
			Driver:     slots[i].Driver,
			Params:     slots[i].Params,
			Config:     slots[i].Config,
			ConfigFile: slots[i].ConfigFile,
			Status:     slots[i].Status,
			Order:      slots[i].Order,
		}

		_, err = device.GetService().CreateSlot(&slot2)
		if err != nil {
			if strings.HasPrefix(err.Error(), "UNIQUE constraint failed:") {
				err = device.GetService().DeleteForce(slot2.ID, &slot2)
				if err != nil {
					return nson.Message{
						consts.CODE:  nson.I32(500),
						consts.ERROR: nson.String(err.Error()),
					}
				}

				goto NEW
			}

			return nson.Message{
				consts.CODE:  nson.I32(500),
				consts.ERROR: nson.String(err.Error()),
			}
		}
	}

	return nson.Message{
		consts.CODE: nson.I32(0),
	}
}

func pushSlot(params, recv nson.Message) nson.Message {
	data, err := params.GetBinary("slot")
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(400),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	slot := device.Slot{}

	err = json.Unmarshal(data, &slot)
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(500),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	slot2, err := device.GetService().GetSlot(slot.ID)
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(500),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	if slot2 != nil {
		slot2.Name = slot.Name
		slot2.Desc = slot.Desc
		slot2.Model = slot.Model
		slot2.Driver = slot.Driver
		slot2.Params = slot.Params
		slot2.Config = slot.Config
		slot2.ConfigFile = slot.ConfigFile
		slot2.Status = slot.Status
		slot2.Order = slot.Order

		_, err := device.GetService().UpdateSlot(slot2)
		if err != nil {
			return nson.Message{
				consts.CODE:  nson.I32(500),
				consts.ERROR: nson.String(err.Error()),
			}
		}
	} else {

	NEW:
		slot3 := device.Slot{
			ID:         slot.ID,
			Name:       slot.Name,
			Desc:       slot.Desc,
			Model:      slot.Model,
			Driver:     slot.Driver,
			Params:     slot.Params,
			Config:     slot.Config,
			ConfigFile: slot.ConfigFile,
			Status:     slot.Status,
			Order:      slot.Order,
		}

		_, err = device.GetService().CreateSlot(&slot3)
		if err != nil {
			if strings.HasPrefix(err.Error(), "UNIQUE constraint failed:") {
				err = device.GetService().DeleteForce(slot3.ID, &slot3)
				if err != nil {
					return nson.Message{
						consts.CODE:  nson.I32(500),
						consts.ERROR: nson.String(err.Error()),
					}
				}

				goto NEW
			}

			return nson.Message{
				consts.CODE:  nson.I32(500),
				consts.ERROR: nson.String(err.Error()),
			}
		}
	}

	return nson.Message{
		consts.CODE: nson.I32(0),
	}
}

func deleteSlot(params, recv nson.Message) nson.Message {
	slotID, err := params.GetString("slotID")
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(400),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	slot, err := device.GetService().GetSlot(slotID)
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(500),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	if slot == nil {
		return nson.Message{
			consts.CODE:  nson.I32(404),
			consts.ERROR: nson.String("Not Found"),
		}
	}

	err = device.GetService().DeleteSlot(slot)
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(500),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	return nson.Message{
		consts.CODE: nson.I32(0),
	}
}

func pushTags(params, recv nson.Message) nson.Message {
	data, err := params.GetBinary("tags")
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(400),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	tags := make([]device.Tag, 0)

	err = json.Unmarshal(data, &tags)
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(500),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	for i := 0; i < len(tags); i++ {
		tag, err := device.GetService().GetTag(tags[i].ID)
		if err != nil {
			return nson.Message{
				consts.CODE:  nson.I32(500),
				consts.ERROR: nson.String(err.Error()),
			}
		}

		if tag != nil {
			tag.Name = tags[i].Name
			tag.Desc = tags[i].Desc
			tag.Unit = tags[i].Unit
			tag.Type = tags[i].Type
			tag.DataType = tags[i].DataType
			tag.Format = tags[i].Format
			tag.Address = tags[i].Address
			tag.Config = tags[i].Config
			tag.Access = tags[i].Access
			tag.Upload = tags[i].Upload
			tag.Save = tags[i].Save
			tag.Visible = tags[i].Visible
			tag.Status = tags[i].Status
			tag.Order = tags[i].Order

			_, err := device.GetService().UpdateTag(tag)
			if err != nil {
				return nson.Message{
					consts.CODE:  nson.I32(500),
					consts.ERROR: nson.String(err.Error()),
				}
			}

			continue
		}

	NEW:
		tag2 := device.Tag{
			ID:       tags[i].ID,
			SlotID:   tags[i].SlotID,
			Name:     tags[i].Name,
			Desc:     tags[i].Desc,
			Unit:     tags[i].Unit,
			Type:     tags[i].Type,
			DataType: tags[i].DataType,
			Format:   tags[i].Format,
			Address:  tags[i].Address,
			Config:   tags[i].Config,
			Access:   tags[i].Access,
			Upload:   tags[i].Upload,
			Save:     tags[i].Save,
			Visible:  tags[i].Visible,
			Status:   tags[i].Status,
			Order:    tags[i].Order,
		}

		_, err = device.GetService().CreateTag(&tag2)
		if err != nil {
			if strings.HasPrefix(err.Error(), "UNIQUE constraint failed:") {
				err = device.GetService().DeleteForce(tag2.ID, &tag2)
				if err != nil {
					return nson.Message{
						consts.CODE:  nson.I32(500),
						consts.ERROR: nson.String(err.Error()),
					}
				}

				goto NEW
			}

			return nson.Message{
				consts.CODE:  nson.I32(500),
				consts.ERROR: nson.String(err.Error()),
			}
		}
	}

	return nson.Message{
		consts.CODE: nson.I32(0),
	}
}

func pushTag(params, recv nson.Message) nson.Message {
	data, err := params.GetBinary("tag")
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(400),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	tag := device.Tag{}
	err = json.Unmarshal(data, &tag)
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(500),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	tag2, err := device.GetService().GetTag(tag.ID)
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(500),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	if tag2 != nil {
		tag2.Name = tag.Name
		tag2.Desc = tag.Desc
		tag2.Unit = tag.Unit
		tag2.Type = tag.Type
		tag2.DataType = tag.DataType
		tag2.Format = tag.Format
		tag2.Address = tag.Address
		tag2.Config = tag.Config
		tag2.Access = tag.Access
		tag2.Upload = tag.Upload
		tag2.Save = tag.Save
		tag2.Visible = tag.Visible
		tag2.Status = tag.Status
		tag2.Order = tag.Order

		_, err := device.GetService().UpdateTag(tag2)
		if err != nil {
			return nson.Message{
				consts.CODE:  nson.I32(500),
				consts.ERROR: nson.String(err.Error()),
			}
		}
	} else {

	NEW:
		tag3 := device.Tag{
			ID:       tag.ID,
			SlotID:   tag.SlotID,
			Name:     tag.Name,
			Desc:     tag.Desc,
			Unit:     tag.Unit,
			Type:     tag.Type,
			DataType: tag.DataType,
			Format:   tag.Format,
			Address:  tag.Address,
			Config:   tag.Config,
			Access:   tag.Access,
			Upload:   tag.Upload,
			Save:     tag.Save,
			Visible:  tag.Visible,
			Status:   tag.Status,
			Order:    tag.Order,
		}

		_, err = device.GetService().CreateTag(&tag3)
		if err != nil {
			if strings.HasPrefix(err.Error(), "UNIQUE constraint failed:") {
				err = device.GetService().DeleteForce(tag3.ID, &tag3)
				if err != nil {
					return nson.Message{
						consts.CODE:  nson.I32(500),
						consts.ERROR: nson.String(err.Error()),
					}
				}

				goto NEW
			}

			return nson.Message{
				consts.CODE:  nson.I32(500),
				consts.ERROR: nson.String(err.Error()),
			}
		}
	}

	return nson.Message{
		consts.CODE: nson.I32(0),
	}
}

func deleteTag(params, recv nson.Message) nson.Message {
	tagID, err := params.GetString("tagID")
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(400),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	tag, err := device.GetService().GetTag(tagID)
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(500),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	if tag == nil {
		return nson.Message{
			consts.CODE:  nson.I32(404),
			consts.ERROR: nson.String("Not Found"),
		}
	}

	err = device.GetService().DeleteTag(tag)
	if err != nil {
		return nson.Message{
			consts.CODE:  nson.I32(500),
			consts.ERROR: nson.String(err.Error()),
		}
	}

	return nson.Message{
		consts.CODE: nson.I32(0),
	}
}
