package slot

import (
	"testing"

	"github.com/danclive/nson-go"
	"github.com/stretchr/testify/assert"
)

func TestTagFormat(t *testing.T) {
	tag1 := Tag{
		Format: ".2f",
		Value:  nson.F32(123.4567),
	}

	assert.Exactly(t, tag1.Value, nson.F32(123.4567))
	tag1.FormatValue()
	assert.Exactly(t, tag1.Value, nson.F32(123.46))

	tag2 := Tag{
		Format: "",
		Value:  nson.F32(123.4567),
	}

	assert.Exactly(t, tag2.Value, nson.F32(123.4567))
	tag2.FormatValue()
	assert.Exactly(t, tag2.Value, nson.F32(123.4567))

	tag3 := Tag{
		Format: ".2f",
		Value:  nson.F64(123.4567),
	}

	t.Log(tag3.Value)
	assert.Exactly(t, tag3.Value, nson.F64(123.4567))
	tag3.FormatValue()
	t.Log(tag3.Value)
	assert.Exactly(t, tag3.Value, nson.F64(123.46))
}
