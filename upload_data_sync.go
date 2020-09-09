package july

import (
	"encoding/json"
	"strings"

	"github.com/danclive/july/dict"
	"github.com/danclive/nson-go"
)

func readyDataSync(crate Crate, upload *Upload) {
	crate.Log().Info("readyDataSync")

	upload.LLacer().Llac(dict.DEV_META, func(recv nson.Message) nson.Message {
		method, err := recv.GetString(dict.METHOD)
		if err != nil {
			crate.Log().Errorf("recv.GetString(dict.METHOD): %s", err)
			return nson.Message{
				dict.CODE:  nson.I32(400),
				dict.ERROR: nson.String(err.Error()),
			}
		}

		params, err := recv.GetMessage(dict.PARAMS)
		if err != nil {
			crate.Log().Errorf("recv.GetMessage(dict.PARAMS): %s", err)
			return nson.Message{
				dict.CODE:  nson.I32(400),
				dict.ERROR: nson.String(err.Error()),
			}
		}

		switch method {
		case "PullSlots":
			return pullSlots(crate, params, recv)
		case "PullSlot":
			return pullSlot(crate, params, recv)
		case "PullTags":
			return pullTags(crate, params, recv)
		case "PullTag":
			return pullTag(crate, params, recv)
		case "PushSlots":
			return pushSlots(crate, params, recv)
		case "PushSlot":
			return pushSlot(crate, params, recv)
		case "DeleteSlot":
			return deleteSlot(crate, params, recv)
		case "PushTags":
			return pushTags(crate, params, recv)
		case "PushTag":
			return pushTag(crate, params, recv)
		case "DeleteTag":
			return deleteTag(crate, params, recv)
		default:
			return nson.Message{
				dict.CODE:  nson.I32(404),
				dict.ERROR: nson.String("Not Found"),
			}
		}
	})
}

func pullSlots(crate Crate, params, recv nson.Message) nson.Message {
	slots, err := crate.SlotService().ListSlotSimple()
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	bytes, err := json.Marshal(slots)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	return nson.Message{
		dict.CODE: nson.I32(0),
		dict.DATA: nson.Binary(bytes),
	}
}

func pullSlot(crate Crate, params, recv nson.Message) nson.Message {
	slotId, err := params.GetString("slotId")
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	slot, err := crate.SlotService().GetSlot(slotId)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	if slot == nil {
		return nson.Message{
			dict.CODE:  nson.I32(404),
			dict.ERROR: nson.String("Not Found"),
		}
	}

	bytes, err := json.Marshal(&slot)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	return nson.Message{
		dict.CODE: nson.I32(0),
		dict.DATA: nson.Binary(bytes),
	}
}

func pullTags(crate Crate, params, recv nson.Message) nson.Message {
	slotId, err := params.GetString("slotId")
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	tags, err := crate.SlotService().ListTagSimple(slotId)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	bytes, err := json.Marshal(tags)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	return nson.Message{
		dict.CODE: nson.I32(0),
		dict.DATA: nson.Binary(bytes),
	}
}

func pullTag(crate Crate, params, recv nson.Message) nson.Message {
	tagId, err := params.GetString("tagId")
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	tag, err := crate.SlotService().GetTag(tagId)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	if tag == nil {
		return nson.Message{
			dict.CODE:  nson.I32(404),
			dict.ERROR: nson.String("Not Found"),
		}
	}

	bytes, err := json.Marshal(&tag)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	return nson.Message{
		dict.CODE: nson.I32(0),
		dict.DATA: nson.Binary(bytes),
	}
}

