package main

import (
	"fmt"
	"log"
	"math"

	"github.com/shibukawa/nanovgo"
)

func debugPrint(i ...interface{}) {
	if true {
		fmt.Println(i)
	}
}

func debugPrint2(i ...interface{}) {
	if true {
		fmt.Println(i)
	}
}

type FunctionVariableReference struct {
	bytecodeIndex      int // index in the bytecode where the variable is used
	variableNumber     uint16
	variableStartIndex int // index of the first byte of the variable in the variables []byte
}

type Macro struct {
	bytecode *Bytecode

	variableReferences   []FunctionVariableReference
	variableSizes        []int
	variableStartIndexes []int
	totalVariablesSize   int
}

func NewMacro() *Macro {
	return &Macro{
		bytecode:           NewBytecode(),
		variableSizes:      []int{},
		totalVariablesSize: 0,
		variableReferences: []FunctionVariableReference{},
	}
}

func (f *Macro) Compile(variables []byte) *Bytecode {
	bytes := f.bytecode.bytes
	for _, variableReference := range f.variableReferences {
		for i := 0; i < f.variableSizes[variableReference.variableNumber]; i++ {
			bytes[variableReference.bytecodeIndex+i] = variables[variableReference.variableStartIndex+i]
		}
	}
	bytecode := NewBytecodeFromBytes(bytes)
	return bytecode
}

type Vec2 struct {
	X float64
	Y float64
	// W float64
}

func (v Vec2) MultiplyFloat(f float64) Vec2 {
	return Vec2{
		X: v.X * f,
		Y: v.Y * f,
	}
}

func (v Vec2) DivideFloat(f float64) Vec2 {
	return Vec2{
		X: v.X / f,
		Y: v.Y / f,
	}
}

// func (h HomoPoint) ToPoint() Point {
// 	return Point{
// 		X: h.X / h.W,
// 		Y: h.Y / h.W,
// 	}
// }

// type Point struct {
// 	X float64
// 	Y float64
// }

// func (p Point) ToHomoPoint() HomoPoint {
// 	return HomoPoint{
// 		X: float64(p.X),
// 		Y: float64(p.Y),
// 		W: 1.0,
// 	}
// }

type Matrix33 struct {
	m00 float64
	m01 float64
	m02 float64
	m10 float64
	m11 float64
	m12 float64
	m20 float64
	m21 float64
	m22 float64
}

func BuildTranslationMatrix(translation Vec2) Matrix33 {
	return Matrix33{
		1, 0, translation.X,
		0, 1, translation.Y,
		0, 0, 1,
	}
}

func BuildRotationMatrix(rotation float64) Matrix33 {
	cos := float64(math.Cos(float64(rotation)))
	sin := float64(math.Sin(float64(rotation)))
	return Matrix33{
		cos, -sin, 0,
		sin, cos, 0,
		0, 0, 1,
	}
}

func BuildScaleMatrix(scale Vec2) Matrix33 {
	return Matrix33{
		scale.X, 0, 0,
		0, scale.Y, 0,
		0, 0, 1,
	}
}

func BuildTransformationMatrix(translation Vec2, rotation float64, scale Vec2) Matrix33 {
	translationMatrix := BuildTranslationMatrix(translation)
	rotationMatrix := BuildRotationMatrix(rotation)
	scaleMatrix := BuildScaleMatrix(scale)
	return translationMatrix.MultiplyMatrix(rotationMatrix).MultiplyMatrix(scaleMatrix)

}

func (t Matrix33) MultiplyMatrix(o Matrix33) Matrix33 {
	return Matrix33{
		m00: t.m00*o.m00 + t.m01*o.m10 + t.m02*o.m20,
		m01: t.m00*o.m01 + t.m01*o.m11 + t.m02*o.m21,
		m02: t.m00*o.m02 + t.m01*o.m12 + t.m02*o.m22,
		m10: t.m10*o.m00 + t.m11*o.m10 + t.m12*o.m20,
		m11: t.m10*o.m01 + t.m11*o.m11 + t.m12*o.m21,
		m12: t.m10*o.m02 + t.m11*o.m12 + t.m12*o.m22,
		m20: t.m20*o.m00 + t.m21*o.m10 + t.m22*o.m20,
		m21: t.m20*o.m01 + t.m21*o.m11 + t.m22*o.m21,
		m22: t.m20*o.m02 + t.m21*o.m12 + t.m22*o.m22,
	}
}

