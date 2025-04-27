package gs2ls

import (
	"database/sql"
	"l2goserver/packets"
	"log"
	"net/netip"
)

const IPTempBan = "INSERT INTO loginserver.ip_ban VALUES ($1, $2) ON CONFLICT(ip) DO UPDATE SET  unix_time = $2"

type ipManager interface {
	AddIpToBan(clientAddr netip.Addr, expiration int)
}

func RequestTempBan(data []byte, db *sql.DB, i ipManager) {
	packet := packets.NewReader(data)
	_ = packet.ReadString() // Логин
	ip := packet.ReadString()
	banTime := int(packet.ReadInt64())

	//haveReason := packet.ReadInt8() != 0
	//if haveReason {
	//	banReason := packet.ReadString()
	//}

	err := banUser(ip, banTime, db, i)
	if err != nil {
		log.Println(err.Error())
	}

}

func banUser(ip string, banTime int, db *sql.DB, i ipManager) error {
	_, err := db.Exec(IPTempBan, ip, banTime)
	if err != nil {
		return err
	}

	err = AddBanForAddress(ip, banTime, i)
	if err != nil {
		return err
	}
	return nil

}

func AddBanForAddress(address string, expiration int, i ipManager) error {
	addr, err := netip.ParseAddr(address)
	if err != nil {
		return err
	}
	i.AddIpToBan(addr, expiration)
	return nil
}
