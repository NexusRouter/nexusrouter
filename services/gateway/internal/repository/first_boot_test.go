package repository

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openFirstBootTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "fb.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, AutoMigrate(db))
	require.NoError(t, EnsureSystemBootstrap(db))
	return db
}

func TestCompleteFirstBoot_SucceedsOnce(t *testing.T) {
	db := openFirstBootTestDB(t)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	err := CompleteFirstBoot(db, now, CompleteFirstBootInput{
		AdminUsername:   "root",
		AdminPassword:   "password1",
		SiteDisplayName: "ACME",
		BcryptCost:      bcrypt.MinCost,
	})
	require.NoError(t, err)
	ok, err := IsSystemInitialized(db)
	require.NoError(t, err)
	require.True(t, ok)
	err = CompleteFirstBoot(db, now, CompleteFirstBootInput{
		AdminUsername: "other",
		AdminPassword: "password2",
		BcryptCost:    bcrypt.MinCost,
	})
	require.ErrorIs(t, err, ErrBootstrapAlreadyCompleted)
}

func TestCompleteFirstBoot_ConcurrentSecondBlocks(t *testing.T) {
	db := openFirstBootTestDB(t)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	err := db.Model(&SystemBootstrapRow{}).Where("id = ?", SystemBootstrapSingletonPK).Updates(map[string]any{
		"init_in_progress": true,
		"init_started_at":  now,
	}).Error
	require.NoError(t, err)
	err = CompleteFirstBoot(db, now.Add(time.Second), CompleteFirstBootInput{
		AdminUsername: "root",
		AdminPassword: "password1",
		BcryptCost:    bcrypt.MinCost,
	})
	require.ErrorIs(t, err, ErrBootstrapInProgress)
}

func TestGetBootstrapStatus_InitializingPhase(t *testing.T) {
	db := openFirstBootTestDB(t)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	require.NoError(t, db.Model(&SystemBootstrapRow{}).Where("id = ?", SystemBootstrapSingletonPK).Updates(map[string]any{
		"init_in_progress": true,
		"init_started_at":  now,
	}).Error)
	st, err := GetBootstrapStatus(db, now.Add(10*time.Second))
	require.NoError(t, err)
	require.False(t, st.Initialized)
	require.Equal(t, BootstrapPhaseInitializing, st.Phase)
}

func TestIsSystemInitialized_NoBootstrapRowMeansNotInitialized(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "norow.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, AutoMigrate(db))
	ok, err := IsSystemInitialized(db)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestResetFirstBoot_ClearsAdmins(t *testing.T) {
	db := openFirstBootTestDB(t)
	now := time.Now().UTC()
	require.NoError(t, CompleteFirstBoot(db, now, CompleteFirstBootInput{
		AdminUsername: "root",
		AdminPassword: "password1",
		BcryptCost:    bcrypt.MinCost,
	}))
	require.NoError(t, ResetFirstBoot(db))
	ok, err := IsSystemInitialized(db)
	require.NoError(t, err)
	require.False(t, ok)
	var n int64
	require.NoError(t, db.Model(&AdminUserModel{}).Count(&n).Error)
	require.Zero(t, n)
}
