package c2ls

import (
	"bytes"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"l2goserver/config"
	"l2goserver/db"
	"l2goserver/loginserver/gameserver"
	"l2goserver/loginserver/gameserver/network/ls2gs"
	"l2goserver/loginserver/models"
	"l2goserver/loginserver/network/ls2c"
	reasons "l2goserver/loginserver/types/reason/clientReasons"
	"l2goserver/loginserver/types/state/clientState"
	"log"
	"math/big"
	"runtime/trace"
	"time"
)

const UserInfoSelect = "SELECT accounts.login,accounts.password,accounts.created_at,accounts.last_active,accounts.access_level,accounts.last_ip,accounts.last_server FROM loginserver.accounts WHERE accounts.login = $1"
const UserLastInfo = "UPDATE loginserver.accounts SET last_ip = $1 , last_active = $2 WHERE login = $3"
const AccountsInsert = "INSERT INTO loginserver.accounts (login, password) VALUES ($1, $2)"

var errNoData = errors.New("errNoData")

type isInLoginInterface interface {
	IsAccountInLoginAndAddIfNot(ClientRequestInterface) bool
	AssignSessionKeyToClient(ClientRequestInterface) *models.SessionKey
	GetGameServerInfoList() []*gameserver.Info
	GetClientCtx(string) ClientRequestInterface
	RemoveAuthedLoginClient(string)
}

type ClientRequestInterface interface {
	Send([]byte) error
	SetState(clientState.ClientCtxState)
	SetSessionKey()
	GetAccountLogin() string
}

func NewRequestAuthLogin(request []byte, client ClientRequestInterface, server isInLoginInterface) error {
	err := validate(request, client)
	if err != nil {
		err = client.Send(ls2c.NewLoginFailPacket(reasons.LoginOrPassWrong))
		return err
	}
	reason := tryCheckinAccount(client, server)

	switch reason {
	default:
		err = client.Send(ls2c.NewLoginFailPacket(reasons.SystemError))
	case clientState.AuthSuccess:
		client.SetState(clientState.AuthedLogin)
		client.SetSessionKey(server.AssignSessionKeyToClient(client))
		err = client.Send(ls2c.NewLoginOkPacket(client))
		sendCharactersOnAccount(client.GetAccountLogin(), server)
	case clientState.AccountBanned:
		err = client.Send(ls2c.NewLoginFailPacket(reasons.Ban))
		client.CloseConnection()
	case clientState.AlreadyOnLs:
		account := client.Account.Login
		err = client.Send(ls2c.NewLoginFailPacket(reasons.AccountInUse))
		oldClient := server.GetClientCtx(account)
		if oldClient != nil {
			err = oldClient.Send(ls2c.AccountKicked(reasons.AccountInUse))
			oldClient.CloseConnection()
			server.RemoveAuthedLoginClient(account)
		}
		client.CloseConnection()
	case clientState.AlreadyOnGs:
		account := client.Account.Login
		err = client.Send(ls2c.NewLoginFailPacket(reasons.AccountInUse))
		gsi := gameserver.GetGameServerInstance().GetAccountOnGameServer(account)
		if gsi != nil {
			if gsi.IsAuthed() {
				_ = gsi.Send(ls2gs.KickPlayer(account))
			}
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func tryCheckinAccount(client *models.ClientCtx, server isInLoginInterface) clientState.ClientAuthState {
	if client.Account.AccessLevel < 0 {
		return clientState.AccountBanned
	}

	ret := clientState.AlreadyOnGs
	if gameserver.IsAccountInGameServer(client.Account.Login) {
		return ret
	}
	ret = clientState.AlreadyOnLs
	if server.IsAccountInLoginAndAddIfNot(client) {
		return ret
	}
	return clientState.AuthSuccess
}

func validate(request []byte, clienst *models.ClientCtx) error {
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

	login := string(trimLogin)
	password := string(trimPassword)

	var account models.Account
	reg := trace.StartRegion(context.Background(), "GetCONN1")
	dbConn1, err := db.GetConn()
	if err != nil {
		log.Println("errConn1")
		return err
	}
	defer dbConn1.Release()
	reg.End()

	reg = trace.StartRegion(context.Background(), "userInfoSelect")
	err = dbConn1.QueryRow(context.Background(), UserInfoSelect, login).
		Scan(&account.Login, &account.Password, &account.CreatedAt, &account.LastActive, &account.AccessLevel, &account.LastIp, &account.LastServer)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) && config.AutoCreateAccounts() {
			err = createAccount(login, password)
			if err != nil {
				return err
			}
			return validate(request, client)
		}

		return err
	}
	reg.End()

	err = bcrypt.CompareHashAndPassword([]byte((account.Password)), []byte(password))
	if err != nil {
		return err
	}

	_, err = dbConn1.Exec(context.Background(), UserLastInfo, client.GetRemoteAddr().String(), time.Now(), login)
	if err != nil {
		log.Println(err)
		return err
	}

	client.Account = account
	return nil
}

func createAccount(clearLogin, clearPassword string) error {
	password, err := bcrypt.GenerateFromPassword([]byte(clearPassword), 10)
	if err != nil {
		return err
	}
	dbConn, err := db.GetConn()
	if err != nil {
		return err
	}
	defer dbConn.Release()
	_, err = dbConn.Exec(context.Background(), AccountsInsert,
		clearLogin, string(password))
	return err
}

func sendCharactersOnAccount(account string, server isInLoginInterface) {
	serverList := server.GetGameServerInfoList()
	for _, gsi := range serverList {
		if gsi.IsAuthed() {
			_ = gsi.Send(ls2gs.RequestCharacter(account))
		}
	}
}
