package c2ls

import (
	"errors"
	"l2goserver/loginserver/models"
	serverpackets2 "l2goserver/loginserver/network/ls2c"
	"l2goserver/loginserver/types/reason/clientReasons"
	"l2goserver/packets"
)

type IsLoginPossibleInterface interface {
	IsLoginPossible(*models.ClientCtx, byte) (bool, error)
}

var errServerOverload = errors.New("serverOverload")

func RequestServerLogin(request []byte, client *models.ClientCtx, server IsLoginPossibleInterface) error {
	var packet = packets.NewReader(request)
	var err error

	key1 := packet.ReadUInt32()
	key2 := packet.ReadUInt32()
	serverId := packet.ReadUInt8()

	_ = serverId

	if key1 != client.SessionKey.LoginOk1 || key2 != client.SessionKey.LoginOk2 {
		err = client.SendBuf(serverpackets2.NewLoginFailPacket(clientReasons.AccessFailed))
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
		err = client.SendBuf(serverpackets2.NewPlayOkPacket(client))
	} else {
		_ = client.SendBuf(serverpackets2.NewPlayFailPacket(clientReasons.ServerOverloaded))
		return errServerOverload
	}

	return err
}
