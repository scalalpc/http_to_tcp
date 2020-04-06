package networks

import (
	"time"
)

type ITerminalEntity interface {
	GetTerminalIden() string
	SetTerminalIden(string)
	GetLastTime() time.Time
	SetLastTime(time.Time)
	GetSocketIden() interface{}
	SetSocketIden(interface{})
	GetRemoteEndpoint() (ip string, port int)
}
