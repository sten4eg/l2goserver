package gs2ls

import (
	"database/sql"
	"l2goserver/packets"
	"log"
)

const AccountIpsUpdate = `UPDATE accounts SET "pcIp" = $1, hop1 = $2, hop2 = $3, hop3 = $4, hop4 = $5 WHERE login = $6`

func PlayerTracert(data []byte, db *sql.DB) {
	packet := packets.NewReader(data)
	account := packet.ReadString()
	pcIp := packet.ReadString()
	hop1 := packet.ReadString()
	hop2 := packet.ReadString()
	hop3 := packet.ReadString()
	hop4 := packet.ReadString()
	err := setAccountLastTracert(account, pcIp, hop1, hop2, hop3, hop4, db)
	if err != nil {
		log.Println(err.Error())
	}

}

func setAccountLastTracert(account string, pcIp string, hop1 string, hop2 string, hop3 string, hop4 string, db *sql.DB) error {
	_, err := db.Exec(AccountIpsUpdate, pcIp, hop1, hop2, hop3, hop4, account)
	if err != nil {
		return err
	}
	return nil
}
