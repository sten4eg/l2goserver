package gameserver

import (
	"crypto/rsa"
	"fmt"
	"l2goserver/config"
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
	"strconv"
)

type Table struct {
	Connection      *net.TCPListener
	gameServersInfo []*Info
	loginServerInfo LoginServInterface
}

var gameServerInstance *Table
var initBlowfishKey = []byte{95, 59, 118, 46, 93, 48, 53, 45, 51, 49, 33, 124, 43, 45, 37, 120, 84, 33, 94, 91, 36, 0}
var uniqId byte

func HandlerInit() error {
	gameServerInstance = new(Table)
	uniqId = 1
	port := config.GetLoginPortForGameServer()
	intPort, err := strconv.Atoi(port)
	if err != nil {
		return err
	}

	addr := new(net.TCPAddr)
	addr.Port = intPort
	addr.IP = net.IP{127, 0, 0, 1}

	listener, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		return err
	}
	gameServerInstance.Connection = listener

	go gameServerInstance.Run()
	return nil
}

func (t *Table) Run() {
	for {
		var err error
		gsi := new(Info)
		gsi.gameServerTable = t
		gsi.uniqId = uniqId //atomic?
		uniqId++

		gsi.SetBlowFishKey(initBlowfishKey)

		gsi.conn, err = t.Connection.AcceptTCP()
		if err != nil {
			log.Println("ошибка при Accept gameserver")
			continue
		}

		gsi.state = gameServer.Connected
		err = gsi.InitRSAKeys()
		if err != nil {
			log.Println("ошибка при создании ключа для геймсервера")
			_ = gsi.conn.Close()
			continue
		}

		t.gameServersInfo = append(t.gameServersInfo, gsi)

		pubKey := make([]byte, 1, 65)
		pubKey = append(pubKey, gsi.privateKey.PublicKey.N.Bytes()...)

		buf := ls2gs.InitLS(pubKey)

		err = gsi.Send(buf)
		if err != nil {
			log.Println("ошибка при отправке в геймсервера")
			gameServerInstance.RemoveGsi(gsi.uniqId)
			continue
		}

		go gsi.Listen()
	}
}

func (t *Table) RemoveGsi(connId byte) {
	for i, gsi := range t.gameServersInfo {
		if gsi.uniqId == connId {
			t.gameServersInfo = append(t.gameServersInfo[:i], t.gameServersInfo[i+1:]...)
		}
	}
}

func (t *Table) GetAccountOnGameServer(account string) *Info {
	for _, gsi := range t.GetGameServerInfoList() {
		if gsi.HasAccountOnGameServer(account) {
			return gsi
		}
	}
	return nil
}

func (t *Table) GetGameServerById(serverId byte) *Info {
	for _, gsi := range t.gameServersInfo {
		if gsi.GetId() == serverId {
			return gsi
		}
	}
	return nil
}

func (gsi *Info) AddAccountOnGameServer(account string) {
	gsi.accounts.mu.Lock()
	gsi.accounts.accounts[account] = true
	gsi.accounts.mu.Unlock()
}

func (gsi *Info) RemoveAccountOnGameServer(account string) {
	gsi.accounts.mu.Lock()
	delete(gsi.accounts.accounts, account)
	gsi.accounts.mu.Unlock()
}

func (gsi *Info) SetInfoGameServerInfo(host []netip.Prefix, hexId []byte, id byte, port int16, maxPlayer int32, authed bool) {
	gsi.host = host //todo unused?
	gsi.hexId = hexId
	gsi.id = id
	gsi.port = port
	gsi.maxPlayer = maxPlayer
	gsi.authed = authed
	gsi.accounts.mu.Lock()
	gsi.accounts.accounts = make(map[string]bool, maxPlayer)
	gsi.accounts.mu.Unlock()
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

func (gsi *Info) HandlePacket(data []byte) error {
	opcode := data[0]
	data = data[1:]
	fmt.Println(opcode)

	switch gsi.state {
	case gameServer.Connected:
		if opcode == 0 {
			gs2ls.BlowFishKey(data, gsi)
		}
	case gameServer.BfConnected:
		if opcode == 1 {
			err := gs2ls.GameServerAuth(data, gsi)
			if err != nil {
				return err
			}
		}
	case gameServer.Authed:

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
	return nil
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

	gsi.mu.Lock()
	_, err := gsi.conn.Write(res)
	gsi.mu.Unlock()

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

func (t *Table) GetGameServerInfoList() []*Info {
	return t.gameServersInfo
}
