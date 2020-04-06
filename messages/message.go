package messages

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"strconv"

	"http_to_tcp/rules"
	"http_to_tcp/utils"
)

type Message struct {
	XMLName xml.Name `xml:"message"`
	Device  string   `xml:"device"`
	Header  Header   `xml:"header"`
	Body    Body     `xml:"body"`
}

func (this *Message) AppendHeaderField(item Item) {
	if len(this.Header.FieldArr) == 0 {
		this.Header.FieldArr = make([]Item, 0)
	}
	this.Header.FieldArr = append(this.Header.FieldArr, item)
}

func (this *Message) AppendBodyField(item Item) {
	if len(this.Body.FieldArr) == 0 {
		this.Body.FieldArr = make([]Item, 0)
	}
	this.Body.FieldArr = append(this.Body.FieldArr, item)
}

func (this *Message) ToXmlString() (xmlStr string, err error) {
	var xmlBytes []byte
	xmlBytes, err = xml.MarshalIndent(this, "", "    ")
	if err == nil {
		xmlStr = string(xmlBytes)
	}
	return
}

func (this *Message) BuildPacketBytes(rule rules.Rule, headerRule rules.Header, bodyRule rules.Bodies) (byteArr []byte, err error) {

	if len(this.Header.FieldArr) != len(headerRule.FieldArr) {
		err = errors.New("Header field data does not match rule")
		return
	}

	for i, messageField := range this.Header.FieldArr {
		if messageField.Name != headerRule.FieldArr[i].Name {
			err = errors.New("Header field name does not match rule")
			return
		}
	}

	var messageType int

	buffer := bytes.Buffer{}

	if len(rule.HeadByteArr) > 0 {
		buffer.Write(rule.HeadByteArr)
	}

	var tmpByteArr []byte
	var fieldRule rules.Field
	for i, messageField := range this.Header.FieldArr {
		fieldRule = headerRule.FieldArr[i]
		if messageField.Name == rules.MessageTypeFieldName {
			switch rules.DataTypeEnum(headerRule.FieldArr[i].DataType) {
			case rules.DataType_Byte:
				messageType, err = processByteField(messageField, 16, &buffer)
				if err != nil {
					return
				}
			case rules.DataType_Uint16:
				messageType, err = processUint16Field(messageField, 16, rule.ByteOrder, &buffer)
				if err != nil {
					return
				}
			}
		} else {
			if messageField.Name == rules.BodySizeFieldName {
				for bodySizeIndex := 0; bodySizeIndex < rule.BodySize.BytesCount; bodySizeIndex++ {
					buffer.WriteByte(0)
				}
			} else {

				switch rules.DataTypeEnum(headerRule.FieldArr[i].DataType) {
				case rules.DataType_Bcd:
					tmpByteArr = utils.BcdStrToBytes(messageField.Value)
					if len(tmpByteArr) > fieldRule.BytesCount {
						err = errors.New(fmt.Sprintf("Value length of %s field does not match rule", messageField.Name))
						return
					}
					_, err = buffer.Write(tmpByteArr)
					if err != nil {
						return
					}
					if len(tmpByteArr) < fieldRule.BytesCount {
						for i := 0; i < fieldRule.BytesCount-len(tmpByteArr); i++ {
							buffer.WriteByte(0)
						}
					}
				case rules.DataType_Byte:
					_, err = processByteField(messageField, 10, &buffer)
					if err != nil {
						return
					}
				case rules.DataType_Bytes:
					tmpByteArr, err = base64.StdEncoding.DecodeString(messageField.Value)
					if err != nil {
						return
					}
					if len(tmpByteArr) > fieldRule.BytesCount {
						err = errors.New(fmt.Sprintf("Value length of %s field does not match rule", messageField.Name))
						return
					}
					_, err = buffer.Write(tmpByteArr)
					if err != nil {
						return
					}
					if len(tmpByteArr) < fieldRule.BytesCount {
						for i := 0; i < fieldRule.BytesCount-len(tmpByteArr); i++ {
							buffer.WriteByte(0)
						}
					}
				case rules.DataType_Float32:
					err = processFloat32Field(messageField, rule.ByteOrder, &buffer)
					if err != nil {
						return
					}
				case rules.DataType_Float64:
					err = processFloat64Field(messageField, rule.ByteOrder, &buffer)
					if err != nil {
						return
					}
				case rules.DataType_Int16:
					err = processInt16Field(messageField, rule.ByteOrder, &buffer)
					if err != nil {
						return
					}
				case rules.DataType_Int32:
					err = processInt32Field(messageField, rule.ByteOrder, &buffer)
					if err != nil {
						return
					}
				case rules.DataType_Int64:
					err = processInt64Field(messageField, rule.ByteOrder, &buffer)
					if err != nil {
						return
					}
				case rules.DataType_String:
					tmpByteArr = []byte(messageField.Value)
					if len(tmpByteArr) > fieldRule.BytesCount {
						err = errors.New(fmt.Sprintf("Value length of %s field does not match rule", messageField.Name))
						return
					}
					_, err = buffer.Write(tmpByteArr)
					if err != nil {
						return
					}
					if len(tmpByteArr) < fieldRule.BytesCount {
						for i := 0; i < fieldRule.BytesCount-len(tmpByteArr); i++ {
							buffer.WriteByte(0)
						}
					}
				case rules.DataType_Uint16:
					_, err = processUint16Field(messageField, 16, rule.ByteOrder, &buffer)
					if err != nil {
						return
					}
				case rules.DataType_Uint32:
					err = processUint32Field(messageField, rule.ByteOrder, &buffer)
					if err != nil {
						return
					}
				case rules.DataType_Uint64:
					err = processUint64Field(messageField, rule.ByteOrder, &buffer)
					if err != nil {
						return
					}
				}
			}
		}
	}

	if messageType < 0 {
		err = errors.New("MessageType field not included in message header.")
		return
	}

	bodySizeValue := 0
	if len(this.Body.FieldArr) > 0 {
		if bodyRule, ok := bodyRule.BodyMap[messageType]; ok {
			if len(this.Body.FieldArr) != len(bodyRule.FieldArr) {
				err = errors.New("Body field data does not match rule")
				return
			}

			for i, messageField := range this.Body.FieldArr {
				if messageField.Name != bodyRule.FieldArr[i].Name {
					err = errors.New("Body field name does not match rule")
					return
				}
			}

			var lastIntValue int
			for i, messageField := range this.Body.FieldArr {
				fieldRule = bodyRule.FieldArr[i]
				switch rules.DataTypeEnum(fieldRule.DataType) {
				case rules.DataType_Bcd:
					if fieldRule.BytesCount < 0 {
						if lastIntValue > 0 {
							fieldRule.BytesCount = lastIntValue
						} else {
							err = errors.New(fmt.Sprintf("Length value for field %s not found.", messageField.Name))
						}
					}
					tmpByteArr = utils.BcdStrToBytes(messageField.Value)
					if fieldRule.BytesCount > 0 && len(tmpByteArr) > fieldRule.BytesCount {
						err = errors.New(fmt.Sprintf("Value length of %s field does not match rule", messageField.Name))
						return
					}
					_, err = buffer.Write(tmpByteArr)
					if err != nil {
						return
					}
					if len(tmpByteArr) < fieldRule.BytesCount {
						for i := 0; i < fieldRule.BytesCount-len(tmpByteArr); i++ {
							buffer.WriteByte(0)
						}
					}
					if fieldRule.BytesCount == 0 {
						break
					}
				case rules.DataType_Byte:
					_, err = processByteField(messageField, 10, &buffer)
					if err != nil {
						return
					}
				case rules.DataType_Bytes:
					if fieldRule.BytesCount < 0 {
						if lastIntValue > 0 {
							fieldRule.BytesCount = lastIntValue
						} else {
							err = errors.New(fmt.Sprintf("Length value for field %s not found.", messageField.Name))
						}
					}
					tmpByteArr, err = base64.StdEncoding.DecodeString(messageField.Value)
					if err != nil {
						return
					}
					if fieldRule.BytesCount > 0 && len(tmpByteArr) > fieldRule.BytesCount {
						err = errors.New(fmt.Sprintf("Value length of %s field does not match rule", messageField.Name))
						return
					}
					_, err = buffer.Write(tmpByteArr)
					if err != nil {
						return
					}
					if fieldRule.BytesCount > 0 && len(tmpByteArr) < fieldRule.BytesCount {
						for i := 0; i < fieldRule.BytesCount-len(tmpByteArr); i++ {
							buffer.WriteByte(0)
						}
					}
					if fieldRule.BytesCount == 0 {
						break
					}
				case rules.DataType_Float32:
					err = processFloat32Field(messageField, rule.ByteOrder, &buffer)
					if err != nil {
						return
					}
				case rules.DataType_Float64:
					err = processFloat64Field(messageField, rule.ByteOrder, &buffer)
					if err != nil {
						return
					}
				case rules.DataType_Int16:
					err = processInt16Field(messageField, rule.ByteOrder, &buffer)
					if err != nil {
						return
					}
				case rules.DataType_Int32:
					err = processInt32Field(messageField, rule.ByteOrder, &buffer)
					if err != nil {
						return
					}
				case rules.DataType_Int64:
					err = processInt64Field(messageField, rule.ByteOrder, &buffer)
					if err != nil {
						return
					}
				case rules.DataType_String:
					if fieldRule.BytesCount < 0 {
						if lastIntValue > 0 {
							fieldRule.BytesCount = lastIntValue
						} else {
							err = errors.New(fmt.Sprintf("Length value for field %s not found.", messageField.Name))
						}
					}
					tmpByteArr = []byte(messageField.Value)
					if fieldRule.BytesCount > 0 && len(tmpByteArr) > fieldRule.BytesCount {
						err = errors.New(fmt.Sprintf("Value length of %s field does not match rule", messageField.Name))
						return
					}
					_, err = buffer.Write(tmpByteArr)
					if err != nil {
						return
					}
					if fieldRule.BytesCount > 0 && len(tmpByteArr) < fieldRule.BytesCount {
						for i := 0; i < fieldRule.BytesCount-len(tmpByteArr); i++ {
							buffer.WriteByte(0)
						}
					}
					if fieldRule.BytesCount == 0 {
						break
					}
				case rules.DataType_Uint16:
					_, err = processUint16Field(messageField, 10, rule.ByteOrder, &buffer)
					if err != nil {
						return
					}
				case rules.DataType_Uint32:
					err = processUint32Field(messageField, rule.ByteOrder, &buffer)
					if err != nil {
						return
					}
				case rules.DataType_Uint64:
					err = processUint64Field(messageField, rule.ByteOrder, &buffer)
					if err != nil {
						return
					}
				}
			}
		} else {
			err = errors.New(fmt.Sprintf("Message type %x not defined", messageType))
			return
		}

		if rule.BodySize.BytesCount > 0 {
			bodySizeValue = buffer.Len() - len(rule.HeadByteArr) - rule.HeaderSize
		}
	}

	for i := 0; i < rule.CheckCode.BytesCount+len(rule.TailByteArr); i++ {
		buffer.WriteByte(0)
	}

	byteArr = buffer.Bytes()

	if rule.BodySize.BytesCount == 2 {
		rule.ByteOrder.PutUint16(byteArr[rule.BodySize.Start:], uint16(bodySizeValue))
	} else {
		byteArr[rule.BodySize.Start] = byte(bodySizeValue)
	}

	err = rule.SetCheckcode(byteArr)
	if err != nil {
		return
	}

	for i := 0; i < len(rule.TailByteArr); i++ {
		byteArr[len(byteArr)-len(rule.TailByteArr)+i] = rule.TailByteArr[i]
	}

	//byteArr = byteArr[:len(byteArr)-(2-checkcodeLen)]

	byteArr, err = rule.ProcessEscapeCode(byteArr)

	return
}

