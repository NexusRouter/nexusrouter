package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/buildinfo"
)

const (
	envHTTPListenAddr = "NEXUSROUTER_HTTP_LISTEN_ADDR"
	envPORT           = "PORT"
	envLogDir         = "NEXUSROUTER_LOG_DIR"
)

type earlyCLIMode int

const (
	earlyCLINone earlyCLIMode = iota
	earlyCLIVersion
	earlyCLIHelp
)

// parseStartupCLI 解析 **-version**、**-help**、**-h**、**-port**、**-log-dir**。**-version** 优先于帮助类标志。
// 返回的端口与日志目录供在 **LoadDotEnv** 之后调用 **applyCLIPort** / **applyCLILogDir**，以便 **`.env`** 优先于命令行。
func parseStartupCLI(argv []string, errOut io.Writer) (earlyCLIMode, int, string, error) {
	name := "gateway"
	if len(argv) > 0 && argv[0] != "" {
		name = filepath.Base(argv[0])
	}
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(errOut)
	showVer := fs.Bool("version", false, "print version and exit")
	showHelp := fs.Bool("help", false, "print usage and exit")
	showHelpShort := fs.Bool("h", false, "print usage and exit (same as -help)")
	cliPort := fs.Int("port", 0, "when listen addr and PORT are unset, set PORT to this port (1-65535)")
	cliLogDir := fs.String("log-dir", "", "when NEXUSROUTER_LOG_DIR is unset, append JSON logs under this directory")
	if e := fs.Parse(argv[1:]); e != nil {
		return earlyCLINone, 0, "", e
	}
	if *showVer {
		return earlyCLIVersion, 0, "", nil
	}
	if *showHelp || *showHelpShort {
		return earlyCLIHelp, 0, "", nil
	}
	return earlyCLINone, *cliPort, strings.TrimSpace(*cliLogDir), nil
}

// exitIfCLIHandled 仅处理 **-version** / **-help** / **-h**（在加载 **`.env`** 之前），与 **parseStartupCLI** 共用同一套标志定义。
func exitIfCLIHandled(argv []string, out io.Writer, errOut io.Writer) (handled bool, err error) {
	mode, _, _, err := parseStartupCLI(argv, errOut)
	if err != nil {
		return false, err
	}
	switch mode {
	case earlyCLIVersion:
		_, err = fmt.Fprintln(out, buildinfo.Version)
		return true, err
	case earlyCLIHelp:
		writeCLIHelp(argv, out)
		return true, nil
	default:
		return false, nil
	}
}

// applyCLIPort 在 **config.LoadDotEnv** 之后调用：当 **NEXUSROUTER_HTTP_LISTEN_ADDR** 与 **PORT** 在环境中仍均为空时，将 **PORT** 设为命令行 **-port**。
func applyCLIPort(port int) error {
	if port == 0 {
		return nil
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid -port %d: must be between 1 and 65535", port)
	}
	if strings.TrimSpace(os.Getenv(envHTTPListenAddr)) != "" {
		return nil
	}
	if strings.TrimSpace(os.Getenv(envPORT)) != "" {
		return nil
	}
	return os.Setenv(envPORT, strconv.Itoa(port))
}

// applyCLILogDir 在 **config.LoadDotEnv** 之后调用：当 **NEXUSROUTER_LOG_DIR** 在环境中仍为空时，将其设为 **-log-dir** 的绝对路径。
func applyCLILogDir(dir string) error {
	if strings.TrimSpace(dir) == "" {
		return nil
	}
	if strings.TrimSpace(os.Getenv(envLogDir)) != "" {
		return nil
	}
	abs, err := filepath.Abs(strings.TrimSpace(dir))
	if err != nil {
		return fmt.Errorf("invalid -log-dir: %w", err)
	}
	return os.Setenv(envLogDir, abs)
}

func writeCLIHelp(argv []string, out io.Writer) {
	bin := "gateway"
	if len(argv) > 0 && argv[0] != "" {
		bin = filepath.Base(argv[0])
	}
	var b strings.Builder
	fmt.Fprintf(&b, "OpenAI 兼容 Chat Completions 网关。\n\n")
	fmt.Fprintf(&b, "用法:\n  %s [选项]\n\n", bin)
	fmt.Fprintf(&b, "选项:\n")
	fmt.Fprintf(&b, "  -version   打印版本并退出\n")
	fmt.Fprintf(&b, "  -h         打印本说明并退出\n")
	fmt.Fprintf(&b, "  -help      同 -h\n")
	fmt.Fprintf(&b, "  -port N    当未设置 %s 且环境 PORT 为空时，将监听端口设为 N\n", envHTTPListenAddr)
	fmt.Fprintf(&b, "  -log-dir D  当未设置 %s 时，将持久化日志目录设为 D（默认 gateway.log；NEXUSROUTER_LOG_DAILY_FILE=true 时为 gateway-YYYYMMDD.log）\n", envLogDir)
	_, _ = io.WriteString(out, b.String())
}
