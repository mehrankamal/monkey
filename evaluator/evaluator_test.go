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
