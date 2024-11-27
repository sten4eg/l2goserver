package network

type Sender interface {
	Send([]byte)
}

type Ls2c interface {
	GetSessionLoginOK1() uint32
	GetSessionLoginOK2() uint32
	GetSessionPlayOK1() uint32
	GetSessionPlayOK2() uint32
	GetSessionId() uint32
	GetScrambleModulus() []byte
	GetBlowFish() []byte
}

type C2ls interface {
	Send([]byte) error
}
