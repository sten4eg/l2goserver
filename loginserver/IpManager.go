package loginserver

import (
	"context"
	"github.com/jackc/pgtype"
	"l2goserver/db"
	"net/netip"
)

var BannedIp []netip.Addr

func LoadBannedIp() {
	dbConn, err := db.GetConn()
	if err != nil {
		panic(err)
	}
	defer dbConn.Release()

	rows, err := dbConn.Query(context.Background(), `SELECT * FROM ip_ban`)
	if err != nil {
		panic(err)
	}
	var i pgtype.Inet

	for rows.Next() {
		err = rows.Scan(&i)
		if err != nil {
			panic(err)
		}
		a := netip.MustParseAddr(i.IPNet.IP.String())
		if a.IsValid() {
			BannedIp = append(BannedIp, a)
		}

	}

}

func IsBannedIp(clientAddr netip.Addr) bool {
	for i := range BannedIp {
		if BannedIp[i].Compare(clientAddr) == 0 {
			return true
		}
	}
	return false
}
