package exelayout

import "golang.org/x/arch/x86/x86asm"

// rowMarker is the 0xBBBBBBBBBBBBBBBB that separates a .dat's rows from its variable data.
const rowMarker = int64(-0x4444444444444445)

// volatile are the registers a Windows x64 call clobbers.
var volatile = []x86asm.Reg{x86asm.RAX, x86asm.RCX, x86asm.RDX, x86asm.R8, x86asm.R9, x86asm.R10, x86asm.R11}

// full maps a sub-register to its 64-bit register.
func full(r x86asm.Reg) x86asm.Reg {
	switch {
	case r >= x86asm.AH && r <= x86asm.BH:
		return x86asm.RAX + (r - x86asm.AH)
	case r >= x86asm.AL && r <= x86asm.R15B:
		i := r - x86asm.AL
		// SPB..DIB and R8B.. come after AH..BH
		if i >= 8 {
			i -= 4
		}
		return x86asm.RAX + i
	case r >= x86asm.AX && r <= x86asm.R15W:
		return x86asm.RAX + (r - x86asm.AX)
	case r >= x86asm.EAX && r <= x86asm.R15L:
		return x86asm.RAX + (r - x86asm.EAX)
	}
	return r
}

func reg(in ins, n int) (x86asm.Reg, bool) {
	if in.bad {
		return 0, false
	}
	r, ok := in.Args[n].(x86asm.Reg)
	if !ok {
		return 0, false
	}
	return full(r), true
}

func imm(in ins, n int) (int64, bool) {
	if in.bad {
		return 0, false
	}
	v, ok := in.Args[n].(x86asm.Imm)
	return int64(v), ok
}

// mem returns the instruction's memory operand, with full-width registers, and its argument index.
func mem(in ins) (x86asm.Mem, int, bool) {
	if in.bad {
		return x86asm.Mem{}, -1, false
	}
	for i, a := range in.Args {
		if a == nil {
			break
		}
		if m, ok := a.(x86asm.Mem); ok {
			m.Base, m.Index = full(m.Base), full(m.Index)
			return m, i, true
		}
	}
	return x86asm.Mem{}, -1, false
}

func isStack(r x86asm.Reg) bool { return r == x86asm.RSP || r == x86asm.RBP }

var noWrite = map[x86asm.Op]bool{
	x86asm.CMP: true, x86asm.TEST: true, x86asm.PUSH: true, x86asm.BT: true, x86asm.NOP: true,
	x86asm.UCOMISS: true, x86asm.UCOMISD: true, x86asm.COMISS: true, x86asm.COMISD: true, x86asm.PTEST: true,
}

// written approximates the registers an instruction writes: arg0 when it is a register, plus common implicit writes.
func written(in ins) []x86asm.Reg {
	if in.bad {
		return nil
	}
	var out []x86asm.Reg
	if r, ok := reg(in, 0); ok && !noWrite[in.Op] && !isJump(in.Op) {
		out = append(out, r)
	}
	switch in.Op {
	case x86asm.XCHG, x86asm.XADD:
		if r, ok := reg(in, 1); ok {
			out = append(out, r)
		}
	case x86asm.MUL, x86asm.DIV, x86asm.IDIV, x86asm.CDQ, x86asm.CQO, x86asm.CWD, x86asm.RDTSC, x86asm.CPUID:
		out = append(out, x86asm.RAX, x86asm.RDX, x86asm.RCX, x86asm.RBX)
	case x86asm.IMUL:
		if in.Args[1] == nil {
			out = append(out, x86asm.RAX, x86asm.RDX)
		}
	case x86asm.CDQE, x86asm.CWDE, x86asm.LAHF:
		out = append(out, x86asm.RAX)
	case x86asm.MOVSB, x86asm.MOVSW, x86asm.MOVSD, x86asm.MOVSQ, x86asm.STOSB, x86asm.STOSW, x86asm.STOSD, x86asm.STOSQ:
		out = append(out, x86asm.RCX, x86asm.RSI, x86asm.RDI)
	}
	return out
}

