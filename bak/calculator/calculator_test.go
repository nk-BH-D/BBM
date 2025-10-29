package calculator

import (
	"errors"
	//"math"
	//"math/big"
	"strings"
	"testing"
)

func TestCalculator(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
		err      error
	}{
		{
			name:     "Simple addition",
			input:    "2 + 3",
			expected: "5",
			err:      nil,
		},
		{
			name:     "Simple subtraction",
			input:    "5 - 2",
			expected: "3",
			err:      nil,
		},
		{
			name:     "Simple multiplication",
			input:    "4 * 6",
			expected: "24",
			err:      nil,
		},
		{
			name:     "Simple division",
			input:    "8 / 2",
			expected: "4",
			err:      nil,
		},
		{
			name:     "Expression with parentheses",
			input:    "(2 + 3) * 4",
			expected: "20",
			err:      nil,
		},
		{
			name:     "Expression with multiple operations",
			input:    "10 + 2 * 3 - 4 / 2",
			expected: "14",
			err:      nil,
		},
		{
			name:     "Complex expression with parentheses",
			input:    "((10 / 2) + 10) * 2",
			expected: "30",
			err:      nil,
		},
		{
			name:     "Division by zero",
			input:    "5 / 0",
			expected: "0",
			err:      errors.New("ошибка вычисления: division by zero"),
		},
		{
			name:     "Mismatched parentheses - opening",
			input:    "(2 + 3",
			expected: "0",
			err:      errors.New("ошибка вычисления: mismatched parentheses"),
		},
		{
			name:     "Mismatched parentheses - closing",
			input:    "2 + 3)",
			expected: "",
			err:      errors.New("ошибка вычисления: mismatched parentheses"),
		},
		{
			name:     "Invalid character",
			input:    "2 + a",
			expected: "",
			err:      errors.New("ошибка вычисления: invalid character: a"),
		},
		{
			name:     "Negative numbers",
			input:    "-2 + 5",
			expected: "3",
			err:      nil,
		},
		{
			name:     "Negative number behind +",
			input:    "2 + -3",
			expected: "-1",
			err:      nil,
		},
		{
			name:     "Negative number behind -",
			input:    "2 --3",
			expected: "5",
			err:      nil,
		},
		{
			name:     "Negative number bihind *",
			input:    "2 *-3",
			expected: "-6",
			err:      nil,
		},
		{
			name:     "Negative nuvber bihind /",
			input:    "2 /-3",
			expected: "-0.666666666666666666667",
			err:      nil,
		},
		{
			name:     "Negative nuvber bihind /",
			input:    "3/10",
			expected: "0.3",
			err:      nil,
		},
		{
			name:     "Float numbers",
			input:    "2.5 + 2.5",
			expected: "5",
			err:      nil,
		},
		{
			name:     "Float number as result",
			input:    "5 / 2",
			expected: "2.5",
			err:      nil,
		},
		{
			name:     "Negative result",
			input:    "2-5",
			expected: "-3",
			err:      nil,
		},
		{
			name:     "Very big numbers",
			input:    "1000000000000 + 2000000000000",
			expected: "3000000000000",
			err:      nil,
		},
		{
			name:     "Very big float number",
			input:    "1000000000000.5 + 2000000000000.5",
			expected: "3000000000001",
			err:      nil,
		},
		{
			name:     "Floating point imprecision",
			input:    "0.1+0.2",
			expected: "0.3",
			err:      nil,
		},
		{
			name:     "Empty expression",
			input:    "",
			expected: "",
			err:      errors.New("ошибка вычисления: invalid expression"),
		},
		{
			name:     "Sqrt",
			input:    "sqrt(25)",
			expected: "5",
			err:      nil,
		},
		{
			name:     "Sum_sqrt",
			input:    "sqrt(8+8)",
			expected: "4",
			err:      nil,
		},
		{
			name:     "Sqrt in expression",
			input:    "sqrt(4)+2",
			expected: "4",
			err:      nil,
		},
		{
			name:     "invalid sqrt",
			input:    "sqrt(9",
			expected: "",
			err:      errors.New("ошибка вычисления: missing closing parenthesis"),
		},
		{
			name:     "Zero Value in sqrt",
			input:    "sqrt()",
			expected: "",
			err:      errors.New("ошибка вычисления: ошибка вычисления: ошибка вычисления: invalid expression"),
		},
		{
			name:     "bigFloat",
			input:    "0.8*0.1",
			expected: "0.08",
			err:      nil,
		},
		{
			name:     "degree value",
			input:    "3**3",
			expected: "27",
			err:      nil,
		},
		{
			name:     "degree negative value",
			input:    "2**(-3)",
			expected: "0.125",
			err:      nil,
		},
		{
			name:     "degree onli negative value",
			input:    "-2**(-3)",
			expected: "-0.125",
			err:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := CalculatorTelegram(tc.input)
			if tc.err != nil {
				if err == nil || err.Error() != tc.err.Error() {
					t.Errorf("expected error '%v', got '%v'", tc.err, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			actual = strings.TrimSpace(strings.ReplaceAll(actual, "\r\n", "\n"))
			expected := strings.TrimSpace(strings.ReplaceAll(tc.expected, "\r\n", "\n"))

			if actual != expected {
				t.Errorf("expected:\n%s\ngot:\n%s", expected, actual)
			}
		})
	}
}
