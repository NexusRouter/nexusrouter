package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSeedOfficialVendors_idempotent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&ModelVendor{}))

	log := zap.NewNop()
	require.NoError(t, SeedOfficialVendors(db, log))

	var n int64
	require.NoError(t, db.Model(&ModelVendor{}).Count(&n).Error)
	require.Equal(t, int64(OfficialVendorSeedCount()), n)

	require.NoError(t, SeedOfficialVendors(db, log))
	require.NoError(t, db.Model(&ModelVendor{}).Count(&n).Error)
	require.Equal(t, int64(OfficialVendorSeedCount()), n)

	var openai ModelVendor
	require.NoError(t, db.Where("vendor_code = ?", "openai").First(&openai).Error)
	require.Equal(t, "OpenAI", openai.VendorName)
	require.Equal(t, int8(1), openai.VendorType)

	var deepseek ModelVendor
	require.NoError(t, db.Where("vendor_code = ?", "deepseek").First(&deepseek).Error)
	require.Equal(t, int8(1), deepseek.Status)
}
