package app

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewApplication_FieldsSet(t *testing.T) {
	log := zap.NewNop()
	eng := gin.New()
	a := NewApplication(log, eng)
	assert.Same(t, log, a.Log)
	assert.Same(t, eng, a.Engine)
}