func (t Matrix33) MultiplyVec2(p Vec2) Vec2 {
	w := t.m20*p.X + t.m21*p.Y + t.m22
	return Vec2{
		X: (t.m00*p.X + t.m01*p.Y + t.m02) / w,
		Y: (t.m10*p.X + t.m11*p.Y + t.m12) / w,
	}
}

// func (t Matrix33) MultiplyPoint(p Point) Point {
// 	return t.MultiplyHomoPoint(p.ToHomoPoint()).ToPoint()
// }

type Node struct {
	renderCode    *Bytecode
	localToGlobal Matrix33
	position      Vec2
	rotation      float64
	scale         Vec2
	parent        *Node
	children      map[*Node]struct{}
}

func NewNode() *Node {
	return &Node{
		children: map[*Node]struct{}{},
		scale:    Vec2{1, 1},
	}
}

func (n *Node) AddChild(child *Node) {
	if child.parent != nil {
		delete(child.parent.children, child)
	}
	child.parent = n
	n.children[child] = struct{}{}
}

func (n *Node) UpdateLocalToGlobalMatrix() {
	transformationMatrix := BuildTransformationMatrix(n.position, n.rotation, n.scale)
	if n.parent != nil {
		n.localToGlobal = n.parent.localToGlobal.MultiplyMatrix(transformationMatrix)
	} else {
		n.localToGlobal = transformationMatrix
	}
}

func (n *Node) TransformPoint(point Vec2) Vec2 {
	return n.localToGlobal.MultiplyVec2(point)
}

func (n *Node) TransformPoints(points []Vec2) []Vec2 {
	for i, p := range points {
		points[i] = n.localToGlobal.MultiplyVec2(p)
	}
	return points
}

type Client struct {
	*Bytecode
	updateOperations [256]func()
	renderOperations [256]func(*Node)
	nvgCtx           *nanovgo.Context
	stack            []*Bytecode
	macros           map[uint16]*Macro
	wipMacro         *Macro
	wipMacroNumber   uint16
	nodes            map[uint16]*Node
	root             *Node
}

func NewClient(nvgCtx *nanovgo.Context) *Client {
	client := Client{
		Bytecode: NewBytecode(),
		nvgCtx:   nvgCtx,
		stack:    []*Bytecode{},
		macros:   map[uint16]*Macro{},
		nodes:    map[uint16]*Node{},
		root:     NewNode(),
	}
	client.updateOperations = [256]func(){
		client.macroDefStart, client.macroDefEnd, client.macroDefOperation, client.macroDefVar8, client.macroDefVar16,
		client.macroDefUseVar8, client.macroDefUseVar16, client.macroDefUseConst8, client.macroDefUseConst16,

		client.nodeCreate, client.nodeSetContent, client.nodeSetParent, client.nodeSetPosition, client.nodeSetRotation, client.nodeSetScale,
	}
	client.renderOperations = [256]func(*Node){
		client.beginPath, client.setFillColor, client.fill, client.rectangle,
		client.macroCall,
	}
	return &client
}

func (c *Client) error(description string) {
	log.Fatal(description)
}

func (c *Client) Update(bytecode *Bytecode) {
	debugPrint2("update bytes: ", bytecode.bytes)
	c.Bytecode = bytecode
	for {
		for c.i < len(c.bytes) {
			c.updateStep()
		}
		if !c.popState() {
			return
		}
	}
}

func (c *Client) updateStep() {
	opcode := c.popOpcode()
	if opcode > uopCodeCount {
		c.error("invalid update opcode: " + fmt.Sprint(opcode))
	}
	debugPrint("i: ", c.i-1, " opcode: ", updateOpcodeNames[opcode])
	c.updateOperations[opcode]()
}

func (c *Client) Render() {
	debugPrint("Render start")
	c.renderNode(c.root)
}

func (c *Client) renderNode(node *Node) {
	if node == nil {
		return
	}
	node.UpdateLocalToGlobalMatrix()
	if node.renderCode != nil {
		c.pushState(node.renderCode)
		for c.i < len(c.bytes) {
			c.renderStep(node)
		}
	}
	for child := range node.children {
		c.renderNode(child)
	}
}

