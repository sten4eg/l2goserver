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
	reasons "l2goserver/loginserver/types/reason"
	"l2goserver/loginserver/types/state"
	"l2goserver/utils"
	"log"
	"math/big"
	"runtime/trace"
	"time"
)

const UserInfoSelect = "SELECT accounts.login,accounts.password,accounts.created_at,accounts.last_active,accounts.access_level,accounts.last_ip,accounts.last_server FROM accounts WHERE accounts.login = $1"
const UserLastInfo = "UPDATE accounts SET last_ip = $1 , last_active = $2 WHERE login = $3"
const AccountsInsert = "INSERT INTO accounts (login, password) VALUES ($1, $2)"

var errNoData = errors.New("errNoData")

type isInLoginInterface interface {
	IsAccountInLoginAndAddIfNot(*models.ClientCtx) bool
	AssignSessionKeyToClient(*models.ClientCtx) *models.SessionKey
	GetGameServerInfoList() []*gameserver.Info
	GetClientCtx(string) *models.ClientCtx
	RemoveAuthedLoginClient(string)
}

func NewRequestAuthLogin(request []byte, client *models.ClientCtx, server isInLoginInterface) error {
	err := validate(request, client)
	if err != nil {
		err = client.SendBuf(ls2c.NewLoginFailPacket(reasons.LoginOrPassWrong))
		return err
	}
	reason := tryCheckinAccount(client, server)

	switch reason {
	default:
		err = client.SendBuf(ls2c.NewLoginFailPacket(reasons.SystemError))
	case reasons.AUTH_SUCCESS:
		client.SetState(state.AuthedLogin)
		client.SetSessionKey(server.AssignSessionKeyToClient(client))
		err = client.SendBuf(ls2c.NewLoginOkPacket(client))
		getCharactersOnAccount(client.Account.Login, server)
	case reasons.ACCOUNT_BANNED:
		err = client.SendBuf(ls2c.NewLoginFailPacket(reasons.Ban))
		client.CloseConnection()
	case reasons.ALREADY_ON_LS:
		account := client.Account.Login
		err = client.SendBuf(ls2c.NewLoginFailPacket(reasons.AccountInUse))
		oldClient := server.GetClientCtx(account)
		if oldClient != nil {
			err = oldClient.SendBuf(ls2c.AccountKicked(reasons.AccountInUse))
			oldClient.CloseConnection()
			server.RemoveAuthedLoginClient(account)
		}
		client.CloseConnection()
	case reasons.ALREADY_ON_GS:
		account := client.Account.Login
		err = client.SendBuf(ls2c.NewLoginFailPacket(reasons.AccountInUse))
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

func tryCheckinAccount(client *models.ClientCtx, server isInLoginInterface) reasons.AuthLoginResult {
	if client.Account.AccessLevel < 0 {
		return reasons.ACCOUNT_BANNED
	}

	ret := reasons.ALREADY_ON_GS
	if gameserver.IsAccountInGameServer(client.Account.Login) {
		return ret
	}
	ret = reasons.ALREADY_ON_LS
	if server.IsAccountInLoginAndAddIfNot(client) {
		return ret
	}
	return reasons.AUTH_SUCCESS
}

func validate(request []byte, client *models.ClientCtx) error {
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
		log.Println("2")
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

	err = bcrypt.CompareHashAndPassword(utils.S2b(account.Password), utils.S2b(password))
	if err != nil {
		return err
	}

	_, err = dbConn1.Exec(context.Background(), UserLastInfo, client.GetRemoteAddr().String(), time.Now(), login)
	if err != nil {
		log.Println('-')
		return err
	}

	if err != nil {
		log.Println('(')
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
		clearLogin, string(password))
	return err
}

func getCharactersOnAccount(account string, server isInLoginInterface) {
	serverList := server.GetGameServerInfoList()
	for _, gsi := range serverList {
		if gsi.IsAuthed() {
			_ = gsi.Send(ls2gs.RequestCharacter(account))
		}
	}
}
