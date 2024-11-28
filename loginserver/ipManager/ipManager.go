package ipManager

import (
	"context"
	"l2goserver/database"
	"net"
	"sync"
)

type ipBanManager struct {
	banned map[string]net.IPNet
	mu     sync.RWMutex
}

var manager = ipBanManager{
	banned: make(map[string]net.IPNet),
	mu:     sync.RWMutex{},
}

func LoadBannedIp(db database.Database) error {
	rows, err := db.Query(context.Background(), `SELECT ip, unix_time FROM ip_ban WHERE unix_time > extract('epoch' from now())::bigint`)
	if err != nil {
		return err
	}
	defer rows.Close()
	manager.mu.Lock()
	defer manager.mu.Unlock()

	for rows.Next() {
		var value int
		var ii string
		err = rows.Scan(&ii, &value)
		if err != nil {
			return err
		}
		_, ipNet, err := net.ParseCIDR(ii)
		if err != nil {
			return err
		}
		manager.banned[ipNet.String()] = *ipNet
	}

	return nil
}

func IsBannedIp(ip string) bool {
	host, _, err := net.SplitHostPort(ip)
	if err != nil {
		return false
	}

	parsedIP := net.ParseIP(host)
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	for _, ipNet := range manager.banned {
		if ipNet.Contains(parsedIP) {
			return true
		}
	}
	return false
}

func AddBannedIp(ip string) error {
	_, ipNet, err := net.ParseCIDR(ip + "/32")
	if err != nil {
		return err
	}
	manager.mu.Lock()
	defer manager.mu.Unlock()
	manager.banned[ipNet.String()] = *ipNet
	return nil
}
