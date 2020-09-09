package july

import (
	"net/url"

	"github.com/danclive/july/util"
	"github.com/danclive/mqtt"
	"github.com/danclive/mqtt/packets"
)

func hook(crate Crate) mqtt.Hooks {
	crate.SlotService().SlotReset(DriverMQTT)

	onConnect := func(client mqtt.Client) (code uint8) {
		clientId := client.OptionsReader().ClientID()
		username := client.OptionsReader().Username()
		password := client.OptionsReader().Password()

		s, err := crate.SlotService().GetSlot(clientId)
		if err != nil {
			return packets.CodeServerUnavaliable
		}

		if s.Driver != DriverMQTT {
			return packets.CodeNotAuthorized
		}

		if s.Status != ON {
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

		crate.SlotService().SlotOnline(clientId)
	}

	onClose := func(client mqtt.Client, err error) {
		clientId := client.OptionsReader().ClientID()

		crate.SlotService().SlotOffline(clientId)
	}

	onStop := func() {
		crate.SlotService().SlotReset(DriverMQTT)
	}

	hooks := mqtt.Hooks{
		OnConnect:   onConnect,
		OnConnected: onConnected,
		OnClose:     onClose,
		OnStop:      onStop,
	}

	return hooks
}
