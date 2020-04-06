// http_to_tcp project main.go
package main

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

	"http_to_tcp/globals"
	"http_to_tcp/messages"
	"http_to_tcp/networks"
	"http_to_tcp/rules"
)

func init() {
	_ = rules.GetRuleContext()
	_ = networks.GetTcpListenContext()
}

func main() {

	go func() {
		err := networks.GetTcpListenContext().Listener.Start(globals.MyConfig.SocketHost, globals.MyConfig.SocketPort)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error starting tcp listener, err: %v", err))
			os.Exit(1)
		}
	}()

	http.HandleFunc("/send", send)
	http.HandleFunc("/webpages/callback.html", callback)

	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", globals.MyConfig.HttpHost, globals.MyConfig.HttpPort), nil); err != nil {
		fmt.Println(fmt.Sprintf("Error starting http listener, err: %v", err))
		os.Exit(1)
	}
}

//for test
func callback(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, `E:\lpc\GolangProj\src\http_to_tcp\webpages\callback.html`)
}

func send(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	xmlStr := r.Form.Get("message")
	if len(xmlStr) == 0 {
		w.WriteHeader(400)
		w.Write([]byte("Sent content cannot be empty."))
		return
	} else {
		message, err := messages.ParseMessage(xmlStr)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Failed to parse message content."))
			return
		}

		device := message.Device
		if len(device) == 0 {
			w.WriteHeader(400)
			w.Write([]byte("Message receiver cannot be empty."))
			return
		}

		ruleContext := rules.GetRuleContext()
		var packetBytes []byte
		packetBytes, err = message.BuildPacketBytes(ruleContext.Rule, ruleContext.Header, ruleContext.Bodies)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		println(fmt.Sprintf("packetBytes: %s", hex.EncodeToString(packetBytes)))

		err = networks.GetTcpListenContext().Listener.Push(packetBytes, device)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
	}

	w.WriteHeader(200)
}
