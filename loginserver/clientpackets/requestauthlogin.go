package clientpackets

import (
	"bytes"
	"context"
	"errors"
	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/bcrypt"
	"l2goserver/db"
	"l2goserver/loginserver/models"
	"l2goserver/loginserver/serverpackets"
	reasons "l2goserver/loginserver/types/reason"
	"l2goserver/loginserver/types/state"
	"math/big"
	"time"
)

type RequestAuthLogin struct {
	Login    string
	Password string
}

func NewRequestAuthLogin(request []byte, client *models.ClientCtx, l []*models.ClientCtx, enableAutoCreateAccount bool) error {
	var result RequestAuthLogin

	payload := request[:128]

	c := new(big.Int).SetBytes(payload)
	decodeData := c.Exp(c, client.PrivateKey.D, client.PrivateKey.N).Bytes()

	trimLogin := bytes.Trim(decodeData[1:14], string(rune(0)))
	trimPassword := bytes.Trim(decodeData[14:28], string(rune(0)))

	result.Login = string(trimLogin)
	result.Password = string(trimPassword)

	reason, err := result.validate(client, l)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	//TODO есть еще проверки на то подключен ли он к гейм серверу

	// есть ли причина кикать?
	switch reason {
	default:
		err = client.Send(serverpackets.NewLoginFailPacket(reasons.SystemError))
	case reasons.NoReason:
		err = client.Send(serverpackets.NewLoginOkPacket(client))
	case reasons.InfoWrong, reasons.Ban, reasons.AccountInUse:
		err = client.Send(serverpackets.NewLoginFailPacket(reason))
	case reasons.LoginOrPassWrong:
		if enableAutoCreateAccount {
			err = createAccount(request, client)
			if err != nil {
				return err
			}
			err = client.Send(serverpackets.NewLoginOkPacket(client))
		} else {
			err = client.Send(serverpackets.NewLoginFailPacket(reason))
		}
	}

	if err != nil {
		return err
	}

	client.State = state.AuthedLogin
	return nil
}

func (r *RequestAuthLogin) validate(client *models.ClientCtx, l []*models.ClientCtx) (reasons.Reason, error) {
	var account models.Account
	dbConn, err := db.GetConn()
	if err != nil {
		return reasons.NoReason, err
	}

	defer dbConn.Release()

	row := dbConn.QueryRow(context.Background(), "SELECT * FROM accounts WHERE login = $1", r.Login)
	err = row.Scan(&account.Login, &account.Password, &account.CreatedAt, &account.LastActive, &account.AccessLevel, &account.LastIp, &account.LastServer)
	if err != nil {
		return reasons.LoginOrPassWrong, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(r.Password))
	if err != nil {
		return reasons.LoginOrPassWrong, err
	}
	if account.AccessLevel < 0 {
		return reasons.Ban, nil
	}

	for _, v := range l {
		if v.State != state.AuthedLogin {
			if v.Account.Login == account.Login {
				return reasons.AccountInUse, nil
			}
		}
	}
	_, err = dbConn.Exec(context.Background(), "UPDATE accounts SET last_ip = $1 , last_active = $2 WHERE login = $3", client.Socket.RemoteAddr().String(), time.Now(), r.Login)
	if err != nil {
		return reasons.InfoWrong, err
	}

	client.Account = account
	return reasons.NoReason, nil
}

func createAccount(request []byte, client *models.ClientCtx) error {
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
	defer dbConn.Release()

	_, err = dbConn.Exec(context.Background(), "INSERT INTO accounts (login, password) VALUES ($1, $2)",
		string(trimLogin), password)
	return err
}
