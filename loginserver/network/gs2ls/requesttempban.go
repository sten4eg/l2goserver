package gs2ls

import (
	"context"
	"l2goserver/db"
	"l2goserver/loginserver/IpManager"
	"l2goserver/packets"
	"log"
	"net/netip"
)

const IPTempBan = "INSERT INTO ip_ban VALUES ($1, $2) ON DUPLICATE KEY UPDATE value=$3"

func RequestTempBan(data []byte) {
	packet := packets.NewReader(data)
	_ = packet.ReadString() // Логин
	ip := packet.ReadString()
	banTime := int(packet.ReadInt64())

	//haveReason := packet.ReadInt8() != 0
	//if haveReason {
	//	banReason := packet.ReadString()
	//}

	err := banUser(ip, banTime)
	if err != nil {
		log.Println(err.Error())
	}

}

func banUser(ip string, banTime int) error {
	dbConn, err := db.GetConn()
	if err != nil {
		return err
	}
	defer dbConn.Release()

	_, err = dbConn.Exec(context.Background(), IPTempBan, ip, banTime, banTime)
	if err != nil {
		return err
	}

	err = AddBanForAddress(ip, banTime)
	if err != nil {
		return err
	}
	return nil

}

func AddBanForAddress(address string, expiration int) error {
	addr, err := netip.ParseAddr(address)
	if err != nil {
		return err
	}
	IpManager.BannedIp[addr] = expiration
	return nil
}
