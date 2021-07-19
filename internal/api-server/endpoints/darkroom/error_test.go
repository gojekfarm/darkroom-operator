package darkroom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	e := Error{
		Message: "Human Readable Message",
		Err:     "system error message",
	}

	assert.Equal(t, "system error message", e.Error())
}
