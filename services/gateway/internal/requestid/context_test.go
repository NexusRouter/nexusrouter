package requestid

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithID_FromContext(t *testing.T) {
	ctx := WithID(context.Background(), "rid-abc")
	assert.Equal(t, "rid-abc", FromContext(ctx))
}

func TestWithID_EmptyNoOp(t *testing.T) {
	ctx := context.Background()
	got := WithID(ctx, "")
	assert.True(t, got == ctx)
	assert.Equal(t, "", FromContext(got))
}

func TestStrict_nilInterfaceContextPanics(t *testing.T) {
	var nilCtx context.Context
	assert.PanicsWithValue(t, "requestid: nil Context", func() {
		FromContext(nilCtx)
	})
	assert.PanicsWithValue(t, "requestid: nil Context", func() {
		_ = WithID(nilCtx, "rid")
	})
	assert.PanicsWithValue(t, "requestid: nil Context", func() {
		_ = WithID(nilCtx, "")
	})
}

func TestChildContext_Inherits(t *testing.T) {
	parent := WithID(context.Background(), "parent-1")
	child, cancel := context.WithTimeout(parent, time.Minute)
	defer cancel()
	assert.Equal(t, "parent-1", FromContext(child))
}
