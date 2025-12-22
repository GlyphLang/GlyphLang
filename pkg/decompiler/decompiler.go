package decompiler

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"github.com/glyphlang/glyph/pkg/vm"
)

// Decompiler converts bytecode back to readable representation
type Decompiler struct {
	bytecode   []byte
	constants  []vm.Value
	offset     int
	codeStart  int
	codeLength int
}

// DecompiledOutput represents the decompiled bytecode
type DecompiledOutput struct {
	Version      uint32
	Constants    []ConstantInfo
	Instructions []InstructionInfo
	Source       string // Reconstructed source (best effort)
}

// ConstantInfo represents a constant in the pool
type ConstantInfo struct {
	Index int
	Type  string
	Value string
}

// InstructionInfo represents a single instruction
type InstructionInfo struct {
	Offset  int
	Opcode  string
	Operand string
	Comment string
}

// NewDecompiler creates a new decompiler instance
func NewDecompiler() *Decompiler {
	return &Decompiler{}
}

// Decompile converts bytecode to a DecompiledOutput
func (d *Decompiler) Decompile(bytecode []byte) (*DecompiledOutput, error) {
	d.bytecode = bytecode
	d.offset = 0
	d.constants = nil

	output := &DecompiledOutput{}

	// Verify and read header
	if len(bytecode) < 4 {
		return nil, fmt.Errorf("invalid bytecode: too short")
	}

	if string(bytecode[0:4]) != "GLYP" {
		return nil, fmt.Errorf("invalid bytecode: bad magic bytes (expected GLYP)")
	}
	d.offset = 4

	// Read version
	if d.offset+4 > len(bytecode) {
		return nil, fmt.Errorf("invalid bytecode: missing version")
	}
	output.Version = binary.LittleEndian.Uint32(bytecode[d.offset : d.offset+4])
	d.offset += 4

	// Read constants
	if d.offset+4 > len(bytecode) {
		return nil, fmt.Errorf("invalid bytecode: missing constant count")
	}
	constCount := binary.LittleEndian.Uint32(bytecode[d.offset : d.offset+4])
	d.offset += 4

	for i := uint32(0); i < constCount; i++ {
		constInfo, err := d.readConstant(int(i))
		if err != nil {
			return nil, fmt.Errorf("error reading constant %d: %w", i, err)
		}
		output.Constants = append(output.Constants, constInfo)
	}

	// Read instruction count
	if d.offset+4 > len(bytecode) {
		return nil, fmt.Errorf("invalid bytecode: missing instruction count")
	}
	d.codeLength = int(binary.LittleEndian.Uint32(bytecode[d.offset : d.offset+4]))
	d.offset += 4
	d.codeStart = d.offset

	// Decompile instructions
	for d.offset < d.codeStart+d.codeLength && d.offset < len(bytecode) {
		instrInfo, err := d.readInstruction()
		if err != nil {
			return nil, fmt.Errorf("error reading instruction at offset %d: %w", d.offset, err)
		}
		output.Instructions = append(output.Instructions, instrInfo)
	}

	// Generate reconstructed source
	output.Source = d.reconstructSource(output)

	return output, nil
}

// readConstant reads and formats a constant
func (d *Decompiler) readConstant(index int) (ConstantInfo, error) {
	info := ConstantInfo{Index: index}

	if d.offset >= len(d.bytecode) {
		return info, fmt.Errorf("unexpected end of bytecode")
	}

	constType := d.bytecode[d.offset]
	d.offset++

	switch constType {
	case 0x00: // Null
		info.Type = "null"
		info.Value = "null"
		d.constants = append(d.constants, vm.NullValue{})

	case 0x01: // Int
		if d.offset+8 > len(d.bytecode) {
			return info, fmt.Errorf("truncated int constant")
		}
		val := int64(binary.LittleEndian.Uint64(d.bytecode[d.offset : d.offset+8]))
		d.offset += 8
		info.Type = "int"
		info.Value = fmt.Sprintf("%d", val)
		d.constants = append(d.constants, vm.IntValue{Val: val})

	case 0x02: // Float
		if d.offset+8 > len(d.bytecode) {
			return info, fmt.Errorf("truncated float constant")
		}
		bits := binary.LittleEndian.Uint64(d.bytecode[d.offset : d.offset+8])
		val := math.Float64frombits(bits)
		d.offset += 8
		info.Type = "float"
		info.Value = fmt.Sprintf("%g", val)
		d.constants = append(d.constants, vm.FloatValue{Val: val})

	case 0x03: // Bool
		if d.offset >= len(d.bytecode) {
			return info, fmt.Errorf("truncated bool constant")
		}
		val := d.bytecode[d.offset] != 0
		d.offset++
		info.Type = "bool"
		info.Value = fmt.Sprintf("%t", val)
		d.constants = append(d.constants, vm.BoolValue{Val: val})

	case 0x04: // String
		if d.offset+4 > len(d.bytecode) {
			return info, fmt.Errorf("truncated string length")
		}
		length := binary.LittleEndian.Uint32(d.bytecode[d.offset : d.offset+4])
		d.offset += 4
		if d.offset+int(length) > len(d.bytecode) {
			return info, fmt.Errorf("truncated string data")
		}
		val := string(d.bytecode[d.offset : d.offset+int(length)])
		d.offset += int(length)
		info.Type = "string"
		info.Value = fmt.Sprintf("%q", val)
		d.constants = append(d.constants, vm.StringValue{Val: val})

	default:
		return info, fmt.Errorf("unknown constant type: 0x%02x", constType)
	}

	return info, nil
}

