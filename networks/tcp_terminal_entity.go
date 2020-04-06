package networks

import (
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var tcpTerminalMap map[string]TcpTerminalEntity
var tcpTerminalLock sync.RWMutex

func init() {
	tcpTerminalMap = make(map[string]TcpTerminalEntity)
}

type TcpTerminalEntity struct {
	TerminalIden string
	Conn         *net.Conn
	LastTime     time.Time
}

func (the TcpTerminalEntity) GetTerminalIden() string {
	return the.TerminalIden
}

func (the TcpTerminalEntity) SetTerminalIden(terminalIden string) {
	the.TerminalIden = terminalIden
}

func (the TcpTerminalEntity) GetLastTime() time.Time {
	return the.LastTime
}

func (the TcpTerminalEntity) SetLastTime(lastTime time.Time) {
	the.LastTime = lastTime
}

func (the TcpTerminalEntity) GetSocketIden() interface{} {
	return the.Conn
}

func (this *TcpTerminalEntity) ShutdownConn() error {
	return (*this.Conn).Close()
}

func (the TcpTerminalEntity) SetSocketIden(socketIden interface{}) {
	the.Conn = socketIden.(*net.Conn)
}

func (the TcpTerminalEntity) GetRemoteEndpoint() (ip string, port int) {
	address := (*the.Conn).RemoteAddr().String()
	arr := strings.SplitN(address, ":", 2)
	ip = arr[0]
	port, _ = strconv.Atoi(arr[1])
	return
}

func RefreshTcpTerminal(terminalEntity TcpTerminalEntity, isNew bool) {
	tcpTerminalLock.Lock()
	defer tcpTerminalLock.Unlock()

	entity, ok := tcpTerminalMap[terminalEntity.GetTerminalIden()]
	if ok && isNew && *entity.Conn != nil {
		(*entity.Conn).Close()
	}
	terminalEntity.LastTime = time.Now()
	tcpTerminalMap[terminalEntity.GetTerminalIden()] = terminalEntity
}

func GetTcpTerminalEntity(terminalIden string) (terminalEntity TcpTerminalEntity) {
	tcpTerminalLock.RLock()
	defer tcpTerminalLock.RUnlock()
	terminalEntity, _ = tcpTerminalMap[terminalIden]
	return
}

func RemoveTcpTerminalEntity(terminalIden string) {
	tcpTerminalLock.RLock()
	defer tcpTerminalLock.RUnlock()
	terminalEntity, ok := tcpTerminalMap[terminalIden]
	if ok {
		(*terminalEntity.Conn).Close()
	}
	return
}
