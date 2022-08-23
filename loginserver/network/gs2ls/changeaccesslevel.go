package gs2ls

import (
	"context"
	"l2goserver/db"
	"l2goserver/packets"
	"log"
)

const AccountAccessLevelUpdate = "UPDATE accounts SET accessLevel = $1 WHERE login = $2"

func ChangeAccessLevel(data []byte) {
	packet := packets.NewReader(data)
	level := packet.ReadInt32()
	account := packet.ReadString()

	err := SetAccountAccessLevel(account, level)
	if err != nil {
		log.Println(err.Error())
	}

}

func SetAccountAccessLevel(account string, level int32) error {
	dbConn, err := db.GetConn()
	if err != nil {
		return err
	}
	defer dbConn.Release()

	_, err = dbConn.Exec(context.Background(), AccountAccessLevelUpdate, level, account)
	if err != nil {
		return err
	}
	return nil
}
