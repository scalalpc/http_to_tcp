package rules

type Field struct {
	Name       string `xml:"name"`
	DataType   string `xml:"dataType"`
	BytesCount int    `xml:"bytesCount"`
}
