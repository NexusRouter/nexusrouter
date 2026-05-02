package accesslog

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/NexusRouter/nexusrouter/services/gateway/internal/runtime"
)

// QueryLimits 单次查询资源上限。
type QueryLimits struct {
	MaxScanBytes int64 // 所有文件合计最多读取的字节数
	MaxLines     int   // 最多返回条数（硬上限）
}

// DefaultQueryLimits 默认保护值。
func DefaultQueryLimits() QueryLimits {
	return QueryLimits{MaxScanBytes: 10 << 20, MaxLines: 500}
}

// LogFilters 与代理访问日志字段 AND 组合筛选。
type LogFilters struct {
	FromRFC3339  string
	ToRFC3339    string
	PathPrefix   string
	StatusMin    int
	StatusMax    int
	APIKeyFP     string
	ClientIP     string
	Limit        int
	Cursor       string
	MaxScanBytes int64
}

// LogRow 单条 JSON 日志解析后的扁平视图（供 API / CSV）。
type LogRow struct {
	Raw map[string]interface{}
}

// DiscoverLogFiles 列出主日志文件及同目录下 lumberjack 轮转文件（不含 .gz）。
func DiscoverLogFiles(primary string) ([]string, error) {
	primary = strings.TrimSpace(primary)
	if primary == "" {
		return nil, fmt.Errorf("accesslog: 未配置日志路径")
	}
	dir := filepath.Dir(primary)
	base := filepath.Base(primary)
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("accesslog: 列出目录: %w", err)
	}
	type fi struct {
		path string
		mt   time.Time
	}
	var out []fi
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if name != base && !strings.HasPrefix(name, base+"-") {
			continue
		}
		if strings.HasSuffix(name, ".gz") {
			continue
		}
		full := filepath.Join(dir, name)
		st, err := os.Stat(full)
		if err != nil || !st.Mode().IsRegular() {
			continue
		}
		out = append(out, fi{path: full, mt: st.ModTime()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].mt.After(out[j].mt) })
	paths := make([]string, 0, len(out))
	for _, x := range out {
		paths = append(paths, x.path)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("accesslog: 未找到日志文件")
	}
	return paths, nil
}

func readFileTail(path string, maxBytes int64) ([]byte, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	sz := st.Size()
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	if sz <= maxBytes {
		return io.ReadAll(f)
	}
	if _, err := f.Seek(sz-maxBytes, 0); err != nil {
		return nil, err
	}
	buf, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if i := bytes.IndexByte(buf, '\n'); i >= 0 && i+1 < len(buf) {
		buf = buf[i+1:]
	}
	return buf, nil
}

func parseLogTime(v interface{}) (time.Time, bool) {
	switch t := v.(type) {
	case string:
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
			if tt, err := time.Parse(layout, t); err == nil {
				return tt.UTC(), true
			}
		}
	case json.Number:
		if ms, err := t.Int64(); err == nil {
			return time.UnixMilli(ms).UTC(), true
		}
	case float64:
		return time.UnixMilli(int64(t)).UTC(), true
	}
	return time.Time{}, false
}

func rowMatches(m map[string]interface{}, f LogFilters, fromT, toT time.Time, hasFrom, hasTo bool) bool {
	if hasFrom || hasTo {
		ts, ok := parseLogTime(m["ts"])
		if !ok {
			return false
		}
		if hasFrom && ts.Before(fromT) {
			return false
		}
		if hasTo && !ts.Before(toT) {
			return false
		}
	}
	if pfx := strings.TrimSpace(f.PathPrefix); pfx != "" {
		path, _ := m["path"].(string)
		if !strings.HasPrefix(path, pfx) {
			return false
		}
	}
	if f.StatusMin > 0 || f.StatusMax > 0 {
		st := 0
		switch v := m["status"].(type) {
		case float64:
			st = int(v)
		case json.Number:
			st, _ = strconv.Atoi(v.String())
		}
		if f.StatusMin > 0 && st < f.StatusMin {
			return false
		}
		if f.StatusMax > 0 && st > f.StatusMax {
			return false
		}
	}
	if want := strings.TrimSpace(f.APIKeyFP); want != "" {
		got, _ := m["api_key_fp"].(string)
		if got != want {
			return false
		}
	}
	if want := strings.TrimSpace(f.ClientIP); want != "" {
		got, _ := m["client_ip"].(string)
		if got != want {
			return false
		}
	}
	return true
}

type cursorPayload struct {
	O int `json:"o"` // 已返回条数累计偏移（基于本次扫描的排序结果）
}

