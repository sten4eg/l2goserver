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
type GameServerInfo struct {
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
}

func (gsi *GameServerInfo) InitRSAKeys() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		panic(err)
	}
	gsi.privateKey = privateKey
}

func (gsi *GameServerInfo) getGameServerInfoPort() int16 {
	return gsi.port
}
func (gsi *GameServerInfo) getGameServerInfoId() byte {
	return gsi.Id
}
func (gsi *GameServerInfo) getGameServerInfoMaxPlayer() int32 {
	return gsi.maxPlayer
}
func (gsi *GameServerInfo) getGameServerInfoAgeLimit() int32 {
	return gsi.ageLimit
}
func (gsi *GameServerInfo) getGameServerInfoType() int32 {
	return gsi.serverType
}

func (gsi *GameServerInfo) getGameServerInfoStatus() int32 {
	return gsi.status
}
func (gsi *GameServerInfo) getGameServerInfoShowBracket() bool {
	return gsi.showBracket
}

func (gsi *GameServerInfo) hasAccountOnGameServer(account string) bool {
	gsi.accounts.mu.Lock()
	defer gsi.accounts.mu.Unlock()
	inGame, ok := gsi.accounts.accounts[account]
	if !ok {
		return false
	}
	return inGame
}

func (gsi *GameServerInfo) IsAuthed() bool {
	return gsi.authed
}