func isJump(op x86asm.Op) bool {
	switch op {
	case x86asm.JMP, x86asm.JA, x86asm.JAE, x86asm.JB, x86asm.JBE, x86asm.JE, x86asm.JNE, x86asm.JG, x86asm.JGE,
		x86asm.JL, x86asm.JLE, x86asm.JS, x86asm.JNS, x86asm.JP, x86asm.JNP, x86asm.JO, x86asm.JNO, x86asm.CALL, x86asm.RET:
		return true
	}
	return false
}

type access struct {
	m             x86asm.Mem
	width         int64
	reads, writes bool
}

// memAccess approximates an instruction's explicit memory access: its operand, width and direction.
func memAccess(in ins) (access, bool) {
	m, idx, ok := mem(in)
	if !ok || in.Op == x86asm.LEA || in.Op == x86asm.NOP || in.Op == x86asm.PREFETCHT0 || in.Op == x86asm.PREFETCHNTA {
		return access{}, false
	}
	a := access{m: m, width: int64(in.MemBytes)}
	if idx != 0 {
		a.reads = true
		return a, true
	}
	switch in.Op {
	case x86asm.CMP, x86asm.TEST, x86asm.PUSH, x86asm.CALL, x86asm.JMP, x86asm.BT:
		a.reads = true
	case x86asm.MOV, x86asm.MOVUPS, x86asm.MOVAPS, x86asm.MOVDQU, x86asm.MOVDQA, x86asm.MOVQ, x86asm.MOVD, x86asm.MOVSD_XMM,
		x86asm.MOVSS, x86asm.MOVUPD, x86asm.MOVAPD, x86asm.MOVLPS, x86asm.MOVHPS, x86asm.MOVNTI, x86asm.VMOVDQU, x86asm.VMOVDQA, x86asm.POP:
		a.writes = true
	default:
		a.reads, a.writes = true, true
	}
	return a, true
}

// sum is the symbolic value rows*a + b.
type sum struct{ a, b int64 }

// sums tracks which registers and stack slots hold a known multiple of the row count.
type sums struct {
	regs  map[x86asm.Reg]sum
	slots map[[2]int64]sum
}

func newSums() *sums { return &sums{map[x86asm.Reg]sum{}, map[[2]int64]sum{}} }

func (s *sums) set(r x86asm.Reg, v sum, ok bool) {
	if ok {
		s.regs[r] = v
	} else {
		delete(s.regs, r)
	}
}

func stackSlot(in ins) ([2]int64, bool) {
	m, _, ok := mem(in)
	if !ok || !isStack(m.Base) || m.Index != 0 {
		return [2]int64{}, false
	}
	return [2]int64{int64(m.Base), m.Disp}, true
}

func (s *sums) known(in ins, n int) (sum, bool) {
	r, ok := reg(in, n)
	if !ok {
		return sum{}, false
	}
	v, ok := s.regs[r]
	return v, ok
}

