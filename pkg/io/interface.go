package io

// DiscreteReader читает дискретный сигнал (команды от PLC)
type DiscreteReader interface {
	Read() (bool, error)
}

// DiscreteWriter пишет дискретный сигнал (фидбек к PLC)
type DiscreteWriter interface {
	Write(value bool) error
}