func pushSlots(crate Crate, params, recv nson.Message) nson.Message {
	data, err := params.GetBinary("slots")
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	slots := make([]Slot, 0)

	err = json.Unmarshal(data, &slots)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	for i := 0; i < len(slots); i++ {
		slot, err := crate.SlotService().GetSlot(slots[i].Id)
		if err != nil {
			return nson.Message{
				dict.CODE:  nson.I32(500),
				dict.ERROR: nson.String(err.Error()),
			}
		}

		if slot != nil {
			slot.Name = slots[i].Name
			slot.Desc = slots[i].Desc
			slot.Driver = slots[i].Driver
			slot.Params = slots[i].Params
			slot.Status = slots[i].Status
			slot.Order = slots[i].Order

			_, err := crate.SlotService().UpdateSlot(slot)
			if err != nil {
				return nson.Message{
					dict.CODE:  nson.I32(500),
					dict.ERROR: nson.String(err.Error()),
				}
			}

			continue
		}

	NEW:
		slot2 := Slot{
			Id:     slots[i].Id,
			Name:   slots[i].Name,
			Desc:   slots[i].Desc,
			Driver: slots[i].Driver,
			Params: slots[i].Params,
			Status: slots[i].Status,
			Order:  slots[i].Order,
		}

		_, err = crate.SlotService().CreateSlot(&slot2)
		if err != nil {
			if strings.HasPrefix(err.Error(), "UNIQUE constraint failed:") {
				err = crate.SlotService().DeleteForce(slot2.Id, &slot2)
				if err != nil {
					return nson.Message{
						dict.CODE:  nson.I32(500),
						dict.ERROR: nson.String(err.Error()),
					}
				}

				goto NEW
			}

			return nson.Message{
				dict.CODE:  nson.I32(500),
				dict.ERROR: nson.String(err.Error()),
			}
		}
	}

	return nson.Message{
		dict.CODE: nson.I32(0),
	}
}

func pushSlot(crate Crate, params, recv nson.Message) nson.Message {
	data, err := params.GetBinary("slot")
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	slot := Slot{}

	err = json.Unmarshal(data, &slot)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	slot2, err := crate.SlotService().GetSlot(slot.Id)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	if slot2 != nil {
		slot2.Name = slot.Name
		slot2.Desc = slot.Desc
		slot2.Driver = slot.Driver
		slot2.Params = slot.Params
		slot2.Status = slot.Status
		slot2.Order = slot.Order

		_, err := crate.SlotService().UpdateSlot(slot2)
		if err != nil {
			return nson.Message{
				dict.CODE:  nson.I32(500),
				dict.ERROR: nson.String(err.Error()),
			}
		}
	} else {

	NEW:
		slot3 := Slot{
			Id:     slot.Id,
			Name:   slot.Name,
			Desc:   slot.Desc,
			Driver: slot.Driver,
			Params: slot.Params,
			Status: slot.Status,
			Order:  slot.Order,
		}

		_, err = crate.SlotService().CreateSlot(&slot3)
		if err != nil {
			if strings.HasPrefix(err.Error(), "UNIQUE constraint failed:") {
				err = crate.SlotService().DeleteForce(slot3.Id, &slot3)
				if err != nil {
					return nson.Message{
						dict.CODE:  nson.I32(500),
						dict.ERROR: nson.String(err.Error()),
					}
				}

				goto NEW
			}

			return nson.Message{
				dict.CODE:  nson.I32(500),
				dict.ERROR: nson.String(err.Error()),
			}
		}
	}

	return nson.Message{
		dict.CODE: nson.I32(0),
	}
}

func deleteSlot(crate Crate, params, recv nson.Message) nson.Message {
	slotId, err := params.GetString("slotId")
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	slot, err := crate.SlotService().GetSlot(slotId)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	if slot == nil {
		return nson.Message{
			dict.CODE:  nson.I32(404),
			dict.ERROR: nson.String("Not Found"),
		}
	}

	err = crate.SlotService().DeleteSlot(slot)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	return nson.Message{
		dict.CODE: nson.I32(0),
	}
}