// QueryJSONLines 从快照配置读取日志尾部，合并多文件后按时间逆序分页。
// 当返回 truncated 为 true 时表示因行数上限未读完所有候选行，结果可能不完整。
func QueryJSONLines(ctx context.Context, snap *runtime.Snapshot, f LogFilters) ([]map[string]interface{}, string, bool, error) {
	if snap == nil || !snap.ProxyAccessLog.Enabled {
		return nil, "", false, fmt.Errorf("accesslog: proxy_access_log 未启用")
	}
	path := strings.TrimSpace(snap.ProxyAccessLog.Path)
	if path == "" {
		return nil, "", false, fmt.Errorf("accesslog: 未配置日志文件路径")
	}
	files, err := DiscoverLogFiles(path)
	if err != nil {
		return nil, "", false, err
	}
	maxScan := f.MaxScanBytes
	if maxScan <= 0 {
		maxScan = DefaultQueryLimits().MaxScanBytes
	}
	limit := f.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > DefaultQueryLimits().MaxLines {
		limit = DefaultQueryLimits().MaxLines
	}
	perFile := maxScan / int64(max(1, len(files)))
	if perFile < 1<<16 {
		perFile = 1 << 16
	}

	var fromT, toT time.Time
	var hasFrom, hasTo bool
	if strings.TrimSpace(f.FromRFC3339) != "" {
		t, err := time.Parse(time.RFC3339Nano, f.FromRFC3339)
		if err != nil {
			t, err = time.Parse(time.RFC3339, f.FromRFC3339)
		}
		if err != nil {
			return nil, "", false, fmt.Errorf("accesslog: from 时间非法: %w", err)
		}
		fromT = t.UTC()
		hasFrom = true
	}
	if strings.TrimSpace(f.ToRFC3339) != "" {
		t, err := time.Parse(time.RFC3339Nano, f.ToRFC3339)
		if err != nil {
			t, err = time.Parse(time.RFC3339, f.ToRFC3339)
		}
		if err != nil {
			return nil, "", false, fmt.Errorf("accesslog: to 时间非法: %w", err)
		}
		toT = t.UTC()
		hasTo = true
	}

	offset := 0
	if strings.TrimSpace(f.Cursor) != "" {
		var cd cursorPayload
		if err := json.Unmarshal([]byte(f.Cursor), &cd); err == nil && cd.O > 0 {
			offset = cd.O
		}
	}

	type hit struct {
		ts time.Time
		m  map[string]interface{}
	}
	var hits []hit
	const maxParsedLines = 20000
	parsed := 0

fileLoop:
	for _, fp := range files {
		select {
		case <-ctx.Done():
			return nil, "", false, ctx.Err()
		default:
		}
		raw, err := readFileTail(fp, perFile)
		if err != nil {
			return nil, "", false, err
		}
		for _, line := range bytes.Split(raw, []byte{'\n'}) {
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			parsed++
			if parsed > maxParsedLines {
				break fileLoop
			}
			var m map[string]interface{}
			dec := json.NewDecoder(bytes.NewReader(line))
			dec.UseNumber()
			if err := dec.Decode(&m); err != nil {
				continue
			}
			if _, ok := m["path"]; !ok {
				continue
			}
			if !rowMatches(m, f, fromT, toT, hasFrom, hasTo) {
				continue
			}
			ts, ok := parseLogTime(m["ts"])
			if !ok {
				continue
			}
			hits = append(hits, hit{ts: ts, m: m})
		}
	}

	sort.SliceStable(hits, func(i, j int) bool { return hits[i].ts.After(hits[j].ts) })
	if offset > len(hits) {
		offset = len(hits)
	}
	end := offset + limit
	if end > len(hits) {
		end = len(hits)
	}
	page := hits[offset:end]
	out := make([]map[string]interface{}, len(page))
	for i := range page {
		out[i] = page[i].m
	}
	next := ""
	if end < len(hits) {
		b, _ := json.Marshal(cursorPayload{O: end})
		next = string(b)
	}
	truncated := parsed >= maxParsedLines
	return out, next, truncated, nil
}

// WriteCSV 将行写入 w（UTF-8），列不含 Authorization 明文。
func WriteCSV(w io.Writer, rows []map[string]interface{}) error {
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"ts", "request_id", "method", "path", "client_ip", "status", "duration_ms", "upstream_id", "upstream_host", "api_key_fp"})
	for _, m := range rows {
		line := []string{
			fmt.Sprint(m["ts"]),
			fmt.Sprint(m["request_id"]),
			fmt.Sprint(m["method"]),
			fmt.Sprint(m["path"]),
			fmt.Sprint(m["client_ip"]),
			fmt.Sprint(m["status"]),
			fmt.Sprint(m["duration_ms"]),
			fmt.Sprint(m["upstream_id"]),
			fmt.Sprint(m["upstream_host"]),
			fmt.Sprint(m["api_key_fp"]),
		}
		if err := cw.Write(line); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