func (c *Client) renderStep(node *Node) {
	opcode := c.popOpcode()
	debugPrint("i: ", c.i-1, " opcode: ", renderOpcodeName[opcode])
	if opcode > ropCodeCount {
		c.error("invalid render opcode: " + fmt.Sprint(opcode))
	}
	c.renderOperations[opcode](node)
}

func (c *Client) pushState(bytecode *Bytecode) {
	debugPrint("push state, c.Bytecode == nil: ", c.Bytecode == nil)
	if bytecode == nil {
		c.error("pushState nil bytecode")
	}
	if c.Bytecode != nil {
		c.stack = append(c.stack, c.Bytecode)
	}
	bytecode.i = 0
	c.Bytecode = bytecode
}

func (c *Client) popState() bool {
	debugPrint("pop state len(c.stack): ", len(c.stack))
	if len(c.stack) == 0 {
		c.stack = []*Bytecode{}
		c.Bytecode = nil
		return false
	}
	topIndex := len(c.stack) - 1
	c.Bytecode = c.stack[topIndex]
	c.stack[topIndex] = nil
	c.stack = c.stack[:topIndex]
	return true
}

func (c *Client) popAndCompileMacro() *Bytecode {
	macroNumber := c.popUint16()
	macro, ok := c.macros[macroNumber]
	if !ok {
		c.error("popAndCompileMacro: invalid macroNumber: " + fmt.Sprint(macroNumber))
	}
	if c.i+macro.totalVariablesSize > len(c.bytes) {
		c.error("popAndCompileMacro: macro variable block out of range")
	}
	variables := c.bytes[c.i : c.i+macro.totalVariablesSize]
	c.i += len(variables)
	return macro.Compile(variables)
}

func (c *Client) popNode() *Node {
	nodeNumber := c.popUint16()
	node, ok := c.nodes[nodeNumber]
	if !ok {
		c.error("popNode: invalid nodeNumber")
	}
	return node
}

// ---------------------UPDATE OPERATIONS---------------------
// ---------------------UPDATE OPERATIONS---------------------
// ---------------------UPDATE OPERATIONS---------------------
// ---------------------UPDATE OPERATIONS---------------------
// ---------------------UPDATE OPERATIONS---------------------
// ---------------------UPDATE OPERATIONS---------------------
// ---------------------UPDATE OPERATIONS---------------------
// ---------------------UPDATE OPERATIONS---------------------

func (c *Client) macroDefStart() {
	if c.wipMacro != nil {
		c.error("macroDefStart: wip function already in progress")
	}
	functionNumber := c.popUint16()
	newWipMacro := NewMacro()
	c.wipMacro = newWipMacro
	c.wipMacroNumber = functionNumber
	debugPrint("funDefStart functionNumber: ", functionNumber)
}

func (c *Client) macroDefEnd() {
	if c.wipMacro == nil {
		c.error("macroDefEnd: nil wip function")
	}
	c.macros[c.wipMacroNumber] = c.wipMacro
	c.wipMacro = nil
}

func (c *Client) macroDefOperation() {
	if c.wipMacro == nil {
		c.error("macroDefOperation: nil wip function")
	}
	opcode := c.popUint8()
	c.wipMacro.bytecode.pushUint8(opcode)
}

func (c *Client) macroDefVar8() {
	if c.wipMacro == nil {
		c.error("macroDefVar8: nil wip function")
	}
	c.wipMacro.variableSizes = append(c.wipMacro.variableSizes, 1)
	c.wipMacro.variableStartIndexes = append(c.wipMacro.variableStartIndexes, c.wipMacro.totalVariablesSize)
	c.wipMacro.totalVariablesSize += 1
}

func (c *Client) macroDefVar16() {
	if c.wipMacro == nil {
		c.error("macroDefVar16: nil wip function")
	}
	c.wipMacro.variableSizes = append(c.wipMacro.variableSizes, 2)
	c.wipMacro.variableStartIndexes = append(c.wipMacro.variableStartIndexes, c.wipMacro.totalVariablesSize)
	c.wipMacro.totalVariablesSize += 2
}

func (c *Client) macroDefUseVar8() {
	if c.wipMacro == nil {
		c.error("macroDefUseVar8: nil wip function")
	}
	variableNumber := c.popUint16()
	variableReference := FunctionVariableReference{
		variableStartIndex: c.wipMacro.variableStartIndexes[variableNumber],
		variableNumber:     variableNumber,
		bytecodeIndex:      len(c.wipMacro.bytecode.bytes),
	}
	c.wipMacro.variableReferences = append(c.wipMacro.variableReferences, variableReference)
	c.wipMacro.bytecode.pushUint8(0)
}

