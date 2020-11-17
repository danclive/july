package queenc

import (
	"github.com/danclive/july/consts"
	"github.com/danclive/july/log"
	"github.com/danclive/nson-go"
	"github.com/danclive/queen-go/client"
)

var _service *Service

func InitService(options client.Options) {
	q, err := newService(options)
	if err != nil {
		log.Suger.Fatal(err)
	}

	_service = q

	AddHandel("ping", nil, false, false, func(_ *client.Client, recv *client.RecvMessage, back *client.SendMessage) {
		log.Suger.Debug(recv)

		if back != nil {
			back.Body().Insert(consts.DATA, nson.Message{consts.CODE: nson.I32(0)})
		}
	})
}

func GetService() *Service {
	return _service
}

type Service struct {
	client  *client.Client
	handles map[string]handle
}

type handle struct {
	ch     string
	label  []string
	share  bool
	attach bool
	fn     func(*client.Client, *client.RecvMessage, *client.SendMessage)
}

func newService(options client.Options) (*Service, error) {
	s := &Service{
		handles: make(map[string]handle),
	}

	c, err := client.NewClient(options)
	if err != nil {
		return nil, err
	}

	c.Recv(s.onRecv)

	s.client = c

	return s, nil
}

func Connect() {
	_service.client.Connect(_service.onConnect)
}

func Close() {
	_service.client.Close()
}

func (s *Service) onConnect(c *client.Client) {
	for _, handle := range s.handles {
		if handle.attach {
			if err := c.Attach(handle.ch, handle.label, handle.share); err != nil {
				log.Suger.Error(err)
			}
		}
	}
}

func (q *Service) onRecv(c *client.Client, recv client.RecvMessage) {
	back := recv.Back()

	if handle, ok := q.handles[recv.Ch]; ok {
		if handle.fn != nil {
			handle.fn(c, &recv, back)
		}
	} else {
		msg := nson.Message{
			consts.CODE:  nson.I32(404),
			consts.ERROR: nson.String("Not Found"),
		}

		if back != nil {
			back.Body().Insert(consts.DATA, msg)
		}
	}

	if back != nil {
		_, err := c.Send(back, 0)
		if err != nil {
			log.Suger.Error(err)
		}
	}
}

func AddHandel(ch string, label []string, share bool, attach bool, fn func(*client.Client, *client.RecvMessage, *client.SendMessage)) {
	_service.handles[ch] = handle{ch, label, share, attach, fn}
}

func GetClient() *client.Client {
	if _service == nil {
		return nil
	}

	return _service.client
}
