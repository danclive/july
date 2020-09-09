package july

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/danclive/july/util"
	"xorm.io/xorm"
)

type Store struct {
	crate   Crate
	dbPath  string
	engines map[string]*xorm.Engine
	lock    sync.Mutex
	close   chan struct{}
	stopped bool
}

type DataPoint struct {
	T uint32  `xorm:"pk 't'" json:"t"`
	V float64 `xorm:"'v'" json:"v"`
}

func (*DataPoint) TableName() string {
	return "data"
}

func newStore(crate Crate, dbPath string) (*Store, error) {
	if !Exist(dbPath) {
		err := os.Mkdir(dbPath, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	h := &Store{
		crate:   crate,
		dbPath:  dbPath,
		engines: make(map[string]*xorm.Engine),
		close:   make(chan struct{}),
		stopped: false,
	}

	return h, nil
}

func (s *Store) store() {
	ticker := time.NewTicker(60 * time.Second)

	for {
		select {
		case <-s.close:
			ticker.Stop()
			return
		case ti := <-ticker.C:
			func(ti time.Time) {
				dbus := s.crate.DBus()

				tags, err := s.crate.SlotService().ListTagForHisData()
				if err != nil {
					suger.Errorf("SlotService().ListTagForHisData(): %s", err)
					return
				}

				suger.Debugf("Save tags: %v", len(tags))

				for i := 0; i < len(tags); i++ {
					value, has := dbus.GetValue(tags[i].Name)
					if !has {
						continue
					}

					if value == nil {
						continue
					}

					value2, ok := util.NsonValueToFloat64(value)
					if !ok {
						continue
					}

					err = s.SaveData(tags[i].Name, ti, value2)
					if err != nil {
						suger.Errorf("store.SaveData: %s", err)
						continue
					}
				}
			}(ti)
		}
	}
}

func (s *Store) getEngine(tagName string) (*xorm.Engine, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.stopped {
		return nil, errors.New("closed")
	}

	if engine, has := s.engines[tagName]; has {
		return engine, nil
	}

	engine, err := xorm.NewEngine("sqlite3", fmt.Sprintf("%s/%s.db?cache=shared&_journal_mode=WAL", s.dbPath, tagName))
	if err != nil {
		return nil, err
	}

	{
		has, err := engine.IsTableExist(DataPoint{})
		if err != nil {
			return nil, err
		}

		if !has {
			err = engine.Sync2(DataPoint{})
			if err != nil {
				return nil, err
			}
		}
	}

	s.engines[tagName] = engine

	return engine, nil
}

func (s *Store) SaveData(tagName string, ti time.Time, value float64) error {
	db, err := s.getEngine(tagName)
	if err != nil {
		return err
	}

	ts := ti.Unix()

	_, err = db.InsertOne(DataPoint{T: uint32(ts / 60 * 60), V: value})
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) QueryData(tagName string, start, end time.Time, step int) ([]DataPoint, error) {
	db, err := s.getEngine(tagName)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`
		SELECT t, AVG(v) as v FROM (
			SELECT CASE WHEN t %% %v = 0 THEN t WHEN t %% %v <> 0 THEN t / %v * %v END as t, v FROM data WHERE t > %v AND t < %v
		) GROUP BY t
	`, step, step, step, step, start.Unix(), end.Unix())

	rows := make([]DataPoint, 0, 100)
	err = db.SQL(sql).Find(&rows)
	// err = db.SQL("SELECT t, AVG(v) as v FROM data WHERE t % 360 = 0 GROUP BY t").Find(&rows)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (s *Store) run() {
	zaplog.Info("starting store")
	go s.store()
}

func (s *Store) stop() error {
	select {
	case <-s.close:
		return nil
	default:
		close(s.close)
	}

	var err error
	s.lock.Lock()
	defer s.lock.Unlock()

	s.stopped = true

	for _, engine := range s.engines {
		e := engine.Close()
		if e != nil {
			err = e
		}
	}

	return err
}

func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

// SELECT CASE WHEN t % 120 = 0 THEN t WHEN t % 120 <> 0 THEN ((t / 120 + 1) * 120) END as t, v FROM data
