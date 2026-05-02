// Package ipaccess 提供 IP 白/黑名单的 CIDR 解析与匹配（IPv4/IPv6）。
package ipaccess

import (
	"fmt"
	"net"
	"strings"
)

// ParseCIDR 规范化并解析单条 CIDR 或裸 IP（视为 /32 或 /128）。
func ParseCIDR(s string) (*net.IPNet, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("ipaccess: 空 CIDR")
	}
	if strings.Contains(s, "/") {
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			return nil, err
		}
		return n, nil
	}
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, fmt.Errorf("ipaccess: 非法 IP %q", s)
	}
	if ip4 := ip.To4(); ip4 != nil {
		return &net.IPNet{IP: ip4, Mask: net.CIDRMask(32, 32)}, nil
	}
	return &net.IPNet{IP: ip.To16(), Mask: net.CIDRMask(128, 128)}, nil
}

// Compile 将 CIDR 字符串列表编译为 IPNet（遇错即返回）。
func Compile(cidrs []string) ([]*net.IPNet, error) {
	out := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		n, err := ParseCIDR(c)
		if err != nil {
			return nil, fmt.Errorf("ipaccess: %q: %w", c, err)
		}
		out = append(out, n)
	}
	return out, nil
}

// Match 判断 IP 是否命中任一 CIDR。
func Match(host string, nets []*net.IPNet) bool {
	ip := net.ParseIP(strings.TrimSpace(host))
	if ip == nil {
		return false
	}
	for _, n := range nets {
		if n != nil && n.Contains(ip) {
			return true
		}
	}
	return false
}