func pushTags(crate Crate, params, recv nson.Message) nson.Message {
	data, err := params.GetBinary("tags")
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	tags := make([]Tag, 0)

	err = json.Unmarshal(data, &tags)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	for i := 0; i < len(tags); i++ {
		tag, err := crate.SlotService().GetTag(tags[i].Id)
		if err != nil {
			return nson.Message{
				dict.CODE:  nson.I32(500),
				dict.ERROR: nson.String(err.Error()),
			}
		}

		if tag != nil {
			tag.Name = tags[i].Name
			tag.Desc = tags[i].Desc
			tag.Unit = tags[i].Unit
			tag.TagType = tags[i].TagType
			tag.DataType = tags[i].DataType
			tag.Format = tags[i].Format
			tag.Address = tags[i].Address
			tag.AccessMode = tags[i].AccessMode
			tag.Upload = tags[i].Upload
			tag.Save = tags[i].Save
			tag.Visible = tags[i].Visible
			tag.Status = tags[i].Status
			tag.Order = tags[i].Order

			_, err := crate.SlotService().UpdateTag(tag)
			if err != nil {
				return nson.Message{
					dict.CODE:  nson.I32(500),
					dict.ERROR: nson.String(err.Error()),
				}
			}

			continue
		}

	NEW:
		tag2 := Tag{
			Id:         tags[i].Id,
			SlotId:     tags[i].SlotId,
			Name:       tags[i].Name,
			Desc:       tags[i].Desc,
			Unit:       tags[i].Unit,
			TagType:    tags[i].TagType,
			DataType:   tags[i].DataType,
			Format:     tags[i].Format,
			Address:    tags[i].Address,
			AccessMode: tags[i].AccessMode,
			Upload:     tags[i].Upload,
			Save:       tags[i].Save,
			Visible:    tags[i].Visible,
			Status:     tags[i].Status,
			Order:      tags[i].Order,
		}

		_, err = crate.SlotService().CreateTag(&tag2)
		if err != nil {
			if strings.HasPrefix(err.Error(), "UNIQUE constraint failed:") {
				err = crate.SlotService().DeleteForce(tag2.Id, &tag2)
				if err != nil {
					return nson.Message{
						dict.CODE:  nson.I32(500),
						dict.ERROR: nson.String(err.Error()),
					}
				}

				goto NEW
			}

			return nson.Message{
				dict.CODE:  nson.I32(500),
				dict.ERROR: nson.String(err.Error()),
			}
		}
	}

	return nson.Message{
		dict.CODE: nson.I32(0),
	}
}

func pushTag(crate Crate, params, recv nson.Message) nson.Message {
	data, err := params.GetBinary("tag")
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	tag := Tag{}
	err = json.Unmarshal(data, &tag)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	tag2, err := crate.SlotService().GetTag(tag.Id)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	if tag2 != nil {
		tag2.Name = tag.Name
		tag2.Desc = tag.Desc
		tag2.Unit = tag.Unit
		tag2.TagType = tag.TagType
		tag2.DataType = tag.DataType
		tag2.Format = tag.Format
		tag2.Address = tag.Address
		tag2.AccessMode = tag.AccessMode
		tag2.Upload = tag.Upload
		tag2.Save = tag.Save
		tag2.Visible = tag.Visible
		tag2.Status = tag.Status
		tag2.Order = tag.Order

		_, err := crate.SlotService().UpdateTag(tag2)
		if err != nil {
			return nson.Message{
				dict.CODE:  nson.I32(500),
				dict.ERROR: nson.String(err.Error()),
			}
		}
	} else {

	NEW:
		tag3 := Tag{
			Id:         tag.Id,
			SlotId:     tag.SlotId,
			Name:       tag.Name,
			Desc:       tag.Desc,
			Unit:       tag.Unit,
			TagType:    tag.TagType,
			DataType:   tag.DataType,
			Format:     tag.Format,
			Address:    tag.Address,
			AccessMode: tag.AccessMode,
			Upload:     tag.Upload,
			Save:       tag.Save,
			Visible:    tag.Visible,
			Status:     tag.Status,
			Order:      tag.Order,
		}

		_, err = crate.SlotService().CreateTag(&tag3)
		if err != nil {
			if strings.HasPrefix(err.Error(), "UNIQUE constraint failed:") {
				err = crate.SlotService().DeleteForce(tag3.Id, &tag3)
				if err != nil {
					return nson.Message{
						dict.CODE:  nson.I32(500),
						dict.ERROR: nson.String(err.Error()),
					}
				}

				goto NEW
			}

			return nson.Message{
				dict.CODE:  nson.I32(500),
				dict.ERROR: nson.String(err.Error()),
			}
		}
	}

	return nson.Message{
		dict.CODE: nson.I32(0),
	}
}

func deleteTag(crate Crate, params, recv nson.Message) nson.Message {
	tagId, err := params.GetString("tagId")
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	tag, err := crate.SlotService().GetTag(tagId)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	if tag == nil {
		return nson.Message{
			dict.CODE:  nson.I32(404),
			dict.ERROR: nson.String("Not Found"),
		}
	}

	err = crate.SlotService().DeleteTag(tag)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	return nson.Message{
		dict.CODE: nson.I32(0),
	}
}
