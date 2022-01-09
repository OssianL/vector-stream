package main

import (
	"log"
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
	bytes      []byte
	i          int
	lastOpcode uint8
}

func (b *Bytecode) error(description string) {
	log.Fatal("i: ", b.i, " lastOpcode: ", b.lastOpcode, " error: ", description)
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

func (b *Bytecode) popUint8() uint8 {
	if b.i >= len(b.bytes) {
		b.error("popUint8 out of range")
	}
	var value uint8 = b.bytes[b.i]
	b.i += 1
	return value
}

func (s *Bytecode) pushUint16(value uint16) {
	high := byte(value >> 8)
	low := byte(value)
	s.bytes = append(s.bytes, low, high)
}

func (b *Bytecode) popUint16() uint16 {
	if (b.i + 1) >= len(b.bytes) {
		b.error("popUint16 out of range")
	}
	low := b.bytes[b.i]
	high := b.bytes[b.i+1]
	value := (uint16(high) << 8) + uint16(low)
	b.i += 2
	return value
}

func (b *Bytecode) popOpcode() uint8 {
	opcode := b.popUint8()
	b.lastOpcode = opcode
	return opcode
}

func (b *Bytecode) pushVec2(point Vec2) {
	b.pushUint16(uint16(point.X))
	b.pushUint16(uint16(point.Y))
}

func (b *Bytecode) popVec2() Vec2 {
	return Vec2{X: float64(b.popUint16()), Y: float64(b.popUint16())}
}

func (b *Bytecode) pushRect(value FixpointRect) {
	b.pushUint16(value.X)
	b.pushUint16(value.Y)
	b.pushUint16(value.W)
	b.pushUint16(value.H)
}

func (b *Bytecode) popRect() Rect {
	x := b.popUint16()
	y := b.popUint16()
	w := b.popUint16()
	h := b.popUint16()
	return Rect{position: Vec2{X: float64(x), Y: float64(y)}, W: float64(w), H: float64(h)}
}

func (b *Bytecode) pushRgba(value nanovgo.Color) {
	b.pushUint8(uint8(value.R * 255))
	b.pushUint8(uint8(value.G * 255))
	b.pushUint8(uint8(value.B * 255))
	b.pushUint8(uint8(value.A * 255))
}

func (b *Bytecode) popRgba() nanovgo.Color {
	color := nanovgo.Color{}
	red := b.popUint8()
	color.R = float32(red) / 255
	green := b.popUint8()
	color.G = float32(green) / 255
	blue := b.popUint8()
	color.B = float32(blue) / 255
	alpha := b.popUint8()
	color.A = float32(alpha) / 255
	return color
}

func (b *Bytecode) pushRotation(rotation float64) {
	rotationUint := uint16(math.Abs(rotation/(math.Pi*2)) * math.MaxUint16)
	b.pushUint16(rotationUint)
}

func (b *Bytecode) popRotation() float64 {
	rotationUint := b.popUint16()
	return float64(rotationUint) / math.MaxUint16 * math.Pi * 2
}

func (b *Bytecode) pushScale(scale Vec2) {
	b.pushVec2(scale.MultiplyFloat(uintScaleFactor))
}

func (b *Bytecode) popScale() Vec2 {
	scale := b.popVec2()
	return scale.DivideFloat(uintScaleFactor)
}
