package gs2ls

import (
	"crypto/rsa"
	"l2goserver/loginserver/types/state/gameServer"
	"l2goserver/packets"
	"math/big"
)

type GsInterfaceBF interface {
	GetPrivateKey() *rsa.PrivateKey
	SetBlowFishKey([]byte)
	SetState(serverState gameServer.GameServerState)
}

func BlowFishKey(data []byte, client GsInterfaceBF) {
	packet := packets.NewReader(data)

	size := packet.ReadInt32()
	tempKey := packet.ReadBytes(int(size))

	c := new(big.Int).SetBytes(tempKey)

	decodeData := c.Exp(c, client.GetPrivateKey().D, client.GetPrivateKey().N).Bytes()

	client.SetBlowFishKey(decodeData)
	client.SetState(gameServer.BfConnected)
}
