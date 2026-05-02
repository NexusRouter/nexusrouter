package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPickChatTarget_priorityAndOfficial(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&ModelVendor{}, &ModelBase{}, &ModelUpstream{}, &ModelInstance{}))

	v := ModelVendor{VendorName: "OpenAI", VendorType: 1, VendorCode: "openai", Status: 1}
	require.NoError(t, db.Create(&v).Error)
	b := ModelBase{ModelName: "GPT-3.5", ModelCode: "gpt-3.5-turbo", ModelType: 1, Status: 1}
	require.NoError(t, db.Create(&b).Error)
	up := ModelUpstream{VendorID: v.ID, UpstreamName: "u1", BaseURL: "https://api.openai.com", APIKey: "sk-test", Timeout: 30, Status: 1}
	require.NoError(t, db.Create(&up).Error)

	// 高优先级 + 官方
	i1 := ModelInstance{
		BaseModelID: b.ID, VendorID: v.ID, UpstreamID: up.ID,
		InstanceName: "official", ProviderModelCode: "gpt-3.5-turbo", Weight: 10, Priority: 1, IsOfficial: 1, Status: 1,
	}
	// 同优先级非官方
	i2 := ModelInstance{
		BaseModelID: b.ID, VendorID: v.ID, UpstreamID: up.ID,
		InstanceName: "3rd", ProviderModelCode: "gpt-3.5-turbo", Weight: 100, Priority: 1, IsOfficial: 0, Status: 1,
	}
	require.NoError(t, db.Create(&i1).Error)
	require.NoError(t, db.Create(&i2).Error)

	res, err := PickChatTarget(db, "gpt-3.5-turbo")
	require.NoError(t, err)
	require.Equal(t, int64(i1.ID), res.InstanceID)
}

func TestRewriteChatBodyToProvider(t *testing.T) {
	in := []byte(`{"model":"gpt-3.5-turbo","messages":[]}`)
	out := RewriteChatBodyToProvider(in, "upstream-real")
	require.Contains(t, string(out), `"model":"upstream-real"`)
}
