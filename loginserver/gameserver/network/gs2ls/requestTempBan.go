package gs2ls

import (
	"context"
	"l2goserver/database"
	"l2goserver/loginserver/ipManager"
	"l2goserver/packets"
	"log"
)

const IPTempBan = "INSERT INTO loginserver.ip_ban VALUES ($1, $2) ON CONFLICT(ip) DO UPDATE SET  unix_time = $2"

func RequestTempBan(data []byte, db database.Database) {
	packet := packets.NewReader(data)
	_ = packet.ReadString() // Логин
	ip := packet.ReadString()
	banTime := int(packet.ReadInt64())

	//haveReason := packet.ReadInt8() != 0
	//if haveReason {
	//	banReason := packet.ReadString()
	//}

	err := banUser(ip, banTime, db)
	if err != nil {
		log.Println(err.Error())
	}

}

func banUser(ip string, banTime int, db database.Database) error {
	_, err := db.Exec(context.Background(), IPTempBan, ip, banTime)
	if err != nil {
		return err
	}

	err = ipManager.AddBannedIp(ip, banTime)
	if err != nil {
		return err
	}
	return nil

}
