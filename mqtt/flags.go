package mqtt

const (
	FORMAT_JSOM     uint16 = 0b0000_0000_0000_0000
	FORMAT_NSOM     uint16 = 0b0000_0000_0000_0001
	DEVICE_DIRECTLY uint16 = 0b0000_0000_0000_0000
	DEVICE_GATEWAY  uint16 = 0b0000_0000_0001_0000
	COMPRESS_GZIP   uint16 = 0b0000_0001_0000_0000
	COMPRESS_ZSTD   uint16 = 0b0000_0010_0000_0000
)
