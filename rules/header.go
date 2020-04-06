package rules

import (
	"encoding/xml"
	"errors"
	"fmt"

	"http_to_tcp/utils"
)

const (
	MessageTypeFieldName string = "messageType"
	BodySizeFieldName    string = "bodySize"
)

type Header struct {
	XMLName  xml.Name `xml:"header"`
	FieldArr []Field  `xml:"field"`
	Size     int
}

func makeHeaderRule() (header Header, err error) {
	defer func() {
		if recerr := recover(); recerr != nil {
			err = errors.New(fmt.Sprintf("%v", recerr))
		}
	}()

	var data []byte
	data, err = utils.ReadFile("/xmls/header_rule.xml")
	if err != nil {
		return
	}
	err = xml.Unmarshal(data, &header)
	if err != nil {
		return
	}

	bytesCount := 0
	foundMessageTypeField := false
	if len(header.FieldArr) > 0 {

		for i, field := range header.FieldArr {
			for j := i + 1; j < len(header.FieldArr); j++ {
				if header.FieldArr[i].Name == header.FieldArr[j].Name {
					err = errors.New(fmt.Sprintf("Field %s cannot be defined repeatedly.", header.FieldArr[i].Name))
					return
				}
			}

			// println(fmt.Sprintf("field.Name: %s, field.DataType: %s, field.BytesCount: %d", field.Name, field.DataType, field.BytesCount))

			if field.Name == MessageTypeFieldName {
				switch DataTypeEnum(field.DataType) {
				case DataType_Byte:
					field.BytesCount = 1
				case DataType_Uint16:
					field.BytesCount = 2
				default:
					err = errors.New("header_rule.xml: dataType of messageType field must be byte or uint16.")
					return
				}

				bytesCount += field.BytesCount

				if !foundMessageTypeField {
					foundMessageTypeField = true
				}
			} else {
				switch DataTypeEnum(field.DataType) {
				case DataType_Byte:
					field.BytesCount = 1
				case DataType_Int16, DataType_Uint16:
					field.BytesCount = 2
				case DataType_Float32, DataType_Int32, DataType_Uint32:
					field.BytesCount = 4
				case DataType_Float64, DataType_Int64, DataType_Uint64:
					field.BytesCount = 8
				case DataType_Bcd, DataType_Bytes, DataType_String:
					if field.BytesCount <= 0 {
						err = errors.New(fmt.Sprintf("header_rule.xml: bytesCount of %s should be greater than 0.", field.Name))
						return
					}
				default:
					err = errors.New(fmt.Sprintf("header_rule.xml: The format of dataType is incorrect, %s: ", field.DataType))
					return
				}
				bytesCount += field.BytesCount
			}
		}
	}

	header.Size = bytesCount

	if !foundMessageTypeField {
		err = errors.New("header_rule.xml: messageType field is not defined.")
		return
	}

	return
}
