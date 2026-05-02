package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv_FromCWD(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(`
# c
NEXUSROUTER_UPSTREAM_HTTP_PROXY=http://dotenv.local:7
export NEXUSROUTER_HTTP_LISTEN_ADDR=:7777
`), 0o644); err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	_ = os.Unsetenv("NEXUSROUTER_UPSTREAM_HTTP_PROXY")
	_ = os.Unsetenv("NEXUSROUTER_HTTP_LISTEN_ADDR")

	LoadDotEnv()
	cfg := Load()
	if cfg.UpstreamHTTPProxy != "http://dotenv.local:7" {
		t.Fatalf("UpstreamHTTPProxy: got %q", cfg.UpstreamHTTPProxy)
	}
	if cfg.HTTPListenAddr != ":7777" {
		t.Fatalf("HTTPListenAddr: got %q", cfg.HTTPListenAddr)
	}
}

func TestLoadDotEnv_SkipsWhenEnvNonEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("NEXUSROUTER_UPSTREAM_HTTP_PROXY=http://from-file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	t.Setenv("NEXUSROUTER_UPSTREAM_HTTP_PROXY", "http://from-os")
	LoadDotEnv()
	cfg := Load()
	if cfg.UpstreamHTTPProxy != "http://from-os" {
		t.Fatalf("UpstreamHTTPProxy: got %q", cfg.UpstreamHTTPProxy)
	}
}

func TestLoadDotEnv_QuotedValue(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(`NEXUSROUTER_UPSTREAM_HTTP_PROXY="http://q:1"`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	_ = os.Unsetenv("NEXUSROUTER_UPSTREAM_HTTP_PROXY")
	LoadDotEnv()
	cfg := Load()
	if cfg.UpstreamHTTPProxy != "http://q:1" {
		t.Fatalf("UpstreamHTTPProxy: got %q", cfg.UpstreamHTTPProxy)
	}
}
