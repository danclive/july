package mqttc

import (
	"sync"
	"time"

	"github.com/danclive/july/log"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var _service *Service

func InitService(config Config) {
	options := mqtt.NewClientOptions()

	for i := 0; i < len(config.Addrs); i++ {
		options.AddBroker(config.Addrs[i])
	}

	options.SetAutoReconnect(true)
	options.SetConnectRetry(true)
	options.SetClientID(config.ClientId)
	options.SetUsername(config.User)
	options.SetPassword(config.Pass)
	options.SetProtocolVersion(4)

	options.SetOnConnectHandler(func(mqttclient mqtt.Client) {
		if _service != nil {
			_service.lock.Lock()
			for i := 0; i < len(_service.onConnect); i++ {
				if _service.onConnect[i] != nil {
					_service.onConnect[i](mqttclient)
				}
			}
			_service.lock.Unlock()
		}
	})

	mqttclient := mqtt.NewClient(options)

	s := &Service{
		client:    mqttclient,
		close:     make(chan struct{}),
		onConnect: make([]func(mqtt.Client), 0),
	}

	_service = s
}

func GetService() *Service {
	return _service
}

type Service struct {
	client    mqtt.Client
	close     chan struct{}
	onConnect []func(mqtt.Client)
	lock      sync.Mutex
}

type Config struct {
	Addrs    []string
	ClientId string
	User     string
	Pass     string
	Debug    bool
}

func Connect() {
	if _service == nil {
		return
	}

	go func() {
		time.Sleep(time.Second * 3)

		for {
			token := _service.client.Connect()

			if token.WaitTimeout(time.Second * 10) {
				break
			}

			err := token.Error()
			if err != nil {
				log.Suger.Error(err)
			}
		}
	}()
}

func Close() {
	if _service != nil {
		_service.client.Disconnect(100)
	}
}

func AddHandle(handle func(mqtt.Client)) {
	if _service != nil {
		_service.lock.Lock()
		_service.onConnect = append(_service.onConnect, handle)
		_service.lock.Unlock()
	}
}

func GetClient() mqtt.Client {
	if _service == nil {
		return nil
	}

	return _service.client
}
