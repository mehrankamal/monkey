package evaluator

import (
	"github.com/mehrankamal/monkey/lexer"
	"github.com/mehrankamal/monkey/object"
	"github.com/mehrankamal/monkey/parser"
	"testing"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-45", -45},
		{"-10", -10},
		{"-(-4)", 4},
		{"1 + 1", 2},
		{"1 - 1", 0},
		{"1 * 1", 1},
		{"1 / 1", 1},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}
	for _, tt := range tests {
		evaluated := evalInput(tt.input)
		assertIntegerObject(t, evaluated, tt.expected)
	}
}

func TestBooleanExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}

	for _, tc := range tests {
		evaluated := evalInput(tc.input)
		assertBooleanObject(t, evaluated, tc.expected)
	}
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}
	for _, tt := range tests {
		evaluated := evalInput(tt.input)
		assertBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, tt := range tests {
		evaluated := evalInput(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			assertIntegerObject(t, evaluated, int64(integer))
		} else {
			assertNullObject(t, evaluated)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
	}
	for idx, tt := range tests {
		evaluated := evalInput(tt.input)
		if !assertIntegerObject(t, evaluated, tt.expected) {
			t.Errorf("Test %d's assertions failed.", idx)
		}
	}
}

func assertNullObject(t *testing.T, evaluated object.Object) bool {
	if evaluated != NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", evaluated, evaluated)
		return false
	}

	return true
}

func evalInput(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	return Eval(program)
}

func assertBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object not Boolean, got %T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("Object has wrong value got=%t, want=%t", result.Value, expected)
		return false
	}

	return true
}

func assertIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d",
			result.Value, expected)
		return false
	}

	return true
}
