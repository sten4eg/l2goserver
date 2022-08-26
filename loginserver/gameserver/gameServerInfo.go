package gameserver

import (
	"crypto/rand"
	"crypto/rsa"
	"l2goserver/loginserver/crypt/blowfish"
	"l2goserver/loginserver/gameserver/network/gs2ls"
	"l2goserver/loginserver/types/gameServerStatuses"
	"l2goserver/loginserver/types/state/gameServer"
	"net"
	"net/netip"
	"sync"
)

type account struct {
	accounts map[string]bool
	mu       sync.Mutex
}

type Info struct {
	host            []netip.Prefix
	hexId           []byte
	id              byte
	port            int16
	maxPlayer       int32
	authed          bool
	status          gameServerStatuses.ServerStatusValues
	serverType      int32
	ageLimit        int32
	showBracket     bool
	state           gameServer.GameServerState
	accounts        account
	privateKey      *rsa.PrivateKey
	conn            *net.TCPConn
	blowfish        *blowfish.Cipher
	mu              sync.Mutex
	gameServerTable *Table
	uniqId          byte
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

func (gsi *Info) GetId() byte {
	return gsi.id
}

func (gsi *Info) getPort() int16 {
	return gsi.port
}

func (gsi *Info) GetMaxPlayer() int32 {
	return gsi.maxPlayer
}

func (gsi *Info) GetCurrentPlayerCount() int32 {
	return int32(len(gsi.accounts.accounts))
}

func (gsi *Info) getAgeLimit() int32 {
	return gsi.ageLimit
}

func (gsi *Info) GetType() int32 {
	return gsi.serverType
}

func (gsi *Info) getStatus() int32 {
	return gsi.status
}

func (gsi *Info) getShowBracket() bool {
	return gsi.showBracket
}

func (gsi *Info) HasAccountOnGameServer(account string) bool {
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

func (gsi *Info) GetStatus() gameServerStatuses.ServerStatusValues {
	return gsi.status
}

func (gsi *Info) GetGsiById(serverId byte) gs2ls.GsiIsAuthInterface {
	gsi_ := gsi.gameServerTable.GetGameServerById(serverId)
	if gsi_ == nil {
		return nil
	}
	return gsi_

}
