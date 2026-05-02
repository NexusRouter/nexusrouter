package repository

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRewriteChatCompletionsModelBody(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, AutoMigrate(db))
	require.NoError(t, CreateModelCatalogEntry(db, &ModelCatalogEntry{ID: "alias-1", DisplayName: "A"}))
	am := "real-model"
	require.NoError(t, CreateModelUpstreamBinding(db, &ModelUpstreamBinding{
		CatalogEntryID: "alias-1",
		UpstreamID:     "default",
		Enabled:        true,
		ActualModel:    &am,
	}))

	raw := []byte(`{"model":"alias-1","messages":[{"role":"user","content":"x"}]}`)
	out := RewriteChatCompletionsModelBody(raw, db, "default")
	var m map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(out, &m))
	var model string
	require.NoError(t, json.Unmarshal(m["model"], &model))
	require.Equal(t, "real-model", model)

	unchanged := RewriteChatCompletionsModelBody(raw, db, "other-upstream")
	require.Equal(t, string(raw), string(unchanged))
}
