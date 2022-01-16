package main

import (
	"math"
	"time"

	"github.com/shibukawa/nanovgo"
)

var colorBlack = nanovgo.Color{R: 0, G: 0, B: 0, A: 1}
var colorRed = nanovgo.Color{R: 1, G: 0, B: 0, A: 1}
var colorGreen = nanovgo.Color{R: 0, G: 1, B: 0, A: 1}
var colorBlue = nanovgo.Color{R: 0, G: 0, B: 1, A: 1}

var testMacro1 MacroNumber
var testMacro2 MacroNumber
var testNode1 NodeNumber
var testNode2 NodeNumber

type Server struct {
	Bytecode

	startTime time.Time

	macroVariableCount uint16
	macroCount         uint16

	nodeCount uint16

	rect           Rect
	rectDirectionX float64
	rectDirectionY float64
}

func NewServer() *Server {
	var server Server = Server{
		Bytecode:       *NewBytecode(),
		rect:           Rect{Vec2{0, 0}, Vec2{30, 30}},
		rectDirectionX: 1,
		rectDirectionY: 1,
	}
	return &server
}

func (s *Server) Init() []byte {
	s.Bytecode = *NewBytecode()
	s.startTime = time.Now()
	testMacro1 = s.defineTestMacro(colorRed)
	testMacro2 = s.defineTestMacro(colorGreen)
	testNode1 = s.createTestNode(testMacro1, s.rect.size)
	testNode2 = s.createTestNode(testMacro2, Vec2{20, 40})
	s.nodeSetParent(testNode2, testNode1)
	s.nodeSetPosition(testNode2, Vec2{40, 40})
	return s.bytes
}

func (s *Server) Update() []byte {
	s.Bytecode = *NewBytecode()

	if s.rect.position.X > windowWidth {
		s.rect.position.X = windowWidth
		s.rectDirectionX = -1
	} else if s.rect.position.X <= 0 {
		s.rect.position.X = 0
		s.rectDirectionX = 1
	}
	if s.rect.position.Y > windowHeight {
		s.rect.position.Y = windowHeight
		s.rectDirectionY = -1
	} else if s.rect.position.Y <= 0 {
		s.rect.position.Y = 0
		s.rectDirectionY = 1
	}
	s.rect.position.X += s.rectDirectionX
	s.rect.position.Y += s.rectDirectionY

	// s.beginPath()
	// s.rectangle(s.rect)

	// r := float32(s.rect.X) / windowWidth
	// g := float32(s.rect.Y) / windowHeight
	// b := float32(0.0)
	// b := float32(math.Mod(float64(s.rect.X*s.rect.Y)/100, 1))

	// s.setFillColor(nanovgo.Color{R: r, G: g, B: b, A: 1.0})
	// s.fill()

	secondsFromStart := time.Since(s.startTime).Seconds()
	sinTime := math.Sin(secondsFromStart*4) + 1
	cosTime := math.Cos(secondsFromStart*4) + 1

	s.nodeSetPosition(testNode1, s.rect.position)
	s.nodeSetRotation(testNode1, secondsFromStart)
	s.nodeSetScale(testNode1, Vec2{sinTime, cosTime})

	// xOffset := sinTime * 70
	s.nodeSetPosition(testNode2, Vec2{20, 20})
	// s.nodeSetPosition(testNode2, Vec2{xOffset, 0})

	return s.bytes
}

func (s *Server) defineTestMacro(color nanovgo.Color) MacroNumber {
	macroNumber := s.macroStart()
	// sizeVar := s.macroVar(sizeOfVec2)

	s.macroOperation(ropBeginPath)
	s.macroOperation(ropMoveTo)
	s.macroUseConstVec2(Vec2{0, 0})
	s.macroOperation(ropLineTo)
	s.macroUseConstVec2(Vec2{100, 0})
	s.macroOperation(ropLineTo)
	s.macroUseConstVec2(Vec2{100, 100})
	s.macroOperation(ropLineTo)
	s.macroUseConstVec2(Vec2{0, 100})
	s.macroOperation(ropClosePath)
	s.macroOperation(ropSetFillColor)
	s.macroUseConstColor(color)
	s.macroOperation(ropFill)

	s.macroEnd()

	return macroNumber
}

func (s *Server) createTestNode(macroNumber MacroNumber, size Vec2) NodeNumber {
	nodeNumber := s.nodeCreate()
	s.nodeSetContent(nodeNumber, macroNumber)
	// s.pushUint16(uint16(width))
	// s.pushUint16(uint16(height))
	return nodeNumber
}

