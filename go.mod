module github.com/danclive/july

go 1.13

require (
	github.com/danclive/mqtt v0.2.1
	github.com/danclive/nson-go v0.3.1
	github.com/danclive/queen-go v0.3.0
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/mattn/go-sqlite3 v1.14.0
	github.com/stretchr/testify v1.6.1
	go.etcd.io/bbolt v1.3.5
	go.uber.org/zap v1.16.0
	xorm.io/xorm v1.0.3
)

// replace github.com/danclive/mqtt => ../mqtt
replace github.com/danclive/queen-go => ../queen-go
