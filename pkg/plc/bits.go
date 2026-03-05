// pkg/plc/bits.go
package plc

// GetBit извлекает бит из DINT
func GetBit(value int32, bit int) bool {
	return (value>>bit)&1 == 1
}

// SetBit устанавливает бит в 1
func SetBit(value int32, bit int) int32 {
	return value | (1 << bit)
}

// ClearBit сбрасывает бит в 0
func ClearBit(value int32, bit int) int32 {
	return value &^ (1 << bit)
}

// SetBitValue устанавливает бит в нужное значение
func SetBitValue(value int32, bit int, bitVal bool) int32 {
	if bitVal {
		return SetBit(value, bit)
	}
	return ClearBit(value, bit)
}
