package calculator

import (
	"errors"
	"fmt"
	"log"
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
	log.Printf("C: result: %v", result)
	if err != nil {
		return nil, err
	}
	//log.Printf("Calc: %f\n", result)
	return result, nil // возвращаем результат без форматирования и вывода в консоль
}

// formatBigFloat - функция форматирования чисел с плавающей точкой из math/big
func formatBigFloat(num *big.Float) string {
	log.Printf("fBF: num: %v", num)
	s := num.Text('f', 21)
	log.Printf("fFB: s: %s", s)
	parts := strings.Split(s, ".")
	if len(parts) == 2 {
		decimalPart := parts[1]
		log.Printf("fBF: decinalPart: %s", decimalPart)
		precision := 0
		for i, r := range decimalPart {
			if r != '0' {
				log.Printf("fBF: precision: %d", precision)
				precision = i + 1
			}
		}
		if precision > 0 {
			log.Printf("fBF: last precision: %d", precision)
			return fmt.Sprintf("%."+fmt.Sprint(precision)+"f", num)
		}
	}
	return s
}

// функция финального форматирования
func finishFormatted(formatted string) string {
	log.Printf("fF: formated: %s", formatted)
	parts := strings.SplitN(formatted, ".", 2)
	if len(parts) != 2 {
		return formatted
	}
	intPart, fracPart := parts[0], parts[1]
	//так как по необъяснимой причине мусорные хвосты всегда содержат 33 повторяющихся символов, дабы максимально увеличить точность вычислений, хвост будет считаться мусорным только если содержит 33 повторяющихся символа
	const repeat = 21
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
	log.Printf("CT: result: %v", result)
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
		if (tokens[i] == "sqrt" ||
			tokens[i] == "sin" ||
			tokens[i] == "cos" ||
			tokens[i] == "tg" ||
			tokens[i] == "ctg" ||
			tokens[i] == "arcsin" ||
			tokens[i] == "arccos" ||
			tokens[i] == "arctg" ||
			tokens[i] == "arcctg") &&
			i+1 < len(tokens) && tokens[i+1] == "(" {
			j := i + 2
			f := tokens[i]
			arg := ""
			final_arg := ""
			foundClosingParen := false
			for j < len(tokens) {
				if tokens[j] == ")" {
					foundClosingParen = true
					ct_arg, err := CalculatorTelegram(arg)
					if err != nil {
						return nil, fmt.Errorf("ошибка вычисления: %v", err)
					}
					final_arg = ct_arg
					break
				}
				arg += tokens[j]
				j++
			}
			if !foundClosingParen {
				return nil, errors.New("missing closing parenthesis")
			}

			tokens = append(tokens[:i], append([]string{fmt.Sprintf("%s(%s)", f, final_arg)}, tokens[j+1:]...)...)
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
			// pop '('
			operators = operators[:len(operators)-1]

			// если сверху — функция, то сразу выталкиваем её в output
			if len(operators) > 0 && isFanc(operators[len(operators)-1]) {
				output = append(output, operators[len(operators)-1])
				operators = operators[:len(operators)-1]
			}
		} else if isFanc(token) {
			// функция — просто кладу в стек операторов (унарный, высокий приоритет)
			operators = append(operators, token)
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
			log.Printf("eP: stack: %v", stack)
		} else if isOperator(token) {
			b := stack[len(stack)-1]
			log.Printf("eP: b: %v", b)
			a := stack[len(stack)-2]
			log.Printf("eP: a: %v", a)
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
				log.Printf("eP: result: %v", result)
			case "**":
				af, _ := a.Float64()
				bf, _ := b.Float64()
				pow := math.Pow(af, bf)
				result.SetFloat64(pow)
			default:
				return nil, fmt.Errorf("unknown operator: %s", token)
			}
			stack = append(stack, result)
		} else if isFanc(token) {
			f := detectFunc(token)
			switch f {
			case "sqrt(":
				numStr := token[5 : len(token)-1]
				num, _, err := big.ParseFloat(numStr, 10, 128, big.ToNearestEven)
				if err != nil {
					return nil, fmt.Errorf("invalid number in %s: %w", f, err)
				}
				result := new(big.Float).Sqrt(num)
				stack = append(stack, result)
			default:
				return nil, fmt.Errorf("unkmown func: %s", token)
			}
		} else {
			return nil, fmt.Errorf("invalid token: %s", token)
		}
	}

	if len(stack) != 1 {
		return nil, errors.New("invalid expression")
	}

	return stack[0], nil
}

// вспомогательная функция calculationFunc, определяет функцию
func detectFunc(token string) string {
	funcs := []string{"sqrt(", "sin(", "cos(", "tg(", "ctg(", "arcsin(", "arccos(", "arctg(", "arcctg("}
	for _, f := range funcs {
		if strings.HasPrefix(token, f) {
			log.Printf("dF: функция: %s", f)
			return f
		}
	}
	return ""
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
	return token == "+" ||
		token == "-" ||
		token == "*" ||
		token == "/" ||
		token == "**"
}

// переезд
// определяем тип токена
func isFanc(token string) bool {
	return strings.HasPrefix(token, "sqrt(") ||
		strings.HasPrefix(token, "sin(") ||
		strings.HasPrefix(token, "cos(") ||
		strings.HasPrefix(token, "tg(") ||
		strings.HasPrefix(token, "tan(") ||
		strings.HasPrefix(token, "ctg(") ||
		strings.HasPrefix(token, "cot(") ||
		strings.HasPrefix(token, "arcsin(") ||
		strings.HasPrefix(token, "arccos(") ||
		strings.HasPrefix(token, "arctg(") ||
		strings.HasPrefix(token, "arctan(") ||
		strings.HasPrefix(token, "arcctg(") ||
		strings.HasPrefix(token, "arccot(")
}

// определяет порядок выполнения действий
func precedence(op string) int {
	if strings.HasPrefix(op, "sqrt(") ||
		strings.HasPrefix(op, "sin(") ||
		strings.HasPrefix(op, "cos(") ||
		strings.HasPrefix(op, "tg(") ||
		strings.HasPrefix(op, "tan(") ||
		strings.HasPrefix(op, "ctg(") ||
		strings.HasPrefix(op, "cot(") ||
		strings.HasPrefix(op, "arcsin(") ||
		strings.HasPrefix(op, "arccos(") ||
		strings.HasPrefix(op, "arctg(") ||
		strings.HasPrefix(op, "arctan(") ||
		strings.HasPrefix(op, "arcctg(") ||
		strings.HasPrefix(op, "arccot(") { // особый приорите что бы избежать несовместимости с "**"
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
