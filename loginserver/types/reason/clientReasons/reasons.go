package clientReasons

type ClientLoginFailed byte

const (
	NoReason               ClientLoginFailed = 0x00
	SystemError            ClientLoginFailed = 0x01
	LoginOrPassWrong       ClientLoginFailed = 0x03
	AccessFailed           ClientLoginFailed = 0x04
	InfoWrong              ClientLoginFailed = 0x05
	AccountInUse           ClientLoginFailed = 0x07
	Ban                    ClientLoginFailed = 0x09
	REASON_MAINTENANCE     ClientLoginFailed = 0x10
	REASON_CHANGE_TMP_PASS ClientLoginFailed = 0x11
	REASON_EXPIRED         ClientLoginFailed = 0x12
	REASON_NO_TIME_LEFT    ClientLoginFailed = 0x13
	ServerOverloaded       ClientLoginFailed = 0x0F
	PermanentlyBanned      ClientLoginFailed = 0x20
)
