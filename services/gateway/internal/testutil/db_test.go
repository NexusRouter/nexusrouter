package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestMustSQLiteMemory_Smoke 内存库迁移与启动顺序辅助可调用（cfg 可为 nil）。
func TestMustSQLiteMemory_Smoke(t *testing.T) {
	db := MustSQLiteMemory(t, nil)
	require.NotNil(t, db)
}
