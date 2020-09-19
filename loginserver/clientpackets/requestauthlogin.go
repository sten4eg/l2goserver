package clientpackets

import (
	"bytes"
	"errors"
	"github.com/jackc/pgx"
	"l2goserver/loginserver/models"
	"l2goserver/loginserver/serverpackets"
	"math/big"
	"time"
)

type RequestAuthLogin struct {
	Login    string
	Password string
}

func NewRequestAuthLogin(request []byte, client *models.Client, db *pgx.Conn) (byte, error) {
	var result RequestAuthLogin

	payload := request[:128]

	c := new(big.Int).SetBytes(payload)
	decodeData := c.Exp(c, client.PrivateKey.D, client.PrivateKey.N).Bytes()

	trimLogin := bytes.Trim(decodeData[1:14], string(rune(0)))
	trimPassword := bytes.Trim(decodeData[14:28], string(rune(0)))

	result.Login = string(trimLogin)
	result.Password = string(trimPassword)

	return result.validate(db, client)
}

func (r *RequestAuthLogin) validate(db *pgx.Conn, client *models.Client) (byte, error) {

	var account models.Account
	row := db.QueryRow("SELECT * FROM accounts WHERE login = $1 AND password = $2", r.Login, r.Password)
	err := row.Scan(&account.Login, &account.Password, &account.CreatedAt, &account.LastActive, &account.AccessLevel, &account.LastIp, &account.LastServer)
	if err != nil {
		return serverpackets.REASON_USER_OR_PASS_WRONG, err
	}
	if account.AccessLevel < 0 {
		return serverpackets.REASON_BAN, errors.New("Ban")
	}
	_, err = db.Exec("UPDATE accounts SET last_ip = $1 , last_active = $2 WHERE login = $3", client.Socket.RemoteAddr().String(), time.Now(), r.Login)
	if err != nil {
		return serverpackets.REASON_INFO_WRONG, err
	}
	client.Account = account
	return 0, nil
}
