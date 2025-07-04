package floodProtecor

import (
	"errors"
	"net"
	"sync"
	"time"
)

const normalConnectionTime = 700 // ms — threshold for "normal" speed
const fastConnectionTime = 350   // ms — threshold for "suspiciously fast" speed
const maxConnectionPerIP = 50    // max. number of connections before immediate blocking
const banTime = time.Minute      // ban duration
const safeConnInterval = 5000    // ms — interval after which the counter is "reset"
type State int64

const (
	StateNormal State = iota
	StateWarn
	StateBlocked
)

type connectionInfo struct {
	connCount    int64
	lastConnTime int64
	state        State
	blockExpire  time.Time
}
type TCPAcceptor interface {
	AcceptTCP() (*net.TCPConn, error)
}

var storage sync.Map

func (ci *connectionInfo) UpdateState(currentTime int64, connectionTime int64) {
	switch ci.state {
	case StateNormal:
		if ci.isSuspicious(connectionTime) {
			ci.state = StateWarn
		}
	case StateWarn:
		if ci.isFlooding(connectionTime) {
			ci.state = StateBlocked
			ci.blockExpire = time.Now().Add(banTime)
		} else if ci.isBackToNormal(currentTime) {
			ci.state = StateNormal
			ci.connCount = 0
		}
	case StateBlocked:
		if time.Now().After(ci.blockExpire) {
			ci.state = StateNormal
			ci.connCount = 0
		}
	}
}

func (ci *connectionInfo) isSuspicious(connectionTime int64) bool {
	return ci.connCount > 2 && connectionTime < fastConnectionTime
}

func (ci *connectionInfo) isFlooding(connectionTime int64) bool {
	return ci.connCount > maxConnectionPerIP || connectionTime < normalConnectionTime
}

func (ci *connectionInfo) isBackToNormal(currentTime int64) bool {
	return currentTime-ci.lastConnTime > safeConnInterval
}

func AcceptTCP(acceptor TCPAcceptor) (*net.TCPConn, error) {
	conn, err := acceptor.AcceptTCP()
	if err != nil {
		return nil, err
	}

	ip, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	curTime := time.Now().UnixMilli()
	ci, ok := storage.Load(ip)
	if !ok {
		ci = connectionInfo{
			state:        StateNormal,
			connCount:    1,
			lastConnTime: curTime,
		}
		storage.Store(ip, ci)
	} else {
		connInfo := ci.(connectionInfo)
		connectionTime := curTime - connInfo.lastConnTime
		connInfo.connCount++
		connInfo.lastConnTime = curTime
		connInfo.UpdateState(curTime, connectionTime)
		storage.Store(ip, connInfo)
	}

	if ci.(connectionInfo).state == StateBlocked {
		_ = conn.Close()
		return nil, errors.New("connection closed by FloodProtection")
	}

	return conn, nil
}
