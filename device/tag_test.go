package device

import (
	"testing"

	"github.com/danclive/march/consts"
	"github.com/danclive/nson-go"
	"github.com/stretchr/testify/assert"
)

func TestDataTypeConvert(t *testing.T) {
	tag := Tag{
		DataType:        TypeF32,
		Value:           nson.F32(50),
		Convert:         consts.ON,
		ConvertDataType: TypeF32,
		HLimit:          100,
		LLimit:          0,
		HValue:          10,
		LValue:          0,
	}

	tag.ReadConvert()
	assert.Exactly(t, tag.Value, nson.F32(5))

	tag.WriteConvert()
	assert.Exactly(t, tag.Value, nson.F32(50))

	tag.ConvertDataType = TypeI32
	tag.HValue = 34567
	tag.LValue = 123

	tag.ReadConvert()
	assert.Exactly(t, tag.Value, nson.I32(17345))

	tag.WriteConvert()
	assert.Exactly(t, tag.Value, nson.F32(50))

	tag.LLimit = 3.45
	tag.LValue = -567
	tag.ReadConvert()
	assert.Exactly(t, tag.Value, nson.I32(16372))
}
