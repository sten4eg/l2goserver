package gs2ls

import (
	"context"
	"l2goserver/database"
	"l2goserver/packets"
	"log"
)

const AccountAccessLevelUpdate = "UPDATE accounts SET accessLevel = $1 WHERE login = $2"

func ChangeAccessLevel(data []byte, db database.Database) {
	packet := packets.NewReader(data)
	level := packet.ReadInt32()
	account := packet.ReadString()

	err := setAccountAccessLevel(account, level, db)
	if err != nil {
		log.Println(err.Error())
	}

}

func setAccountAccessLevel(account string, level int32, db database.Database) error {
	_, err := db.Exec(context.Background(), AccountAccessLevelUpdate, level, account)
	if err != nil {
		return err
	}
	return nil
}
