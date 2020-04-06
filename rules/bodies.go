package rules

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"

	"http_to_tcp/utils"
)

type Bodies struct {
	XMLName xml.Name `xml:"bodies"`
	BodyArr []Body   `xml:"body"`
	BodyMap map[int]Body
}

func makeBodiesRule() (bodies Bodies, err error) {
	defer func() {
		if recerr := recover(); recerr != nil {
			err = errors.New(fmt.Sprintf("%v", recerr))
		}
	}()

	var data []byte
	data, err = utils.ReadFile("/xmls/body_rule.xml")
	if err != nil {
		return
	}

	err = xml.Unmarshal(data, &bodies)
	if err != nil {
		return
	}

	bodies.BodyMap = make(map[int]Body)

	if len(bodies.BodyArr) > 0 {
		var messageType int64
		for _, body := range bodies.BodyArr {
			for i := 0; i < len(body.FieldArr); i++ {
				for j := i + 1; j < len(body.FieldArr); j++ {
					if body.FieldArr[i].Name == body.FieldArr[j].Name {
						err = errors.New(fmt.Sprintf("body_rule.xml: Field %s of message %d cannot be defined repeatedly.", body.FieldArr[i].Name, body.MessageType))
						return
					}
				}
			}

			messageType, _ = strconv.ParseInt(strconv.Itoa(body.MessageType), 16, 32)
			body.MessageType = int(messageType)
			bodies.BodyMap[int(messageType)] = body
		}
	}

	return
}
