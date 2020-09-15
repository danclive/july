package july

import (
	"sync"
	"time"

	"github.com/danclive/queen-go/client"
	"github.com/danclive/queen-go/conn"
)

type QueenService struct {
	crate   Crate
	client  *client.Client
	lock    sync.Mutex
	close   chan struct{}
	stopped bool
}

func newQueenService(crate Crate, config conn.Config) (*QueenService, error) {
	s := &QueenService{
		crate: crate,
		close: make(chan struct{}),
	}
	go func() {
		s.lock.Lock()
		defer s.lock.Unlock()

		for {
			c, err := client.NewClient(config)
			if err != nil {
				// return nil, err
				crate.Log().Error(err)
				time.Sleep(time.Second * 10)
				continue
			}

			c.OnConnect(s.onConnect)

			s.client = c

			return
		}
	}()

	return s, nil
}

func (q *QueenService) run() {
	go func() {
		q.lock.Lock()
		q.lock.Unlock()

		recvChan := q.client.Recv()

		for {
			select {
			case <-q.close:
				return
			case recvMsg := <-recvChan:
				go func() {
					backMsg := recvMsg.Back()

					q.recv(recvMsg, backMsg)

					if backMsg != nil {
						q.client.Send(backMsg, 0)
					}
				}()
			}
		}
	}()
}

func (q *QueenService) stop() {
	go func() {
		select {
		case <-q.close:
			return
		default:
			close(q.close)
		}

		q.lock.Lock()
		defer q.lock.Unlock()

		q.stopped = true

		q.client.Close()
	}()
}

func (q *QueenService) onConnect() {
	q.client.Attach("ping", nil, 0)
	// q.client.Attach(dict.DEV_META, nil, 0)
}

func (q *QueenService) Client() *client.Client {
	return q.client
}