func ParseMessage(xmlStr string) (message Message, err error) {
	err = xml.Unmarshal([]byte(xmlStr), &message)
	return
}

func processByteField(messageField Item, intBase int, buffer *bytes.Buffer) (value int, err error) {
	var tmpIntValue uint64
	tmpIntValue, err = strconv.ParseUint(messageField.Value, intBase, 8)
	if err != nil {
		return
	}
	if tmpIntValue < 0 || tmpIntValue > 255 {
		err = errors.New(fmt.Sprintf("Data type of %s field does not match rule", messageField.Name))
		return
	}
	err = buffer.WriteByte(byte(tmpIntValue))
	if err != nil {
		return
	}
	value = int(tmpIntValue)
	return
}

func processFloat32Field(messageField Item, byteOrder binary.ByteOrder, buffer *bytes.Buffer) (err error) {
	var tmpFloatValue float64
	tmpFloatValue, err = strconv.ParseFloat(messageField.Value, 32)
	if err != nil {
		return
	}
	bits := math.Float32bits(float32(tmpFloatValue))
	tmpByteArr := make([]byte, 4)
	byteOrder.PutUint32(tmpByteArr, bits)
	_, err = buffer.Write(tmpByteArr)
	return
}

func processFloat64Field(messageField Item, byteOrder binary.ByteOrder, buffer *bytes.Buffer) (err error) {
	var tmpFloatValue float64
	tmpFloatValue, err = strconv.ParseFloat(messageField.Value, 64)
	if err != nil {
		return
	}
	bits := math.Float64bits(tmpFloatValue)
	tmpByteArr := make([]byte, 8)
	byteOrder.PutUint64(tmpByteArr, bits)
	_, err = buffer.Write(tmpByteArr)
	return
}

