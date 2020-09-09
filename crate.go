package july

import (
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
)

var zaplog *zap.Logger
var suger *zap.SugaredLogger

func init() {
	zaplog = zap.NewNop()
	suger = zaplog.Sugar()
}

type Crate interface {
	Log() *zap.SugaredLogger
	SlotService() *SlotService
	BboltDB() *bbolt.DB
	DBus() DBus
	Collect() Collect
	MqttService() *MqttService
	Store() *Store
	Upload() *Upload
	Run() error
	Stop() error
}

var _ Crate = &crate{}

type crate struct {
	slotService *SlotService
	bboltDB     *bbolt.DB
	dbus        DBus
	collect     Collect
	mqttService *MqttService
	store       *Store
	upload      *Upload
	close       chan struct{}
	option      Options
}

type Options struct {
	Log                   *zap.Logger
	ConfigDBPath          string
	BboltDBPath           string
	CollectReadInterval   int
	CollectKeepalive      int
	HisStoreEnable        bool
	HisStorePath          string
	MqttTcpAddrs          []string
	MqttWsAddrs           []string
	UploadEnable          bool
	UploadAddrs           []string
	UploadClientId        string
	UploadUser            string
	UploadPass            string
	UploadOnConnect       func(Crate, *Upload)
	UploadMqttClientDebug bool
}

func NewCrate(options Options) (Crate, error) {
	if options.Log == nil {
		options.Log = zap.NewNop()
	}

	if options.ConfigDBPath == "" {
		options.ConfigDBPath = "collect.db"
	}

	if options.BboltDBPath == "" {
		options.BboltDBPath = "bblot.db"
	}

	if options.CollectReadInterval == 0 {
		options.CollectReadInterval = 6
	}

	if options.CollectKeepalive == 0 {
		options.CollectKeepalive = 60
	}

	if options.HisStorePath == "" {
		options.HisStorePath = "store/"
	}

	if len(options.MqttTcpAddrs) == 0 {
		options.MqttTcpAddrs = []string{":1884"}
	}

	if len(options.MqttWsAddrs) == 0 {
		options.MqttWsAddrs = []string{":8084"}
	}

	if len(options.UploadAddrs) == 0 {
		options.UploadAddrs = []string{"127.0.0.1:1883"}
	}

	if options.UploadOnConnect == nil {
		options.UploadOnConnect = readyDataSync
	}

	zaplog = options.Log
	suger = zaplog.Sugar()

	// bboltDB
	bboltDB, err := bbolt.Open(options.BboltDBPath, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = bboltDB.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(DBUS_BUCKET)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// crate
	crate := &crate{
		bboltDB: bboltDB,
		close:   make(chan struct{}),
		option:  options,
	}

	// slot service
	slotService, err := newSlotService(crate, options.ConfigDBPath)
	if err != nil {
		return nil, err
	}

	crate.slotService = slotService

	// dbus
	crate.dbus = newDBus(crate)

	// collect
	crate.collect = newCollect(crate, options.CollectReadInterval, options.CollectKeepalive)

	// mqtt
	mqttService, err := newMqttService(crate, options.MqttTcpAddrs, options.MqttWsAddrs)
	if err != nil {
		return nil, err
	}

	crate.mqttService = mqttService

	// store
	his, err := newStore(crate, options.HisStorePath)
	if err != nil {
		return nil, err
	}
	crate.store = his

	// upload
	crate.upload = newUpload(
		crate,
		options.UploadAddrs,
		options.UploadClientId,
		options.UploadUser,
		options.UploadPass,
		options.UploadOnConnect,
		options.UploadMqttClientDebug,
	)

	return crate, nil
}

func (c *crate) Log() *zap.SugaredLogger {
	return suger
}

func (c *crate) SlotService() *SlotService {
	return c.slotService
}

func (c *crate) BboltDB() *bbolt.DB {
	return c.bboltDB
}

func (c *crate) DBus() DBus {
	return c.dbus
}

func (c *crate) Collect() Collect {
	return c.collect
}

func (c *crate) MqttService() *MqttService {
	return c.mqttService
}

func (c *crate) Store() *Store {
	return c.store
}

func (c *crate) Upload() *Upload {
	return c.upload
}

func (c *crate) Run() error {

	c.collect.Run()

	if c.option.HisStoreEnable {
		c.store.run()
	}

	c.mqttService.run()

	if c.option.UploadEnable {
		c.upload.run()
	}

	return nil
}

func (c *crate) Stop() error {
	select {
	case <-c.close:
		return nil
	default:
		close(c.close)
	}

	var err error

	if c.option.UploadEnable {
		if err2 := c.upload.stop(); err2 != nil {
			err = err2
		}
	}

	if err2 := c.mqttService.stop(); err2 != nil {
		err = err2
	}

	if c.option.HisStoreEnable {
		if err2 := c.store.stop(); err2 != nil {
			err = err2
		}
	}

	if err2 := c.collect.Stop(); err2 != nil {
		err = err2
	}

	if err2 := c.bboltDB.Close(); err2 != nil {
		err = err2
	}

	if err2 := c.slotService.engine.Close(); err2 != nil {
		err = err2
	}

	suger.Sync()
	zaplog.Sync()

	return err
}
