package rules

type CheckCode struct {
	Exists    bool   `xml:"exists,attr"`
	Start     int    `xml:"start,attr"`
	Algorithm string `xml:"algorithm,attr"`
	Begin     int    `xml:"begin,attr"`

	BytesCount int
}

func (this *CheckCode) verifyFields() {
	if this.Exists {
		if this.Algorithm != string(AlgorithmType_Xor) && this.Algorithm != string(AlgorithmType_Crc16) {
			panic("Checkcode's algorithm type must be xor or crc16")
		}

		if this.Algorithm == string(AlgorithmType_Crc16) {
			this.BytesCount = 2
		} else {
			this.BytesCount = 1
		}
	}
}
