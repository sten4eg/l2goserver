package gameserver

import (
	"crypto/rsa"
	"fmt"
	"l2goserver/config"
	"l2goserver/loginserver/crypt"
	"l2goserver/loginserver/crypt/blowfish"
	"l2goserver/loginserver/network/gs2ls"
	"l2goserver/loginserver/network/ls2gs"
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
	"log"
	"net"
	"strconv"
)

type GS struct {
	Connection      *net.TCPListener
	gameServersInfo []*GameServerInfo
	loginServerInfo LoginServInterface
}

var gameServerInstance *GS
var initBlowfishKey = []byte{95, 59, 118, 46, 93, 48, 53, 45, 51, 49, 33, 124, 43, 45, 37, 120, 84, 33, 94, 91, 36, 0}

func (gsi *GameServerInfo) AddAccountOnGameServer(account string) {
	gsi.accounts.mu.Lock()
	gsi.accounts.accounts[account] = true
	gsi.accounts.mu.Unlock()
}

func (gsi *GameServerInfo) RemoveAccountOnGameServer(account string) {
	gsi.accounts.mu.Lock()
	delete(gsi.accounts.accounts, account)
	gsi.accounts.mu.Unlock()
}

func (gsi *GameServerInfo) SetInfoGameServerInfo(host string, hexId []byte, id byte, port int16, maxPlayer int32, authed bool) {
	gsi.host = host
	gsi.hexId = hexId
	gsi.Id = id
	gsi.port = port
	gsi.maxPlayer = maxPlayer
	gsi.authed = authed
	gsi.accounts.mu.Lock()
	gsi.accounts.accounts = make(map[string]bool, maxPlayer)
	gsi.accounts.mu.Unlock()
}

func (gsi *GameServerInfo) SetCharactersOnServer(account string, charsNum uint8, timeToDel []int64) {
	client := gsi.gs.loginServerInfo.GetAccount(account)

	if client == nil {
		return
	}

	if charsNum > 0 {
		client.SetCharsOnServer(gsi.Id, charsNum)
	}

	if len(timeToDel) > 0 {
		client.SetCharsWaitingDelOnServer(gsi.Id, timeToDel)
	}
}

func GameServerHandlerInit() {
	gameServerInstance = new(GS)

	port := config.GetLoginPortForGameServer()
	intPort, err := strconv.Atoi(port)
	if err != nil {
		panic(err)
	}

	addr := new(net.TCPAddr)
	addr.Port = intPort
	addr.IP = net.IP{127, 0, 0, 1}

	listener, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		panic(err)
	}
	gameServerInstance.Connection = listener

	go gameServerInstance.Run()
}

func (gs *GS) Run() {
	for {
		var err error
		gsi := new(GameServerInfo)
		gsi.gs = gs

		gsi.SetBlowFishKey(initBlowfishKey)

		gsi.conn, err = gs.Connection.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}

		gsi.state = state.CONNECTED
		gsi.InitRSAKeys()

		gs.gameServersInfo = append(gs.gameServersInfo, gsi)

		pubKey := make([]byte, 1, 65)
		pubKey = append(pubKey, gsi.privateKey.PublicKey.N.Bytes()...)

		buf := ls2gs.InitLS(pubKey)

		gsi.Send(buf)
		go gsi.Listen()
	}
}

func (gsi *GameServerInfo) Listen() {
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
			_ = gsi.conn.Close()
			return
		}
		gsi.HandlePackage(data)
	}
}
func (gsi *GameServerInfo) HandlePackage(data []byte) {
	opcode := data[0]
	data = data[1:]
	fmt.Println(opcode)

	switch gsi.state {
	case state.CONNECTED:
		if opcode == 0 {
			gs2ls.BlowFishKey(data, gsi)
		}
	case state.BfConnected:
		if opcode == 1 {
			gs2ls.GameServerAuth(data, gsi)
		}
	case state.AUTHED:

		switch opcode {
		case 0x02:
			gs2ls.PlayerInGame(data, gsi)
		case 0x03:
			gs2ls.PlayerLogout(data, gsi)
		case 0x04:
			gs2ls.ChangeAccessLevel(data)
		case 0x05:
			gs2ls.PlayerAuthRequest(data, gsi)
		case 0x06:
			gs2ls.ServerStatus(data, gsi)
		case 0x07:
			gs2ls.PlayerTracert(data)
		case 0x08:
			gs2ls.ReplyCharacters(data, gsi)
		case 0x0A:
			gs2ls.RequestTempBan(data)
		}

	}
}
func (gsi *GameServerInfo) Send(buf *packets.Buffer) {
	size := buf.Len() + 4
	size = (size + 8) - (size % 8) // padding

	data := make([]byte, 200)
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

	gsi.mu.Lock()
	_, err := gsi.conn.Write(res)
	gsi.mu.Unlock()

	if err != nil {
		panic(err)
	}
}

func (gsi *GameServerInfo) GetPrivateKey() *rsa.PrivateKey {
	return gsi.privateKey
}
func (gsi *GameServerInfo) SetBlowFishKey(key []byte) {
	localKey := make([]byte, len(key))
	copy(localKey, key)
	cipher, err := blowfish.NewCipher(localKey)
	if err != nil {
		panic(err)
	}
	gsi.blowfish = cipher
}
func (gsi *GameServerInfo) SetState(state state.GameServerState) {
	gsi.state = state
}
func (gsi *GameServerInfo) ForceClose(reason state.LoginServerFail) {
	gsi.Send(ls2gs.LoginServerFail(reason))
	err := gsi.conn.Close()
	if err != nil {
		log.Println(err)
	}

}

func (gsi *GameServerInfo) SetStatus(status int32) {
	gsi.status = status
}
func (gsi *GameServerInfo) SetShowBracket(showBracket bool) {
	gsi.showBracket = showBracket
}
func (gsi *GameServerInfo) SetMaxPlayer(maxPlayer int32) {
	gsi.maxPlayer = maxPlayer
}
func (gsi *GameServerInfo) SetServerType(serverType int32) {
	gsi.serverType = serverType
}
func (gsi *GameServerInfo) SetAgeLimit(ageLimit int32) {
	gsi.ageLimit = ageLimit
}

func (gsi *GameServerInfo) GetServerInfoId() byte {
	return gsi.Id
}

func (gs *GS) GetGameServerInfoList() []*GameServerInfo {
	return gs.gameServersInfo
}
