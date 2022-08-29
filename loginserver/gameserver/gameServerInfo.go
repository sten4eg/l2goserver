package gameserver

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/puzpuzpuz/xsync"
	"l2goserver/loginserver/crypt"
	"l2goserver/loginserver/crypt/blowfish"
	"l2goserver/loginserver/gameserver/network/gs2ls"
	"l2goserver/loginserver/gameserver/network/ls2gs"
	"l2goserver/loginserver/types/gameServerStatuses"
	"l2goserver/loginserver/types/reason/loginServer"
	"l2goserver/loginserver/types/state/gameServer"
	"l2goserver/packets"
	"log"
	"net"
	"net/netip"
)

type Info struct {
	showBracket     bool
	authed          bool
	id              byte
	state           gameServer.GameServerState
	port            int16
	uniqId          uint32
	maxPlayer       int32
	ageLimit        int32
	serverType      int32
	status          gameServerStatuses.ServerStatusValues
	privateKey      *rsa.PrivateKey
	conn            *net.TCPConn
	blowfish        *blowfish.Cipher
	gameServerTable *Table
	host            []netip.Prefix
	hexId           []byte
	accounts        *xsync.MapOf[bool]
}

func InitGameServerInfo() (*Info, error) {
	gsi := new(Info)
	gsi.accounts = xsync.NewMapOf[bool]()
	err := gsi.InitRSAKeys()
	return gsi, err
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
	return int32(gsi.accounts.Size())
}

func (gsi *Info) getAgeLimit() int32 {
	return gsi.ageLimit
}

func (gsi *Info) GetType() int32 {
	return gsi.serverType
}

func (gsi *Info) getShowBracket() bool {
	return gsi.showBracket
}

func (gsi *Info) HasAccountOnGameServer(account string) bool {
	inGame, ok := gsi.accounts.Load(account)
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

func (gsi *Info) AddAccountOnGameServer(account string) {
	gsi.accounts.Store(account, true)
}

func (gsi *Info) RemoveAccountOnGameServer(account string) {
	gsi.accounts.Delete(account)
}

func (gsi *Info) SetInfoGameServerInfo(host []netip.Prefix, hexId []byte, id byte, port int16, maxPlayer int32, authed bool) {
	gsi.host = host // todo unused?
	gsi.hexId = hexId
	gsi.id = id
	gsi.port = port
	gsi.maxPlayer = maxPlayer
	gsi.authed = authed
}

func (gsi *Info) SetCharactersOnServer(account string, charsNum uint8, timeToDel []int64) {
	accountInfo := gsi.gameServerTable.loginServerInfo.GetAccount(account)

	if accountInfo == nil {
		return
	}

	if charsNum > 0 {
		accountInfo.SetCharsOnServer(gsi.GetId(), charsNum)
	}

	if len(timeToDel) > 0 {
		accountInfo.SetCharsWaitingDelOnServer(gsi.GetId(), timeToDel)
	}
}

func (gsi *Info) Listen() {
	defer gameServerInstance.RemoveGsi(gsi.uniqId)

	for {
		header := make([]byte, 2)

		n, err := gsi.conn.Read(header)
		if err != nil {
			log.Println(err)
			return
		}
		if n != 2 {
			log.Println("Должно быть 2 байта размера")
			return
		}

		dataSize := (int(header[0]) | int(header[1])<<8) - 2

		data := make([]byte, dataSize)
		n, err = gsi.conn.Read(data)
		if err != nil {
			panic(err)
		}
		if n != dataSize {
			log.Println("Прочитанно байт меньше чем объявлено в размере пакета")
			return
		}

		for i := 0; i < dataSize; i += 8 {
			gsi.blowfish.Decrypt(data, data, i, i)
		}

		ok := crypt.VerifyCheckSum(data, dataSize)
		if !ok {
			fmt.Println("Неверная контрольная сумма пакета, закрытие соединения.")
			return
		}
		err = gsi.HandlePacket(data)
		if err != nil {
			return
		}
	}
}

func (gsi *Info) Send(buf *packets.Buffer) error {
	size := buf.Len() + 4
	size = (size + 8) - (size % 8) // padding

	data := make([]byte, size)
	copy(data, buf.Bytes())
	packets.Put(buf)

	rs := crypt.AppendCheckSum(data, size)

	for i := 0; i < size; i += 8 {
		gsi.blowfish.Encrypt(rs, rs, i, i)
	}

	rs = rs[:size]
	leng := len(rs) + 2

	s, f := byte(leng>>8), byte(leng&0xff)
	res := append([]byte{f, s}, rs...)

	_, err := gsi.conn.Write(res)

	if err != nil {
		return err
	}
	return err
}

func (gsi *Info) GetPrivateKey() *rsa.PrivateKey {
	return gsi.privateKey
}

func (gsi *Info) SetBlowFishKey(key []byte) {
	localKey := make([]byte, len(key))
	copy(localKey, key)
	cipher, err := blowfish.NewCipher(localKey)
	if err != nil {
		panic(err)
	}
	gsi.blowfish = cipher
}

func (gsi *Info) SetState(state gameServer.GameServerState) {
	gsi.state = state
}

func (gsi *Info) ForceClose(reason loginServer.FailReason) {
	_ = gsi.Send(ls2gs.LoginServerFail(reason))
	err := gsi.conn.Close()
	if err != nil {
		log.Println(err)
	}

}

func (gsi *Info) SetStatus(status gameServerStatuses.ServerStatusValues) {
	gsi.status = status
}

func (gsi *Info) SetShowBracket(showBracket bool) {
	gsi.showBracket = showBracket
}

func (gsi *Info) SetMaxPlayer(maxPlayer int32) {
	gsi.maxPlayer = maxPlayer
}

func (gsi *Info) SetServerType(serverType int32) {
	gsi.serverType = serverType
}

func (gsi *Info) SetAgeLimit(ageLimit int32) {
	gsi.ageLimit = ageLimit
}
