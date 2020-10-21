package mqtt

import (
	"context"
	"errors"
	"sync"

	"net"
	"net/http"

	"github.com/danclive/july/log"
	"github.com/danclive/mqtt"
)

var _service *Service
var once sync.Once

func InitService(tcpAddrs []string, wsAddrs []string) {
	s, err := newService(tcpAddrs, wsAddrs)
	if err != nil {
		log.Suger.Fatal(err)
	}

	_service = s
}

func GetService() *Service {
	return _service
}

type Service struct {
	server mqtt.Server
}

func newService(tcpAddrs []string, wsAddrs []string) (*Service, error) {
	if len(tcpAddrs) == 0 && len(wsAddrs) == 0 {
		return nil, errors.New("addrs cannot be null")
	}

	s := &Service{}

	server := mqtt.NewServer(
		mqtt.WithHook(hook()),
		mqtt.WithPlugin(&MqttRecv{}),
		mqtt.WithLogger(log.Logger),
	)

	for _, addr := range tcpAddrs {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, err
		}

		server.Init(mqtt.WithTCPListener(ln))
	}

	for _, addr := range wsAddrs {
		ws := &mqtt.WsServer{
			Server: &http.Server{Addr: addr},
			Path:   "/ws",
		}

		server.Init(mqtt.WithWebsocketServer(ws))
	}

	s.server = server

	return s, nil
}

func Run() {
	once.Do(func() {
		_service.server.Run()
	})
}

func Stop() error {
	return _service.server.Stop(context.Background())
}

func (m *Service) Server() mqtt.Server {
	return m.server
}
