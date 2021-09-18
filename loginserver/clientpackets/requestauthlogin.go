package clientpackets

import (
	"bytes"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"l2goserver/db"
	"l2goserver/loginserver/models"
	"l2goserver/loginserver/serverpackets"
	"math/big"
	"time"
)

type RequestAuthLogin struct {
	Login    string
	Password string
}

func NewRequestAuthLogin(request []byte, client *models.Client, l []*models.Client) (byte, error) {
	var result RequestAuthLogin

	payload := request[:128]

	c := new(big.Int).SetBytes(payload)
	decodeData := c.Exp(c, client.PrivateKey.D, client.PrivateKey.N).Bytes()

	trimLogin := bytes.Trim(decodeData[1:14], string(rune(0)))
	trimPassword := bytes.Trim(decodeData[14:28], string(rune(0)))

	result.Login = string(trimLogin)
	result.Password = string(trimPassword)

	return result.validate(client, l)
}

func CreateAccount(request []byte, client *models.Client) error {
	payload := request[:128]
	c := new(big.Int).SetBytes(payload)
	decodeData := c.Exp(c, client.PrivateKey.D, client.PrivateKey.N).Bytes()
	trimLogin := bytes.Trim(decodeData[1:14], string(rune(0)))
	trimPassword := bytes.Trim(decodeData[14:28], string(rune(0)))
	password, err := bcrypt.GenerateFromPassword(trimPassword, 10)
	if err != nil {
		return err
	}
	dbConn, err := db.GetConn()
	if err != nil {
		panic(err.Error())
	}
	_, err = dbConn.Exec(context.Background(), "INSERT INTO accounts (login, password, access_level) VALUES ($1, $2, 0) ",
		string(trimLogin), password)
	return err
}

func (r *RequestAuthLogin) validate(client *models.Client, l []*models.Client) (byte, error) {
	var account models.Account
	dbConn, err := db.GetConn()
	if err != nil {
		panic(err.Error())
	}
	row := dbConn.QueryRow(context.Background(), "SELECT * FROM accounts WHERE login = $1", r.Login)
	err = row.Scan(&account.Login, &account.Password, &account.CreatedAt, &account.LastActive, &account.AccessLevel, &account.LastIp, &account.LastServer)
	if err != nil {
		return serverpackets.REASON_USER_OR_PASS_WRONG, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(r.Password))
	if err != nil {
		return serverpackets.REASON_USER_OR_PASS_WRONG, err
	}
	if account.AccessLevel < 0 {
		return serverpackets.REASON_BAN, errors.New("Ban")
	}

	for _, v := range l {
		if v.Account.Login == account.Login {
			return serverpackets.REASON_ACCOUNT_IN_USE, errors.New("account used")
		}
	}
	_, err = dbConn.Exec(context.Background(), "UPDATE accounts SET last_ip = $1 , last_active = $2 WHERE login = $3", client.Socket.RemoteAddr().String(), time.Now(), r.Login)
	if err != nil {
		return serverpackets.REASON_INFO_WRONG, err
	}

	client.Account = account
	return 0, nil
}
