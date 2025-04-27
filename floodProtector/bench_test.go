package floodProtecor

import (
	"errors"
	"net"

	"testing"
	"time"
)

type fakeAcceptor struct {
	conn *net.TCPConn
	err  error
}

func (f *fakeAcceptor) AcceptTCP() (*net.TCPConn, error) {
	return f.conn, f.err
}

func clearStorage() {
	storage.Range(func(key, value interface{}) bool {
		storage.Delete(key)
		return true
	})
}

// createFakeConn starts a TCP listener on localhost and dials into it, returning the server side TCPConn
func createFakeConn(t testing.TB) (*net.TCPConn, func()) {
	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("ListenTCP failed: %v", err)
	}
	clientCh := make(chan *net.TCPConn, 1)
	go func() {
		client, err := net.DialTCP("tcp", nil, l.Addr().(*net.TCPAddr))
		if err != nil {
			t.Errorf("DialTCP failed: %v", err)
			return
		}
		clientCh <- client
	}()
	server, err := l.AcceptTCP()
	if err != nil {
		t.Fatalf("AcceptTCP on listener failed: %v", err)
	}
	client := <-clientCh
	cleanup := func() {
		client.Close()
		server.Close()
		l.Close()
	}
	return server, cleanup
}

func TestAcceptTCP_ErrorFromAcceptor(t *testing.T) {
	clearStorage()
	expectedErr := errors.New("test error")
	fa := &fakeAcceptor{conn: nil, err: expectedErr}
	conn, err := AcceptTCP(fa)
	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
	if conn != nil {
		t.Errorf("expected nil conn, got %v", conn)
	}
}

func TestAcceptTCP_FirstConnection(t *testing.T) {
	clearStorage()
	server, cleanup := createFakeConn(t)
	defer cleanup()
	fa := &fakeAcceptor{conn: server, err: nil}
	conn, err := AcceptTCP(fa)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if conn != server {
		t.Fatalf("expected returned conn to be server")
	}
	// check storage
	ip, _, _ := net.SplitHostPort(server.RemoteAddr().String())
	v, ok := storage.Load(ip)
	if !ok {
		t.Fatalf("expected storage entry for %s", ip)
	}
	ci := v.(connectionInfo)
	if ci.connCount != 1 {
		t.Errorf("expected connCount 1, got %d", ci.connCount)
	}
	if ci.state != StateNormal {
		t.Errorf("expected state StateNormal, got %v", ci.state)
	}
}

