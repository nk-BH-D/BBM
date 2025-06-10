package linear

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/nikitakutergin59/BBM/bak/calculator"
	"github.com/nikitakutergin59/BBM/bak/equations/TW" // директория TW реализует пакет equations
)

// инвертируем уравнение
func InvertedEquations(tokens []equations.Token) ([]equations.Token, error) {
	equationsIndex := -1
	for i, token := range tokens {
		if token.Type == equations.TOKEN_OPERATOR && token.Value == "=" {
			equationsIndex = i
			break
		}
	}
	if equationsIndex == -1 {
		return nil, fmt.Errorf("оператор '=' не найден")
	}

	leftSide := tokens[:equationsIndex]
	rightSide := tokens[equationsIndex+1:]
	if len(rightSide) > 0 {
		if rightSide[0].Type == equations.TOKEN_NUMBER && !strings.HasPrefix(rightSide[0].Value, "-") {
			rightSide = append([]equations.Token{{Type: equations.TOKEN_OPERATOR, Value: "+"}}, rightSide...)
		}
		if rightSide[0].Type == equations.TOKEN_PARENT_OPEN && !strings.HasPrefix(rightSide[0].Value, "-") {
			rightSide = append([]equations.Token{{Type: equations.TOKEN_OPERATOR, Value: "+"}}, rightSide...)
		}
		if rightSide[0].Type == equations.TOKEN_VARIABLE && !strings.HasPrefix(rightSide[0].Value, "-") {
			rightSide = append([]equations.Token{{Type: equations.TOKEN_OPERATOR, Value: "+"}}, rightSide...)
		}
	}

	// инвертируем знаки в правой части
	invertedRightSide := make([]equations.Token, len(rightSide))
	for j, token := range rightSide {
		invertedToken := token
		switch token.Type {
		case equations.TOKEN_OPERATOR: // инвертация операторов
			switch token.Value {
			case "+":
				invertedToken.Value = "-"
			case "-":
				invertedToken.Value = "+"
			default:
				invertedToken = token // оставляем как есть
			}
		case equations.TOKEN_MULT:
			invertedToken = token // оставляем как есть
		case equations.TOKEN_NUMBER: // инвертация чисел
			if strings.HasPrefix(invertedToken.Value, "-") {
				invertedToken.Value = strings.Replace(invertedToken.Value, "-", "+", 1)
			} else {
				invertedToken.Value = strings.Replace(invertedToken.Value, "+", "-", 1)
			}
		case equations.TOKEN_PARENT_OPEN: // инвертация скобок
			if strings.HasPrefix(invertedToken.Value, "-") {
				invertedToken.Value = strings.Replace(invertedToken.Value, "-", "+", 1)
			} else {
				invertedToken.Value = strings.Replace(invertedToken.Value, "+", "-", 1)
			}
		case equations.TOKEN_VARIABLE: // инвертация переменных
			if strings.HasPrefix(invertedToken.Value, "-") {
				invertedToken.Value = strings.Replace(invertedToken.Value, "-", "+", 1)
			} else {
				invertedToken.Value = strings.Replace(invertedToken.Value, "+", "-", 1)
			}
		default:
			invertedToken = token // просто копируем токен, если это переменная или скобка
		}
		invertedRightSide[j] = invertedToken
	}
	// форматирование результатов
	result_inverted := make([]equations.Token, 0, len(leftSide)+len(invertedRightSide)+1)
	result_inverted = append(result_inverted, leftSide...)
	result_inverted = append(result_inverted, invertedRightSide...)
	log.Println("результат инвертации", result_inverted)
	return result_inverted, nil
}

