package ipManager

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net"
	"net/netip"
)

type IpManager interface {
	IsBannedIp(clientAddr netip.Addr) bool
	AddIpToBan(clientAddr netip.Addr, expiration int)
}
type bannedIpsMap map[netip.Addr]int

var bannedIp bannedIpsMap

func (a bannedIpsMap) IsBannedIp(clientAddr netip.Addr) bool {
	_, ok := a[clientAddr]
	return ok
}

func (a bannedIpsMap) AddIpToBan(clientAddr netip.Addr, expiration int) {
	a[clientAddr] = expiration
}

func LoadBannedIp(db *sql.DB) (IpManager, error) {
	bannedIp = make(map[netip.Addr]int, 100)

	rows, err := db.Query(`SELECT ip, unix_time FROM ip_ban WHERE unix_time > extract('epoch' from now())::bigint`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var i CIDR

	for rows.Next() {
		var value int
		err = rows.Scan(&i, &value)
		if err != nil {
			return nil, err
		}
		a := netip.MustParseAddr(i.IPNet.IP.String())
		if a.IsValid() {
			bannedIp[a] = value
		}

	}

	return bannedIp, nil
}

type CIDR struct {
	net.IPNet
}

func (c *CIDR) Scan(value interface{}) error {
	if value == nil {
		return errors.New("CIDR cannot be null")
	}

	var s string
	switch v := value.(type) {
	case []byte:
		s = string(v)
	case string:
		s = v
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}

	_, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %w", err)
	}

	c.IPNet = *ipnet
	return nil
}

func (c CIDR) Value() (driver.Value, error) {
	return c.String(), nil
}
