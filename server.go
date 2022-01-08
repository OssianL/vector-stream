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

var testMacro1 uint16
var testMacro2 uint16
var testNode1 uint16
var testNode2 uint16

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
		rect:           Rect{Vec2{0, 0}, 30, 30},
		rectDirectionX: 1,
		rectDirectionY: 1,
	}
	return &server
}

func (s *Server) Init() []byte {
	s.Bytecode = *NewBytecode()
	s.startTime = time.Now()
	testMacro1 = s.defineTestMacro(20, 30)
	testMacro2 = s.defineTestMacro(10, 40)
	testNode1 = s.createTestNode(testMacro1, s.rect.W, s.rect.H)
	testNode2 = s.createTestNode(testMacro2, 20, 40)
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

	xOffset := sinTime * 70
	s.nodeSetPosition(testNode2, Vec2{xOffset, 0})

	return s.bytes
}

func (s *Server) defineTestMacro(x uint16, y uint16) uint16 {
	macroNumber := s.macroStart()
	width := s.macroVar16()
	height := s.macroVar16()

	s.macroOperation(ropBeginPath)
	s.macroOperation(ropRectangle)
	s.macroUseConst16(x) // x
	s.macroUseConst16(y) // y
	s.macroUseVar16(width)
	s.macroUseVar16(height)
	s.macroOperation(ropSetFillColor)
	s.macroUseConstColor(colorRed)
	s.macroOperation(ropFill)

	s.macroEnd()

	return macroNumber
}

func (s *Server) createTestNode(macroNumber uint16, width float64, height float64) uint16 {
	nodeNumber := s.nodeCreate()
	s.nodeSetContent(nodeNumber, macroNumber)
	s.pushUint16(uint16(width))
	s.pushUint16(uint16(height))
	return nodeNumber
}

func (s *Server) macroUseConstColor(color nanovgo.Color) {
	s.macroUseConst8(uint8(color.R * 255))
	s.macroUseConst8(uint8(color.R * 255))
	s.macroUseConst8(uint8(color.R * 255))
	s.macroUseConst8(uint8(color.R * 255))
}

//-----------------UPDATE OPERATIONS----------------------------
//-----------------UPDATE OPERATIONS----------------------------
//-----------------UPDATE OPERATIONS----------------------------
//-----------------UPDATE OPERATIONS----------------------------
//-----------------UPDATE OPERATIONS----------------------------
//-----------------UPDATE OPERATIONS----------------------------

func (s *Server) macroStart() uint16 {
	s.pushUint8(uopMacroStart)
	macroNumber := s.macroCount
	s.pushUint16(macroNumber)
	s.macroCount++
	return macroNumber
}

func (s *Server) macroEnd() {
	s.pushUint8(uopMacroEnd)
	s.macroVariableCount = 0
}

func (s *Server) macroOperation(opcode uint8) {
	s.pushUint8(uopMacroOperation)
	s.pushUint8(opcode)
}

func (s *Server) macroVar8() uint16 {
	s.pushUint8(uopMacroVar8)
	variableNumber := s.macroVariableCount
	s.macroVariableCount++
	return variableNumber
}

func (s *Server) macroVar16() uint16 {
	s.pushUint8(uopMacroVar16)
	variableNumber := s.macroVariableCount
	s.macroVariableCount++
	return variableNumber
}

func (s *Server) macroUseVar8(variableNumber uint16) {
	s.pushUint8(uopMacroUseVar8)
	s.pushUint16(variableNumber)
}

func (s *Server) macroUseVar16(variableNumber uint16) {
	s.pushUint8(uopMacroUseVar16)
	s.pushUint16(variableNumber)
}

func (s *Server) macroUseConst8(const8 uint8) {
	s.pushUint8(uopMacroUseConst8)
	s.pushUint8(const8)
}

func (s *Server) macroUseConst16(const16 uint16) {
	s.pushUint8(uopMacroUseConst16)
	s.pushUint16(const16)
}

func (s *Server) nodeCreate() uint16 {
	s.pushUint8(uopNodeCreate)
	nodeNumber := s.nodeCount
	s.pushUint16(nodeNumber)
	s.nodeCount++
	return nodeNumber
}

func (s *Server) nodeSetContent(nodeNumber uint16, macroNumber uint16) {
	s.pushUint8(uopNodeSetContent)
	s.pushUint16(nodeNumber)
	s.pushUint16(macroNumber)
}

func (s *Server) nodeSetParent(nodeNumber uint16, parentNumber uint16) {
	s.pushUint8(uopNodeSetParent)
	s.pushUint16(nodeNumber)
	s.pushUint16(parentNumber)
}

func (s *Server) nodeSetPosition(nodeNumber uint16, position Vec2) {
	s.pushUint8(uopNodeSetPosition)
	s.pushUint16(nodeNumber)
	s.pushVec2(position)
}

func (s *Server) nodeSetRotation(nodeNumber uint16, rotation float64) {
	s.pushUint8(uopNodeSetRotation)
	s.pushUint16(nodeNumber)
	s.pushRotation(rotation)
}

func (s *Server) nodeSetScale(nodeNumber uint16, scale Vec2) {
	s.pushUint8(uopNodeSetScale)
	s.pushUint16(nodeNumber)
	s.pushScale(scale)
}

//-------------------------RENDER OPERATIONS---------------------------
//-------------------------RENDER OPERATIONS---------------------------
//-------------------------RENDER OPERATIONS---------------------------
//-------------------------RENDER OPERATIONS---------------------------
//-------------------------RENDER OPERATIONS---------------------------
//-------------------------RENDER OPERATIONS---------------------------

func (s *Server) beginPath() {
	s.pushUint8(ropBeginPath)
}

func (s *Server) setFillColor(color nanovgo.Color) {
	s.pushUint8(ropSetFillColor)
	s.pushRgba(color)
}

func (s *Server) fill() {
	s.pushUint8(ropFill)
}

func (s *Server) rectangle(rect FixpointRect) {
	s.pushUint8(ropRectangle)
	s.pushRect(rect)
}

func (s *Server) macroCall(macroNumber uint16) {
	s.pushUint8(ropMacroCall)
	s.pushUint16(macroNumber)
}
