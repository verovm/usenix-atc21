package vm

import (
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"github.com/holiman/uint256"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"

	"os"
)

var LiveCount uint64
var TotalCount uint64

var LogChan = make(chan CallLog)
var Done = make(chan struct{})

type CallInfo struct {
	block   int64
	txIndex int
	depth   int
	caller  string
	address string
}
type CallLog struct {
	callInfo  CallInfo
	totalInst uint64
	liveInst  uint64
	totalGas  uint64
	liveGas   uint64
}

func PassLog(log CallLog) {
	LogChan <- log
}
func WriteLog(fn string, LogChan <-chan CallLog, Done <-chan struct{}) {
	if fn == "" {
		return
	}
	f, err := os.Create(fn)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.WriteString("block,totalInst,liveInst,totalGas,liveGas,txIndex,depth,caller,self\n")
	for {
		select {
		case <-Done:
			break
		default:
			log := <-LogChan
			f.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%s,%s\n", log.callInfo.block, log.totalInst, log.liveInst, log.totalGas, log.liveGas,
				log.callInfo.txIndex, log.callInfo.depth, log.callInfo.caller, log.callInfo.address))
		}
	}
}

type JumpStack struct {
	nodes   []*Node
	jumpPos []uint64
}

func newJumpStack() *JumpStack {
	return &JumpStack{nodes: make([]*Node, 0, 100), jumpPos: make([]uint64, 0, 100)}
}

func (js *JumpStack) push(nd *Node, pos uint256.Int) {
	js.nodes = append(js.nodes, nd)
	js.jumpPos = append(js.jumpPos, pos.Uint64())
}
func (js *JumpStack) pop() (*Node, uint64) {
	jnode := js.nodes[len(js.nodes)-1]
	jdest := js.jumpPos[len(js.jumpPos)-1]
	js.nodes = js.nodes[:len(js.nodes)-1]
	js.jumpPos = js.jumpPos[:len(js.jumpPos)-1]
	return jnode, jdest
}
func (js *JumpStack) len() int { return len(js.nodes) }

func newMemoryTracer() *MemoryTracer {
	return &MemoryTracer{memory: make([]*Node, 0, 5), offset: make([]int64, 0, 5), length: make([]int64, 0, 5)}
}

type MemoryTracer struct {
	memory []*Node
	offset []int64
	length []int64
}

func (rec *MemoryTracer) store(nd *Node, off int64, l int64) {
	rec.memory = append(rec.memory, nd)
	rec.offset = append(rec.offset, off)
	rec.length = append(rec.length, l)
}

func (rec *MemoryTracer) loadRecent(off int64, l int64, index int) (*Node, int64, int64, int) {
	for i := index; i >= 0; i-- {
		if rec.offset[i] < off+l && off < rec.offset[i]+rec.length[i] {
			return rec.memory[i], rec.offset[i], rec.length[i], i
		}
	}
	return nil, 0, 0, 0
}

func (rec *MemoryTracer) recursiveLoad(nodes *[]*Node, off int64, l int64, index int) {
	if l == 0 {
		return
	}

	node, start, size, index := rec.loadRecent(off, l, index-1)
	end := start + size
	if node != nil {
		*nodes = append(*nodes, node)
		if (start - off) > 0 {
			rec.recursiveLoad(nodes, off, start-off, index-1)
		}
		if (end - (off + l)) < 0 {
			rec.recursiveLoad(nodes, end, (off+l)-end, index-1)
		}
	}
}

func (rec *MemoryTracer) load(off int64, l int64) []*Node {
	var nodes []*Node
	rec.recursiveLoad(&nodes, off, l, len(rec.memory))

	return nodes

}

var liveOps = []string{"SSTORE", "SELFDESTRUCT", "CREATE", "CREATE2", "RETURN", "REVERT",
	"LOG0", "LOG1", "LOG2", "LOG3", "LOG4"}
var callOps = []string{"CALL", "CALLCODE", "DELEGATECALL"}
var jumpOps = []string{"JUMP", "JUMPI", "JUMPDEST"}

