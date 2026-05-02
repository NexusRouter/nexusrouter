package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestModelCatalog_CRUD_and_Published(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, AutoMigrate(db))

	require.NoError(t, CreateModelCatalogEntry(db, &ModelCatalogEntry{
		ID:          "gpt-test",
		DisplayName: "GPT Test",
		OwnedBy:     "acme",
	}))
	got, err := GetModelCatalogEntry(db, "gpt-test")
	require.NoError(t, err)
	require.Equal(t, "GPT Test", got.DisplayName)

	require.NoError(t, CreateModelUpstreamBinding(db, &ModelUpstreamBinding{
		CatalogEntryID: "gpt-test",
		UpstreamID:     "u1",
		Enabled:        true,
		Priority:       0,
	}))
	valid := map[string]struct{}{"u1": {}}
	rows, err := ListPublishedModels(db, valid)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "gpt-test", rows[0].CatalogID)

	valid2 := map[string]struct{}{"missing": {}}
	rows2, err := ListPublishedModels(db, valid2)
	require.NoError(t, err)
	require.Len(t, rows2, 0)

	require.NoError(t, DeleteModelCatalogEntry(db, "gpt-test"))
	_, err = GetModelCatalogEntry(db, "gpt-test")
	require.Error(t, err)
}

func TestModelCatalog_DuplicateEntry(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, AutoMigrate(db))
	row := &ModelCatalogEntry{ID: "dup", DisplayName: "A"}
	require.NoError(t, CreateModelCatalogEntry(db, row))
	err = CreateModelCatalogEntry(db, &ModelCatalogEntry{ID: "dup", DisplayName: "B"})
	require.Error(t, err)
}
