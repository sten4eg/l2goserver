package c2ls

import (
	"bytes"
	"context"
	"crypto/rsa"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
	"l2goserver/config"
	"l2goserver/database"
	"l2goserver/loginserver/gameserver/network/ls2gs"
	"l2goserver/loginserver/network/ls2c"
	reasons "l2goserver/loginserver/types/reason/clientReasons"
	"l2goserver/loginserver/types/state/clientState"
	"log"
	"math/big"
	"net"
	"runtime/trace"
	"time"
)

const UserInfoSelect = "SELECT accounts.password,accounts.created_at,accounts.last_active,accounts.access_level,accounts.last_ip,accounts.last_server FROM loginserver.accounts WHERE accounts.login = $1"
const UserLastInfo = "UPDATE loginserver.accounts SET last_ip = $1 , last_active = $2 WHERE login = $3"
const AccountsInsert = "INSERT INTO loginserver.accounts (login, password) VALUES ($1, $2)"

var errNoData = errors.New("errNoData")

type loginServerInterface interface {
	IsAccountInLoginAndAddIfNot(ClientRequestInterface) bool
	AssignSessionKeyToClient(GAL) (uint32, uint32, uint32, uint32)
	GetGameServerInfoList() []GameServerInfoInterface
	GetClientCtx(string) ClientRequestInterface
	RemoveAuthedLoginClient(string)
}

type GameServerInfoInterface interface {
	IsAuthed() bool
	SendSlice([]byte) error
}

type GAL interface {
	GetAccountLogin() string
}
type ClientRequestInterface interface {
	Send([]byte) error
	SetState(clientState.ClientCtxState)
	SetSessionKey(uint32, uint32, uint32, uint32)
	GetAccountLogin() string
	GetPrivateKey() *rsa.PrivateKey
	GetSessionLoginOK1() uint32
	GetSessionLoginOK2() uint32
	CloseConnection()
	GetAccountAccessLevel() int8
	GetRemoteAddr() net.Addr
	SetAccount(string, string, pgtype.Timestamp, pgtype.Timestamp, int8, int8, sql.NullString)
}
type gameServerInterface interface {
	GetAccountOnGameServer(string) GameServerInfoInterface
	IsAccountInGameServer(account string) bool
}

func NewRequestAuthLogin(request []byte, client ClientRequestInterface, loginServer loginServerInterface, gameServer gameServerInterface, db database.Database) error {
	err := validate(request, client, db)
	if err != nil {
		err = client.Send(ls2c.NewLoginFailPacket(reasons.LoginOrPassWrong))
		return err
	}
	reason := tryCheckinAccount(client, loginServer, gameServer)

	switch reason {
	default:
		err = client.Send(ls2c.NewLoginFailPacket(reasons.SystemError))
	case clientState.AuthSuccess:
		client.SetState(clientState.AuthedLogin)
		client.SetSessionKey(loginServer.AssignSessionKeyToClient(client))
		err = client.Send(ls2c.NewLoginOkPacket(client))
		sendCharactersOnAccount(client.GetAccountLogin(), loginServer)
	case clientState.AccountBanned:
		err = client.Send(ls2c.NewLoginFailPacket(reasons.Ban))
		client.CloseConnection()
	case clientState.AlreadyOnLs:
		account := client.GetAccountLogin()
		err = client.Send(ls2c.NewLoginFailPacket(reasons.AccountInUse))
		oldClient := loginServer.GetClientCtx(account)
		if oldClient != nil {
			err = oldClient.Send(ls2c.AccountKicked(reasons.AccountInUse))
			if err != nil {
				loginServer.RemoveAuthedLoginClient(account)
				return err
			}
			oldClient.CloseConnection()
			loginServer.RemoveAuthedLoginClient(account)
		}
		client.CloseConnection()
	case clientState.AlreadyOnGs:
		account := client.GetAccountLogin()
		err = client.Send(ls2c.NewLoginFailPacket(reasons.AccountInUse))
		gsi := gameServer.GetAccountOnGameServer(account)
		if gsi != nil {
			if gsi.IsAuthed() {
				_ = gsi.SendSlice(ls2gs.KickPlayer(account))
			}
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func tryCheckinAccount(client ClientRequestInterface, server loginServerInterface, gameServer gameServerInterface) clientState.ClientAuthState {
	if client.GetAccountAccessLevel() < 0 {
		return clientState.AccountBanned
	}

	if gameServer.IsAccountInGameServer(client.GetAccountLogin()) {
		return clientState.AlreadyOnGs
	}

	if server.IsAccountInLoginAndAddIfNot(client) {
		return clientState.AlreadyOnLs
	}
	return clientState.AuthSuccess
}

func validate(request []byte, client ClientRequestInterface, db database.Database) error {
	if cap(request) < 128 {
		return errNoData
	}
	payload := request[:128]

	c := new(big.Int).SetBytes(payload)
	privateKey := client.GetPrivateKey()
	decodeData := c.Exp(c, privateKey.D, privateKey.N).Bytes()

	if cap(decodeData) < 28 {
		return errNoData
	}

	trimLogin := bytes.Trim(decodeData[1:14], string(rune(0)))
	trimPassword := bytes.Trim(decodeData[14:28], string(rune(0)))

	login := string(trimLogin)
	password := string(trimPassword)

	var accountPassword string
	var accountCreatedAt, accountLastActive pgtype.Timestamp
	var accountAccessLevel, accountLastServer int8
	var accountLastIp sql.NullString

	reg := trace.StartRegion(context.Background(), "userInfoSelect")
	err := db.QueryRow(context.Background(), UserInfoSelect, login).
		Scan(&accountPassword, &accountCreatedAt, &accountLastActive, &accountAccessLevel, &accountLastIp, &accountLastServer)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) && config.AutoCreateAccounts() {
			err = createAccount(login, password, db)
			if err != nil {
				return err
			}
			return validate(request, client, db)
		}

		return err
	}
	reg.End()

	err = bcrypt.CompareHashAndPassword([]byte((accountPassword)), []byte(password))
	if err != nil {
		return err
	}

	_, err = db.Exec(context.Background(), UserLastInfo, client.GetRemoteAddr().String(), time.Now(), login)
	if err != nil {
		log.Println(err)
		return err
	}

	client.SetAccount(login, accountPassword, accountCreatedAt, accountLastActive, accountAccessLevel, accountLastServer, accountLastIp)
	return nil
}

func createAccount(clearLogin, clearPassword string, db database.Database) error {
	password, err := bcrypt.GenerateFromPassword([]byte(clearPassword), 10)
	if err != nil {
		return err
	}

	_, err = db.Exec(context.Background(), AccountsInsert, clearLogin, string(password))
	return err
}

func sendCharactersOnAccount(account string, server loginServerInterface) {
	serverList := server.GetGameServerInfoList()
	for _, gsi := range serverList {
		if gsi.IsAuthed() {
			_ = gsi.SendSlice(ls2gs.RequestCharacter(account))
		}
	}
}
