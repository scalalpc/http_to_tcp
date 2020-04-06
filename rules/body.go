package rules

type Body struct {
	MessageType int     `xml:"messageType,attr"`
	FieldArr    []Field `xml:"field"`
}
