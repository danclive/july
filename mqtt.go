package july

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/danclive/mqtt"
	"github.com/danclive/mqtt/ext"
)

type MqttService struct {
	crate  Crate
	server mqtt.Server
}

func newMqttService(crate Crate, tcpAddrs []string, wsAddrs []string) (*MqttService, error) {
	if len(tcpAddrs) == 0 && len(wsAddrs) == 0 {
		return nil, errors.New("地址不能为空")
	}

	s := &MqttService{
		crate: crate,
	}

	server := mqtt.NewServer(
		mqtt.WithHook(hook(crate)),
		mqtt.WithPlugin(&MqttRecv{crate: crate}),
		mqtt.WithPlugin(ext.NewMqttCall(zaplog)),
		mqtt.WithLogger(zaplog),
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

func (m *MqttService) run() {
	m.server.Run()
}

func (m *MqttService) stop() error {
	return m.server.Stop(context.Background())
}

func (m *MqttService) MqttServer() mqtt.Server {
	return m.server
}
