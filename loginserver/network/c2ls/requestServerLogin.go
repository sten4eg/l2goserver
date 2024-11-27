package c2ls

import (
	"errors"
	"l2goserver/loginserver/network/ls2c"
	"l2goserver/loginserver/types/reason/clientReasons"
	"l2goserver/packets"
)

type IsLoginPossibleInterface interface {
	IsLoginPossible(ClientServerLogin, byte) (bool, error)
}

type ClientServerLogin interface {
	Send([]byte) error
	GetSessionKey() (uint32, uint32, uint32, uint32)
	SetJoinedGS(bool)
	GetSessionPlayOK1() uint32
	GetSessionPlayOK2() uint32
	GetAccountAccessLevel() int8
	GetLastServer() int8
	GetAccountLogin() string
}

var errServerOverload = errors.New("serverOverload")

func RequestServerLogin(request []byte, client ClientServerLogin, server IsLoginPossibleInterface) error {
	var packet = packets.NewReader(request)
	var err error

	key1 := packet.ReadUInt32()
	key2 := packet.ReadUInt32()
	serverId := packet.ReadUInt8()

	loginOk1, loginOk2, _, _ := client.GetSessionKey()

	if key1 != loginOk1 || key2 != loginOk2 {
		err = client.Send(ls2c.NewLoginFailPacket(clientReasons.AccessFailed))
		if err != nil {
			return err
		}
		return errServerOverload
	}
	loginOk, err := server.IsLoginPossible(client, serverId)
	if err != nil {
		return err
	}

	if loginOk {
		client.SetJoinedGS(true)
		err = client.Send(ls2c.NewPlayOkPacket(client))
	} else {
		err = client.Send(ls2c.NewPlayFailPacket(clientReasons.ServerOverloaded))
		return errServerOverload
	}

	return err
}