func (c *Client) macroDefUseVar16() {
	if c.wipMacro == nil {
		c.error("macroDefUseVar16: nil wip function")
	}
	variableNumber := c.popUint16()
	variableReference := FunctionVariableReference{
		variableStartIndex: c.wipMacro.variableStartIndexes[variableNumber],
		variableNumber:     variableNumber,
		bytecodeIndex:      len(c.wipMacro.bytecode.bytes),
	}
	c.wipMacro.variableReferences = append(c.wipMacro.variableReferences, variableReference)
	c.wipMacro.bytecode.pushUint16(0)
}

func (c *Client) macroDefUseConst8() {
	if c.wipMacro == nil {
		c.error("macroDefUseConst8: nil wip function")
	}
	const8 := c.popUint8()
	c.wipMacro.bytecode.pushUint8(const8)
}

func (c *Client) macroDefUseConst16() {
	if c.wipMacro == nil {
		c.error("macroDefUseConst16: nil wip function")
	}
	const16 := c.popUint16()
	c.wipMacro.bytecode.pushUint16(const16)
}

func (c *Client) nodeCreate() {
	nodeNumber := c.popUint16()
	newNode := NewNode()
	if _, ok := c.nodes[nodeNumber]; ok {
		c.error("nodeCreate: a node with nodeNumber already exists")
	}
	c.nodes[nodeNumber] = newNode
	c.root.AddChild(newNode)
}

func (c *Client) nodeSetContent() {
	node := c.popNode()
	macroBytecode := c.popAndCompileMacro()
	if node == nil || macroBytecode == nil {
		return
	}
	node.renderCode = macroBytecode
}

func (c *Client) nodeSetParent() {
	node := c.popNode()
	if node == nil {
		return
	}
	parentNodeNumber := c.popUint16()
	parentNode, ok := c.nodes[parentNodeNumber]
	if !ok {
		c.error("nodeSetParent: invalid parentNodeNumber")
	}
	parentNode.AddChild(node)
}

func (c *Client) nodeSetPosition() {
	node := c.popNode()
	if node == nil {
		return
	}
	newPosition := c.popVec2()
	node.position = newPosition
}

func (c *Client) nodeSetRotation() {
	node := c.popNode()
	if node == nil {
		return
	}
	rotation := c.popRotation()
	node.rotation = rotation
}

func (c *Client) nodeSetScale() {
	node := c.popNode()
	if node == nil {
		return
	}
	scale := c.popScale()
	node.scale = scale
}

// ------------------------- RENDER OPERATIONS --------------------------------
// ------------------------- RENDER OPERATIONS --------------------------------
// ------------------------- RENDER OPERATIONS --------------------------------
// ------------------------- RENDER OPERATIONS --------------------------------
// ------------------------- RENDER OPERATIONS --------------------------------
// ------------------------- RENDER OPERATIONS --------------------------------
// ------------------------- RENDER OPERATIONS --------------------------------
// ------------------------- RENDER OPERATIONS --------------------------------

func (c *Client) beginPath(node *Node) {
	c.nvgCtx.BeginPath()
}

func (c *Client) setFillColor(node *Node) {
	color := c.popRgba()
	c.nvgCtx.SetFillColor(color)
}

func (c *Client) fill(node *Node) {
	c.nvgCtx.Fill()
}

func (c *Client) rectangle(node *Node) {
	rect := c.popRect()
	corners := rect.GetCorners()
	globalCorners := node.TransformPoints(corners[:])
	c.nvgCtx.MoveTo(float32(globalCorners[0].X), float32(globalCorners[0].Y))
	c.nvgCtx.LineTo(float32(globalCorners[1].X), float32(globalCorners[1].Y))
	c.nvgCtx.LineTo(float32(globalCorners[2].X), float32(globalCorners[2].Y))
	c.nvgCtx.LineTo(float32(globalCorners[3].X), float32(globalCorners[3].Y))
	c.nvgCtx.ClosePath()
}

func (c *Client) macroCall(node *Node) {
	macroBytecode := c.popAndCompileMacro()
	c.pushState(macroBytecode)
}