func processInt16Field(messageField Item, byteOrder binary.ByteOrder, buffer *bytes.Buffer) (err error) {
	var tmpIntValue uint64
	tmpIntValue, err = strconv.ParseUint(messageField.Value, 10, 16)
	if err != nil {
		return
	}
	tmpByteArr := make([]byte, 2)
	byteOrder.PutUint16(tmpByteArr, uint16(tmpIntValue))
	_, err = buffer.Write(tmpByteArr)
	return
}

func processInt32Field(messageField Item, byteOrder binary.ByteOrder, buffer *bytes.Buffer) (err error) {
	var tmpIntValue uint64
	tmpIntValue, err = strconv.ParseUint(messageField.Value, 10, 32)
	if err != nil {
		return
	}
	tmpByteArr := make([]byte, 4)
	byteOrder.PutUint32(tmpByteArr, uint32(tmpIntValue))
	_, err = buffer.Write(tmpByteArr)
	return
}

func processInt64Field(messageField Item, byteOrder binary.ByteOrder, buffer *bytes.Buffer) (err error) {
	var tmpIntValue uint64
	tmpIntValue, err = strconv.ParseUint(messageField.Value, 10, 64)
	if err != nil {
		return
	}
	tmpByteArr := make([]byte, 8)
	byteOrder.PutUint64(tmpByteArr, uint64(tmpIntValue))
	_, err = buffer.Write(tmpByteArr)
	return
}