// readInstruction reads and formats an instruction
func (d *Decompiler) readInstruction() (InstructionInfo, error) {
	info := InstructionInfo{
		Offset: d.offset - d.codeStart,
	}

	if d.offset >= len(d.bytecode) {
		return info, fmt.Errorf("unexpected end of bytecode")
	}

	opcode := vm.Opcode(d.bytecode[d.offset])
	d.offset++

	info.Opcode = opcodeToString(opcode)

	// Handle operands
	if hasOperand(opcode) {
		if d.offset+4 > len(d.bytecode) {
			return info, fmt.Errorf("truncated operand")
		}
		operand := binary.LittleEndian.Uint32(d.bytecode[d.offset : d.offset+4])
		d.offset += 4
		info.Operand = fmt.Sprintf("%d", operand)

		// Add helpful comments
		info.Comment = d.getOperandComment(opcode, operand)
	}

	return info, nil
}

// getOperandComment provides context for operands
func (d *Decompiler) getOperandComment(opcode vm.Opcode, operand uint32) string {
	switch opcode {
	case vm.OpPush, vm.OpLoadVar, vm.OpStoreVar:
		if int(operand) < len(d.constants) {
			val := d.constants[operand]
			switch v := val.(type) {
			case vm.StringValue:
				return fmt.Sprintf("; {%s}", v.Val)
			case vm.IntValue:
				return fmt.Sprintf("; {%d}", v.Val)
			case vm.FloatValue:
				return fmt.Sprintf("; {%g}", v.Val)
			case vm.BoolValue:
				return fmt.Sprintf("; {%t}", v.Val)
			case vm.NullValue:
				return "; {null}"
			default:
				return fmt.Sprintf("; const[%d]", operand)
			}
		}
	case vm.OpJump, vm.OpJumpIfFalse, vm.OpJumpIfTrue:
		return fmt.Sprintf("; -> offset %d", operand)
	case vm.OpBuildObject:
		return fmt.Sprintf("; %d fields", operand)
	case vm.OpBuildArray:
		return fmt.Sprintf("; %d elements", operand)
	case vm.OpCall:
		return fmt.Sprintf("; %d args", operand)
	case vm.OpIterNext:
		if operand != 0 {
			return "; with key"
		}
		return "; value only"
	}
	return ""
}

// reconstructSource attempts to reconstruct GlyphLang source
func (d *Decompiler) reconstructSource(output *DecompiledOutput) string {
	var sb strings.Builder

	sb.WriteString("# Decompiled GlyphLang Source\n")
	sb.WriteString(fmt.Sprintf("# Version: %d\n", output.Version))
	sb.WriteString(fmt.Sprintf("# Constants: %d\n", len(output.Constants)))
	sb.WriteString(fmt.Sprintf("# Instructions: %d\n\n", len(output.Instructions)))

	// Extract route path from constants if possible
	routePath := "/"
	for _, c := range d.constants {
		if sv, ok := c.(vm.StringValue); ok {
			if len(sv.Val) > 0 && sv.Val[0] == '/' {
				routePath = sv.Val
				break
			}
		}
	}

	sb.WriteString(fmt.Sprintf("@ route %s\n", routePath))

	// Track variables and build pseudo-source
	variables := make(map[uint32]string)
	indent := "  "

	for _, instr := range output.Instructions {
		switch instr.Opcode {
		case "STORE_VAR":
			if idx, err := parseOperand(instr.Operand); err == nil {
				if int(idx) < len(d.constants) {
					if sv, ok := d.constants[idx].(vm.StringValue); ok {
						variables[idx] = sv.Val
						// Don't output internal variables
						if !strings.HasPrefix(sv.Val, "__") {
							sb.WriteString(fmt.Sprintf("%s$ %s = ...\n", indent, sv.Val))
						}
					}
				}
			}
		case "RETURN":
			sb.WriteString(fmt.Sprintf("%s> ...\n", indent))
		}
	}

	sb.WriteString("\n# --- Bytecode Disassembly ---\n\n")

	// Add constant pool
	sb.WriteString("# Constant Pool:\n")
	for _, c := range output.Constants {
		sb.WriteString(fmt.Sprintf("#   [%d] %s: %s\n", c.Index, c.Type, c.Value))
	}
	sb.WriteString("\n")

	// Add instructions
	sb.WriteString("# Instructions:\n")
	for _, instr := range output.Instructions {
		line := fmt.Sprintf("#   %04d: %-15s", instr.Offset, instr.Opcode)
		if instr.Operand != "" {
			line += fmt.Sprintf(" %s", instr.Operand)
		}
		if instr.Comment != "" {
			line += fmt.Sprintf("  %s", instr.Comment)
		}
		sb.WriteString(line + "\n")
	}

	return sb.String()
}

