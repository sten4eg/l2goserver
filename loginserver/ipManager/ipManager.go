package ipManager

import (
	"context"
	"github.com/jackc/pgtype"
	"l2goserver/database"
	"net/netip"
)

var bannedIp map[netip.Addr]int

func LoadBannedIp(db database.Database) error {
	bannedIp = make(map[netip.Addr]int, 100)

	rows, err := db.Query(context.Background(), `SELECT ip, unix_time FROM ip_ban WHERE unix_time > extract('epoch' from now())::bigint`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var i pgtype.Inet

	for rows.Next() {
		var value int
		err = rows.Scan(&i, &value)
		if err != nil {
			return err
		}
		a := netip.MustParseAddr(i.IPNet.IP.String())
		if a.IsValid() {
			bannedIp[a] = value
		}

	}

	return nil
}

func IsBannedIp(clientAddr netip.Addr) bool {
	_, ok := bannedIp[clientAddr]
	return ok
}

func AddBannedIp(ip string, banTime int) error {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return err
	}
	bannedIp[addr] = banTime
	return nil
}
