package c2ls

import (
	"bytes"
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"l2goserver/config"
	"l2goserver/loginserver/gameserver"
	"l2goserver/loginserver/gameserver/network/ls2gs"
	"l2goserver/loginserver/models"
	"l2goserver/loginserver/network/ls2c"
	reasons "l2goserver/loginserver/types/reason/clientReasons"
	"l2goserver/loginserver/types/state/clientState"
	"log"
	"math/big"
	"time"
)

const UserInfoSelect = "SELECT accounts.login,accounts.password,accounts.created_at,accounts.last_active,accounts.role,accounts.last_ip,accounts.last_server FROM accounts WHERE accounts.login = $1"
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

func RequestAuthLogin(request []byte, client *models.ClientCtx, server isInLoginInterface, db *sql.DB) error {
	err := validate(request, client, db, 1)
	if err != nil {
		err = client.SendBuf(ls2c.NewLoginFailPacket(reasons.LoginOrPassWrong))
		return err
	}
	reason := tryCheckinAccount(client, server)

	switch reason {
	default:
		err = client.SendBuf(ls2c.NewLoginFailPacket(reasons.SystemError))
	case clientState.AuthSuccess:
		client.SetState(clientState.AuthedLogin)
		client.SetSessionKey(server.AssignSessionKeyToClient(client))
		err = client.SendBuf(ls2c.NewLoginOkPacket(client))
		sendCharactersOnAccount(client.Account.Login, server)
	case clientState.AccountBanned:
		err = client.SendBuf(ls2c.NewLoginFailPacket(reasons.Ban))
		client.CloseConnection()
	case clientState.AlreadyOnLs:
		account := client.Account.Login
		err = client.SendBuf(ls2c.NewLoginFailPacket(reasons.AccountInUse))
		oldClient := server.GetClientCtx(account)
		if oldClient != nil {
			err = oldClient.SendBuf(ls2c.AccountKicked(reasons.AccountInUse))
			oldClient.CloseConnection()
			server.RemoveAuthedLoginClient(account)
		}
		client.CloseConnection()
	case clientState.AlreadyOnGs:
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

func tryCheckinAccount(client *models.ClientCtx, server isInLoginInterface) clientState.ClientAuthState {

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

func validate(request []byte, client *models.ClientCtx, db *sql.DB, attempt int) error {
	if attempt > 2 {
		return errors.New("too many attempts")
	}
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

	err := db.QueryRow(UserInfoSelect, login).
		Scan(&account.Login, &account.Password, &account.CreatedAt, &account.LastActive, &account.AccessLevel, &account.LastIp, &account.LastServer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) && config.AutoCreateAccounts() {
			err = createAccount(login, password, db)
			if err != nil {
				return err
			}
			return validate(request, client, db, attempt+1)
		}

		return err
	}

	if config.IsOnlyForAdmin() {
		if account.AccessLevel != "admin" {
			return errors.New("login server only for admin")
		}
	}
	err = bcrypt.CompareHashAndPassword([]byte((account.Password)), []byte(password))
	if err != nil {
		return err
	}

	_, err = db.Exec(UserLastInfo, client.GetRemoteAddr().String(), time.Now(), login)
	if err != nil {
		log.Println(err)
		return err
	}

	client.Account = account
	return nil
}

func createAccount(clearLogin, clearPassword string, db *sql.DB) error {
	password, err := bcrypt.GenerateFromPassword([]byte(clearPassword), 10)
	if err != nil {
		return err
	}
	_, err = db.Exec(AccountsInsert,
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
