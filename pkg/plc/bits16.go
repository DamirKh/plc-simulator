package plc

func GetBit16(value int16, bit int) bool {
	return (value>>bit)&1 == 1
}

func SetBit16(value int16, bit int) int16 {
	return value | (1 << bit)
}

func ClearBit16(value int16, bit int) int16 {
	return value &^ (1 << bit)
}

func SetBitValue16(value int16, bit int, bitVal bool) int16 {
	if bitVal {
		return SetBit16(value, bit)
	}
	return ClearBit16(value, bit)
}
