package loginserver

import (
	"context"
	"github.com/jackc/pgtype"
	"l2goserver/db"
	"net/netip"
)

var BannedIp []netip.Addr

func LoadBannedIp() error {
	dbConn, err := db.GetConn()
	if err != nil {
		return err
	}
	defer dbConn.Release()

	rows, err := dbConn.Query(context.Background(), `SELECT * FROM ip_ban`)
	if err != nil {
		return err
	}
	var i pgtype.Inet

	for rows.Next() {
		err = rows.Scan(&i)
		if err != nil {
			return err
		}
		a := netip.MustParseAddr(i.IPNet.IP.String())
		if a.IsValid() {
			BannedIp = append(BannedIp, a)
		}

	}

	return nil
}

func IsBannedIp(clientAddr netip.Addr) bool {
	for i := range BannedIp {
		if BannedIp[i].Compare(clientAddr) == 0 {
			return true
		}
	}
	return false
}
