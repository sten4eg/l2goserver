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
const maxConnectionPerIP = 2
const banTime = time.Minute

type connection struct {
	connNum    uint8
	lastConn   int64
	isFlooding bool
	banExpire  time.Time
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
			floodProtection.Store(x, &connection{1, time.Now().UnixMilli(), false, time.Now()})
		} else {
			if fConn.isFlooding {
				if time.Now().After(fConn.banExpire) {
					fConn.isFlooding = false
					fConn.banExpire = time.Now()
					fConn.connNum = 1
					fConn.lastConn = time.Now().UnixMilli()
				} else {
					_ = conn.Close()
					continue
				}
			}

			fConn.connNum++
			curTime := time.Now().UnixMilli()
			connectionTime := curTime - fConn.lastConn
			fConn.lastConn = curTime

			if (fConn.connNum > 2 && connectionTime < fastConnectionTime) ||
				fConn.connNum > maxConnectionPerIP || connectionTime < normalConnectionTime {
				if connectionTime > 5000 {
					fConn.isFlooding = false
					fConn.banExpire = time.Now()
					fConn.connNum = 1
				} else {
					_ = conn.Close()
					fConn.isFlooding = true
					fConn.banExpire = time.Now().Add(banTime)
					continue
				}

			}
		}
		return conn, nil
	}

}

func InitializeFloodProtection() {
	floodProtection = xsync.NewMapOf[*connection]()
}
