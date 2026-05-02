package handler

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewOpenAIModelItem_JSONShape(t *testing.T) {
	m := newOpenAIModelItem("gpt-test", "acme", 1626777600)
	b, err := json.Marshal(m)
	require.NoError(t, err)
	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(b, &raw))
	require.Contains(t, raw, "permission")
	require.Contains(t, raw, "root")
	var perms []map[string]any
	require.NoError(t, json.Unmarshal(raw["permission"], &perms))
	require.GreaterOrEqual(t, len(perms), 1)
	require.Equal(t, "model_permission", perms[0]["object"])
	var root string
	require.NoError(t, json.Unmarshal(raw["root"], &root))
	require.Equal(t, "gpt-test", root)
	_, hasParent := raw["parent"]
	require.False(t, hasParent)
}