func TestAcceptTCP_StateTransitions(t *testing.T) {
	clearStorage()
	var server *net.TCPConn
	var cleanup func()
	// 3 quick connections to trigger StateWarn
	for i := 0; i < 3; i++ {
		server, cleanup = createFakeConn(t)
		fa := &fakeAcceptor{conn: server, err: nil}
		_, err := AcceptTCP(fa)
		cleanup()
		if err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	// after 3: state should be Warn
	ip, _, _ := net.SplitHostPort(server.RemoteAddr().String())
	v, ok := storage.Load(ip)
	if !ok {
		t.Fatalf("expected storage entry for %s", ip)
	}
	ci := v.(connectionInfo)
	if ci.state != StateWarn {
		t.Errorf("expected state StateWarn after 3 quick conns, got %v", ci.state)
	}
	// 4th connection triggers storage to Blocked but returns conn
	server, cleanup = createFakeConn(t)
	fa := &fakeAcceptor{conn: server, err: nil}
	conn, err := AcceptTCP(fa)
	cleanup()
	if err != nil {
		t.Fatalf("expected no error on 4th conn, got %v", err)
	}
	if conn != server {
		t.Fatalf("expected conn on 4th conn")
	}
	v2, _ := storage.Load(ip)
	ci2 := v2.(connectionInfo)
	if ci2.state != StateBlocked {
		t.Errorf("expected state StateBlocked after flooding, got %v", ci2.state)
	}
}

func TestAcceptTCP_BlockedBehavior(t *testing.T) {
	clearStorage()
	// prepare storage to be in Blocked state
	ip := "127.0.0.1"
	storage.Store(ip, connectionInfo{state: StateBlocked, blockExpire: time.Now().Add(1 * time.Minute)})
	server, cleanup := createFakeConn(t)
	defer cleanup()
	fa := &fakeAcceptor{conn: server, err: nil}
	conn, err := AcceptTCP(fa)
	if err == nil || err.Error() != "соединение закрыто FloodProtection" {
		t.Fatalf("expected FloodProtection error, got %v", err)
	}
	if conn != nil {
		t.Errorf("expected nil conn when blocked, got %v", conn)
	}
}

func TestAcceptTCP_ExpiredBlock(t *testing.T) {
	clearStorage()
	ip := "127.0.0.1"
	// expired block
	storage.Store(ip, connectionInfo{state: StateBlocked, blockExpire: time.Now().Add(-1 * time.Second)})
	// first call: still blocked due to old ci, but storage should update to Normal
	server1, cleanup1 := createFakeConn(t)
	defer cleanup1()
	fa1 := &fakeAcceptor{conn: server1, err: nil}
	conn1, err1 := AcceptTCP(fa1)
	if err1 == nil || err1.Error() != "соединение закрыто FloodProtection" {
		t.Fatalf("expected blocked error on expired block, got %v", err1)
	}
	if conn1 != nil {
		t.Errorf("expected nil conn when expired block, got %v", conn1)
	}
	// storage updated
	v, _ := storage.Load(ip)
	ci := v.(connectionInfo)
	if ci.state != StateNormal {
		t.Errorf("expected state StateNormal after expired block, got %v", ci.state)
	}
	// second call: should succeed
	server2, cleanup2 := createFakeConn(t)
	defer cleanup2()
	fa2 := &fakeAcceptor{conn: server2, err: nil}
	conn2, err2 := AcceptTCP(fa2)
	if err2 != nil {
		t.Fatalf("expected success after expired block, got %v", err2)
	}
	if conn2 != server2 {
		t.Errorf("expected returned conn to be server2")
	}
}

// Benchmarks for AcceptTCP

func BenchmarkAcceptTCP_Error(b *testing.B) {
	fa := &fakeAcceptor{conn: nil, err: errors.New("benchmark error")}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AcceptTCP(fa)
	}
}

func BenchmarkAcceptTCP_Normal(b *testing.B) {
	// setup one persistent connection
	clearStorage()
	server, cleanup := createFakeConn(b)
	defer cleanup()
	fa := &fakeAcceptor{conn: server, err: nil}
	ip, _, _ := net.SplitHostPort(server.RemoteAddr().String())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clearStorage()
		AcceptTCP(fa)
		// ensure storage cleared for next iteration
		clearStorage()
		// avoid buildup of entries
		storage.Delete(ip)
	}
}

func BenchmarkAcceptTCP_WarnToBlock(b *testing.B) {
	// persistent connection and initial Warn state setup each iteration
	server, cleanup := createFakeConn(b)
	defer cleanup()
	fa := &fakeAcceptor{conn: server, err: nil}
	ip, _, _ := net.SplitHostPort(server.RemoteAddr().String())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clearStorage()
		// pre-populate as Warn
		storage.Store(ip, connectionInfo{state: StateWarn, connCount: 3, lastConnTime: time.Now().UnixMilli()})
		AcceptTCP(fa)
		storage.Delete(ip)
	}
}

func BenchmarkAcceptTCP_Blocked(b *testing.B) {
	// persistent connection and Blocked state pre-set
	server, cleanup := createFakeConn(b)
	defer cleanup()
	fa := &fakeAcceptor{conn: server, err: nil}
	ip, _, _ := net.SplitHostPort(server.RemoteAddr().String())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clearStorage()
		// pre-populate as Blocked
		storage.Store(ip, connectionInfo{state: StateBlocked, blockExpire: time.Now().Add(1 * time.Minute)})
		AcceptTCP(fa)
		storage.Delete(ip)
	}
}
