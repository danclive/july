package july

import (
	"log"

	"github.com/danclive/july/dict"
	"github.com/danclive/nson-go"
	"github.com/danclive/queen-go/client"
)

func (q *QueenService) recv(recvMsg *client.RecvMessage, backMsg *client.SendMessage) {
	log.Println(recvMsg)

	switch recvMsg.Ch {
	case "ping":
		if backMsg != nil {
			backMsg.Body().Insert(dict.DATA, nson.Message{dict.CODE: nson.I32(0)})
		}
	case dict.DEV_META:
		recv, err := recvMsg.Body.GetMessage(dict.DATA)
		if err != nil {
			return
		}

		back := handleDevMeta(q.crate, recv)

		if backMsg != nil {
			backMsg.Body().Insert(dict.DATA, back)
		}
	default:
		msg := nson.Message{
			dict.CODE:  nson.I32(404),
			dict.ERROR: nson.String("Not Found"),
		}

		if backMsg != nil {
			backMsg.Body().Insert(dict.DATA, msg)
		}
	}
}
