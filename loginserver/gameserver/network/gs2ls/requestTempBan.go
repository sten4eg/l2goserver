package gs2ls

import (
	"context"
	"l2goserver/database"
	"l2goserver/ipManager"
	"l2goserver/packets"
	"log"
)

const IPTempBan = `INSERT INTO ip_ban (ip) VALUES ($1)`

func RequestTempBan(data []byte, db database.Database) {
	packet := packets.NewReader(data)
	_ = packet.ReadString() // Логин
	ip := packet.ReadString()
	_ = packet.ReadInt64() //banTime

	//haveReason := packet.ReadInt8() != 0
	//if haveReason {
	//	banReason := packet.ReadString()
	//}

	err := banUser(ip, db)
	if err != nil {
		log.Println(err.Error())
	}

}

func banUser(ip string, db database.Database) error {
	_, err := db.Exec(context.Background(), IPTempBan, ip)
	if err != nil {
		return err
	}

	err = ipManager.AddBannedIp(ip)
	if err != nil {
		return err
	}
	return nil

}
