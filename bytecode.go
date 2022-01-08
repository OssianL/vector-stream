package main

import (
	"errors"
	"math"

	"github.com/shibukawa/nanovgo"
)

const uintScaleFactor = 100

// render operations
const (
	ropBeginPath = iota
	ropSetFillColor
	ropFill
	ropRectangle

	ropMacroCall

	ropCodeCount
)

// update operations
const (
	uopMacroStart = iota
	uopMacroEnd
	uopMacroOperation
	uopMacroVar8
	uopMacroVar16
	uopMacroUseVar8
	uopMacroUseVar16
	uopMacroUseConst8
	uopMacroUseConst16

	uopNodeCreate
	uopNodeSetContent
	uopNodeSetParent
	uopNodeSetPosition
	uopNodeSetRotation
	uopNodeSetScale
	// opNodeSetPseudoParent

	// opCreatePseudoNode

	// opContextCreate

	uopCodeCount
)

var updateOpcodeNames = [254]string{
	"uopFunDefStart", "uopFunDefEnd", "uopFunDefOperation", "uopFunDefVar8", "uopFunDefVar16",
	"uopFunDefUseVar8", "uopFunDefUseVar16", "uopFunDefUseConst8", "uopFunDefUseConst16",

	"uopNodeCreate", "uopNodeSetContent", "uopNodeSetParent", "uopNodeSetPosition", "uopNodeSetRotation", "uopNodeSetScale",
}

var renderOpcodeName = [256]string{
	"ropBeginPath", "ropSetFillColor", "ropFill", "ropRectangle", "uopFunCall",
}

type Bytecode struct {
	bytes []byte
	i     int
}

func NewBytecode() *Bytecode {
	bytecode := Bytecode{
		bytes: []byte{},
		i:     0,
	}
	return &bytecode
}

func NewBytecodeFromBytes(bytes []byte) *Bytecode {
	bytecode := Bytecode{
		bytes: bytes,
		i:     0,
	}
	return &bytecode
}

func (b *Bytecode) pushUint8(value uint8) {
	b.bytes = append(b.bytes, value)
}

func (c *Bytecode) popUint8() (uint8, error) {
	if c.i >= len(c.bytes) {
		return 0, errors.New("popUint8 out of range")
	}
	var value uint8 = c.bytes[c.i]
	c.i += 1
	return value, nil
}

func (s *Bytecode) pushUint16(value uint16) {
	high := byte(value >> 8)
	low := byte(value)
	s.bytes = append(s.bytes, low, high)
}

func (b *Bytecode) popUint16() (uint16, error) {
	if (b.i + 1) >= len(b.bytes) {
		return 0, errors.New("popUint16 out of range")
	}
	low := b.bytes[b.i]
	high := b.bytes[b.i+1]
	value := (uint16(high) << 8) + uint16(low)
	b.i += 2
	return value, nil
}

func (b *Bytecode) pushVec2(point Vec2) {
	b.pushUint16(uint16(point.X))
	b.pushUint16(uint16(point.Y))
}

func (b *Bytecode) popVec2() (Vec2, error) {
	x, err := b.popUint16()
	if err != nil {
		return Vec2{}, err
	}
	y, err := b.popUint16()
	if err != nil {
		return Vec2{}, err
	}
	return Vec2{X: float64(x), Y: float64(y)}, nil
}

func (b *Bytecode) pushRect(value FixpointRect) {
	b.pushUint16(value.X)
	b.pushUint16(value.Y)
	b.pushUint16(value.W)
	b.pushUint16(value.H)
}

func (b *Bytecode) popRect() (Rect, error) {
	x, err := b.popUint16()
	if err != nil {
		return Rect{}, err
	}
	y, err := b.popUint16()
	if err != nil {
		return Rect{}, err
	}
	w, err := b.popUint16()
	if err != nil {
		return Rect{}, err
	}
	h, err := b.popUint16()
	if err != nil {
		return Rect{}, err
	}
	return Rect{position: Vec2{X: float64(x), Y: float64(y)}, W: float64(w), H: float64(h)}, nil
}

func (b *Bytecode) pushRgba(value nanovgo.Color) {
	b.pushUint8(uint8(value.R * 255))
	b.pushUint8(uint8(value.G * 255))
	b.pushUint8(uint8(value.B * 255))
	b.pushUint8(uint8(value.A * 255))
}

func (b *Bytecode) popRgba() (nanovgo.Color, error) {
	color := nanovgo.Color{}
	red, err := b.popUint8()
	if err != nil {
		return color, err
	}
	color.R = float32(red) / 255
	green, err := b.popUint8()
	if err != nil {
		return color, err
	}
	color.G = float32(green) / 255
	blue, err := b.popUint8()
	if err != nil {
		return color, err
	}
	color.B = float32(blue) / 255
	alpha, err := b.popUint8()
	if err != nil {
		return color, err
	}
	color.A = float32(alpha) / 255
	return color, nil
}

func (b *Bytecode) pushRotation(rotation float64) {
	rotationUint := uint16(math.Abs(rotation/(math.Pi*2)) * math.MaxUint16)
	b.pushUint16(rotationUint)
}

func (b *Bytecode) popRotation() (float64, error) {
	rotationUint, err := b.popUint16()
	if err != nil {
		return 0, err
	}
	return float64(rotationUint) / math.MaxUint16 * math.Pi * 2, nil
}

func (b *Bytecode) pushScale(scale Vec2) {
	b.pushVec2(scale.MultiplyFloat(uintScaleFactor))
}

func (b *Bytecode) popScale() (Vec2, error) {
	scale, err := b.popVec2()
	if err != nil {
		return Vec2{}, err
	}
	return scale.DivideFloat(uintScaleFactor), nil
}
