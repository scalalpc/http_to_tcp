package rules

type BodySize struct {
	Exists   bool   `xml:"exists,attr"`
	DataType string `xml:"dataType,attr"`
	Start    int    `xml:"start,attr"`

	BytesCount int
}

func (this *BodySize) verifyFields() {
	if this.Exists {
		if this.DataType != string(DataType_Byte) && this.DataType != string(DataType_Uint16) {
			panic("Bodysize's data type must be byte or uint16")
		}

		if this.DataType == string(DataType_Uint16) {
			this.BytesCount = 2
		} else {
			this.BytesCount = 1
		}
	}
}
