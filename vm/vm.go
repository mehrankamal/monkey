package vm

import (
	"fmt"
	"github.com/mehrankamal/monkey/code"
	"github.com/mehrankamal/monkey/compiler"
	"github.com/mehrankamal/monkey/object"
)

const StackSize = 2048
const GlobalsSize = 65536

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

type VirtualMachine struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]

	globals []object.Object
}

func New(bytecode *compiler.Bytecode) *VirtualMachine {
	return &VirtualMachine{
		constants:    bytecode.Constants,
		instructions: bytecode.Instructions,

		stack: make([]object.Object, StackSize),
		sp:    0,

		globals: make([]object.Object, GlobalsSize),
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

func (vm *VirtualMachine) Run() error {

	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpPop:
			_, err := vm.pop()
			if err != nil {
				return err
			}

		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

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
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2

			condition, err := vm.pop()
			if err != nil {
				return err
			}

			if !isTruthy(condition) {
				ip = pos - 1
			}
		case code.OpJump:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip = pos - 1

		case code.OpSetGlobal:
			globalIdx := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			obj, err := vm.pop()
			if err != nil {
				return err
			}

			vm.globals[globalIdx] = obj
		case code.OpGetGlobal:
			idx := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.globals[idx])
			if err != nil {
				return err
			}
		}

	}

	return nil
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

	if leftType == object.INTEGER && rightType == object.INTEGER {
		return vm.executeBinaryIntegerOperation(op, left, right)
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s",
		leftType, rightType)
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
