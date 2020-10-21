package mqttc

import (
	"time"

	"github.com/danclive/july/log"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var _client *client

type client struct {
	client    mqtt.Client
	close     chan struct{}
	onConnect []func(*client)
}

type Config struct {
	Addrs                []string
	ClientId             string
	User                 string
	Pass                 string
	Keepalive            int
	ConnectTimeout       int
	MaxReconnectInterval int
	PingTimeout          int
	Debug                bool
}

func Connect(config Config) {
	options := mqtt.NewClientOptions()

	for i := 0; i < len(config.Addrs); i++ {
		options.AddBroker(config.Addrs[i])
	}

	options.SetAutoReconnect(true)
	options.SetClientID(config.ClientId)
	options.SetUsername(config.User)
	options.SetPassword(config.Pass)
	options.SetKeepAlive(time.Duration(config.Keepalive))
	options.SetMaxReconnectInterval(time.Duration(config.MaxReconnectInterval))
	options.SetPingTimeout(time.Duration(config.PingTimeout))
	options.SetProtocolVersion(4)

	options.SetOnConnectHandler(func(_ mqtt.Client) {
		if _client != nil {
			if _client.onConnect != nil {
				for i := 0; i < len(_client.onConnect); i++ {
					if _client.onConnect[i] != nil {
						_client.onConnect[i](_client)
					}
				}
			}
		}
	})

	mqttclient := mqtt.NewClient(options)

	go func() {
		for {
			token := mqttclient.Connect()

			if token.WaitTimeout(time.Second * 10) {
				break
			}

			err := token.Error()
			if err != nil {
				log.Suger.Error(err)
			}
		}
	}()

	c := &client{
		client:    mqttclient,
		close:     make(chan struct{}),
		onConnect: make([]func(*client), 0),
	}

	_client = c
}

func GetClient() mqtt.Client {
	if _client == nil {
		return nil
	}

	return _client.client
}

func Close() {
	if _client != nil {
		_client.client.Disconnect(100)
	}
}
