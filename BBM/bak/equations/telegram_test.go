package telegram

import (
	"testing"
)

func TestTelegram(t *testing.T) {
	testCase := []struct {
		name     string
		input    string
		expected string
		err      error
	}{
		{
			name:  "very пиздец equations",
			input: "2(3x+(4-2(x-1)))-5(x+2)=7(2-x)+(3+x)",
			expected: `Уравнение: 2*(3*x+1*(4-2*(x-1)))-5*(x+2)=7*(2-x)+1*(3+x)
Тип уравнения: Линейное
Инвертированное уравнение: 2*(3*x+1*(4-2*(x-1)))-5*(x+2)-7*(2+x)-1*(3-x)=0
Упрощенное уравнение: 6*x+8-2*x-4-5*x+10-14+7*x-3-1*x=0`,
			err: nil,
		},
	}
	for _, test := range testCase {
		t.Run(test.name, func(t *testing.T) {
			actual, err := EquationTelegram(test.input)
			if test.err != nil {
				if err == nil || err.Error() != test.err.Error() {
					t.Errorf("expected error %v, got %v", test.err, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if actual != test.expected {
				t.Errorf("expected %s\ngot\n%s", test.expected, actual)
			}
		})
	}
}
