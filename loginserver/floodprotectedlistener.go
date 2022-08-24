package loginserver

import (
	"net"
	"net/netip"
	"time"
)

const fastConnectionLimit = 15
const normalConnectionTime = 700
const fastConnectionTime = 350
const maxConnectionPerIP = 50

type connection struct {
	connNum    uint8
	lastConn   int64
	isFlooding bool
}

var floodProtection map[netip.Addr]*connection

func (ls *LoginServer) AcceptTCPWithFloodProtection() (*net.TCPConn, error) {
	for {
		conn, err := ls.clientsListener.AcceptTCP()
		if err != nil {
			continue
		}

		addr, err := netip.ParseAddrPort(conn.RemoteAddr().String())
		if err != nil {
			panic(err)
		}

		fConn := floodProtection[addr.Addr()]

		if fConn == nil {
			floodProtection[addr.Addr()] = &connection{1, time.Now().UnixMilli(), false}
		} else {
			fConn.connNum++
			connectionTime := time.Now().UnixMilli() - fConn.lastConn
			if fConn.connNum > fastConnectionLimit && connectionTime < normalConnectionTime || connectionTime < fastConnectionTime || fConn.connNum > maxConnectionPerIP {
				fConn.lastConn = time.Now().UnixMilli()
				conn.Close()
				fConn.connNum--
				fConn.isFlooding = true
				continue
			}

			if fConn.isFlooding {
				fConn.isFlooding = false
			}
		}
		return conn, nil
	}

}

func InitializeFloodProtection() {
	floodProtection = make(map[netip.Addr]*connection, 10)
}
