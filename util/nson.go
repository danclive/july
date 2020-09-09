package util

import (
	"github.com/danclive/nson-go"
)

func NsonValueToFloat64(value nson.Value) (float64, bool) {
	switch value.Tag() {
	case nson.TAG_I32:
		return float64(value.(nson.I32)), true
	case nson.TAG_U32:
		return float64(value.(nson.U32)), true
	case nson.TAG_I64:
		return float64(value.(nson.I64)), true
	case nson.TAG_U64:
		return float64(value.(nson.U64)), true
	case nson.TAG_F32:
		return float64(value.(nson.F32)), true
	case nson.TAG_F64:
		return float64(value.(nson.F64)), true
	case nson.TAG_BOOL:
		if value.(nson.Bool) {
			return 1, true
		}

		return 0, true
	}

	return 0, false
}