func isLive(op string) bool {
	for _, item := range liveOps {
		if op == item {
			return true
		}
	}
	for _, item := range callOps {
		if op == item {
			return true
		}
	}
	return false
}
func isCall(op string) bool {
	for _, item := range callOps {
		if op == item {
			return true
		}
	}
	return false
}
func isJump(op string) bool {
	for _, item := range jumpOps {
		if op == item {
			return true
		}
	}
	return false
}

type Node struct {
	id     int64
	opcode string
	pc     uint64
	gas    uint64
	islive bool
	value  *ValueNode
}
type ValueNode struct {
	value string
}

func (nd Node) ID() int64   { return nd.id }
func (nd Node) op() string  { return nd.opcode }
func (nd Node) live() bool  { return nd.islive }
func (nd *Node) setLive()   { nd.islive = true }
func (nd Node) isHex() bool { return strings.Contains(nd.opcode, "0x") }

type NodeStack struct {
	graph    *simple.DirectedGraph
	nodes    []*Node
	id       int64
	log      CallLog
	stopFlag bool //////////
	g        *graphviz.Graphviz
}

func newNodeStack(ci CallInfo) *NodeStack {
	return &NodeStack{log: CallLog{callInfo: ci}}
}

func (st *NodeStack) newGraph(draw bool) {
	st.graph = simple.NewDirectedGraph()
	st.nodes = make([]*Node, 0, 100)
	st.id = 0
	st.stopFlag = false
	if draw {
		st.g = graphviz.New()
	}
}

func (st *NodeStack) GetLog() *CallLog { return &st.log }

//Mark x as live and continue with its predecessor
func (st *NodeStack) markPredecessor(x *Node) {
	if !x.live() {
		x.setLive()
		if !x.isHex() {
			st.log.liveInst++
			st.log.liveGas += x.gas
			atomic.AddUint64(&LiveCount, 1)
		}

		//check if x is used by SWAPn
		toNodes := st.graph.From(x.ID())
		for toNodes.Next() {
			to := toNodes.Node().(*Node)
			if !to.live() && strings.Contains(to.op(), "SWAP") {
				to.setLive()
				st.log.liveInst++
				st.log.liveGas += to.gas
				atomic.AddUint64(&LiveCount, 1)
			}
		}

		fromNodes := st.graph.To(x.ID())
		for fromNodes.Next() {
			st.markPredecessor(fromNodes.Node().(*Node))
		}
	}
}

func (st *NodeStack) createCallEdge(from, to *Node, isInput bool) graph.Edge {
	e := st.graph.NewEdge(from, to)
	st.graph.SetEdge(e)
	//if isInput {
	st.markPredecessor(from)
	//}
	return e
}

func (st *NodeStack) createEdge(from, to *Node) graph.Edge {
	e := st.graph.NewEdge(from, to)
	st.graph.SetEdge(e)
	if to.live() {
		st.markPredecessor(from)
	}
	return e
}

func (st *NodeStack) createValue(value string) *ValueNode {
	nd := ValueNode{value}
	return &nd
}

func (st *NodeStack) createNode(op string, pc uint64, cost uint64, vn *ValueNode) *Node {
	nd := Node{st.id, op, pc, cost, isLive(op), vn}
	st.log.totalInst++
	st.log.totalGas += nd.gas
	atomic.AddUint64(&TotalCount, 1)
	if nd.live() {
		st.log.liveInst++
		st.log.liveGas += nd.gas
		atomic.AddUint64(&LiveCount, 1)
	}
	st.id++
	st.graph.AddNode(&nd)
	return &nd
}

func (st *NodeStack) createValueOp(opCode OpCode, pc uint64, cost uint64, vn *ValueNode) error {
	opNd := st.createNode(opCode.String(), pc, cost, vn)
	if st.len() > 0 {
		prev := st.peek()
		if isJump(prev.op()) {
			st.pop()
			st.createEdge(prev, opNd)
		}
	}
	st.push(opNd)

	return nil
}
func (st *NodeStack) createOpComponent(popSize int, opCode OpCode, pc uint64, cost uint64) error {
	opNd := st.createNode(opCode.String(), pc, cost, nil)
	if st.len() > popSize {
		prev := st.peek()
		if isJump(prev.op()) {
			st.pop()
			st.createEdge(prev, opNd)
		}
	}
	for i := 0; i < popSize; i++ {
		nd := st.pop()
		if isCall(opNd.op()) {
			st.createCallEdge(nd, opNd, (i+2 < popSize))
		} else {
			st.createEdge(nd, opNd)
		}
	}
	st.push(opNd)

	return nil
}

