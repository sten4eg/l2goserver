package gs2ls

import (
	"crypto/rsa"
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
	"math/big"
)

type GsInterfaceBF interface {
	GetPrivateKey() *rsa.PrivateKey
	SetBlowFishKey([]byte)
	SetState(state.GameServerState)
}

func BlowFishKey(data []byte, client GsInterfaceBF) {
	packet := packets.NewReader(data)
	_ = packet.ReadSingleByte() // пропускаем опкод

	size := packet.ReadInt32()
	tempKey := packet.ReadBytes(int(size))

	c := new(big.Int).SetBytes(tempKey)

	decodeData := c.Exp(c, client.GetPrivateKey().D, client.GetPrivateKey().N).Bytes()

	client.SetBlowFishKey(decodeData)
	client.SetState(state.BfConnected)
}
