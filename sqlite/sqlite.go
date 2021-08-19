package sqlite

import (
	"fmt"

	"github.com/danclive/july/log"
	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
	"xorm.io/xorm/caches"
	xormlog "xorm.io/xorm/log"
)

var _engine *xorm.Engine

func Connect(file string, debug bool) {
	// ?cache=shared&_busy_timeout=10000
	engine, err := xorm.NewEngine("sqlite3", fmt.Sprintf("%s?cache=shared&_journal_mode=WAL", file))
	if err != nil {
		log.Suger.Fatal("Can't open db: ", err)
	}

	if debug {
		engine.ShowSQL(true)
		engine.Logger().SetLevel(xormlog.LOG_DEBUG)
	}

	// engine.SetMaxOpenConns(5)
	// engine.SetMaxIdleConns(2)

	// fmt.Printf("%+v/n", engine.DB().Stats())

	engine.Ping()

	cacher := caches.NewLRUCacher(caches.NewMemoryStore(), 1500)
	engine.SetDefaultCacher(cacher)

	_engine = engine
}

func GetEngine() *xorm.Engine {
	return _engine
}

func Close() {
	_engine.Close()
}
