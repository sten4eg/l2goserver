package reason

type Reason byte

const (
	NoReason               Reason = 0x00
	SystemError            Reason = 0x01
	LoginOrPassWrong       Reason = 0x03
	AccessFailed           Reason = 0x04
	InfoWrong              Reason = 0x05
	AccountInUse           Reason = 0x07
	Ban                    Reason = 0x09
	REASON_MAINTENANCE     Reason = 0x10
	REASON_CHANGE_TMP_PASS Reason = 0x11
	REASON_EXPIRED         Reason = 0x12
	REASON_NO_TIME_LEFT    Reason = 0x13
)

const (
	ServerOverloaded Reason = 0x0F
)
