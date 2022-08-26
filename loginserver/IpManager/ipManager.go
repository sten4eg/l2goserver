package IpManager

import (
	"context"
	"github.com/jackc/pgtype"
	"l2goserver/db"
	"net/netip"
)

var BannedIp map[netip.Addr]int

func LoadBannedIp() error {
	BannedIp = make(map[netip.Addr]int, 100)
	dbConn, err := db.GetConn()
	if err != nil {
		return err
	}
	defer dbConn.Release()

	rows, err := dbConn.Query(context.Background(), `SELECT ip, "value" FROM ip_ban WHERE "value" > extract('epoch' from CURRENT_TIMESTAMP)::bigint`)
	if err != nil {
		return err
	}
	var i pgtype.Inet

	for rows.Next() {
		var value int
		err = rows.Scan(&i, &value)
		if err != nil {
			return err
		}
		a := netip.MustParseAddr(i.IPNet.IP.String())
		if a.IsValid() {
			BannedIp[a] = value
		}

	}

	return nil
}

func IsBannedIp(clientAddr netip.Addr) bool {
	_, ok := BannedIp[clientAddr]
	return ok
}
