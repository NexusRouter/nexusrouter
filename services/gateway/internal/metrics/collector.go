// Package metrics 提供进程内网关业务指标聚合（线程安全、有界）。
package metrics

import (
	"strconv"
	"sync"
	"time"
)

const maxErrCodes = 64

// Collector 进程内 Chat 路径与其它网关错误的轻量统计。
type Collector struct {
	mu sync.Mutex

	totalReq   uint64
	successReq uint64
	sumLatMs   int64
	nLat       int64

	errMu     sync.Mutex
	errCounts map[string]uint64

	// 按 UTC 日界滚动的今日/昨日请求量（用于对比）。
	dayUTC           string
	todayReq         uint64
	todaySuccess     uint64
	yesterdayReq     uint64
	yesterdaySucc    uint64
	todayErrByCode   map[string]uint64
	prevDayErrByCode map[string]uint64

	// 最近若干整秒的请求计数（用于估算 RPS，键为 Unix 秒）。
	perSecond map[int64]uint64
}

// NewCollector 构造采集器。
func NewCollector() *Collector {
	return &Collector{
		errCounts:        make(map[string]uint64),
		todayErrByCode:   make(map[string]uint64),
		prevDayErrByCode: make(map[string]uint64),
		perSecond:        make(map[int64]uint64),
		dayUTC:           time.Now().UTC().Format("2006-01-02"),
	}
}

func (c *Collector) rollDayIfNeeded(now time.Time) {
	day := now.UTC().Format("2006-01-02")
	if c.dayUTC == day {
		return
	}
	c.yesterdayReq = c.todayReq
	c.yesterdaySucc = c.todaySuccess
	c.prevDayErrByCode = c.todayErrByCode
	c.todayReq = 0
	c.todaySuccess = 0
	c.todayErrByCode = make(map[string]uint64)
	c.dayUTC = day
}

func bumpErrMap(m map[string]uint64, code string) {
	if code == "" {
		code = "UNKNOWN"
	}
	if len(m) >= maxErrCodes && m[code] == 0 {
		m["OTHER"]++
		return
	}
	m[code]++
}

// RecordChat 记录一次 Chat 代理路径完成后的结果（defer 调用）。
func (c *Collector) RecordChat(httpStatus int, durationMs int64, gatewayCode string) {
	if c == nil {
		return
	}
	now := time.Now()
	c.mu.Lock()
	c.rollDayIfNeeded(now)
	c.totalReq++
	c.todayReq++
	sec := now.Unix()
	c.perSecond[sec]++
	for k := range c.perSecond {
		if k < sec-120 {
			delete(c.perSecond, k)
		}
	}

	if httpStatus >= 200 && httpStatus < 300 {
		c.successReq++
		c.todaySuccess++
	} else {
		code := gatewayCode
		if code == "" {
			code = "HTTP_" + strconv.Itoa(httpStatus)
		}
		c.errMu.Lock()
		bumpErrMap(c.errCounts, code)
		c.errMu.Unlock()
		bumpErrMap(c.todayErrByCode, code)
	}
	if durationMs >= 0 {
		c.sumLatMs += durationMs
		c.nLat++
	}
	c.mu.Unlock()
}

// RecordGatewayError 记录未进入 Chat 代理的网关错误（鉴权、限流等）。
func (c *Collector) RecordGatewayError(gatewayCode string) {
	if c == nil {
		return
	}
	now := time.Now()
	c.mu.Lock()
	c.rollDayIfNeeded(now)
	c.totalReq++
	c.todayReq++
	sec := now.Unix()
	c.perSecond[sec]++
	for k := range c.perSecond {
		if k < sec-120 {
			delete(c.perSecond, k)
		}
	}
	code := gatewayCode
	if code == "" {
		code = "UNKNOWN"
	}
	c.errMu.Lock()
	bumpErrMap(c.errCounts, code)
	c.errMu.Unlock()
	bumpErrMap(c.todayErrByCode, code)
	c.mu.Unlock()
}

// SummaryJSON 返回供管理 API 使用的聚合视图。
func (c *Collector) SummaryJSON() map[string]any {
	if c == nil {
		return map[string]any{}
	}
	now := time.Now()
	c.mu.Lock()
	c.rollDayIfNeeded(now)
	total := c.totalReq
	ok := c.successReq
	sum := c.sumLatMs
	n := c.nLat
	today := c.todayReq
	todayOK := c.todaySuccess
	yest := c.yesterdayReq
	yestOK := c.yesterdaySucc
	nowSec := now.Unix()
	var last60 uint64
	for k, v := range c.perSecond {
		if k >= nowSec-60 {
			last60 += v
		}
	}
	c.mu.Unlock()

	c.errMu.Lock()
	ec := cloneErr(c.errCounts)
	c.errMu.Unlock()

	c.mu.Lock()
	te := cloneErr(c.todayErrByCode)
	pe := cloneErr(c.prevDayErrByCode)
	c.mu.Unlock()

	avg := float64(0)
	if n > 0 {
		avg = float64(sum) / float64(n)
	}
	rate := float64(last60) / 60.0
	successRate := float64(0)
	if total > 0 {
		successRate = float64(ok) / float64(total)
	}

	return map[string]any{
		"server_time":              now.UTC().Format(time.RFC3339Nano),
		"online":                   true,
		"requests_total":           total,
		"success_total":            ok,
		"success_rate":             successRate,
		"avg_latency_ms":           avg,
		"current_rps_estimate":     rate,
		"requests_today":           today,
		"requests_yesterday":       yest,
		"success_today":            todayOK,
		"success_yesterday":        yestOK,
		"errors_by_code":           ec,
		"errors_today_by_code":     te,
		"errors_yesterday_by_code": pe,
	}
}

func cloneErr(m map[string]uint64) map[string]uint64 {
	out := make(map[string]uint64, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
