package slot

type Driver interface {
	Connect(params string) (Driver, error)
	Close() error
	Name() string
	Read([]Tag) error
	Write([]Tag) error
}

var _drivers = make(map[string]Driver)

func RegisterDriver(name string, driver Driver) {
	_drivers[name] = driver
}
