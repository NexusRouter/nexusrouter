package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/buildinfo"
	"github.com/stretchr/testify/require"
)

func TestExitIfCLIHandled_version(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	ok, err := exitIfCLIHandled([]string{"gateway", "-version"}, &out, &errOut)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, buildinfo.Version+"\n", out.String())
	require.Empty(t, errOut.String())
}

func TestExitIfCLIHandled_noFlag(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	ok, err := exitIfCLIHandled([]string{"gateway"}, &out, &errOut)
	require.NoError(t, err)
	require.False(t, ok)
	require.Empty(t, out.String())
}

func TestExitIfCLIHandled_help(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	ok, err := exitIfCLIHandled([]string{"gateway", "-help"}, &out, &errOut)
	require.NoError(t, err)
	require.True(t, ok)
	require.Contains(t, out.String(), "用法:")
	require.Contains(t, out.String(), "-version")
	require.Empty(t, errOut.String())
}

func TestExitIfCLIHandled_h(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	ok, err := exitIfCLIHandled([]string{"/bin/nexusrouter-api", "-h"}, &out, &errOut)
	require.NoError(t, err)
	require.True(t, ok)
	require.Contains(t, out.String(), "nexusrouter-api")
	require.True(t, strings.Contains(out.String(), "-help") || strings.Contains(out.String(), "同 -h"))
	require.Empty(t, errOut.String())
}

func TestExitIfCLIHandled_versionWinsOverHelp(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	ok, err := exitIfCLIHandled([]string{"gateway", "-version", "-help"}, &out, &errOut)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, buildinfo.Version+"\n", out.String())
}

func TestParseStartupCLI_port(t *testing.T) {
	var errOut bytes.Buffer
	mode, port, logDir, err := parseStartupCLI([]string{"gateway", "-port", "3000"}, &errOut)
	require.NoError(t, err)
	require.Equal(t, earlyCLINone, mode)
	require.Equal(t, 3000, port)
	require.Empty(t, logDir)
	require.Empty(t, errOut.String())
}

func TestParseStartupCLI_logDir(t *testing.T) {
	var errOut bytes.Buffer
	mode, port, logDir, err := parseStartupCLI([]string{"gateway", "-log-dir", "/tmp/logs"}, &errOut)
	require.NoError(t, err)
	require.Equal(t, earlyCLINone, mode)
	require.Equal(t, 0, port)
	require.Equal(t, "/tmp/logs", logDir)
}

func TestApplyCLILogDir_setsEnvWhenEligible(t *testing.T) {
	t.Setenv(envLogDir, "")
	dir := t.TempDir()
	require.NoError(t, applyCLILogDir(dir))
	require.Equal(t, dir, os.Getenv(envLogDir))
}

func TestApplyCLILogDir_skipsWhenEnvSet(t *testing.T) {
	t.Setenv(envLogDir, "/existing")
	require.NoError(t, applyCLILogDir("/other"))
	require.Equal(t, "/existing", os.Getenv(envLogDir))
}

func TestApplyCLIPort_setsPORTWhenEligible(t *testing.T) {
	t.Setenv(envHTTPListenAddr, "")
	t.Setenv(envPORT, "")
	require.NoError(t, applyCLIPort(9090))
	require.Equal(t, "9090", os.Getenv(envPORT))
}

func TestApplyCLIPort_skipsWhenListenAddrSet(t *testing.T) {
	t.Setenv(envHTTPListenAddr, ":4000")
	t.Setenv(envPORT, "")
	require.NoError(t, applyCLIPort(9090))
	require.Empty(t, os.Getenv(envPORT))
}

func TestApplyCLIPort_skipsWhenPORTSet(t *testing.T) {
	t.Setenv(envHTTPListenAddr, "")
	t.Setenv(envPORT, "8080")
	require.NoError(t, applyCLIPort(9090))
	require.Equal(t, "8080", os.Getenv(envPORT))
}

func TestApplyCLIPort_invalid(t *testing.T) {
	t.Setenv(envHTTPListenAddr, "")
	t.Setenv(envPORT, "")
	require.Error(t, applyCLIPort(70000))
	require.Error(t, applyCLIPort(-1))
}
