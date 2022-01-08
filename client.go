package main

import (
	"errors"
	"fmt"
	"math"

	"github.com/shibukawa/nanovgo"
)

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

func (f *Macro) Compile(functionNumber uint16, variables []byte) *Bytecode {
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
	updateOperations [256]func() error
	renderOperations [256]func(*Node) error
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
	client.updateOperations = [256]func() error{
		client.macroDefStart, client.macroDefEnd, client.macroDefOperation, client.macroDefVar8, client.macroDefVar16,
		client.macroDefUseVar8, client.macroDefUseVar16, client.macroDefUseConst8, client.macroDefUseConst16,

		client.nodeCreate, client.nodeSetContent, client.nodeSetParent, client.nodeSetPosition, client.nodeSetRotation, client.nodeSetScale,
	}
	client.renderOperations = [256]func(*Node) error{
		client.beginPath, client.setFillColor, client.fill, client.rectangle,
		client.macroCall,
	}
	return &client
}

func (c *Client) Update(bytecode *Bytecode) error {
	fmt.Println("update bytes: ", bytecode.bytes)
	c.Bytecode = bytecode
	for {
		for c.i < len(c.bytes) {
			if err := c.updateStep(); err != nil {
				return errors.New("error while excecuting instruction, i: " + fmt.Sprint(c.i) + " error: " + err.Error())
			}
		}
		if !c.popState() {
			return nil
		}
	}
}

func (c *Client) updateStep() error {
	opcode, err := c.popUint8()
	if err != nil {
		return err
	}
	fmt.Println("i: ", c.i-1, " opcode: ", updateOpcodeNames[opcode])
	if opcode > uopCodeCount {
		return errors.New("invalid update opcode: " + fmt.Sprint(opcode))
	}
	err = c.updateOperations[opcode]()
	return err
}

func (c *Client) Render() error {
	fmt.Println("Render start")
	if err := c.renderNode(c.root); err != nil {
		return err
	}
	return nil
}

