package c2ls

import (
	"errors"
	serverpackets2 "l2goserver/loginserver/network/ls2c"
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
		err = client.Send(serverpackets2.NewLoginFailPacket(clientReasons.AccessFailed))
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
		err = client.Send(serverpackets2.NewPlayOkPacket(client))
	} else {
		err = client.Send(serverpackets2.NewPlayFailPacket(clientReasons.ServerOverloaded))
		return errServerOverload
	}

	return err
}
