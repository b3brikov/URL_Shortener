package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	first := generate(5)
	second := generate(5)

	assert.NotEqual(t, first, second)
	assert.Len(t, first, 5)
	assert.Len(t, second, 5)
}
