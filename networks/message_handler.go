package networks

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"io"

	//	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"http_to_tcp/globals"
	"http_to_tcp/messages"
	"http_to_tcp/rules"
	"http_to_tcp/utils"
)

func handleMessage(srcTerminalIden string, remoteAddr net.Addr, buffer []byte) (terminalIden string, remainLength int, replyBytes []byte, err error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(fmt.Sprintf("Error processing message, err: %v", err))
		}
	}()

	terminalIden = srcTerminalIden

	rule := rules.GetRuleContext().Rule
	header := rules.GetRuleContext().Header
	bodies := rules.GetRuleContext().Bodies

	if len(buffer) < rule.MinPacketSize {
		remainLength = len(buffer)
		return
	}

	var packetBytes []byte

	beginIndex := 0
	endIndex := 0
	offset := 0

	if len(rule.HeadByteArr) > 0 {
		foundHead := false
		for i := 0; i < len(buffer)-rule.MinPacketSize+1; i++ {
			var j int
			for j = 0; j < len(rule.HeadByteArr); j++ {
				if buffer[i+j] != rule.HeadByteArr[j] {
					break
				}
			}
			if j == len(rule.HeadByteArr) {
				beginIndex = i
				foundHead = true
				break
			} else {
				continue
			}
		}
		if !foundHead {
			remainLength = len(buffer)
			return
		}
		offset += len(rule.HeadByteArr)
	}

	if (len(rule.EscapeChar.ItemArr) > 0 || !rule.BodySize.Exists) && len(rule.TailByteArr) > 0 {
		foundTail := false
		for i := beginIndex + rule.MinPacketSize - len(rule.TailByteArr); i < len(buffer)-len(rule.TailByteArr)+1; i++ {
			var j int
			for j = 0; j < len(rule.TailByteArr); j++ {
				if buffer[i+j] != rule.TailByteArr[j] {
					break
				}
			}
			if j == len(rule.TailByteArr) {
				endIndex = i + len(rule.TailByteArr)
				foundTail = true
				break
			} else {
				continue
			}
		}
		if !foundTail {
			remainLength = len(buffer)
			return
		}
		packetBytes = buffer[beginIndex:endIndex]

		if len(rule.EscapeChar.ItemArr) > 0 {
			packetBytes, err = rule.ProcessUnescapeCode(packetBytes)
			if err != nil {
				fmt.Println(fmt.Sprintf("Unescape failed, %v", err))
				remainLength = len(buffer)
				return
			}
		}
	} else if len(rule.EscapeChar.ItemArr) == 0 && rule.BodySize.Exists {
		bodySizeValue := 0
		if rule.BodySize.Start+rule.BodySize.BytesCount > len(buffer) {
			fmt.Println("return 1, ")
			remainLength = len(buffer)
			return
		}
		if rule.BodySize.BytesCount == 2 {
			bodySizeValue = int(rule.ByteOrder.Uint16(buffer[rule.BodySize.Start:]))
		} else {
			bodySizeValue = int(buffer[rule.BodySize.Start])
		}
		endIndex = beginIndex + len(rule.HeadByteArr) + rule.HeaderSize + bodySizeValue + rule.CheckCode.BytesCount + len(rule.TailByteArr)
		if endIndex > len(buffer) {
			remainLength = len(buffer)
			fmt.Println(fmt.Sprintf("return 2, endIndex: %d, len(buffer): %d", endIndex, len(buffer)))
			return
		}
		packetBytes = buffer[beginIndex:endIndex]

		fmt.Println(fmt.Sprintf("beginIndex: %d, endIndex: %d, len(packetBytes): %d", beginIndex, endIndex, len(packetBytes)))
	} else {
		fmt.Println("What should not happen while processing messages.")
		return
	}

	var verified bool
	verified, err = rule.VerifyCheckcode(packetBytes[:])
	if err != nil {
		fmt.Println(fmt.Sprintf("Error verifying verification code, err: %v", err))
		remainLength = len(buffer)
		return
	}
	if !verified {
		fmt.Println("Incorrect check code.")
	} else {
		messageBytes := packetBytes[0 : len(packetBytes)-rule.CheckCode.BytesCount]
		//println(fmt.Sprintf("message bytes: %v", hex.EncodeToString(messageBytes)))

		messageObj := messages.Message{
			Device: srcTerminalIden,
		}

		offset = len(rule.HeadByteArr)
		messageTypeValue := 0
		var endianBuffer *bytes.Buffer
		for _, field := range header.FieldArr {
			messageItem := messages.Item{}
			messageItem.Name = field.Name
			switch rules.DataTypeEnum(field.DataType) {
			case rules.DataType_Bcd:
				messageItem.Value = strings.Trim(utils.BcdToString(messageBytes[offset:offset+field.BytesCount]), "\x00")
			case rules.DataType_Byte:
				messageItem.Value = strconv.Itoa(int(messageBytes[offset]))
			case rules.DataType_Bytes:
				messageItem.Value = base64.StdEncoding.EncodeToString(messageBytes[offset : offset+field.BytesCount])
			case rules.DataType_Float32:
				var floatValue float32
				endianBuffer = bytes.NewBuffer(messageBytes[offset : offset+field.BytesCount])
				binary.Read(endianBuffer, rule.ByteOrder, &floatValue)
				messageItem.Value = fmt.Sprintf("%d", floatValue)
			case rules.DataType_Float64:
				var floatValue float64
				endianBuffer = bytes.NewBuffer(messageBytes[offset : offset+field.BytesCount])
				binary.Read(endianBuffer, rule.ByteOrder, &floatValue)
				messageItem.Value = fmt.Sprintf("%d", floatValue)
			case rules.DataType_Int16:
				messageItem.Value = fmt.Sprintf("%d", rule.ByteOrder.Uint16(messageBytes[offset:offset+field.BytesCount]))
			case rules.DataType_Int32:
				messageItem.Value = fmt.Sprintf("%d", rule.ByteOrder.Uint32(messageBytes[offset:offset+field.BytesCount]))
			case rules.DataType_Int64:
				messageItem.Value = fmt.Sprintf("%d", rule.ByteOrder.Uint64(messageBytes[offset:offset+field.BytesCount]))
			case rules.DataType_String:
				messageItem.Value = strings.Trim(string(messageBytes[offset:offset+field.BytesCount]), "\x00")
			case rules.DataType_Uint16:
				messageItem.Value = fmt.Sprintf("%d", rule.ByteOrder.Uint16(messageBytes[offset:offset+field.BytesCount]))
			case rules.DataType_Uint32:
				messageItem.Value = fmt.Sprintf("%d", rule.ByteOrder.Uint32(messageBytes[offset:offset+field.BytesCount]))
			case rules.DataType_Uint64:
				messageItem.Value = fmt.Sprintf("%d", rule.ByteOrder.Uint64(messageBytes[offset:offset+field.BytesCount]))
			}
			offset += field.BytesCount
			if field.Name == string(rules.MessageTypeFieldName) {
				messageTypeValue, err = strconv.Atoi(messageItem.Value)
				if err != nil {
					fmt.Println(fmt.Sprintf("%v", err))
					remainLength = endIndex
					return
				}
			}

			messageObj.AppendHeaderField(messageItem)
		}

		if bodyRule, ok := bodies.BodyMap[messageTypeValue]; ok {
			if len(bodyRule.FieldArr) > 0 {
				var curIntStrValue string
				var curIntValue int
				var lastIntValue int
				for _, field := range bodyRule.FieldArr {
					messageItem := messages.Item{}
					messageItem.Name = field.Name
					curIntStrValue = ""
					curIntValue = 0
					lastIntValue = 0
					switch rules.DataTypeEnum(field.DataType) {
					case rules.DataType_Bcd:
						if field.BytesCount < 0 && lastIntValue > 0 {
							field.BytesCount = lastIntValue
						}
						if field.BytesCount > 0 {
							messageItem.Value = strings.Trim(utils.BcdToString(messageBytes[offset:offset+field.BytesCount]), "\x00")
						}
					case rules.DataType_Byte:
						curIntValue = int(messageBytes[offset])
						messageItem.Value = strconv.Itoa(curIntValue)
						lastIntValue = curIntValue
					case rules.DataType_Bytes:
						if field.BytesCount < 0 && lastIntValue > 0 {
							field.BytesCount = lastIntValue
						}
						if field.BytesCount > 0 {
							messageItem.Value = base64.StdEncoding.EncodeToString(messageBytes[offset : offset+field.BytesCount])
						}
					case rules.DataType_Float32:
						var floatValue float32
						endianBuffer = bytes.NewBuffer(messageBytes[offset : offset+field.BytesCount])
						binary.Read(endianBuffer, rule.ByteOrder, &floatValue)
						messageItem.Value = fmt.Sprintf("%d", floatValue)
					case rules.DataType_Float64:
						var floatValue float64
						endianBuffer = bytes.NewBuffer(messageBytes[offset : offset+field.BytesCount])
						binary.Read(endianBuffer, rule.ByteOrder, &floatValue)
						messageItem.Value = fmt.Sprintf("%d", floatValue)
					case rules.DataType_Int16:
						curIntStrValue = fmt.Sprintf("%d", rule.ByteOrder.Uint16(messageBytes[offset:offset+field.BytesCount]))
						messageItem.Value = curIntStrValue
						lastIntValue, _ = strconv.Atoi(curIntStrValue)
					case rules.DataType_Int32:
						curIntStrValue = fmt.Sprintf("%d", rule.ByteOrder.Uint32(messageBytes[offset:offset+field.BytesCount]))
						messageItem.Value = curIntStrValue
						lastIntValue, _ = strconv.Atoi(curIntStrValue)
					case rules.DataType_Int64:
						curIntStrValue = fmt.Sprintf("%d", rule.ByteOrder.Uint64(messageBytes[offset:offset+field.BytesCount]))
						messageItem.Value = curIntStrValue
						lastIntValue, _ = strconv.Atoi(curIntStrValue)
					case rules.DataType_String:
						if field.BytesCount < 0 && lastIntValue > 0 {
							field.BytesCount = lastIntValue
						}
						if field.BytesCount > 0 {
							messageItem.Value = strings.Trim(string(messageBytes[offset:offset+field.BytesCount]), "\x00")
						}
						lastIntValue = 0
					case rules.DataType_Uint16:
						curIntStrValue = fmt.Sprintf("%d", rule.ByteOrder.Uint16(messageBytes[offset:offset+field.BytesCount]))
						messageItem.Value = curIntStrValue
						lastIntValue, _ = strconv.Atoi(curIntStrValue)
					case rules.DataType_Uint32:
						curIntStrValue = fmt.Sprintf("%d", rule.ByteOrder.Uint32(messageBytes[offset:offset+field.BytesCount]))
						messageItem.Value = curIntStrValue
						lastIntValue, _ = strconv.Atoi(curIntStrValue)
					case rules.DataType_Uint64:
						curIntStrValue = fmt.Sprintf("%d", rule.ByteOrder.Uint64(messageBytes[offset:offset+field.BytesCount]))
						messageItem.Value = curIntStrValue
						lastIntValue, _ = strconv.Atoi(curIntStrValue)
					}

					if field.BytesCount == 0 {
						break
					} else {
						offset += field.BytesCount
					}

					messageObj.AppendBodyField(messageItem)
				}
			}
		}

		var xmlContent string
		xmlContent, err = messageObj.ToXmlString()
		if err != nil {
			fmt.Println(fmt.Sprintf("Error converting object instance to XML, err: %v", err))
			remainLength = endIndex
			return
		}
		println(fmt.Sprintf("message xml: %s", xmlContent))

		if len(globals.MyConfig.CallbackUrl) > 0 {
			replyBytes, err = postMessageToCallbackUrl(globals.MyConfig.CallbackUrl, xmlContent)

			println(fmt.Sprintf("replyBytes: %s", string(replyBytes)))

			if err != nil {
				fmt.Println(fmt.Sprintf("Error sending message to callback URL, err: %s", err))
				return
			}

			if len(srcTerminalIden) == 0 {
				if len(replyBytes) > 0 {
					var replyMessage messages.Message
					replyMessage, err = messages.ParseMessage(string(replyBytes))
					if err != nil {
						fmt.Println(fmt.Sprintf("Error parsing message XML, err: %s", err))
						return
					}
					terminalIden = replyMessage.Device
				}
			}
		} else {
			err = errors.New("Callback URL does not exist.")
			return
		}

		remainLength = endIndex
		return
	}
	return
}

func postMessageToCallbackUrl(callbackUrl, xmlContent string) (replyBytes []byte, err error) {

	rsp, err := http.Post(callbackUrl, "application/xml", strings.NewReader(xmlContent))
	if rsp != nil && rsp.Body != nil {
		defer rsp.Body.Close()
	}
	if err != nil {
		return
	}
	if rsp.StatusCode == 200 {
		if rsp.ContentLength > 0 {
			replyBytes = make([]byte, rsp.ContentLength)
			_, err = rsp.Body.Read(replyBytes)
			if err != nil {
				if err == io.EOF {
					err = nil
				} else {
					return
				}
			}
		}
	} else {
		err = errors.New(fmt.Sprintf("Error posting callback url, status code: %d", rsp.StatusCode))
		return
	}
	return
}
