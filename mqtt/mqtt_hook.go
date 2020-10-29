package mqtt

import (
	"net/url"

	"github.com/danclive/july/consts"
	"github.com/danclive/july/device"
	"github.com/danclive/july/util"
	"github.com/danclive/mqtt"
	"github.com/danclive/mqtt/packets"
)

func hook() mqtt.Hooks {
	device.GetService().SlotReset(device.DriverMQTT)

	onConnect := func(client mqtt.Client) (code uint8) {
		clientId := client.OptionsReader().ClientID()
		username := client.OptionsReader().Username()
		password := client.OptionsReader().Password()

		s, err := device.GetService().GetSlot(clientId)
		if err != nil {
			return packets.CodeServerUnavaliable
		}

		if s.Driver != device.DriverMQTT {
			return packets.CodeNotAuthorized
		}

		if s.Status != consts.ON {
			return packets.CodeNotAuthorized
		}

		var config struct {
			User string `cfg:"user"`
			Pass string `cfg:"pass"`
		}

		u, err := url.ParseQuery(s.Params)
		if err != nil {
			return packets.CodeNotAuthorized
		}

		err = util.MapConfig(&config, u)
		if err != nil {
			return packets.CodeNotAuthorized
		}

		if username != config.User || password != config.Pass {
			return packets.CodeBadUsernameorPsw
		}

		return packets.CodeAccepted
	}

	onConnected := func(client mqtt.Client) {
		clientId := client.OptionsReader().ClientID()

		device.GetService().SlotOnline(clientId)
	}

	onClose := func(client mqtt.Client, err error) {
		clientId := client.OptionsReader().ClientID()

		device.GetService().SlotOffline(clientId)
	}

	onStop := func() {
		device.GetService().SlotReset(device.DriverMQTT)
	}

	hooks := mqtt.Hooks{
		OnConnect:   onConnect,
		OnConnected: onConnected,
		OnClose:     onClose,
		OnStop:      onStop,
	}

	return hooks
}
