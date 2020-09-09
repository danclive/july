package july

import (
	"bytes"
	"encoding/json"

	"github.com/danclive/july/dict"
	"github.com/danclive/july/util"
	"github.com/danclive/mqtt"
	"github.com/danclive/mqtt/packets"
	"github.com/danclive/nson-go"
)

type MqttRecv struct {
	crate  Crate
	server mqtt.Server
}

var _ mqtt.Plugin = &MqttRecv{}

func (m *MqttRecv) Name() string {
	return "mqtt_recv"
}

func (m *MqttRecv) Load(server mqtt.Server) error {
	m.server = server
	return nil
}

func (m *MqttRecv) Unload() error {
	return nil
}

func (m *MqttRecv) Hooks() mqtt.Hooks {
	return mqtt.Hooks{
		OnMsgArrived: m.msgArrived,
	}
}

func (m *MqttRecv) msgArrived(client mqtt.Client, msg packets.Message) (valid bool) {
	clientId := client.OptionsReader().ClientID()

	// suger.Debug(
	// 	"msgArrived",
	// 	zap.String("clientId", clientId),
	// 	zap.String("topic", msg.Topic()),
	// )

	if msg.Topic() == dict.DEV_DATA {
		reader := bytes.NewBuffer(msg.Payload())

		flags, err := util.ReadUint16(reader)
		if err != nil {
			suger.Errorf("read flags: %s", err)
		}

		client.Set("FLAGS", nson.U32(flags))

		if flags == 0 {
			message := map[string]interface{}{}

			err = json.Unmarshal(reader.Bytes(), &message)
			if err != nil {
				suger.Errorf("json.Unmarshal(reader.Bytes(): %s", err)
				return
			}

			if data, ok := message[dict.DATA]; ok {
				if dataMessage, ok := data.(map[string]interface{}); ok {
					// fmt.Println(dataMessage)

					// 查询设备
					s, err := m.crate.SlotService().GetSlot(clientId)
					if err != nil {
						suger.Errorf("SlotService{}.GetSlot(clientId): %s", err)
						return
					}

					if s.Status != ON {
						return
					}

					for k, v := range dataMessage {
						// tag, err := m.crate.SlotService().GetTagByName(k)
						tag, err := m.crate.SlotService().GetTagBySlotIdAndAddress(clientId, k)
						if err != nil {
							suger.Errorf("SlotService{}.GetTagByName(k): %s", err)
							return
						}

						if tag == nil {
							continue
						}

						if tag.Status != ON || tag.TagType != TagTypeIO {
							continue
						}

						var value2 nson.Value

						switch tag.DataType {
						case TypeI8, TypeI16, TypeI32:
							if value, ok := v.(float64); ok {
								value2 = nson.I32(value)
							}
						case TypeU8, TypeU16, TypeU32:
							if value, ok := v.(float64); ok {
								value2 = nson.U32(value)
							}
						case TypeI64:
							if value, ok := v.(float64); ok {
								value2 = nson.I64(value)
							}
						case TypeU64:
							if value, ok := v.(float64); ok {
								value2 = nson.U64(value)
							}
						case TypeF32:
							if value, ok := v.(float64); ok {
								value2 = nson.F32(value)
							}
						case TypeF64:
							if value, ok := v.(float64); ok {
								value2 = nson.F64(value)
							}
						case TypeBool:
							if value, ok := v.(bool); ok {
								value2 = nson.Bool(value)
							}
						case TypeString:
							if value, ok := v.(string); ok {
								value2 = nson.String(value)
							}
						default:
							continue
						}

						if value2 != nil {
							// 缓存
							m.crate.DBus().SetValue(tag.Name, value2, false)
						}
					}

					// fmt.Printf("%#v\n", Bucket.buses)
				}
			}

		} else if flags == 1 {
			nsonValue, err := nson.Message{}.Decode(reader)
			if err != nil {
				suger.Errorf("nson.Message{}.Decode: %s", err)
				return
			}

			message, ok := nsonValue.(nson.Message)
			if !ok {
				return
			}

			data, err := message.GetMessage(dict.DATA)
			if err != nil {
				suger.Errorf("message.GetMessage(dict.DATA): %s", err)
				return
			}

			s, err := m.crate.SlotService().GetSlot(clientId)
			if err != nil {
				suger.Errorf("SlotService{}.GetSlot(clientId): %s", err)
				return
			}

			if s.Status != ON {
				return
			}

			// fmt.Println(data.String())
			// fmt.Println(clientId)

			for k, v := range data {
				tag, err := m.crate.SlotService().GetTagBySlotIdAndAddress(clientId, k)
				if err != nil {
					suger.Errorf("SlotService{}.GetTagByName(k): %s", err)
					return
				}

				// fmt.Println(tag)

				if tag == nil {
					continue
				}

				if tag.Status != ON {
					continue
				}

				switch tag.DataType {
				case TypeI8, TypeI16, TypeI32:
					if v.Tag() != nson.TAG_I32 {
						continue
					}
				case TypeU8, TypeU16, TypeU32:
					if v.Tag() != nson.TAG_U32 {
						continue
					}
				case TypeI64:
					if v.Tag() != nson.TAG_I64 {
						continue
					}
				case TypeU64:
					if v.Tag() != nson.TAG_U64 {
						continue
					}
				case TypeF32:
					if v.Tag() != nson.TAG_F32 {
						continue
					}
				case TypeF64:
					if v.Tag() != nson.TAG_F64 {
						continue
					}
				case TypeBool:
					if v.Tag() != nson.TAG_BOOL {
						continue
					}
				case TypeString:
					if v.Tag() != nson.TAG_STRING {
						continue
					}
				default:
					continue
				}

				// 缓存
				m.crate.DBus().SetValue(tag.Name, v, false)
			}
		}

		return
	}

	return true
}