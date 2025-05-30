package models

import (
	"database/sql"
	"time"
)

type Account struct {
	Login           string
	Password        string
	CreatedAt       time.Time
	LastActive      *time.Time
	AccessLevel     string
	LastIp          sql.NullString
	LastServer      int8
	CharacterCount  map[uint8]uint8
	CharactersToDel map[uint8][]int64
}

func (account *Account) SetCharsOnServer(serverId uint8, chars uint8) {
	if account.CharacterCount == nil {
		account.CharacterCount = make(map[uint8]uint8)
	}
	account.CharacterCount[serverId] = chars
}

func (account *Account) SetCharsWaitingDelOnServer(serverId uint8, charsToDel []int64) {
	if account.CharactersToDel == nil {
		account.CharactersToDel = make(map[uint8][]int64)
	}
	account.CharactersToDel[serverId] = charsToDel
}
