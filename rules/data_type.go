package rules

type DataTypeEnum string

const (
	DataType_Bcd     DataTypeEnum = "bcd"
	DataType_Byte    DataTypeEnum = "byte"
	DataType_Bytes   DataTypeEnum = "bytes"
	DataType_Float32 DataTypeEnum = "float32"
	DataType_Float64 DataTypeEnum = "float64"
	DataType_Int16   DataTypeEnum = "int16"
	DataType_Int32   DataTypeEnum = "int32"
	DataType_Int64   DataTypeEnum = "int64"
	DataType_String  DataTypeEnum = "string"
	DataType_Uint16  DataTypeEnum = "uint16"
	DataType_Uint32  DataTypeEnum = "uint32"
	DataType_Uint64  DataTypeEnum = "uint64"
)
