package keystore

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestLoadRecordsFromFile_ParseAndExpiry(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "keys.json")
	raw := `[
	  {"id":"a","secret":"sk-live","disabled":false,"expires_at":null},
	  {"id":"b","secret":"sk-old","disabled":false,"expires_at":"2000-01-01T00:00:00Z"}
	]`
	require.NoError(t, os.WriteFile(p, []byte(raw), 0o600))

	recs, err := LoadRecordsFromFile(p)
	require.NoError(t, err)
	require.Len(t, recs, 2)
	require.Equal(t, "sk-live", recs[0].Secret)
	require.Nil(t, recs[0].ExpiresAt)
	require.Equal(t, "sk-old", recs[1].Secret)
	require.NotNil(t, recs[1].ExpiresAt)

	s := &Store{log: zap.NewNop(), path: p}
	s.replace(recs)
	require.True(t, s.ValidateBearer("sk-live"))
	require.False(t, s.ValidateBearer("sk-old"))
}

func TestNew_LegacyEnvKeys(t *testing.T) {
	cfg := &config.Config{GatewayAPIKeys: []string{"  a ", "b"}}
	s, err := New(cfg, zap.NewNop(), nil)
	require.NoError(t, err)
	require.True(t, s.ValidateBearer("a"))
	require.True(t, s.ValidateBearer("b"))
	require.False(t, s.ValidateBearer("c"))
}

func TestValidateBearer_DisabledAndExpiry(t *testing.T) {
	future := time.Now().UTC().Add(time.Hour).Format(time.RFC3339Nano)
	past := time.Now().UTC().Add(-time.Hour).Format(time.RFC3339Nano)
	dir := t.TempDir()
	path := filepath.Join(dir, "k.json")
	raw := `[
	  {"id":"1","secret":"ok","disabled":false,"expires_at":"` + future + `"},
	  {"id":"2","secret":"off","disabled":true,"expires_at":null},
	  {"id":"3","secret":"late","disabled":false,"expires_at":"` + past + `"}
	]`
	require.NoError(t, os.WriteFile(path, []byte(raw), 0o600))

	recs, err := LoadRecordsFromFile(path)
	require.NoError(t, err)
	require.Len(t, recs, 3)

	s := &Store{log: zap.NewNop(), path: path}
	s.replace(recs)
	require.True(t, s.ValidateBearer("ok"))
	require.False(t, s.ValidateBearer("off"))
	require.False(t, s.ValidateBearer("late"))
}
