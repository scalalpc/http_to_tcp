package rules

import (
	"strconv"
	"strings"
)

type EscapeChar struct {
	ItemArr []EscapeCharItem
}

type EscapeCharItem struct {
	SrcByteArr  []byte
	DestByteArr []byte
}

func (this *EscapeChar) verifyFields() {
	for _, item := range this.ItemArr {
		if len(item.SrcByteArr) == 0 {
			panic("src of escape char cannot be empty.")
		}
		if len(item.DestByteArr) == 0 {
			panic("dest of escape char cannot be empty.")
		}
		if len(item.DestByteArr) < len(item.SrcByteArr) {
			panic("src length of escape char cannot be less than dest length.")
		}
	}

}

type RawEscapeChar struct {
	ItemArr []RawEscapeCharItem `xml:"item"`
}

type RawEscapeCharItem struct {
	Src  string `xml:"src,attr"`
	Dest string `xml:"dest,attr"`
}

func (this *RawEscapeCharItem) parseSrc() (byteArr []byte) {
	return this.parseSrcOrDest(strings.TrimSpace(this.Src))
}

func (this *RawEscapeCharItem) parseDest() (byteArr []byte) {
	return this.parseSrcOrDest(strings.TrimSpace(this.Dest))
}

func (this *RawEscapeCharItem) parseSrcOrDest(str string) (byteArr []byte) {
	byteArr = make([]byte, 0)
	if len(str) > 0 {
		itemArr := strings.Split(str, ",")
		var value int64
		var err error
		for _, itemStr := range itemArr {
			if len(strings.TrimSpace(itemStr)) > 0 {
				value, err = strconv.ParseInt(strings.Replace(strings.TrimSpace(itemStr), "0x", "", -1), 16, 0)
				if err != nil {
					panic(err)
				}
				byteArr = append(byteArr, byte(value))
			}
		}
	}
	return
}
