package loginserver

import (
	"github.com/puzpuzpuz/xsync"
	"net"
	"strings"
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

var floodProtection *xsync.MapOf[*connection]

func (ls *LoginServer) AcceptTCPWithFloodProtection() (*net.TCPConn, error) {
	for {
		conn, err := ls.clientsListener.AcceptTCP()
		if err != nil {
			continue
		}

		x, _, ok := strings.Cut(conn.RemoteAddr().String(), ":")
		if !ok {
			continue
		}

		fConn, ok := floodProtection.Load(x)

		if !ok {
			floodProtection.Store(x, &connection{1, time.Now().UnixMilli(), false})
		} else {
			fConn.connNum++
			curTime := time.Now().UnixMilli()
			connectionTime := curTime - fConn.lastConn
			fConn.lastConn = curTime

			if (fConn.connNum > fastConnectionLimit && connectionTime < normalConnectionTime) ||
				connectionTime < fastConnectionTime ||
				fConn.connNum > maxConnectionPerIP {
				_ = conn.Close()
				fConn.connNum--
				fConn.isFlooding = true
				continue
			}

			fConn.isFlooding = false
		}
		return conn, nil
	}

}

func InitializeFloodProtection() {
	floodProtection = xsync.NewMapOf[*connection]()
}
