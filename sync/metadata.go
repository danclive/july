package sync

import (
	"encoding/json"
	"strings"

	"github.com/danclive/july/dict"
	"github.com/danclive/july/log"
	slotpkg "github.com/danclive/july/slot"
	"github.com/danclive/nson-go"
	"github.com/danclive/queen-go/client"
)

func SyncMetadata(_ *client.Client, recv *client.RecvMessage, back *client.SendMessage) {
	//fmt.Println(recv)

	data, err := recv.Body.GetMessage(dict.DATA)
	if err != nil {
		return
	}

	back2 := syncMetadata(data)

	if back != nil {
		back.Body().Insert(dict.DATA, back2)
	}
}

func syncMetadata(recv nson.Message) nson.Message {
	method, err := recv.GetString(dict.METHOD)
	if err != nil {
		log.Suger.Errorf("recv.GetString(dict.METHOD): %s", err)
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	params, err := recv.GetMessage(dict.PARAMS)
	if err != nil {
		log.Suger.Errorf("recv.GetMessage(dict.PARAMS): %s", err)
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
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
			dict.CODE:  nson.I32(404),
			dict.ERROR: nson.String("Not Found"),
		}
	}
}

func pullSlots(params, recv nson.Message) nson.Message {
	slots, err := slotpkg.GetService().ListSlotSimple()
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

func pullSlot(params, recv nson.Message) nson.Message {
	slotId, err := params.GetString("slotId")
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	slot, err := slotpkg.GetService().GetSlot(slotId)
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

func pullTags(params, recv nson.Message) nson.Message {
	slotId, err := params.GetString("slotId")
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	tags, err := slotpkg.GetService().ListTagSimple(slotId)
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

func pullTag(params, recv nson.Message) nson.Message {
	tagId, err := params.GetString("tagId")
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	tag, err := slotpkg.GetService().GetTag(tagId)
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

func pushSlots(params, recv nson.Message) nson.Message {
	data, err := params.GetBinary("slots")
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	slots := make([]slotpkg.Slot, 0)

	err = json.Unmarshal(data, &slots)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	for i := 0; i < len(slots); i++ {
		slot, err := slotpkg.GetService().GetSlot(slots[i].Id)
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

			_, err := slotpkg.GetService().UpdateSlot(slot)
			if err != nil {
				return nson.Message{
					dict.CODE:  nson.I32(500),
					dict.ERROR: nson.String(err.Error()),
				}
			}

			continue
		}

	NEW:
		slot2 := slotpkg.Slot{
			Id:     slots[i].Id,
			Name:   slots[i].Name,
			Desc:   slots[i].Desc,
			Driver: slots[i].Driver,
			Params: slots[i].Params,
			Status: slots[i].Status,
			Order:  slots[i].Order,
		}

		_, err = slotpkg.GetService().CreateSlot(&slot2)
		if err != nil {
			if strings.HasPrefix(err.Error(), "UNIQUE constraint failed:") {
				err = slotpkg.GetService().DeleteForce(slot2.Id, &slot2)
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

func pushSlot(params, recv nson.Message) nson.Message {
	data, err := params.GetBinary("slot")
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	slot := slotpkg.Slot{}

	err = json.Unmarshal(data, &slot)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	slot2, err := slotpkg.GetService().GetSlot(slot.Id)
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

		_, err := slotpkg.GetService().UpdateSlot(slot2)
		if err != nil {
			return nson.Message{
				dict.CODE:  nson.I32(500),
				dict.ERROR: nson.String(err.Error()),
			}
		}
	} else {

	NEW:
		slot3 := slotpkg.Slot{
			Id:     slot.Id,
			Name:   slot.Name,
			Desc:   slot.Desc,
			Driver: slot.Driver,
			Params: slot.Params,
			Status: slot.Status,
			Order:  slot.Order,
		}

		_, err = slotpkg.GetService().CreateSlot(&slot3)
		if err != nil {
			if strings.HasPrefix(err.Error(), "UNIQUE constraint failed:") {
				err = slotpkg.GetService().DeleteForce(slot3.Id, &slot3)
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

func deleteSlot(params, recv nson.Message) nson.Message {
	slotId, err := params.GetString("slotId")
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	slot, err := slotpkg.GetService().GetSlot(slotId)
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

	err = slotpkg.GetService().DeleteSlot(slot)
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

func pushTags(params, recv nson.Message) nson.Message {
	data, err := params.GetBinary("tags")
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	tags := make([]slotpkg.Tag, 0)

	err = json.Unmarshal(data, &tags)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	for i := 0; i < len(tags); i++ {
		tag, err := slotpkg.GetService().GetTag(tags[i].Id)
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

			_, err := slotpkg.GetService().UpdateTag(tag)
			if err != nil {
				return nson.Message{
					dict.CODE:  nson.I32(500),
					dict.ERROR: nson.String(err.Error()),
				}
			}

			continue
		}

	NEW:
		tag2 := slotpkg.Tag{
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

		_, err = slotpkg.GetService().CreateTag(&tag2)
		if err != nil {
			if strings.HasPrefix(err.Error(), "UNIQUE constraint failed:") {
				err = slotpkg.GetService().DeleteForce(tag2.Id, &tag2)
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

func pushTag(params, recv nson.Message) nson.Message {
	data, err := params.GetBinary("tag")
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	tag := slotpkg.Tag{}
	err = json.Unmarshal(data, &tag)
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(500),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	tag2, err := slotpkg.GetService().GetTag(tag.Id)
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

		_, err := slotpkg.GetService().UpdateTag(tag2)
		if err != nil {
			return nson.Message{
				dict.CODE:  nson.I32(500),
				dict.ERROR: nson.String(err.Error()),
			}
		}
	} else {

	NEW:
		tag3 := slotpkg.Tag{
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

		_, err = slotpkg.GetService().CreateTag(&tag3)
		if err != nil {
			if strings.HasPrefix(err.Error(), "UNIQUE constraint failed:") {
				err = slotpkg.GetService().DeleteForce(tag3.Id, &tag3)
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

func deleteTag(params, recv nson.Message) nson.Message {
	tagId, err := params.GetString("tagId")
	if err != nil {
		return nson.Message{
			dict.CODE:  nson.I32(400),
			dict.ERROR: nson.String(err.Error()),
		}
	}

	tag, err := slotpkg.GetService().GetTag(tagId)
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

	err = slotpkg.GetService().DeleteTag(tag)
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