// функция для раскрытия скобок
func OpenParent(result_inverted []equations.Token) ([]equations.Token, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Паника перехвачена: %v", r)
		}
	}()

	for i, openTheDoor := range result_inverted {
		if openTheDoor.Type == equations.TOKEN_PARENT_OPEN && openTheDoor.Value == "(" {
			// найти индекс закрывающей скобки
			closeIndex := -1
			openParent := 1
			for j := i + 1; j < len(result_inverted); j++ {
				if result_inverted[j].Type == equations.TOKEN_PARENT_OPEN && result_inverted[j].Value == "(" {
					openParent++
				} else if result_inverted[j].Type == equations.TOKEN_PARENT_CLOSE && result_inverted[j].Value == ")" {
					openParent--
					if openParent == 0 {
						closeIndex = j
						break
					}
				}
			}
			if closeIndex == -1 {
				return nil, fmt.Errorf("не найдена закрывающая скобка")
			}

			// рекурсивно раскрываем содержимое скобок
			parentSlice := result_inverted[i+1 : closeIndex]
			processedParentSlice, err := OpenParent(parentSlice)
			if err != nil {
				return nil, err
			}

			// ищем множитель перед скобкой: шаблон [NUMBER, *, (]
			multiplier := 1.0
			multiplierIndex := -1
			if i >= 2 {
				if result_inverted[i-2].Type == equations.TOKEN_NUMBER && result_inverted[i-1].Type == equations.TOKEN_MULT {
					val, err := strconv.ParseFloat(result_inverted[i-2].Value, 64)
					if err == nil {
						multiplier = val
						multiplierIndex = i - 2
						log.Printf("✅ Найден множитель: %f", multiplier)
					}
				}
			}
			// определяем тип числа коэфицент это ли просто число
			invertedTypeSlice := equations.WhatCfOrNumber(processedParentSlice)
			//приводим подобные слагаемые
			similarNumberSlice, similarCfSlice := equations.SimilarMembers(invertedTypeSlice)
			// проверяем что числа приведенны правильно
			if len(similarNumberSlice) > 2 {
				valueSlise := make([]string, 0)
				filtrSimilarNumberSlice := similarNumberSlice
				if similarNumberSlice[0].Value == "+" || similarNumberSlice[0].Value == "*" || similarNumberSlice[0].Value == "/" {
					filtrSimilarNumberSlice = similarNumberSlice[1:]
				}
				for i := range filtrSimilarNumberSlice {
					valueSlise = append(valueSlise, filtrSimilarNumberSlice[i].Value)
					if len(valueSlise) == len(filtrSimilarNumberSlice) {
						str_valueSlice := strings.Join(valueSlise, "")
						result_calculator, err := calculator.CalculatorTelegram(str_valueSlice)
						if err != nil {
							return nil, fmt.Errorf("ошибка вычисления: %w", err)
						}
						newFiltrSimilarNumberSlice, err := equations.Tokenize_BH(result_calculator)
						if err != nil {
							return nil, fmt.Errorf("ошибка повторной токенизации: %w", err)
						}
						log.Println("OP: получилось по итогу у чисел:", newFiltrSimilarNumberSlice)
					}
				}
			}
			// проверяем что коэффициенты приведенны правильно
			if len(similarCfSlice) >= 6 {
				valueSlise := make([]string, 0)
				filtrSimilarCfSlice := similarCfSlice
				if similarCfSlice[0].Value == "+" || similarCfSlice[0].Value == "*" || similarCfSlice[0].Value == "/" {
					filtrSimilarCfSlice = similarCfSlice[1:]
				}
				notUsedValue := 0
				for i := range filtrSimilarCfSlice {
					if filtrSimilarCfSlice[i].Type != equations.TOKEN_CF && filtrSimilarCfSlice[i].Type != equations.TOKEN_OPERATOR {
						notUsedValue++
					} else if filtrSimilarCfSlice[i].Type == equations.TOKEN_CF || filtrSimilarCfSlice[i].Type == equations.TOKEN_OPERATOR {
						valueSlise = append(valueSlise, filtrSimilarCfSlice[i].Value)
					}
				}
				//проверяем условие после завершения цикла
				if len(valueSlise) == len(filtrSimilarCfSlice)-notUsedValue {
					str_valueSlice := strings.Join(valueSlise, "")
					result_calculator, err := calculator.CalculatorTelegram(str_valueSlice)
					if err != nil {
						return nil, fmt.Errorf("ошибка вычисления: %w", err)
					}
					result_calculator = result_calculator + "x"
					newFiltrSimilarCfSlice, err := equations.Tokenize_BH(result_calculator)
					if err != nil {
						return nil, fmt.Errorf("ошибка повторной токенизации: %w", err)
					}
					log.Println("OP: получилось по итогу у кф:", newFiltrSimilarCfSlice)
				}
			}
			// умножаем содержимое скобок
			multipliedInnerSlice, err := equations.MultiplyInnerSlice(invertedTypeSlice, multiplier)
			if err != nil {
				return nil, fmt.Errorf("ошибка MultiplyInnerSlice: %w", err)
			}

			// собираем новое выражение
			var before []equations.Token
			if multiplierIndex != -1 {
				before = result_inverted[:multiplierIndex] // убираем и число, и *
			} else {
				before = result_inverted[:i]
			}

			var tail []equations.Token
			if closeIndex+1 < len(result_inverted) {
				tail = result_inverted[closeIndex+1:]
			} else {
				tail = []equations.Token{}
			}

			open_result_inverted := append(before, multipliedInnerSlice...)
			open_result_inverted = append(open_result_inverted, tail...)
			//log.Println("Результат раскрытия скобок", open_result_inverted)

			// запускаем рекурсивно, если есть другие скобки
			return OpenParent(open_result_inverted)
		}
	}
	return result_inverted, nil
}

// функция проверяет есть ли ещё скобки в уравнении, если есть, то перезапускает OpenParent
func OpenAllParent(all_open_result_inverted []equations.Token) ([]equations.Token, error) {
	result_inverted := all_open_result_inverted
	//log.Println("НЕ МЕНИЕ ВАЖНО", all_open_result_inverted)
	for {
		allLen := len(result_inverted)

		var err error
		result_inverted, err = OpenParent(result_inverted)
		if err != nil {
			return nil, err
		}
		//log.Println("ОЧЕНЬ ВАЖНО", result_inverted)

		if len(result_inverted) == allLen {
			break
		}
	}
	log.Println("конечный результат", result_inverted)
	return result_inverted, nil
}