func parseOperand(s string) (uint32, error) {
	var v uint32
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

// opcodeToString converts an opcode to its string name
func opcodeToString(op vm.Opcode) string {
	names := map[vm.Opcode]string{
		vm.OpPush:            "PUSH",
		vm.OpPop:             "POP",
		vm.OpAdd:             "ADD",
		vm.OpSub:             "SUB",
		vm.OpMul:             "MUL",
		vm.OpDiv:             "DIV",
		vm.OpEq:              "EQ",
		vm.OpNe:              "NE",
		vm.OpLt:              "LT",
		vm.OpGt:              "GT",
		vm.OpGe:              "GE",
		vm.OpLe:              "LE",
		vm.OpAnd:             "AND",
		vm.OpOr:              "OR",
		vm.OpNot:             "NOT",
		vm.OpNeg:             "NEG",
		vm.OpLoadVar:         "LOAD_VAR",
		vm.OpStoreVar:        "STORE_VAR",
		vm.OpJump:            "JUMP",
		vm.OpJumpIfFalse:     "JUMP_IF_FALSE",
		vm.OpJumpIfTrue:      "JUMP_IF_TRUE",
		vm.OpGetIter:         "GET_ITER",
		vm.OpIterNext:        "ITER_NEXT",
		vm.OpIterHasNext:     "ITER_HAS_NEXT",
		vm.OpGetIndex:        "GET_INDEX",
		vm.OpReturn:          "RETURN",
		vm.OpCall:            "CALL",
		vm.OpBuildObject:     "BUILD_OBJECT",
		vm.OpGetField:        "GET_FIELD",
		vm.OpBuildArray:      "BUILD_ARRAY",
		vm.OpHttpReturn:      "HTTP_RETURN",
		vm.OpWsSend:          "WS_SEND",
		vm.OpWsBroadcast:     "WS_BROADCAST",
		vm.OpWsBroadcastRoom: "WS_BROADCAST_ROOM",
		vm.OpWsJoinRoom:      "WS_JOIN_ROOM",
		vm.OpWsLeaveRoom:     "WS_LEAVE_ROOM",
		vm.OpWsClose:         "WS_CLOSE",
		vm.OpWsGetRooms:      "WS_GET_ROOMS",
		vm.OpWsGetClients:    "WS_GET_CLIENTS",
		vm.OpWsGetConnCount:  "WS_GET_CONN_COUNT",
		vm.OpWsGetUptime:     "WS_GET_UPTIME",
		vm.OpHalt:            "HALT",
	}

	if name, ok := names[op]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN_0x%02X", byte(op))
}

// hasOperand returns true if the opcode has an operand
func hasOperand(op vm.Opcode) bool {
	withOperand := map[vm.Opcode]bool{
		vm.OpPush:        true,
		vm.OpLoadVar:     true,
		vm.OpStoreVar:    true,
		vm.OpJump:        true,
		vm.OpJumpIfFalse: true,
		vm.OpJumpIfTrue:  true,
		vm.OpIterNext:    true,
		vm.OpCall:        true,
		vm.OpBuildObject: true,
		vm.OpBuildArray:  true,
	}
	return withOperand[op]
}

// Format returns a formatted string representation
func (o *DecompiledOutput) Format() string {
	return o.Source
}

// FormatDisassembly returns only the disassembly portion
func (o *DecompiledOutput) FormatDisassembly() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("GlyphLang Bytecode v%d\n", o.Version))
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	sb.WriteString("CONSTANT POOL:\n")
	sb.WriteString(strings.Repeat("-", 30) + "\n")
	for _, c := range o.Constants {
		sb.WriteString(fmt.Sprintf("  [%3d] %-8s %s\n", c.Index, c.Type, c.Value))
	}

	sb.WriteString("\nINSTRUCTIONS:\n")
	sb.WriteString(strings.Repeat("-", 30) + "\n")
	for _, instr := range o.Instructions {
		line := fmt.Sprintf("  %04d: %-18s", instr.Offset, instr.Opcode)
		if instr.Operand != "" {
			line += fmt.Sprintf(" %-6s", instr.Operand)
		} else {
			line += "       "
		}
		if instr.Comment != "" {
			line += fmt.Sprintf(" %s", instr.Comment)
		}
		sb.WriteString(line + "\n")
	}

	return sb.String()
}
