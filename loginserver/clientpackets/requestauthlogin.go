package clientpackets

import (
	"bytes"
	"context"
	"errors"
	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/bcrypt"
	"l2goserver/config"
	"l2goserver/db"
	"l2goserver/loginserver/models"
	"l2goserver/loginserver/serverpackets"
	reasons "l2goserver/loginserver/types/reason"
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
	"l2goserver/utils"
	"math/big"
	"sync"
	"time"
)

const UserInfoSelect = "SELECT * FROM accounts WHERE login = $1"
const UserLastInfo = "UPDATE accounts SET last_ip = $1 , last_active = $2 WHERE login = $3"
const AccountsInsert = "INSERT INTO accounts (login, password) VALUES ($1, $2)"

var errNoData = errors.New("errNoData")

func NewRequestAuthLogin(request []byte, client *models.ClientCtx, l *sync.Map) error {
	buff := packets.Get()

	err := validate(request, client, l)
	if err != nil {
		err = client.SendBuf(serverpackets.NewLoginFailPacket(reasons.LoginOrPassWrong, buff))
		return err
	}

	//TODO есть еще проверки на то подключен ли он к гейм серверу
	reason := reasons.NoReason
	// есть ли причина кикать?
	switch reason {
	default:
		err = client.SendBuf(serverpackets.NewLoginFailPacket(reasons.SystemError, buff))
	case reasons.NoReason:
		err = client.SendBuf(serverpackets.NewLoginOkPacket(client, buff))
	case reasons.InfoWrong, reasons.Ban, reasons.AccountInUse, reasons.LoginOrPassWrong:
		err = client.SendBuf(serverpackets.NewLoginFailPacket(reason, buff))
	}

	if err != nil {
		return err
	}

	client.State = state.AuthedLogin
	return nil
}

func tryCheckinAccount(client *models.ClientCtx) reasons.AuthLoginResult {
	if client.Account.AccessLevel < 0 {
		return reasons.ACCOUNT_BANNED
	}

	ret := reasons.ALREADY_ON_GS
	return ret
}

func validate(request []byte, client *models.ClientCtx, l *sync.Map) error {
	if cap(request) < 128 {
		return errNoData
	}
	payload := request[:128]

	c := new(big.Int).SetBytes(payload)
	decodeData := c.Exp(c, client.PrivateKey.D, client.PrivateKey.N).Bytes()

	if cap(decodeData) < 28 {
		return errNoData
	}

	trimLogin := bytes.Trim(decodeData[1:14], string(rune(0)))
	trimPassword := bytes.Trim(decodeData[14:28], string(rune(0)))

	login := utils.B2s(trimLogin)
	password := utils.B2s(trimPassword)

	var account models.Account
	dbConn, err := db.GetConn()
	if err != nil {
		return err
	}

	defer dbConn.Release()

	row := dbConn.QueryRow(context.Background(), UserInfoSelect, login)
	err = row.Scan(&account.Login, &account.Password, &account.CreatedAt, &account.LastActive, &account.AccessLevel, &account.LastIp, &account.LastServer)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) && config.AutoCreateAccounts() {
			err = createAccount(login, password)
			if err != nil {
				panic(err)
			}
			return validate(request, client, l)
		}

		return err
	}
	err = bcrypt.CompareHashAndPassword(utils.S2b(account.Password), utils.S2b(password))
	if err != nil {
		return err
	}

	//if account.AccessLevel < 0 {
	//	return reasons.Ban, nil
	//}

	_, err = dbConn.Exec(context.Background(), UserLastInfo, client.Socket.RemoteAddr().String(), time.Now(), login)
	if err != nil {
		return err
	}

	client.Account = account
	return nil
}

func createAccount(clearLogin, clearPassword string) error {
	password, err := bcrypt.GenerateFromPassword(utils.S2b(clearPassword), 10)
	if err != nil {
		return err
	}
	dbConn, err := db.GetConn()
	if err != nil {
		return err
	}
	defer dbConn.Release()

	_, err = dbConn.Exec(context.Background(), AccountsInsert,
		clearLogin, password)
	return err
}