// step applies in to the tracked values. A 32-bit load from a non-stack pointer is taken to be the row count.
func (s *sums) step(in ins) {
	if in.bad {
		return
	}
	dst, hasDst := reg(in, 0)
	switch {
	case (in.Op == x86asm.MOV || in.Op == x86asm.MOVZX || in.Op == x86asm.MOVSXD || in.Op == x86asm.MOVSX) && hasDst:
		var v sum
		ok := false
		switch a := in.Args[1].(type) {
		case x86asm.Reg:
			v, ok = s.known(in, 1)
		case x86asm.Mem:
			if slot, isSlot := stackSlot(in); isSlot {
				v, ok = s.slots[slot]
			} else {
				b := full(a.Base)
				if in.Op == x86asm.MOV && in.MemBytes == 4 && a.Index == 0 && !isStack(b) && b != 0 && b != x86asm.RIP {
					v, ok = sum{1, 0}, true
				}
			}
		case x86asm.Imm:
			v, ok = sum{0, int64(a)}, true
		}
		s.set(dst, v, ok)
	case in.Op == x86asm.MOV && !hasDst:
		if slot, ok := stackSlot(in); ok {
			if v, ok := s.known(in, 1); ok {
				s.slots[slot] = v
			} else {
				delete(s.slots, slot)
			}
		}
	case in.Op == x86asm.SHL && hasDst:
		v, ok1 := s.regs[dst]
		n, ok2 := imm(in, 1)
		s.set(dst, sum{v.a << n, v.b << n}, ok1 && ok2)
	case in.Op == x86asm.IMUL && hasDst && in.Args[2] != nil:
		v, ok1 := s.known(in, 1)
		n, ok2 := imm(in, 2)
		s.set(dst, sum{v.a * n, v.b * n}, ok1 && ok2)
	case in.Op == x86asm.LEA && hasDst:
		m, _, _ := mem(in)
		ok := m.Base != x86asm.RIP
		v := sum{0, m.Disp}
		for _, p := range []struct {
			r x86asm.Reg
			k int64
		}{{m.Base, 1}, {m.Index, int64(m.Scale)}} {
			if p.r != 0 && ok {
				q, has := s.regs[p.r]
				ok = has
				v = sum{v.a + q.a*p.k, v.b + q.b*p.k}
			}
		}
		s.set(dst, v, ok)
	case (in.Op == x86asm.ADD || in.Op == x86asm.SUB) && hasDst:
		sign := int64(1)
		if in.Op == x86asm.SUB {
			sign = -1
		}
		var by sum
		ok := false
		if n, isImm := imm(in, 1); isImm {
			by, ok = sum{0, n}, true
		} else {
			by, ok = s.known(in, 1)
		}
		v, has := s.regs[dst]
		s.set(dst, sum{v.a + sign*by.a, v.b + sign*by.b}, ok && has)
	case in.Op == x86asm.CALL:
		for _, r := range volatile {
			delete(s.regs, r)
		}
	default:
		for _, r := range written(in) {
			delete(s.regs, r)
		}
	}
}

// loaderPattern finds a table loader's check that the row count times the row size lands on the 0xBB marker.
// It returns that row size and the function called with it in r8, which fixes up the rows' pointers.
func loaderPattern(insns []ins) (int, uint32, bool) {
	for i, in := range insns {
		if in.bad || in.Op != x86asm.MOV {
			continue
		}
		if v, ok := imm(in, 1); !ok || v != rowMarker {
			continue
		}
		marker, _ := reg(in, 0)
		check := -1
		for j := i + 1; j < min(len(insns), i+6); j++ {
			c := insns[j]
			if c.bad || c.Op != x86asm.CMP {
				continue
			}
			if _, isMem := c.Args[0].(x86asm.Mem); !isMem {
				continue
			}
			if r, ok := reg(c, 1); ok && r == marker {
				check = j
				break
			}
		}
		if check < 0 {
			continue
		}
		s := newSums()
		for _, p := range insns[max(0, i-60):check] {
			s.step(p)
		}
		m, _, _ := mem(insns[check])
		var sizes []int64
		for _, p := range []struct {
			r x86asm.Reg
			k int64
		}{{m.Base, 1}, {m.Index, int64(m.Scale)}} {
			if v, ok := s.regs[p.r]; ok && p.r != 0 && v.a != 0 {
				sizes = append(sizes, v.a*p.k)
			}
		}
		if len(sizes) != 1 || sizes[0] <= 0 {
			continue
		}
		size := sizes[0]
		var fixup uint32
		for _, p := range insns[check:] {
			if !p.bad && p.Op == x86asm.CALL {
				if t, ok := branchTarget(p); ok && s.regs[x86asm.R8] == (sum{size, 0}) {
					fixup = t
					break
				}
			}
			s.step(p)
		}
		return int(size), fixup, true
	}
	return 0, 0, false
}
