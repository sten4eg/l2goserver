package gameserverpackets

import (
	"crypto/rsa"
	"l2goserver/loginserver/types/state"
	"l2goserver/packets"
	"math/big"
)

type GsInterfaceBF interface {
	GetPrivKey() *rsa.PrivateKey
	SetBlowFishKey([]byte)
	SetState(state.GameServerState)
}

func BlowFishKey(data []byte, client GsInterfaceBF) {
	packet := packets.NewReader(data)
	_ = packet.ReadSingleByte() // пропускаем опкод

	size := packet.ReadInt32()
	tempKey := packet.ReadBytes(int(size))

	c := new(big.Int).SetBytes(tempKey)

	decodeData := c.Exp(c, client.GetPrivKey().D, client.GetPrivKey().N).Bytes()

	client.SetBlowFishKey(decodeData)
	client.SetState(state.BF_CONNECTED)
}
