package networks

type INetListener interface {
	Start(host string, port int) error
	Stop() error
	RefreshTerminal(terminalEntity ITerminalEntity, isNewConn bool)
	GetTerminalEntity(terminalIden string) ITerminalEntity
	Send(byteArr []byte, socketIden interface{}) error
	Push(byteArr []byte, terminalIden string) error
}
