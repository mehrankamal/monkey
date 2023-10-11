package vm

import (
	"fmt"
	"github.com/mehrankamal/monkey/code"
	"github.com/mehrankamal/monkey/compiler"
	"github.com/mehrankamal/monkey/object"
)

const StackSize = 2048
const GlobalsSize = 65536
const MaxFrames = 1024

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

type VirtualMachine struct {
	constants []object.Object

	stack []object.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]

	globals []object.Object

	frames     []*Frame
	frameIndex int
}

func New(bytecode *compiler.Bytecode) *VirtualMachine {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(mainFn)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VirtualMachine{
		constants: bytecode.Constants,

		stack: make([]object.Object, StackSize),
		sp:    0,

		globals: make([]object.Object, GlobalsSize),

		frames:     frames,
		frameIndex: 1,
	}
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VirtualMachine {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VirtualMachine) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}

	return vm.stack[vm.sp-1]
}

func (vm *VirtualMachine) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}

func (vm *VirtualMachine) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()

		op = code.Opcode(ins[ip])

		switch op {
		case code.OpPop:
			_, err := vm.pop()
			if err != nil {
				return err
			}

		case code.OpConstant:
			constIndex := code.ReadUint16(vm.currentFrame().Instructions()[vm.currentFrame().ip+1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}

		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}

		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}
		case code.OpNegate:
			err := vm.executeNegateOperator()
			if err != nil {
				return err
			}

		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}

		case code.OpJumpFalsy:
			pos := int(code.ReadUint16(vm.currentFrame().Instructions()[vm.currentFrame().ip+1:]))
			vm.currentFrame().ip += 2

			condition, err := vm.pop()
			if err != nil {
				return err
			}

			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}
		case code.OpJump:
			pos := int(code.ReadUint16(vm.currentFrame().Instructions()[vm.currentFrame().ip+1:]))
			vm.currentFrame().ip = pos - 1

		case code.OpSetGlobal:
			globalIdx := code.ReadUint16(vm.currentFrame().Instructions()[vm.currentFrame().ip+1:])
			vm.currentFrame().ip += 2

			obj, err := vm.pop()
			if err != nil {
				return err
			}

			vm.globals[globalIdx] = obj
		case code.OpGetGlobal:
			idx := code.ReadUint16(vm.currentFrame().Instructions()[vm.currentFrame().ip+1:])
			vm.currentFrame().ip += 2

			err := vm.push(vm.globals[idx])
			if err != nil {
				return err
			}

		case code.OpArray:
			arraySize := int(code.ReadUint16(vm.currentFrame().Instructions()[vm.currentFrame().ip+1:]))
			vm.currentFrame().ip += 2

			array := vm.buildArray(vm.sp-arraySize, vm.sp)
			vm.sp -= arraySize

			err := vm.push(array)
			if err != nil {
				return err
			}
		case code.OpHash:
			numHashPairs := int(code.ReadUint16(vm.currentFrame().Instructions()[vm.currentFrame().ip+1:]))
			vm.currentFrame().ip += 2

			hash, err := vm.buildHash(vm.sp-(numHashPairs*2), numHashPairs)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - (numHashPairs * 2)

			err = vm.push(hash)
			if err != nil {
				return err
			}
		case code.OpIndex:
			index, err := vm.pop()
			if err != nil {
				return err
			}
			left, err := vm.pop()
			if err != nil {
				return err
			}

			err = vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func (vm *VirtualMachine) pushFrame(f *Frame) {
	vm.frames[vm.frameIndex] = f
	vm.frameIndex++
}

func (vm *VirtualMachine) popFrame() *Frame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
}

func isTruthy(condition object.Object) bool {
	switch condition := condition.(type) {
	case *object.Boolean:
		return condition.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

func (vm *VirtualMachine) executeBinaryOperation(op code.Opcode) error {
	right, err := vm.pop()
	if err != nil {
		return err
	}
	left, err := vm.pop()
	if err != nil {
		return err
	}

	leftType := left.Type()
	rightType := right.Type()

	switch {
	case leftType == object.INTEGER && rightType == object.INTEGER:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case leftType == object.STRING && rightType == object.STRING:
		return vm.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s",
			leftType, rightType)
	}
}

func (vm *VirtualMachine) push(o object.Object) error {

	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VirtualMachine) pop() (object.Object, error) {
	if vm.sp == 0 {
		return nil, fmt.Errorf("stack empty")
	}

	obj := vm.stack[vm.sp-1]
	vm.sp -= 1

	return obj, nil
}

func (vm *VirtualMachine) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VirtualMachine) executeBinaryIntegerOperation(op code.Opcode, left, right object.Object) error {
	rightVal := right.(*object.Integer).Value
	leftVal := left.(*object.Integer).Value

	var result int64

	switch op {
	case code.OpAdd:
		result = leftVal + rightVal
	case code.OpSub:
		result = leftVal - rightVal
	case code.OpMul:
		result = leftVal * rightVal
	case code.OpDiv:
		result = leftVal / rightVal
	default:
		return fmt.Errorf("unknown integer operation: %d", op)
	}

	return vm.push(&object.Integer{Value: result})

}

func (vm *VirtualMachine) executeComparison(op code.Opcode) error {
	right, err := vm.pop()
	if err != nil {
		return err
	}

	left, err := vm.pop()
	if err != nil {
		return err
	}

	if left.Type() == object.INTEGER && right.Type() == object.INTEGER {
		return vm.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)",
			op, left.Type(), right.Type())
	}
}

func nativeBoolToBooleanObject(b bool) object.Object {
	if b {
		return True
	}

	return False
}

func (vm *VirtualMachine) executeIntegerComparison(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue == leftValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue != leftValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VirtualMachine) executeBangOperator() error {
	operand, err := vm.pop()
	if err != nil {
		return err
	}

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VirtualMachine) executeNegateOperator() error {
	operand, err := vm.pop()
	if err != nil {
		return err
	}

	if operand.Type() != object.INTEGER {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	value := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -value})
}

func (vm *VirtualMachine) executeBinaryStringOperation(op code.Opcode, left, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	return vm.push(&object.String{Value: leftValue + rightValue})
}

func (vm *VirtualMachine) buildArray(start int, end int) object.Object {
	elems := make([]object.Object, end-start)

	for i := start; i < end; i++ {
		elems[i-start] = vm.stack[i]
	}

	return &object.Array{Elements: elems}
}

func (vm *VirtualMachine) buildHash(start int, size int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)

	for i := 0; i < size; i += 1 {
		key := vm.stack[start+(i*2)]
		value := vm.stack[start+(i*2)+1]

		pair := object.HashPair{Key: key, Value: value}
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		hashedPairs[hashKey.HashKey()] = pair
	}

	return &object.Hash{Pairs: hashedPairs}, nil
}

func (vm *VirtualMachine) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY && index.Type() == object.INTEGER:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VirtualMachine) executeArrayIndex(array object.Object, index object.Object) error {
	arrayObj := array.(*object.Array)
	idxValue := index.(*object.Integer).Value

	max := int64(len(arrayObj.Elements) - 1)

	if idxValue < 0 || idxValue > max {
		return vm.push(Null)
	}

	return vm.push(arrayObj.Elements[idxValue])

}

func (vm *VirtualMachine) executeHashIndex(hash, index object.Object) error {
	hashObj := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObj.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}
