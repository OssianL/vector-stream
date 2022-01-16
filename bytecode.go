package main

import (
	"fmt"
	"log"
	"math"

	"github.com/shibukawa/nanovgo"
)

const uintScaleFactor = 100
const fixpointMultiplier = 2 << 15

// build in anchors
const (
	anchorContextTopLeft = iota
	anchorContextTopRight
	anchorContextBottomLeft
	anchorContextBottomRight
	anchorContextCenter

	anchorFrameTopLeft
	anchorFrameTopRight
	anchorFrameBottomLeft
	anchorFrameBottomRight
	anchorFrameCenter

	anchorMouse
)

// render operations
const (
	ropBeginPath = iota
	ropSetFillColor
	ropFill
	ropMoveTo
	ropLineTo
	ropClosePath

	ropMacroCall

	ropUseAnchor

	ropCodeCount
)

// update operations
const (
	uopMacroStart = iota
	uopMacroEnd
	uopMacroOperation
	uopMacroVar
	uopMacroUseVar
	uopMacroUseConst

	uopNodeCreate
	uopNodeSetContent
	uopNodeSetParent
	uopNodeSetPosition
	uopNodeSetRotation
	uopNodeSetScale

	uopAnchorCreate

	// opCreatePseudoNode

	// opContextCreate

	uopCodeCount
)

var updateOpcodeNames = [254]string{
	"uopMacroDefStart", "uopMacroDefEnd", "uopMacroDefOperation", "uopMacroDefVar",
	"uopMacroDefUseVar", "uopMacroDefUseConst",

	"uopNodeCreate", "uopNodeSetContent", "uopNodeSetParent", "uopNodeSetPosition", "uopNodeSetRotation", "uopNodeSetScale",
}

var renderOpcodeName = [256]string{
	"ropBeginPath", "ropSetFillColor", "ropFill", "ropMoveTo", "ropLineTo", "ropClosePath", "ropMacroCall",
}

type AnchorNumber uint16
type NodeNumber uint16
type MacroNumber uint16

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
	high := uint8(value >> 8)
	low := uint8(value)
	s.bytes = append(s.bytes, low, high)
}

func (b *Bytecode) popUint16() uint16 {
	if (b.i + 1) >= len(b.bytes) {
		b.error("popUint16 out of range")
	}
	low := b.bytes[b.i]
	high := b.bytes[b.i+1]
	b.i += 2
	return (uint16(high) << 8) + uint16(low)
}

func (b *Bytecode) pushNodeNumber(number NodeNumber) {
	b.pushUint16(uint16(number))
}

func (b *Bytecode) popNodeNumber() NodeNumber {
	return NodeNumber(b.popUint16())
}

func (b *Bytecode) pushMacroNumber(number MacroNumber) {
	b.pushUint16(uint16(number))
}

func (b *Bytecode) popMacroNumber() MacroNumber {
	return MacroNumber(b.popUint16())
}

func (b *Bytecode) pushAnchorNumber(number AnchorNumber) {
	b.pushUint16(uint16(number))
}

func (b *Bytecode) popAnchorNumber() AnchorNumber {
	return AnchorNumber(b.popUint16())
}

func (b *Bytecode) pushUint32(value uint32) {
	b1 := uint8(value >> 24) // highest, most significant
	b2 := uint8(value >> 16)
	b3 := uint8(value >> 8)
	b4 := uint8(value) // lowest, least significant
	b.bytes = append(b.bytes, b4, b3, b2, b1)
}

func (b *Bytecode) popUint32() uint32 {
	if (b.i + 3) >= len(b.bytes) {
		b.error("popUint32 out of range")
	}
	b1 := b.bytes[b.i+3]
	b2 := b.bytes[b.i+2]
	b3 := b.bytes[b.i+1]
	b4 := b.bytes[b.i]
	b.i += 4
	return (uint32(b1) << 24) + (uint32(b2) << 16) + (uint32(b3) << 8) + uint32(b4)
}

func (b *Bytecode) pushInt32(value int32) {
	b1 := uint8(value >> 24) // highest, most significant
	b2 := uint8(value >> 16)
	b3 := uint8(value >> 8)
	b4 := uint8(value) // lowest, least significant
	b.bytes = append(b.bytes, b4, b3, b2, b1)
}

func (b *Bytecode) popInt32() int32 {
	if (b.i + 3) >= len(b.bytes) {
		b.error("popUint32 out of range")
	}
	b1 := b.bytes[b.i+3]
	b2 := b.bytes[b.i+2]
	b3 := b.bytes[b.i+1]
	b4 := b.bytes[b.i]
	b.i += 4
	fmt.Println("popInt32: b1: ", b1, " b2: ", b2, " b3: ", b3, " b4: ", b4)
	return (int32(b1) << 24) + (int32(b2) << 16) + (int32(b3) << 8) + int32(b4)
}

func (b *Bytecode) pushFloat64(value float64) {
	b.pushInt32(int32(value * fixpointMultiplier))
}

func (b *Bytecode) popFloat64() float64 {
	return float64(b.popInt32()) / fixpointMultiplier
}

func (b *Bytecode) pushOpcode(opcode uint8) {
	b.pushUint8(opcode)
}

func (b *Bytecode) popOpcode() uint8 {
	opcode := b.popUint8()
	b.lastOpcode = opcode
	return opcode
}

func (b *Bytecode) pushSize(size int) {
	if size > math.MaxUint8 {
		b.error("pushSize: size uint8 overflow")
	}
	b.pushUint8(uint8(size))
}

func (b *Bytecode) pushVec2(point Vec2) {
	b.pushFloat64(point.X)
	b.pushFloat64(point.Y)
}

func (b *Bytecode) popVec2() Vec2 {
	return Vec2{X: b.popFloat64(), Y: b.popFloat64()}
}

// func (b *Bytecode) pushRect(value FixpointRect) {
// 	b.pushUint16(value.X)
// 	b.pushUint16(value.Y)
// 	b.pushUint16(value.W)
// 	b.pushUint16(value.H)
// }

func (b *Bytecode) popRect() Rect {
	position := b.popVec2()
	size := b.popVec2()
	return Rect{position, size}
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
