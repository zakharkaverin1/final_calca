package tests

import (
	"testing"
)

func TestValid_ValidExpression(t *testing.T) {
	valid := Valid("3+(2*4)-5")
	if !valid {
		t.Errorf("Ожидалось true для валидного выражения")
	}
}

func TestValid_InvalidCharacters(t *testing.T) {
	valid := Valid("3+2a")
	if valid {
		t.Errorf("Ожидалось false для выражения с недопустимыми символами")
	}
}

func TestValid_UnbalancedParentheses(t *testing.T) {
	valid := Valid("(3+2")
	if valid {
		t.Errorf("Ожидалось false при несбалансированных скобках")
	}
}

func TestValid_InvalidOperatorPlacement(t *testing.T) {
	valid := Valid("3++2")
	if valid {
		t.Errorf("Ожидалось false при двойных операторах")
	}
}

func TestValid_InvalidLastCharacter(t *testing.T) {
	valid := Valid("3+2(")
	if valid {
		t.Errorf("Ожидалось false при неправильном последнем символе")
	}
}
