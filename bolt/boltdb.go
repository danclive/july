package bolt

import (
	"github.com/danclive/july/log"
	"go.etcd.io/bbolt"
)

var _boltDB *bbolt.DB

var CFG_BUCKET = []byte("cfg")

func Connect(file string) {
	bboltDB, err := bbolt.Open(file, 0600, nil)
	if err != nil {
		log.Suger.Fatal("Can't open db: ", err)
	}

	err = bboltDB.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(CFG_BUCKET)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Suger.Fatal("Can't update db: ", err)
	}

	_boltDB = bboltDB
}

func GetBoltDB() *bbolt.DB {
	return _boltDB
}

func Close() {
	_boltDB.Close()
}
