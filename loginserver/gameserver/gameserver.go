package gameserver

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"l2goserver/config"
	"l2goserver/loginserver/crypt"
	"l2goserver/loginserver/crypt/blowfish"
	gameserverpackets2 "l2goserver/loginserver/network/gameserverpackets"
	loginserverpackets2 "l2goserver/loginserver/network/loginserverpackets"
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
	"log"
	"net"
	"sync"
)

type GS struct {
	Connection      net.Listener
	privateKey      *rsa.PrivateKey
	blowfish        *blowfish.Cipher
	mu              sync.Mutex
	conn            net.Conn
	state           state.GameServerState
	gameServersInfo GameServerInfo
}

var gameServerInstance *GS

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
	accounts    account
}

func GetGameServerInstance() *GS {
	return gameServerInstance
}

func (gs *GS) AddAccountOnGameServer(account string) {
	gs.gameServersInfo.accounts.mu.Lock()
	gs.gameServersInfo.accounts.accounts[account] = true
	gs.gameServersInfo.accounts.mu.Unlock()
}
func (gs *GS) SetInfoGameServerInfo(host string, hexId []byte, id byte, port int16, maxPlayer int32, authed bool) {
	gs.gameServersInfo.host = host
	gs.gameServersInfo.hexId = hexId
	gs.gameServersInfo.Id = id
	gs.gameServersInfo.port = port
	gs.gameServersInfo.maxPlayer = maxPlayer
	gs.gameServersInfo.authed = authed
	gs.gameServersInfo.accounts.mu.Lock()
	gs.gameServersInfo.accounts.accounts = make(map[string]bool, maxPlayer)
	gs.gameServersInfo.accounts.mu.Unlock()
}

func (gs *GS) InitRSAKeys() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		panic(err)
	}
	gs.privateKey = privateKey

}

func GameServerHandlerInit() {
	gameServerInstance = new(GS)
	gameServerInstance.InitRSAKeys()

	port := config.GetLoginPortForGameServer()

	blowfishKey := []byte{95, 59, 118, 46, 93, 48, 53, 45, 51, 49, 33, 124, 43, 45, 37, 120, 84, 33, 94, 91, 36, 0}

	gameServerInstance.SetBlowFishKey(blowfishKey)

	listener, err := net.Listen("tcp4", ":"+port)
	if err != nil {
		panic(err)
	}
	gameServerInstance.Connection = listener

	go gameServerInstance.Run()
}

func (gs *GS) Run() {
	for {
		var err error

		gs.conn, err = gs.Connection.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		gs.state = state.CONNECTED

		pubKey := make([]byte, 1, 65)
		pubKey = append(pubKey, gs.privateKey.PublicKey.N.Bytes()...)

		buf := loginserverpackets2.InitLS(pubKey)

		gs.Send(buf)
		go gs.Listen()
	}
}

func (gs *GS) Listen() {
	for {
		header := make([]byte, 2)

		n, err := gs.conn.Read(header)
		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				return
			}
			log.Println(err)
		}
		dataSize := (int(header[0]) | int(header[1])<<8) - 2

		data := make([]byte, dataSize)
		n, err = gs.conn.Read(data)
		if err != nil {
			panic(err)
		}
		if n != dataSize {
			panic("qweqwedsaasdcg")
		}

		for i := 0; i < dataSize; i += 8 {
			gs.blowfish.Decrypt(data, data, i, i)
		}

		ok := crypt.VerifyCheckSum(data, dataSize)
		if !ok {
			fmt.Println("Неверная контрольная сумма пакета, закрытие соединения.")
			gs.conn.Close()
			return
		}
		gs.HandlePackage(data)
	}
}
func (gs *GS) HandlePackage(data []byte) {
	opcode := data[0]

	switch gs.state {
	case state.CONNECTED:
		if opcode == 0 {
			gameserverpackets2.BlowFishKey(data, gs)
		}
	case state.BF_CONNECTED:
		if opcode == 1 {
			gameserverpackets2.GameServerAuth(data, gs)
		}
	case state.AUTHED:
		switch opcode {
		case 0x02:
			gameserverpackets2.PlayerInGame(data, gs)
		case 0x06:
			gameserverpackets2.ServerStatus(data, gs)

		}

	}
}
func (gs *GS) Send(buf *packets.Buffer) {
	size := buf.Len() + 4
	size = (size + 8) - (size % 8) // padding

	data := make([]byte, 200)
	copy(data, buf.Bytes())
	packets.Put(buf)

	rs := crypt.AppendCheckSum(data, size)

	for i := 0; i < size; i += 8 {
		gs.blowfish.Encrypt(rs, rs, i, i)
	}

	rs = rs[:size]
	leng := len(rs) + 2

	s, f := byte(leng>>8), byte(leng&0xff)
	res := append([]byte{f, s}, rs...)

	gs.mu.Lock()
	_, err := gs.conn.Write(res)
	gs.mu.Unlock()

	if err != nil {
		panic(err)
	}
}

func (gs *GS) GetPrivateKey() *rsa.PrivateKey {
	return gs.privateKey
}
func (gs *GS) SetBlowFishKey(key []byte) {
	cipher, err := blowfish.NewCipher(key)
	if err != nil {
		panic(err)
	}
	gs.blowfish = cipher
}
func (gs *GS) SetState(state state.GameServerState) {
	gs.state = state
}
func (gs *GS) ForceClose(reason state.LoginServerFail) {
	gs.Send(loginserverpackets2.LoginServerFail(reason))
	err := gs.conn.Close()
	if err != nil {
		log.Println(err)
	}

}
func (gs *GS) GetGameServersInfoHexId() []byte {
	return gs.gameServersInfo.hexId
}
func (gs *GS) SetStatus(status int32) {
	gs.gameServersInfo.status = status
}
func (gs *GS) SetShowBracket(showBracket bool) {
	gs.gameServersInfo.showBracket = showBracket
}
func (gs *GS) SetMaxPlayer(maxPlayer int32) {
	gs.gameServersInfo.maxPlayer = maxPlayer
}
func (gs *GS) SetServerType(serverType int32) {
	gs.gameServersInfo.serverType = serverType
}
func (gs *GS) SetAgeLimit(ageLimit int32) {
	gs.gameServersInfo.ageLimit = ageLimit
}
func (gs *GS) GetServerInfoId() byte {
	return gs.gameServersInfo.Id
}
func (gs *GS) HasAccountOnGameServer(account string) bool {
	gs.gameServersInfo.accounts.mu.Lock()
	inGame, ok := gs.gameServersInfo.accounts.accounts[account]
	if !ok {
		return false
	}
	gs.gameServersInfo.accounts.mu.Unlock()
	return inGame
}