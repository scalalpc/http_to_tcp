package rules

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

var ruleContext *RuleContext
var ruleContextOnce sync.Once

func GetRuleContext() *RuleContext {
	ruleContextOnce.Do(func() {
		header, err := makeHeaderRule()
		if err != nil {
			fmt.Println(fmt.Sprintf("%v", err))
			os.Exit(1)
		}

		bodies, err := makeBodiesRule()
		if err != nil {
			fmt.Println(fmt.Sprintf("%v", err))
			os.Exit(1)
		}

		rule, err := makeRule()
		if err != nil {
			fmt.Println(fmt.Sprintf("%v", err))
			os.Exit(1)
		}

		err = checkBodySizeRule(rule, header)
		if err != nil {
			fmt.Println(fmt.Sprintf("%v", err))
			os.Exit(1)
		}

		rule.HeaderSize = header.Size

		ruleContext = &RuleContext{
			Header: header,
			Bodies: bodies,
			Rule:   rule,
		}
	})

	return ruleContext
}

type RuleContext struct {
	Header Header
	Bodies Bodies
	Rule   Rule
}

func checkBodySizeRule(rule Rule, header Header) (err error) {
	if rule.BodySize.BytesCount > 0 {
		found := false
		prevBytesCount := len(rule.HeadByteArr)
		for _, headerField := range header.FieldArr {
			if headerField.Name == BodySizeFieldName {
				found = true
				if headerField.DataType == string(DataType_Byte) || headerField.DataType == string(DataType_Uint16) {
					if rule.BodySize.Start != prevBytesCount {
						err = errors.New("The starting position of bodySize field in package rule does not match the position of bodySize field in message header rule.")
					}
				} else {
					err = errors.New("The data type of bodySize field in header rule can only be byte or uint16.")
				}
				break
			} else {
				prevBytesCount += headerField.BytesCount
			}
		}
		if !found {
			err = errors.New("Package rule requires bodySize field to be defined in message header.")
		}
	}
	return
}
