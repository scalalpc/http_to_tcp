package rules

import (
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"http_to_tcp/utils"
)

type Rule struct {
	HeadByteArr []byte
	TailByteArr []byte
	EscapeChar  EscapeChar
	BodySize    BodySize
	CheckCode   CheckCode
	ByteOrder   binary.ByteOrder

	HeaderSize    int
	MinPacketSize int
}

func (this *Rule) validFields() {
	if len(this.EscapeChar.ItemArr) > 0 && len(this.TailByteArr) == 0 {
		panic(errors.New("If escapeChar is set, tail must be set."))
	}
	if len(this.TailByteArr) == 0 && !this.BodySize.Exists {
		panic(errors.New("tail and bodySize must be set to one."))
	}

	this.MinPacketSize = len(this.HeadByteArr) + len(this.TailByteArr) + this.HeaderSize + this.CheckCode.BytesCount
}

func (this *Rule) ProcessEscapeCode(packetBytes []byte) (escapeBytes []byte, err error) {
	defer func() {
		if recerr := recover(); recerr != nil {
			err = errors.New(fmt.Sprintf("%v", recerr))
		}
	}()

	var offset int
	if len(this.EscapeChar.ItemArr) > 0 && len(packetBytes)-len(this.HeadByteArr)-len(this.TailByteArr) > 0 {
		for _, escapeCharItem := range this.EscapeChar.ItemArr {
			if len(escapeBytes) > 0 {
				escapeBytes = nil
			}
			offset = 0
			for i := len(this.HeadByteArr); i < len(packetBytes)-len(this.TailByteArr)-len(escapeCharItem.SrcByteArr)+1; i++ {
				var j int
				for j = 0; j < len(escapeCharItem.SrcByteArr); j++ {
					if packetBytes[i+j] != escapeCharItem.SrcByteArr[j] {
						break
					}
				}
				if j == len(escapeCharItem.SrcByteArr) {
					if len(escapeBytes) == 0 {
						escapeBytes = make([]byte, 0)
					}
					escapeBytes = append(escapeBytes, packetBytes[offset:i]...)
					escapeBytes = append(escapeBytes, escapeCharItem.DestByteArr[:]...)
					offset = i + len(escapeCharItem.SrcByteArr)
				}
			}
			if len(escapeBytes) > 0 {
				if offset < len(packetBytes) {
					escapeBytes = append(escapeBytes, packetBytes[offset:]...)
				}
				packetBytes = make([]byte, len(escapeBytes))
				copy(packetBytes, escapeBytes)
			}
		}
	}

	if len(escapeBytes) == 0 {
		escapeBytes = packetBytes
	}

	return
}

func (this *Rule) ProcessUnescapeCode(packetBytes []byte) (escapeBytes []byte, err error) {
	defer func() {
		if recerr := recover(); recerr != nil {
			err = errors.New(fmt.Sprintf("%v", recerr))
		}
	}()

	var offset int
	if len(this.EscapeChar.ItemArr) > 0 && len(packetBytes)-len(this.HeadByteArr)-len(this.TailByteArr) > 0 {
		reversedItemArr := make([]EscapeCharItem, len(this.EscapeChar.ItemArr))
		copy(reversedItemArr, this.EscapeChar.ItemArr)
		for _, escapeCharItem := range reversedItemArr {
			if len(escapeBytes) > 0 {
				escapeBytes = nil
			}
			offset = 0
			for i := len(this.HeadByteArr); i < len(packetBytes)-len(this.TailByteArr)-len(escapeCharItem.DestByteArr)+1; i++ {
				var j int
				for j = 0; j < len(escapeCharItem.DestByteArr); j++ {
					if packetBytes[i+j] != escapeCharItem.DestByteArr[j] {
						break
					}
				}
				if j == len(escapeCharItem.DestByteArr) {
					if len(escapeBytes) == 0 {
						escapeBytes = make([]byte, 0)
					}
					escapeBytes = append(escapeBytes, packetBytes[offset:i]...)
					escapeBytes = append(escapeBytes, escapeCharItem.SrcByteArr[:]...)
					offset = i + len(escapeCharItem.DestByteArr)
				}
			}
			if len(escapeBytes) > 0 {
				if offset < len(packetBytes) {
					escapeBytes = append(escapeBytes, packetBytes[offset:]...)
				}
				packetBytes = make([]byte, len(escapeBytes))
				copy(packetBytes, escapeBytes)
			}
		}
	}
	if len(escapeBytes) == 0 {
		escapeBytes = packetBytes
	}
	return
}

