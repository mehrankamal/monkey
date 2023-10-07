package vm

import (
	"fmt"
	"github.com/mehrankamal/monkey/code"
	"github.com/mehrankamal/monkey/compiler"
	"github.com/mehrankamal/monkey/object"
)

const StackSize = 2048

type VirtualMachine struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]
}

func New(bytecode *compiler.Bytecode) *VirtualMachine {
	return &VirtualMachine{
		constants:    bytecode.Constants,
		instructions: bytecode.Instructions,

		stack: make([]object.Object, StackSize),
		sp:    0,
	}
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
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpPop:
			_, err := vm.pop()
			if err != nil {
				return err
			}
		}
	}

	return nil
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

	fmt.Printf("Input %d %d %d = %d\n", leftVal, op, rightVal, result)

	return vm.push(&object.Integer{Value: result})

}
