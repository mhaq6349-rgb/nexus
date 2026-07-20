package types

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

const (
	LangRust   = 0
	LangGo     = 1
	LangPython = 2
	LangTS     = 3
)

const (
	MsgCall   = 0x0001
	MsgReturn = 0x0002
	MsgError  = 0x0003
	MsgPing   = 0x00FE
	MsgPong   = 0x00FF
	MsgLog    = 0x0100
	MsgMetric = 0x0101
)

type ValueType uint8

const (
	ValNull   ValueType = 0
	ValBool   ValueType = 1
	ValI64    ValueType = 2
	ValU64    ValueType = 3
	ValF64    ValueType = 4
	ValString ValueType = 5
	ValBytes  ValueType = 6
	ValList   ValueType = 7
	ValMap    ValueType = 8
	ValNdArray ValueType = 9
)

type Value struct {
	Type   ValueType
	Bool   bool
	I64    int64
	U64    uint64
	F64    float64
	String string
	Bytes  []byte
	List   []Value
	Map    map[string]Value
	NdArray *NdArray
}

type NdArray struct {
	DType uint8
	Shape []uint64
	Data  []byte
}

type CallMessage struct {
	CallID    uint64
	Function  string
	Args      []Value
	TimeoutMs uint64
}

type ReturnMessage struct {
	CallID uint64
	Result Value
}

type ErrorMessage struct {
	CallID  uint64
	Code    int32
	Message string
}

type FunctionHandler func(args []Value) (Value, error)

type FunctionInfo struct {
	Name    string
	Handler FunctionHandler
	Lang    uint8
	Latency time.Duration
	Count   int64
}

func ValF64From(v float64) Value   { return Value{Type: ValF64, F64: v} }
func ValStr(s string) Value        { return Value{Type: ValString, String: s} }
func ValI64From(v int64) Value     { return Value{Type: ValI64, I64: v} }
func ValNullValue() Value          { return Value{Type: ValNull} }
func ValBytes(b []byte) Value      { return Value{Type: ValBytes, Bytes: b} }
func ValBool(b bool) Value         { return Value{Type: ValBool, Bool: b} }

func (v Value) AsF64() (float64, bool) {
	if v.Type == ValF64 { return v.F64, true }
	return 0, false
}

func (v Value) AsStr() (string, bool) {
	if v.Type == ValString { return v.String, true }
	return "", false
}

func (v Value) AsI64() (int64, bool) {
	if v.Type == ValI64 { return v.I64, true }
	return 0, false
}

func MarshalValue(v Value) ([]byte, error) {
	switch v.Type {
	case ValNull:
		return []byte{0}, nil
	case ValBool:
		if v.Bool { return []byte{1, 1}, nil }
		return []byte{1, 0}, nil
	case ValI64:
		b := make([]byte, 9)
		b[0] = 2
		binary.LittleEndian.PutUint64(b[1:], uint64(v.I64))
		return b, nil
	case ValF64:
		b := make([]byte, 9)
		b[0] = 4
		binary.LittleEndian.PutUint64(b[1:], math.Float64bits(v.F64))
		return b, nil
	case ValString:
		b := make([]byte, 9+len(v.String))
		b[0] = 5
		binary.LittleEndian.PutUint64(b[1:9], uint64(len(v.String)))
		copy(b[9:], v.String)
		return b, nil
	default:
		return MarshalValue(ValStr(fmt.Sprintf("%v", v)))
	}
}

func UnmarshalValue(b []byte) (Value, int, error) {
	if len(b) < 1 { return Value{}, 0, fmt.Errorf("empty") }
	switch b[0] {
	case 0:
		return ValNullValue(), 1, nil
	case 1:
		return ValBool(b[1] != 0), 2, nil
	case 2:
		return ValI64From(int64(binary.LittleEndian.Uint64(b[1:9]))), 9, nil
	case 4:
		return ValF64From(math.Float64frombits(binary.LittleEndian.Uint64(b[1:9]))), 9, nil
	case 5:
		l := binary.LittleEndian.Uint64(b[1:9])
		return ValStr(string(b[9:9+l])), 9 + int(l), nil
	default:
		return ValNullValue(), 1, nil
	}
}
