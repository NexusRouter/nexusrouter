package ipaccess

import (
	"net"
	"testing"
)

func TestParseCIDR(t *testing.T) {
	n, err := ParseCIDR("203.0.113.10")
	if err != nil {
		t.Fatal(err)
	}
	if !n.Contains(net.ParseIP("203.0.113.10")) {
		t.Fatal("expected host in /32")
	}
	if n.Contains(net.ParseIP("203.0.113.11")) {
		t.Fatal("expected /32 exclude neighbor")
	}

	n2, err := ParseCIDR("203.0.113.0/24")
	if err != nil {
		t.Fatal(err)
	}
	if !n2.Contains(net.ParseIP("203.0.113.1")) {
		t.Fatal("expected /24")
	}
	if n2.Contains(net.ParseIP("198.51.100.1")) {
		t.Fatal("expected outside /24")
	}
}

func TestCompileMatch(t *testing.T) {
	nets, err := Compile([]string{"10.0.0.0/8", "192.168.1.5"})
	if err != nil {
		t.Fatal(err)
	}
	if !Match("10.1.2.3", nets) {
		t.Fatal("10.x should match")
	}
	if !Match("192.168.1.5", nets) {
		t.Fatal("single IP should match")
	}
	if Match("8.8.8.8", nets) {
		t.Fatal("8.8.8.8 should not match")
	}
}
