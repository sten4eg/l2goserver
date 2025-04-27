package gs2ls

import (
	"database/sql"
	"l2goserver/packets"
	"log"
)

const AccountAccessLevelUpdate = "UPDATE accounts SET accessLevel = $1 WHERE login = $2"

func ChangeAccessLevel(data []byte, db *sql.DB) {
	packet := packets.NewReader(data)
	level := packet.ReadInt32()
	account := packet.ReadString()

	err := setAccountAccessLevel(account, level, db)
	if err != nil {
		log.Println(err.Error())
	}

}

func setAccountAccessLevel(account string, level int32, db *sql.DB) error {
	_, err := db.Exec(AccountAccessLevelUpdate, level, account)
	if err != nil {
		return err
	}
	return nil
}