func (this *Rule) VerifyCheckcode(packetBytes []byte) (valid bool, err error) {
	defer func() {
		if recerr := recover(); recerr != nil {
			err = errors.New(fmt.Sprintf("%v", recerr))
		}
	}()

	if this.CheckCode.Exists {
		start := this.CheckCode.Start
		if start < 0 {
			start = start + len(packetBytes)
		}
		if this.CheckCode.Algorithm == string(AlgorithmType_Crc16) {
			calcValue := utils.UsMBCRC16(packetBytes, this.CheckCode.Begin, start)
			srcValue := this.ByteOrder.Uint16(packetBytes[start : start+2])
			valid = calcValue == srcValue
			return
		} else {
			calcValue := packetBytes[this.CheckCode.Begin]
			for i := this.CheckCode.Begin + 1; i < start; i++ {
				calcValue ^= packetBytes[i]
			}
			valid = calcValue == packetBytes[start]
			return
		}
	} else {
		valid = true
		return
	}
}

func (this *Rule) SetCheckcode(packetBytes []byte) (err error) {
	defer func() {
		if recerr := recover(); recerr != nil {
			err = errors.New(fmt.Sprintf("%v", recerr))
		}
	}()

	if this.CheckCode.Exists {
		start := this.CheckCode.Start
		if start < 0 {
			start = start + len(packetBytes)
		}
		if this.CheckCode.Algorithm == string(AlgorithmType_Crc16) {
			calcValue := utils.UsMBCRC16(packetBytes, this.CheckCode.Begin, start)
			this.ByteOrder.PutUint16(packetBytes[start:], calcValue)
		} else {
			calcValue := packetBytes[this.CheckCode.Begin]
			for i := this.CheckCode.Begin + 1; i < start; i++ {
				calcValue ^= packetBytes[i]
			}
			packetBytes[start] = calcValue
		}
	}
	return
}

type RawRule struct {
	XMLName       xml.Name      `xml:"rule"`
	Head          string        `xml:"head"`
	Tail          string        `xml:"tail"`
	RawEscapeChar RawEscapeChar `xml:"escapeChar"`
	BodySize      BodySize      `xml:"bodySize"`
	CheckCode     CheckCode     `xml:"checkCode"`
	ByteOrder     string        `xml:"byteOrder"`
}

func (this *RawRule) parseHead() (byteArr []byte) {
	return this.parseHeadOrTail(strings.TrimSpace(this.Head))
}

func (this *RawRule) parseTail() (byteArr []byte) {
	return this.parseHeadOrTail(strings.TrimSpace(this.Tail))
}

func (this *RawRule) parseHeadOrTail(str string) (byteArr []byte) {
	byteArr = make([]byte, 0)
	if len(str) > 0 {
		itemArr := strings.Split(str, ",")
		var value int64
		var err error
		for _, itemStr := range itemArr {
			if len(strings.TrimSpace(itemStr)) > 0 {
				value, err = strconv.ParseInt(strings.Replace(strings.TrimSpace(itemStr), "0x", "", -1), 16, 8)
				if err != nil {
					panic(err)
				}
				byteArr = append(byteArr, byte(value))
			}
		}
	}
	return
}

func makeRule() (rule Rule, err error) {
	defer func() {
		if recerr := recover(); recerr != nil {
			err = errors.New(fmt.Sprintf("%v", recerr))
		}
	}()

	var data []byte
	data, err = utils.ReadFile("/xmls/packet_rule.xml")
	if err != nil {
		return
	}
	rawRule := RawRule{}
	err = xml.Unmarshal(data, &rawRule)
	if err != nil {
		return
	}

	rule.BodySize = rawRule.BodySize
	rule.BodySize.verifyFields()

	if rawRule.ByteOrder == string(ByteOrder_Little) {
		rule.ByteOrder = binary.LittleEndian
	} else {
		rule.ByteOrder = binary.BigEndian
	}

	rule.CheckCode = rawRule.CheckCode
	rule.CheckCode.verifyFields()

	rule.HeadByteArr = rawRule.parseHead()

	rule.TailByteArr = rawRule.parseTail()

	rule.EscapeChar.ItemArr = make([]EscapeCharItem, len(rawRule.RawEscapeChar.ItemArr))
	for i, escapeCharItem := range rawRule.RawEscapeChar.ItemArr {
		rule.EscapeChar.ItemArr[i].SrcByteArr = escapeCharItem.parseSrc()
		rule.EscapeChar.ItemArr[i].DestByteArr = escapeCharItem.parseDest()
	}
	rule.EscapeChar.verifyFields()

	rule.validFields()

	return
}
