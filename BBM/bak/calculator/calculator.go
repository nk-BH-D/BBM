package calculator

import (
	"errors"
	"fmt"
	//"log"
	"math"
	"math/big"
	"strconv"
	"strings"
)

func Calc(expression string) (*big.Float, error) {
	tokens, err := tokenize(expression)
	if err != nil {
		return nil, err
	}
	postfix, err := infixToPostfix(tokens)
	if err != nil {
		return nil, err
	}
	result, err := evaluatePostfix(postfix)
	if err != nil {
		return nil, err
	}
	//log.Printf("Calc: %f\n", result)
	return result, nil // возвращаем результат без форматирования и вывода в консоль
}

// formatBigFloat - функция форматирования чисел с плавающей точкой из math/big
func formatBigFloat(num *big.Float) string {
	s := num.Text('f', -1)
	parts := strings.Split(s, ".")
	if len(parts) == 2 {
		decimalPart := parts[1]
		precision := 0
		for i, r := range decimalPart {
			if r != '0' {
				precision = i + 1
			}
		}
		if precision > 0 {
			return fmt.Sprintf("%."+fmt.Sprint(precision)+"f", num)
		}
	}
	return s
}

// функция финального форматирования
func finishFormatted(formatted string) string {
	parts := strings.SplitN(formatted, ".", 2)
	if len(parts) != 2 {
		return formatted
	}
	intPart, fracPart := parts[0], parts[1]
	//так как по необъяснимой причине мусорные хвосты всегда содержат 33 повторяющихся символов, дабы максимально увеличить точность вычислений, хвост будет считаться мусорным только если содержит 33 повторяющихся символа
	const repeat = 33
	for i := 0; i <= len(fracPart)-repeat; i++ {
		allSame := true
		for j := 1; j < repeat; j++ {
			if fracPart[i] != fracPart[i+j] {
				allSame = false
				break
			}

		}
		if allSame {
			fracPart = fracPart[:i]
			break
		}
	}

	if fracPart == "" {
		return intPart
	}
	// возращаем окончательный результат, если всё прошло хорошо
	fF := intPart + "." + fracPart
	return fF
}

// обвязка с телеграммом
func CalculatorTelegram(expression string) (string, error) {
	result, err := Calc(expression)
	if err != nil {
		return "", fmt.Errorf("ошибка вычисления: %w", err)
	}
	formatted := formatBigFloat(result)

	fF := finishFormatted(formatted)

	return fF, nil
}

// функция токенизации разбивает входную строку expr на отдельные токены. Она использует strings.Builder для эффективного построения токенов
func tokenize(expr string) ([]string, error) {
	var tokens []string
	var currentToken strings.Builder
	previousToken := ""

	for i := 0; i < len(expr); i++ {
		char := rune(expr[i])
		switch {
		case char == ' ':
			continue
		case string(char) == "(":
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, string(char))
		case string(char) == ")":
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, string(char))
		case i+1 < len(expr) && string(expr[i:i+2]) == "**":
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, "**")
			i++
			previousToken = "**"
			continue
		case char == '+' || char == '*' || char == '/' || string(char) == "sqrt":
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, string(char))
		case char == '-':
			if previousToken == "" || previousToken == "(" || previousToken == "+" || previousToken == "*" || previousToken == "/" || previousToken == "-" {
				currentToken.WriteRune(char)
			} else {
				if currentToken.Len() > 0 {
					tokens = append(tokens, currentToken.String())
					currentToken.Reset()
				}
				tokens = append(tokens, string(char))
			}
		default:
			currentToken.WriteRune(char)
		}
		if char != ' ' && string(expr[i:i+1]) != "**" {
			if currentToken.Len() > 0 {
				previousToken = currentToken.String()
			} else {
				previousToken = string(char)
			}
		}
	}

	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}
	i := 0
	for i < len(tokens) {
		if tokens[i] == "sqrt" && i+1 < len(tokens) && tokens[i+1] == "(" {
			j := i + 2
			arg := ""
			foundClosingParen := false
			for j < len(tokens) {
				if tokens[j] == ")" {
					foundClosingParen = true
					break
				}
				arg += tokens[j]
				j++
			}
			if !foundClosingParen {
				return nil, errors.New("missing closing parenthesis")
			}

			tokens = append(tokens[:i], append([]string{fmt.Sprintf("sqrt(%s)", arg)}, tokens[j+1:]...)...)
			i = 0
			continue
		}
		i++
	}
	return tokens, nil
}

