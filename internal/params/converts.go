package params

import (
	"strconv"
)

func (p *Params) GetInt(name string) (int, error) {
	return strconv.Atoi(p.Get(name))
}

func (p *Params) GetInt8(name string) (int8, error) {
	result, err := strconv.ParseInt(p.Get(name), 10, 8)
	return int8(result), err
}

func (p *Params) GetUint8(name string) (uint8, error) {
	result, err := strconv.ParseUint(p.Get(name), 10, 8)
	return uint8(result), err
}

func (p *Params) GetInt16(name string) (int16, error) {
	result, err := strconv.ParseInt(p.Get(name), 10, 16)
	return int16(result), err
}

func (p *Params) GetUint16(name string) (uint16, error) {
	result, err := strconv.ParseUint(p.Get(name), 10, 16)
	return uint16(result), err
}

func (p *Params) GetUint32(name string) (uint32, error) {
	result, err := strconv.ParseUint(p.Get(name), 10, 32)
	return uint32(result), err
}

func (p *Params) GetInt32(name string) (int32, error) {
	result, err := strconv.ParseInt(p.Get(name), 10, 32)
	return int32(result), err
}

func (p *Params) GetInt64(name string) (int64, error) {
	return strconv.ParseInt(p.Get(name), 10, 64)
}

func (p *Params) GetUint64(name string) (uint64, error) {
	return strconv.ParseUint(p.Get(name), 10, 64)
}

func (p *Params) GetFloat(name string) (float64, error) {
	return strconv.ParseFloat(p.Get(name), 64)
}

func (p *Params) GetBool(name string) (bool, error) {
	return strconv.ParseBool(p.Get(name))
}