func (c *Client) renderNode(node *Node) error {
	if node == nil {
		return nil
	}
	transformationMatrix := BuildTransformationMatrix(node.position, node.rotation, node.scale)
	fmt.Println("transformation matrix: ", transformationMatrix)
	if node.parent != nil {
		node.localToGlobal = node.parent.localToGlobal.MultiplyMatrix(transformationMatrix)
	} else {
		node.localToGlobal = transformationMatrix
	}
	if node.renderCode != nil {
		if err := c.pushState(node.renderCode); err != nil {
			return err
		}
		for c.i < len(c.bytes) {
			if err := c.renderStep(node); err != nil {
				return err
			}
		}
	}
	for child := range node.children {
		if err := c.renderNode(child); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) renderStep(node *Node) error {
	opcode, err := c.popUint8()
	if err != nil {
		return err
	}
	fmt.Println("i: ", c.i-1, " opcode: ", renderOpcodeName[opcode])
	if opcode > ropCodeCount {
		return errors.New("invalid render opcode: " + fmt.Sprint(opcode))
	}
	if err = c.renderOperations[opcode](node); err != nil {
		return err
	}
	return nil
}

func (c *Client) pushState(bytecode *Bytecode) error {
	fmt.Println("push state, c.Bytecode == nil: ", c.Bytecode == nil)
	if bytecode == nil {
		return errors.New("pushState: nil bytecode")
	}
	if c.Bytecode != nil {
		c.stack = append(c.stack, c.Bytecode)
	}
	bytecode.i = 0
	c.Bytecode = bytecode
	return nil
}

func (c *Client) popState() bool {
	fmt.Println("pop state len(c.stack): ", len(c.stack))
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

func (c *Client) popAndCompileMacro() (*Bytecode, error) {
	macroNumber, err := c.popUint16()
	if err != nil {
		return nil, err
	}
	macro, ok := c.macros[macroNumber]
	if !ok {
		return nil, errors.New("popAndCompileMacro: invalid macroNumber: " + fmt.Sprint(macroNumber))
	}
	if c.i+macro.totalVariablesSize > len(c.bytes) {
		return nil, errors.New("popAndCompileMacro: macro variable block out of range")
	}
	variables := c.bytes[c.i : c.i+macro.totalVariablesSize]
	c.i += len(variables)
	return macro.Compile(macroNumber, variables), nil
}

func (c *Client) popNode() (*Node, error) {
	nodeNumber, err := c.popUint16()
	if err != nil {
		return nil, err
	}
	node, ok := c.nodes[nodeNumber]
	if !ok {
		return nil, errors.New("nodeSetPosition: invalid nodeNumber")
	}
	return node, nil
}

// ---------------------UPDATE OPERATIONS---------------------
// ---------------------UPDATE OPERATIONS---------------------
// ---------------------UPDATE OPERATIONS---------------------
// ---------------------UPDATE OPERATIONS---------------------
// ---------------------UPDATE OPERATIONS---------------------
// ---------------------UPDATE OPERATIONS---------------------
// ---------------------UPDATE OPERATIONS---------------------
// ---------------------UPDATE OPERATIONS---------------------

func (c *Client) macroDefStart() error {
	if c.wipMacro != nil {
		return errors.New("funDefStart: wip function already exists")
	}
	functionNumber, err := c.popUint16()
	if err != nil {
		return err
	}
	newWipMacro := NewMacro()
	c.wipMacro = newWipMacro
	c.wipMacroNumber = functionNumber
	fmt.Println("funDefStart functionNumber: ", functionNumber)
	return nil
}

func (c *Client) macroDefEnd() error {
	if c.wipMacro == nil {
		return errors.New("funDefEnd: nil wip function")
	}
	c.macros[c.wipMacroNumber] = c.wipMacro
	c.wipMacro = nil
	return nil
}

func (c *Client) macroDefOperation() error {
	if c.wipMacro == nil {
		return errors.New("funOperation: nil wipFunction")
	}
	opcode, err := c.popUint8()
	if err != nil {
		return err
	}
	fmt.Println("operation: ", renderOpcodeName[opcode])
	c.wipMacro.bytecode.pushUint8(opcode)
	return nil
}

func (c *Client) macroDefVar8() error {
	if c.wipMacro == nil {
		return errors.New("funVarDef8: nil wipFunction")
	}
	c.wipMacro.variableSizes = append(c.wipMacro.variableSizes, 1)
	c.wipMacro.variableStartIndexes = append(c.wipMacro.variableStartIndexes, c.wipMacro.totalVariablesSize)
	c.wipMacro.totalVariablesSize += 1
	return nil
}

func (c *Client) macroDefVar16() error {
	if c.wipMacro == nil {
		return errors.New("funVarDef16: nil wipFunction")
	}
	c.wipMacro.variableSizes = append(c.wipMacro.variableSizes, 2)
	c.wipMacro.variableStartIndexes = append(c.wipMacro.variableStartIndexes, c.wipMacro.totalVariablesSize)
	c.wipMacro.totalVariablesSize += 2
	return nil
}

func (c *Client) macroDefUseVar8() error {
	if c.wipMacro == nil {
		return errors.New("pushFunVar8: nil wipFunction")
	}
	variableNumber, err := c.popUint16()
	if err != nil {
		return err
	}
	variableReference := FunctionVariableReference{
		variableStartIndex: c.wipMacro.variableStartIndexes[variableNumber],
		variableNumber:     variableNumber,
		bytecodeIndex:      len(c.wipMacro.bytecode.bytes),
	}
	c.wipMacro.variableReferences = append(c.wipMacro.variableReferences, variableReference)
	c.wipMacro.bytecode.pushUint8(0)
	return nil
}

func (c *Client) macroDefUseVar16() error {
	if c.wipMacro == nil {
		return errors.New("pushFunVar16: nil wipFunction")
	}
	variableNumber, err := c.popUint16()
	if err != nil {
		return err
	}
	variableReference := FunctionVariableReference{
		variableStartIndex: c.wipMacro.variableStartIndexes[variableNumber],
		variableNumber:     variableNumber,
		bytecodeIndex:      len(c.wipMacro.bytecode.bytes),
	}
	c.wipMacro.variableReferences = append(c.wipMacro.variableReferences, variableReference)
	c.wipMacro.bytecode.pushUint16(0)
	return nil
}

func (c *Client) macroDefUseConst8() error {
	if c.wipMacro == nil {
		return errors.New("pushFunConst8: nil wipFunction")
	}
	const8, err := c.popUint8()
	if err != nil {
		return err
	}
	c.wipMacro.bytecode.pushUint8(const8)
	return nil
}

func (c *Client) macroDefUseConst16() error {
	if c.wipMacro == nil {
		return errors.New("pushFunConst16: nil wipFunction")
	}
	const16, err := c.popUint16()
	if err != nil {
		return err
	}
	c.wipMacro.bytecode.pushUint16(const16)
	return nil
}

func (c *Client) nodeCreate() error {
	nodeNumber, err := c.popUint16()
	if err != nil {
		return err
	}
	newNode := NewNode()
	if _, ok := c.nodes[nodeNumber]; ok {
		return errors.New("nodeCreate: a node with nodeNumber already exists")
	}
	c.nodes[nodeNumber] = newNode
	c.root.AddChild(newNode)
	return nil
}

func (c *Client) nodeSetContent() error {
	node, err := c.popNode()
	if err != nil {
		return err
	}
	macroBytecode, err := c.popAndCompileMacro()
	if err != nil {
		return err
	}
	node.renderCode = macroBytecode
	return nil
}

func (c *Client) nodeSetParent() error {
	node, err := c.popNode()
	if err != nil {
		return err
	}
	parentNodeNumber, err := c.popUint16()
	if err != nil {
		return err
	}
	parentNode, ok := c.nodes[parentNodeNumber]
	if !ok {
		return errors.New("nodeSetParent: invalid parentNodeNumber")
	}
	parentNode.AddChild(node)
	return nil
}

func (c *Client) nodeSetPosition() error {
	node, err := c.popNode()
	if err != nil {
		return err
	}
	newPosition, err := c.popVec2()
	if err != nil {
		return err
	}
	node.position = newPosition
	return nil
}

func (c *Client) nodeSetRotation() error {
	node, err := c.popNode()
	if err != nil {
		return err
	}
	rotation, err := c.popRotation()
	if err != nil {
		return err
	}
	node.rotation = rotation
	return nil
}

func (c *Client) nodeSetScale() error {
	node, err := c.popNode()
	if err != nil {
		return err
	}
	scale, err := c.popScale()
	if err != nil {
		return err
	}
	node.scale = scale
	return nil
}

// ------------------------- RENDER OPERATIONS --------------------------------
// ------------------------- RENDER OPERATIONS --------------------------------
// ------------------------- RENDER OPERATIONS --------------------------------
// ------------------------- RENDER OPERATIONS --------------------------------
// ------------------------- RENDER OPERATIONS --------------------------------
// ------------------------- RENDER OPERATIONS --------------------------------
// ------------------------- RENDER OPERATIONS --------------------------------
// ------------------------- RENDER OPERATIONS --------------------------------

func (c *Client) beginPath(node *Node) error {
	c.nvgCtx.BeginPath()
	return nil
}

func (c *Client) setFillColor(node *Node) error {
	color, err := c.popRgba()
	if err != nil {
		return err
	}
	c.nvgCtx.SetFillColor(color)
	return nil
}

func (c *Client) fill(node *Node) error {
	c.nvgCtx.Fill()
	return nil
}

func (c *Client) rectangle(node *Node) error {
	rect, err := c.popRect()
	if err != nil {
		return err
	}
	corners := rect.GetCorners()
	globalCorners := node.TransformPoints(corners[:])
	c.nvgCtx.MoveTo(float32(globalCorners[0].X), float32(globalCorners[0].Y))
	c.nvgCtx.LineTo(float32(globalCorners[1].X), float32(globalCorners[1].Y))
	c.nvgCtx.LineTo(float32(globalCorners[2].X), float32(globalCorners[2].Y))
	c.nvgCtx.LineTo(float32(globalCorners[3].X), float32(globalCorners[3].Y))
	c.nvgCtx.ClosePath()
	return nil
}

func (c *Client) macroCall(node *Node) error {
	fmt.Println("macroCall")
	macroBytecode, err := c.popAndCompileMacro()
	if err != nil {
		return err
	}
	c.pushState(macroBytecode)
	return nil
}
