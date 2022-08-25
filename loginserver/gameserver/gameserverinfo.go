package gameserver

import (
	"crypto/rand"
	"crypto/rsa"
	"l2goserver/loginserver/crypt/blowfish"
	"l2goserver/loginserver/types/state"
	"net"
	"sync"
)

type account struct {
	accounts map[string]bool
	mu       sync.Mutex
}
type Info struct {
	host        string
	hexId       []byte
	Id          byte
	port        int16
	maxPlayer   int32
	authed      bool
	status      int32
	serverType  int32
	ageLimit    int32
	showBracket bool
	state       state.GameServerState
	accounts    account
	privateKey  *rsa.PrivateKey
	conn        *net.TCPConn
	blowfish    *blowfish.Cipher
	mu          sync.Mutex
	gs          *GS
	uniqId      byte
}

func (gsi *Info) InitRSAKeys() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		return err
	}
	gsi.privateKey = privateKey
	return nil
}

func (gsi *Info) IsAuthed() bool {
	return gsi.authed
}
func (gsi *Info) GetGameServerInfoId() byte {
	return gsi.Id
}

func (gsi *Info) getGameServerInfoPort() int16 {
	return gsi.port
}
func (gsi *Info) getGameServerInfoMaxPlayer() int32 {
	return gsi.maxPlayer
}
func (gsi *Info) getGameServerInfoAgeLimit() int32 {
	return gsi.ageLimit
}
func (gsi *Info) getGameServerInfoType() int32 {
	return gsi.serverType
}
func (gsi *Info) getGameServerInfoStatus() int32 {
	return gsi.status
}
func (gsi *Info) getGameServerInfoShowBracket() bool {
	return gsi.showBracket
}
func (gsi *Info) hasAccountOnGameServer(account string) bool {
	gsi.accounts.mu.Lock()
	defer gsi.accounts.mu.Unlock()
	inGame, ok := gsi.accounts.accounts[account]
	if !ok {
		return false
	}
	return inGame
}
func (gsi *Info) getGameServerConn() *net.TCPConn {
	return gsi.conn
}
