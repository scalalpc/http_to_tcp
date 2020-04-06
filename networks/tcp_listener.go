package networks

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"runtime/debug"
	"time"

	"http_to_tcp/globals"
)

type myTcpListener struct {
	listener net.Listener
	stopSign globals.IStopSign
}

func newTcpListener() INetListener {
	instance := myTcpListener{}
	return &instance
}

func (this *myTcpListener) Start(host string, port int) (err error) {
	if this.stopSign == nil {
		this.stopSign = globals.NewStopSign()
	} else {
		this.stopSign.Reset()
	}

	this.listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return
	}
	defer this.listener.Close()

	for {
		var conn net.Conn
		conn, err = this.listener.Accept()
		if err != nil {
			fmt.Println(fmt.Sprintf("Error accepting connection request, err: %v", err))
			time.Sleep(time.Second * 1)
			conn.Close()
			continue
		}

		terminalEntity := TcpTerminalEntity{
			Conn:     &conn,
			LastTime: time.Now(),
		}

		if globals.ByteArrayPool.FreeCount() > 0 {
			go this.receiveMessage(terminalEntity)
		} else {
			fmt.Println("Insufficient buffer pool resources, remote connection not accepted")
			time.Sleep(time.Second * 1)
			continue
		}
	}
	return
}

func (this *myTcpListener) Stop() error {
	this.stopSign.Sign()
	return this.listener.Close()
}

func (this *myTcpListener) RefreshTerminal(terminalEntity ITerminalEntity, isNewConn bool) {
	RefreshTcpTerminal(terminalEntity.(TcpTerminalEntity), isNewConn)
}

func (this *myTcpListener) GetTerminalEntity(terminalIden string) (terminalEntity ITerminalEntity) {
	return GetTcpTerminalEntity(terminalIden)
}

func (this *myTcpListener) Send(byteArr []byte, socketIden interface{}) (err error) {
	defer func() {
		if recerr := recover(); recerr != nil {
			fmt.Println(fmt.Sprintf("myTcpListener.Send, recover, error: %v", recerr))
			debug.PrintStack()
			err = errors.New(fmt.Sprintf("%v", recerr))
		}
	}()

	conn := socketIden.(*net.Conn)
	_, err = (*conn).Write(byteArr)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error sending message, error: %v", err))
	}
	return
}

func (this *myTcpListener) Push(byteArr []byte, terminalIden string) (err error) {
	defer func() {
		if recerr := recover(); recerr != nil {
			fmt.Println(fmt.Sprintf("myTcpListener.Push, recover, error: %v", recerr))
			debug.PrintStack()
			err = errors.New(fmt.Sprintf("%v", recerr))
		}
	}()

	terminalEntity := GetTcpTerminalEntity(terminalIden)
	if len(terminalEntity.TerminalIden) == 0 {
		err = errors.New("Device offline.")
		return
	}
	if terminalEntity.LastTime.Before(time.Now().Add(-time.Second * time.Duration(globals.MyConfig.OnlineOvertimeSeconds))) {
		err = errors.New("Device offline.")
		return
	}
	return this.Send(byteArr, terminalEntity.Conn)
}

func (this *myTcpListener) receiveMessage(terminalEntity ITerminalEntity) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(fmt.Sprintf("Error processing message, err: %v", err))
		}
		conn := terminalEntity.GetSocketIden().(*net.Conn)
		if *conn != nil {
			(*conn).Close()
		}
	}()

	tcpTerminalEntity := terminalEntity.(TcpTerminalEntity)

	var readSize int
	terminalIden := tcpTerminalEntity.TerminalIden
	isNewConn := len(terminalIden) == 0
	remainLength := 0
	reader := bufio.NewReader(*tcpTerminalEntity.Conn)

	var bytesEntity globals.IEntity
	var err error
	for {
		if this.stopSign.Signed() {
			this.stopSign.Deal("")
			return
		}

		bytesEntity, err = globals.ByteArrayPool.Take()
		if err != nil {
			fmt.Println(fmt.Sprintf("Error getting resources from resource pool, err: %v", err))
			time.Sleep(time.Second * 1)
			continue
		}
		break
	}
	defer globals.ByteArrayPool.Return(bytesEntity)

	buffer := bytesEntity.(globals.IByteArrayEntity).Bytes()

	for {
		if this.stopSign.Signed() {
			this.stopSign.Deal("")
			return
		}

		err = (*tcpTerminalEntity.Conn).SetReadDeadline(time.Now().Add(time.Second * time.Duration(globals.MyConfig.ReceiveOvertimeSeconds)))
		if err != nil {
			fmt.Println(fmt.Sprintf("Error setting data receive timeout, err: %v", err))
			break
		}

		readSize, err = reader.Read(buffer[remainLength:globals.MyConfig.BufferSize])
		if err != nil {
			if err != io.EOF {
				fmt.Println(fmt.Sprintf("Error reading data from buffer, err: %v", err))
				break
			}
		}

		if readSize <= 0 {
			break
		}

		remainLength += readSize
		oldRemainLength := remainLength
		remoteAddr := (*tcpTerminalEntity.Conn).RemoteAddr()
		var replyBytes []byte
		terminalIden, remainLength, replyBytes, err = handleMessage(tcpTerminalEntity.TerminalIden, remoteAddr, buffer[:remainLength])

		if err != nil {
			fmt.Println(fmt.Sprintf("Error processing message, err: %v", err))
			break
		}

		if len(terminalIden) == 0 {
			break
		}

		if len(tcpTerminalEntity.TerminalIden) == 0 {
			tcpTerminalEntity.TerminalIden = terminalIden
		}

		if len(replyBytes) > 0 {
			err = this.Send(replyBytes, tcpTerminalEntity.Conn)
			if err != nil {
				fmt.Println(fmt.Sprintf("HandleMessage error: %v", err))
				break
			}
		}

		// println(fmt.Sprintf("remainLength: %d", remainLength))

		if remainLength >= globals.MyConfig.BufferSize {
			remainLength = 0
		} else {
			copy(buffer[:remainLength], buffer[oldRemainLength-remainLength:oldRemainLength])
		}

		if len(terminalIden) > 0 {
			RefreshTcpTerminal(tcpTerminalEntity, isNewConn)
			isNewConn = false
		}
	}

	if len(terminalIden) > 0 {
		RemoveTcpTerminalEntity(terminalIden)
	} else {
		tcpTerminalEntity.ShutdownConn()
	}
}
