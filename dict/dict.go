package dict

// 请求相关
const (
	METHOD    = "method"
	PARAMS    = "params"
	CODE      = "code"
	ERROR     = "error"
	DATA      = "data"
	DATA_SIZE = "data_size"
)

// 数据
const (
	DEV_DATA     = "dev.data"     // 数据
	DEV_DATA_GET = "dev.data.get" // 读数据
	DEV_DATA_SET = "dev.data.set" // 写数据
	DEV_META     = "dev.meta"     // 元数据
	DEV_META_GET = "dev.meta.get" // 读元数据
	DEV_META_SET = "dev.meta.set" // 写元数据
)
