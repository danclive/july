package upload

import (
	"bytes"
	"sync"
	"time"

	"github.com/danclive/july/cache"
	"github.com/danclive/july/device"
	"github.com/danclive/july/log"
	"github.com/danclive/july/mqttc"
	"github.com/danclive/march/consts"
	"github.com/danclive/march/packet"
	"github.com/danclive/nson-go"
)

var _service *Service
var once sync.Once

type Service struct {
	lock    sync.Mutex
	close   chan struct{}
	stopped bool
}

func InitService() {
	s := &Service{
		close: make(chan struct{}),
	}

	_service = s
}

func GetService() *Service {
	return _service
}

func Run(interval int) {
	once.Do(func() {
		log.Suger.Info("run upload")
		go _service.run(interval)
	})
}

func Stop() {
	s := _service

	select {
	case <-s.close:
		return
	default:
		close(s.close)
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.stopped = true
}

func (s *Service) run(interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	for {
		select {
		case <-s.close:
			ticker.Stop()
			return
		case ts := <-ticker.C:
			func(ts time.Time) {
				cache := cache.GetService()
				mqttclient := mqttc.GetClient()
				if mqttclient == nil {
					log.Suger.Error("mqtt client not connect")
					return
				}

				tags, err := device.GetService().ListTagAndUpload()
				if err != nil {
					log.Suger.Errorf("ListTagForUpload(): %s", err)
					return
				}

				log.Suger.Debugf("Upload tags: %v", len(tags))

				array := nson.Array{}

				array.Push(nson.U32(0))
				array.Push(nson.Bool(true))

				for i := 0; i < len(tags); i++ {
					err := cache.GetValue(&tags[i])
					if err != nil {
						log.Suger.Error(err)
						continue
					}

					k, err := nson.MessageIdFromHex(tags[i].ID)
					if err != nil {
						continue
					}

					array.Push(k)
					array.Push(tags[i].Value)
				}

				if len(array) == 0 {
					return
				}

				pack := packet.NewPacket()
				pack.Header.SetContentType(1)

				buffer := new(bytes.Buffer)

				err = pack.Encode(buffer)
				if err != nil {
					log.Suger.Error(err)
					return
				}

				err = array.Encode(buffer)
				if err != nil {
					log.Suger.Error(err)
					return
				}

				token := mqttclient.Publish(consts.DEV_DATA, 0, false, buffer.Bytes())
				if !token.WaitTimeout(time.Second * 10) {
					log.Suger.Error("Publish timeout")
				}

				err = token.Error()
				if err != nil {
					log.Suger.Errorf("mqttclient.Publish: %v", err.Error())
				}
			}(ts)
		}
	}
}
