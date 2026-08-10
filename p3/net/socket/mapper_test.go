package socket

import (
	"net"
	"testing"

	lookup "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_sockets_ip_name_lookup"
	sockets "github.com/jamesstocktonj1/componentize-sdk/p3/gen/wasi_sockets_types"
)

func TestParseIpAddress(t *testing.T) {
	t.Run("ipv4", func(t *testing.T) {
		addr, ok := parseIpAddress("192.168.1.42")
		if !ok {
			t.Fatalf("parseIpAddress() ok = false, want true")
		}
		if addr.Tag() != sockets.IpAddressIpv4 {
			t.Fatalf("parseIpAddress() tag = %v, want ipv4", addr.Tag())
		}
		want := sockets.Ipv4Address{F0: 192, F1: 168, F2: 1, F3: 42}
		if addr.Ipv4() != want {
			t.Errorf("parseIpAddress() = %+v, want %+v", addr.Ipv4(), want)
		}
	})

	t.Run("ipv6", func(t *testing.T) {
		addr, ok := parseIpAddress("2001:db8::abcd")
		if !ok {
			t.Fatalf("parseIpAddress() ok = false, want true")
		}
		if addr.Tag() != sockets.IpAddressIpv6 {
			t.Fatalf("parseIpAddress() tag = %v, want ipv6", addr.Tag())
		}
		want := sockets.Ipv6Address{F0: 0x2001, F1: 0x0db8, F2: 0, F3: 0, F4: 0, F5: 0, F6: 0, F7: 0xabcd}
		if addr.Ipv6() != want {
			t.Errorf("parseIpAddress() = %+v, want %+v", addr.Ipv6(), want)
		}
	})

	t.Run("hostname is not an IP literal", func(t *testing.T) {
		if _, ok := parseIpAddress("example.com"); ok {
			t.Errorf("parseIpAddress(\"example.com\") ok = true, want false")
		}
	})
}

func TestMapAddressFamily(t *testing.T) {
	tests := []struct {
		name string
		host string
		want sockets.IpAddressFamily
	}{
		{"ipv4", "10.0.0.1", sockets.IpAddressFamilyIpv4},
		{"ipv6", "::1", sockets.IpAddressFamilyIpv6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, ok := parseIpAddress(tt.host)
			if !ok {
				t.Fatalf("parseIpAddress(%q) ok = false", tt.host)
			}
			if got := mapAddressFamily(addr); got != tt.want {
				t.Errorf("mapAddressFamily() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapIpSocketAddress(t *testing.T) {
	t.Run("ipv4", func(t *testing.T) {
		addr, _ := parseIpAddress("127.0.0.1")
		got := mapIpSocketAddress(addr, 8080)
		if got.Tag() != sockets.IpSocketAddressIpv4 {
			t.Fatalf("mapIpSocketAddress() tag = %v, want ipv4", got.Tag())
		}
		v4 := got.Ipv4()
		if v4.Port != 8080 || v4.Address != (sockets.Ipv4Address{F0: 127, F1: 0, F2: 0, F3: 1}) {
			t.Errorf("mapIpSocketAddress() = %+v, want port 8080 addr 127.0.0.1", v4)
		}
	})

	t.Run("ipv6", func(t *testing.T) {
		addr, _ := parseIpAddress("::1")
		got := mapIpSocketAddress(addr, 443)
		if got.Tag() != sockets.IpSocketAddressIpv6 {
			t.Fatalf("mapIpSocketAddress() tag = %v, want ipv6", got.Tag())
		}
		v6 := got.Ipv6()
		if v6.Port != 443 || v6.Address != (sockets.Ipv6Address{F0: 0, F1: 0, F2: 0, F3: 0, F4: 0, F5: 0, F6: 0, F7: 1}) {
			t.Errorf("mapIpSocketAddress() = %+v, want port 443 addr ::1", v6)
		}
	})
}

func TestMapNetAddr(t *testing.T) {
	t.Run("ipv4", func(t *testing.T) {
		addr, _ := parseIpAddress("203.0.113.5")
		sockAddr := mapIpSocketAddress(addr, 9000)

		got, ok := mapNetAddr(sockAddr).(*net.TCPAddr)
		if !ok {
			t.Fatalf("mapNetAddr() did not return *net.TCPAddr")
		}
		if !got.IP.Equal(net.IPv4(203, 0, 113, 5)) || got.Port != 9000 {
			t.Errorf("mapNetAddr() = %v, want 203.0.113.5:9000", got)
		}
	})

	t.Run("ipv6", func(t *testing.T) {
		addr, _ := parseIpAddress("fe80::1")
		sockAddr := mapIpSocketAddress(addr, 22)

		got, ok := mapNetAddr(sockAddr).(*net.TCPAddr)
		if !ok {
			t.Fatalf("mapNetAddr() did not return *net.TCPAddr")
		}
		if !got.IP.Equal(net.ParseIP("fe80::1")) || got.Port != 22 {
			t.Errorf("mapNetAddr() = %v, want fe80::1:22", got)
		}
	})
}

func TestMapIP(t *testing.T) {
	tests := []struct {
		name string
		host string
	}{
		{"ipv4", "8.8.8.8"},
		{"ipv6", "2001:4860:4860::8888"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, ok := parseIpAddress(tt.host)
			if !ok {
				t.Fatalf("parseIpAddress(%q) ok = false", tt.host)
			}
			want := net.ParseIP(tt.host)
			if got := mapIP(addr); !got.Equal(want) {
				t.Errorf("mapIP() = %v, want %v", got, want)
			}
		})
	}
}

// TestIpAddressRoundTrip verifies parseIpAddress and mapIP are inverses of
// one another for both address families.
func TestIpAddressRoundTrip(t *testing.T) {
	ips := []string{
		"0.0.0.0",
		"255.255.255.255",
		"192.168.1.1",
		"::",
		"::1",
		"2001:db8::abcd:1234",
	}

	for _, ip := range ips {
		t.Run(ip, func(t *testing.T) {
			want := net.ParseIP(ip)
			addr, ok := parseIpAddress(ip)
			if !ok {
				t.Fatalf("parseIpAddress(%q) ok = false", ip)
			}
			if got := mapIP(addr); !got.Equal(want) {
				t.Errorf("round trip for %s = %v, want %v", ip, got, want)
			}
		})
	}
}

func TestMapLookupErrorCode(t *testing.T) {
	tests := []struct {
		name string
		code lookup.ErrorCode
	}{
		{"access denied", lookup.MakeErrorCodeAccessDenied()},
		{"name unresolvable", lookup.MakeErrorCodeNameUnresolvable()},
		{"temporary resolver failure", lookup.MakeErrorCodeTemporaryResolverFailure()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := mapLookupErrorCode(tt.code); err == nil {
				t.Errorf("mapLookupErrorCode() = nil, want non-nil error")
			}
		})
	}
}