func processUint16Field(messageField Item, intBase int, byteOrder binary.ByteOrder, buffer *bytes.Buffer) (value int, err error) {
	var tmpIntValue uint64
	tmpIntValue, err = strconv.ParseUint(messageField.Value, intBase, 16)
	if err != nil {
		return
	}
	tmpByteArr := make([]byte, 2)
	byteOrder.PutUint16(tmpByteArr, uint16(tmpIntValue))
	_, err = buffer.Write(tmpByteArr)
	if err != nil {
		return
	}
	value = int(tmpIntValue)
	return
}

func processUint32Field(messageField Item, byteOrder binary.ByteOrder, buffer *bytes.Buffer) (err error) {
	var tmpIntValue uint64
	tmpIntValue, err = strconv.ParseUint(messageField.Value, 10, 32)
	if err != nil {
		return
	}
	tmpByteArr := make([]byte, 4)
	byteOrder.PutUint32(tmpByteArr, uint32(tmpIntValue))
	_, err = buffer.Write(tmpByteArr)
	return
}

func processUint64Field(messageField Item, byteOrder binary.ByteOrder, buffer *bytes.Buffer) (err error) {
	var tmpIntValue uint64
	tmpIntValue, err = strconv.ParseUint(messageField.Value, 10, 64)
	if err != nil {
		return
	}
	tmpByteArr := make([]byte, 8)
	byteOrder.PutUint64(tmpByteArr, uint64(tmpIntValue))
	_, err = buffer.Write(tmpByteArr)
	return
}
