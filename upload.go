package july

import (
	"bytes"
	"log"
	"os"
	"time"

	"github.com/danclive/july/dict"
	"github.com/danclive/july/util"
	"github.com/danclive/mqtt/ext"
	"github.com/danclive/nson-go"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Upload struct {
	crate     Crate
	client    mqtt.Client
	close     chan struct{}
	llac      *ext.MqttLlac
	onConnect func(Crate, *Upload)
}

type UploadConfig struct {
	Addrs     []string
	ClientId  string
	User      string
	Pass      string
	OnConnect func(Crate, *Upload)
	Debug     bool
}

func newUpload(crate Crate, config UploadConfig) *Upload {
	if config.Debug {
		mqtt.DEBUG = log.New(os.Stdout, "", log.LstdFlags)
		mqtt.ERROR = log.New(os.Stdout, "", log.LstdFlags)
		mqtt.WARN = log.New(os.Stdout, "", log.LstdFlags)
	}

	options := mqtt.NewClientOptions()

	if len(config.Addrs) == 0 {
		options.AddBroker("127.0.0.1:1883")
	} else {
		for i := 0; i < len(config.Addrs); i++ {
			options.AddBroker(config.Addrs[i])
		}
	}

	options.SetClientID(config.ClientId)
	options.SetUsername(config.User)
	options.SetPassword(config.Pass)
	options.SetMaxReconnectInterval(time.Second * 30)

	client := mqtt.NewClient(options)

	upload := &Upload{
		crate:     crate,
		client:    client,
		close:     make(chan struct{}),
		onConnect: config.OnConnect,
	}

	return upload
}

func (u *Upload) LLacer() *ext.MqttLlac {
	return u.llac
}

func (u *Upload) run() error {
	go u.start()
	return nil
}

func (u *Upload) start() error {
	zaplog.Info("starting upload")

	for {
		select {
		case <-u.close:
			return nil
		default:
		}

		zaplog.Debug("upload client connect...")

		token := u.client.Connect()

		if token.WaitTimeout(time.Second * 5) {
			break
		}

		zaplog.Error("upload client connect timeout")

		err := token.Error()
		if err != nil {
			zaplog.Error(err.Error())
		}
	}

	zaplog.Debug("upload client connected")

	u.llac = ext.NewMqttLlac(u.client, zaplog)

	if u.onConnect != nil {
		u.onConnect(u.crate, u)
	}

	for {
		ticker := time.NewTicker(60 * time.Second)

		for {
			select {
			case <-u.close:
				ticker.Stop()
				return nil
			case <-ticker.C:
				func() {
					dbus := u.crate.DBus()

					tags, err := u.crate.SlotService().ListTagForUpload()
					if err != nil {
						suger.Errorf("SlotService().ListTagForUpload(): %s", err)
						return
					}

					suger.Debugf("Upload tags: %v", len(tags))

					data := nson.Message{}

					for i := 0; i < len(tags); i++ {
						value, has := dbus.GetValue(tags[i].Name)
						if !has {
							continue
						}

						if value == nil {
							continue
						}

						data[tags[i].Name] = value
					}

					msg := nson.Message{
						"data": data,
					}

					buffer := new(bytes.Buffer)
					util.WriteUint16(1, buffer)

					err = msg.Encode(buffer)
					if err != nil {
						suger.Errorf("msg.Encode(buffer): %v", err.Error())
						return
					}

					token := u.client.Publish(dict.DEV_DATA, 1, false, buffer.Bytes())
					if !token.WaitTimeout(time.Second * 10) {
						suger.Error("Upload timeout")
					}

					err = token.Error()
					if err != nil {
						suger.Errorf(" u.client.Publish(dict.DEV_DATA: %v", err.Error())
					}
				}()
			}
		}
	}
}

func (u *Upload) stop() error {
	select {
	case <-u.close:
		return nil
	default:
		close(u.close)
	}

	u.client.Disconnect(100)

	return nil
}
