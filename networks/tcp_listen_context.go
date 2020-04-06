package networks

import (
	"sync"
)

var tcpListenContext *TcpListenContext
var tcpListenContextOnce sync.Once

func GetTcpListenContext() *TcpListenContext {
	tcpListenContextOnce.Do(func() {
		tcpListenContext = &TcpListenContext{
			Listener: newTcpListener(),
		}
	})

	return tcpListenContext
}

type TcpListenContext struct {
	Listener INetListener
}
