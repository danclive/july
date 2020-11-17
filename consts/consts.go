package consts

const (
	ON  = 1
	OFF = -1
)

const (
	RW = 1
	RO = -1
)

// parmas
const (
	METHOD    = "method"
	PARAMS    = "params"
	CODE      = "code"
	ERROR     = "error"
	DATA      = "data"
	DATA_SIZE = "datase"
	SLOT      = "slot"
	FLAGS     = "flags"
)

// chan
const (
	DEV_DATA     = "dev.data"     // 数据
	DEV_DATA_GET = "dev.data.get" // 读数据
	DEV_DATA_SET = "dev.data.set" // 写数据
	DEV_META     = "dev.meta"     // 元数据
	DEV_META_GET = "dev.meta.get" // 读元数据
	DEV_META_SET = "dev.meta.set" // 写元数据
)