func (st *NodeStack) peek() *Node {
	return st.nodes[st.len()-1]
}

func (st *NodeStack) push(nd *Node) {
	st.nodes = append(st.nodes, nd)
}

func (st *NodeStack) pushN(nds ...*Node) {
	st.nodes = append(st.nodes, nds...)
}

func (st *NodeStack) pop() *Node {
	ret := st.nodes[len(st.nodes)-1]
	st.nodes = st.nodes[:len(st.nodes)-1]
	return ret
}

func (st *NodeStack) countLive() (uint64, uint64) {
	return st.log.totalInst, st.log.liveInst
}

func (st *NodeStack) len() int {
	return len(st.nodes)
}

func (st *NodeStack) swap(n int, pc uint64, cost uint64) {
	opNd := st.createNode(OpCode(byte(SWAP1)+byte(n-1)).String(), pc, cost, nil)
	if st.len() > 0 {
		prev := st.peek()
		if isJump(prev.op()) {
			st.pop()
			st.createEdge(prev, opNd)
		}
	}
	st.createEdge(st.nodes[st.len()-n], opNd)
	st.createEdge(st.nodes[st.len()-1], opNd)
	st.nodes[st.len()-n], st.nodes[st.len()-1] = st.nodes[st.len()-1], st.nodes[st.len()-n]

	if st.nodes[st.len()-n].live() || st.nodes[st.len()-1].live() {
		opNd.setLive()
		st.log.liveInst++
		st.log.liveGas += opNd.gas
		atomic.AddUint64(&LiveCount, 1)
	}
}

func (st *NodeStack) dup(n int, pc uint64, cost uint64) {
	opNd := st.createNode(OpCode(byte(DUP1)+byte(n-1)).String(), pc, cost, nil)
	if st.len() > 0 {
		prev := st.peek()
		if isJump(prev.op()) {
			st.pop()
			st.createEdge(prev, opNd)
		}
	}
	st.createEdge(st.nodes[st.len()-n], opNd)
	st.push(opNd)
}

func (nd Node) makeLabel() string {
	var str string
	if nd.op() == "PUSH1" {
		str = fmt.Sprintf("%d.%s", nd.ID(), "PUSH")
	} else {
		str = fmt.Sprintf("%d.%s", nd.ID(), nd.op())
	}
	return str
}

var TxIndex int
var BlockNum uint64

func (st *NodeStack) GenerateGraph(path string) error {
	i, j := st.countLive()
	fmt.Printf("Generated a graph for tx #%d in block #%d\n", TxIndex, BlockNum)
	fmt.Printf("Total node: %d, Live node: %d\n", i, j)

	vgraph, _ := st.g.Graph(graphviz.Directed)
	vnodes := make([]*cgraph.Node, 0, 2048)

	//Copy nodes
	nodes := st.graph.Nodes()
	for nodes.Next() {
		n := nodes.Node().(*Node)
		vnode, _ := vgraph.CreateNode(string(n.ID()))
		vnode.SetLabel(n.makeLabel())
		if n.live() && !n.isHex() {
			if n.op() == "RETURN" || n.op() == "REVERT" {
				vnode.SetColor("blue")
			} else {
				vnode.SetColor("blue")
			}
		}
		vnodes = append(vnodes, vnode)
	}

	//Copy edges
	edges := st.graph.Edges()
	for edges.Next() {
		e := edges.Edge()
		from, to := e.From().(*Node), e.To().(*Node)
		fstr, tstr := from.makeLabel(), to.makeLabel()

		var vf, vt *cgraph.Node
		for _, v := range vnodes {
			if fstr == v.Get("label") {
				vf = v
			} else if tstr == v.Get("label") {
				vt = v
			}
		}

		if vf == nil || vt == nil {
			fmt.Println("NIL")
			fmt.Println(fstr, tstr)
		}
		vgraph.CreateEdge("", vf, vt)
	}

	return st.g.RenderFilename(vgraph, graphviz.PNG, path)
}