func (s *Server) macroUseConstColor(color nanovgo.Color) {
	s.macroUseConstUint8(uint8(color.R * 255))
	s.macroUseConstUint8(uint8(color.G * 255))
	s.macroUseConstUint8(uint8(color.B * 255))
	s.macroUseConstUint8(uint8(color.A * 255))
}

//-----------------UPDATE OPERATIONS----------------------------
//-----------------UPDATE OPERATIONS----------------------------
//-----------------UPDATE OPERATIONS----------------------------
//-----------------UPDATE OPERATIONS----------------------------
//-----------------UPDATE OPERATIONS----------------------------
//-----------------UPDATE OPERATIONS----------------------------

func (s *Server) macroStart() MacroNumber {
	s.pushOpcode(uopMacroStart)
	macroNumber := MacroNumber(s.macroCount)
	s.pushMacroNumber(macroNumber)
	s.macroCount++
	return macroNumber
}

func (s *Server) macroEnd() {
	s.pushOpcode(uopMacroEnd)
	s.macroVariableCount = 0
}

func (s *Server) macroOperation(opcode uint8) {
	s.pushOpcode(uopMacroOperation)
	s.pushUint8(opcode)
}

func (s *Server) macroVar(varSize int) uint16 {
	s.pushOpcode(uopMacroVar)
	s.pushSize(varSize)
	variableNumber := s.macroVariableCount
	s.macroVariableCount++
	return variableNumber
}

func (s *Server) macroUseVar(variableNumber uint16) {
	s.pushOpcode(uopMacroUseVar)
	s.pushUint16(variableNumber)
}

func (s *Server) macroUseConstUint8(const8 uint8) {
	s.pushOpcode(uopMacroUseConst)
	s.pushSize(1)
	s.pushUint8(const8)
}

func (s *Server) macroUseConstUint16(const16 uint16) {
	s.pushOpcode(uopMacroUseConst)
	s.pushSize(2)
	s.pushUint16(const16)
}

func (s *Server) macroUseConstVec2(constVec2 Vec2) {
	s.pushOpcode(uopMacroUseConst)
	s.pushSize(sizeOfVec2)
	s.pushVec2(constVec2)
}

func (s *Server) nodeCreate() NodeNumber {
	s.pushOpcode(uopNodeCreate)
	nodeNumber := NodeNumber(s.nodeCount)
	s.pushNodeNumber(nodeNumber)
	s.nodeCount++
	return nodeNumber
}

func (s *Server) nodeSetContent(nodeNumber NodeNumber, macroNumber MacroNumber) {
	s.pushOpcode(uopNodeSetContent)
	s.pushNodeNumber(nodeNumber)
	s.pushMacroNumber(macroNumber)
}

func (s *Server) nodeSetParent(nodeNumber NodeNumber, parentNumber NodeNumber) {
	s.pushOpcode(uopNodeSetParent)
	s.pushNodeNumber(nodeNumber)
	s.pushNodeNumber(parentNumber)
}

func (s *Server) nodeSetPosition(nodeNumber NodeNumber, position Vec2) {
	s.pushOpcode(uopNodeSetPosition)
	s.pushNodeNumber(nodeNumber)
	s.pushVec2(position)
}

func (s *Server) nodeSetRotation(nodeNumber NodeNumber, rotation float64) {
	s.pushOpcode(uopNodeSetRotation)
	s.pushNodeNumber(nodeNumber)
	s.pushRotation(rotation)
}

func (s *Server) nodeSetScale(nodeNumber NodeNumber, scale Vec2) {
	s.pushOpcode(uopNodeSetScale)
	s.pushNodeNumber(nodeNumber)
	s.pushScale(scale)
}

//-------------------------RENDER OPERATIONS---------------------------
//-------------------------RENDER OPERATIONS---------------------------
//-------------------------RENDER OPERATIONS---------------------------
//-------------------------RENDER OPERATIONS---------------------------
//-------------------------RENDER OPERATIONS---------------------------
//-------------------------RENDER OPERATIONS---------------------------

func (s *Server) beginPath() {
	s.pushOpcode(ropBeginPath)
}

func (s *Server) setFillColor(color nanovgo.Color) {
	s.pushOpcode(ropSetFillColor)
	s.pushRgba(color)
}

func (s *Server) fill() {
	s.pushOpcode(ropFill)
}

func (s *Server) macroCall(macroNumber uint16) {
	s.pushOpcode(ropMacroCall)
	s.pushUint16(macroNumber)
}