// эта функция — сердце алгоритма она преобразует токены из инфиксной нотации в постфиксную она использует стек (operators) для хранения операторов.
func infixToPostfix(tokens []string) ([]string, error) {
	var output []string
	var operators []string

	for _, token := range tokens {
		if isNumber(token) {
			output = append(output, token)
		} else if token == "(" {
			operators = append(operators, token)
		} else if token == ")" {
			for len(operators) > 0 && operators[len(operators)-1] != "(" {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			if len(operators) == 0 {
				return nil, errors.New("mismatched parentheses")
			}
			operators = operators[:len(operators)-1]
		} else if isOperator(token) {
			for len(operators) > 0 && precedence(operators[len(operators)-1]) >= precedence(token) {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
			operators = append(operators, token)
		} else {
			return nil, fmt.Errorf("invalid character: %s", token)
		}
	}

	for len(operators) > 0 {
		if operators[len(operators)-1] == "(" {
			return nil, errors.New("mismatched parentheses")
		}
		output = append(output, operators[len(operators)-1])
		operators = operators[:len(operators)-1]
	}

	return output, nil
}

// эта функция вычисляет значение арифметического выражения представленного в постфиксной нотации (postfix).
func evaluatePostfix(postfix []string) (*big.Float, error) {
	var stack []*big.Float

	for _, token := range postfix {
		if isNumber(token) {
			num, _, err := big.ParseFloat(token, 10, 128, big.ToNearestEven)
			if err != nil {
				return nil, fmt.Errorf("invalid number: %w", err)
			}
			stack = append(stack, num)
		} else if isOperator(token) {
			if len(stack) < 2 && !strings.HasPrefix(token, "sqrt(") { // проверка на sqrt(x)
				return nil, errors.New("invalid expression")
			}
			if strings.HasPrefix(token, "sqrt(") {
				numStr := token[5 : len(token)-1]
				num, _, err := big.ParseFloat(numStr, 10, 128, big.ToNearestEven)
				if err != nil {
					return nil, fmt.Errorf("invalid number in sqrt: %w", err)
				}
				result := new(big.Float).Sqrt(num)
				stack = append(stack, result)
				continue // переходим к следующему токену после обработки sqrt
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			result := new(big.Float)
			switch token {
			case "+":
				result.Add(a, b)
			case "-":
				result.Sub(a, b)
			case "*":
				result.Mul(a, b)
			case "/":
				if b.Cmp(big.NewFloat(0)) == 0 {
					return nil, errors.New("division by zero")
				}
				result.Quo(a, b)
			case "**":
				af, _ := a.Float64()
				bf, _ := b.Float64()
				pow := math.Pow(af, bf)
				result.SetFloat64(pow)
			default:
				return nil, fmt.Errorf("unknown operator: %s", token)
			}
			stack = append(stack, result)
		} else {
			return nil, fmt.Errorf("invalid token: %s", token)
		}
	}

	if len(stack) != 1 {
		return nil, errors.New("invalid expression")
	}

	return stack[0], nil
}

// определяет тип токена
func isNumber(token string) bool {
	if _, err := strconv.ParseFloat(token, 64); err == nil {
		return true
	}
	return false
}

// определяет тип токена
func isOperator(token string) bool {
	return token == "+" || token == "-" || token == "*" || token == "/" || token == "**" || strings.HasPrefix(token, "sqrt(")
}

// определяет порядок выполнения действий
func precedence(op string) int {
	if strings.HasPrefix(op, "sqrt(") { // особый приорите что бы избежать несовместимости с "**"
		return 3
	}
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	case "**":
		return 3
	default:
		return 0
	}
}
