package ipManager

import (
	"maps"
	"net/netip"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAddAndIsBannedIp(t *testing.T) {
	var m bannedIpsMap = make(map[netip.Addr]int)
	ip := netip.MustParseAddr("203.0.113.1")

	// Initially not banned
	if m.IsBannedIp(ip) {
		t.Errorf("expected IP %v to not be banned", ip)
	}

	// Add to ban with expiration
	exp := 12345
	m.AddIpToBan(ip, exp)

	// Should be banned now
	if !m.IsBannedIp(ip) {
		t.Errorf("expected IP %v to be banned", ip)
	}

	// Check expiration value stored
	if got := m[ip]; got != exp {
		t.Errorf("expected expiration %d, got %d", exp, got)
	}
}

func TestLoadBannedIp(t *testing.T) {
	// Set up sqlmock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock: %v", err)
	}
	defer db.Close()

	// Prepare mock rows
	rows := sqlmock.NewRows([]string{"ip", "unix_time"}).
		AddRow("198.51.100.42/32", 99999).
		AddRow("2001:db8::1/128", 123456)

	eq := mock.ExpectQuery(`SELECT ip, unix_time FROM ip_ban WHERE unix_time > extract\('epoch' from now\(\)\)::bigint`)
	eq.WillReturnRows(rows)

	mgr, err := LoadBannedIp(db)
	if err != nil {
		t.Fatalf("LoadBannedIp returned error: %v", err)
	}

	// Type assertion
	m, ok := mgr.(bannedIpsMap)
	if !ok {
		t.Fatalf("expected bannedIpsMap, got %T", mgr)
	}

	// Expected IPs
	want := map[netip.Addr]int{
		netip.MustParseAddr("198.51.100.42"): 99999,
		netip.MustParseAddr("2001:db8::1"):   123456,
	}

	if !maps.Equal(m, want) {
		t.Errorf("loaded map mismatch:\n got  %v\n want %v", m, want)
	}

	// Ensure expectations met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func BenchmarkAddIpToBan(b *testing.B) {
	m := make(map[netip.Addr]int, b.N)
	//base := netip.MustParseAddr("192.0.2.1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// invent an IP by flipping low bits
		addr := netip.AddrFrom4([4]byte{192, 0, 2, byte(i & 0xFF)})
		m[addr] = i
	}
}

func BenchmarkIsBannedIp(b *testing.B) {
	m := make(map[netip.Addr]int, b.N)
	ip := netip.MustParseAddr("203.0.113.5")
	// prepopulate
	m[ip] = 1
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m[ip]
	}
}
