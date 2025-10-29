package crar

import (
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
)

// среднее арифметическое, дисперсия,
func cr_ar(array []float64) (float64, float64, float64) {
	sum := 0.0
	no_sum := float64(len(array))
	for i := 0; i < len(array); i++ {
		sum += array[i]
	}
	cRaR := sum / no_sum

	sum_sq_diff := 0.0
	for i := 0; i < len(array); i++ {
		diff := array[i] - cRaR
		sum_sq_diff += diff * diff
	}
	variance := sum_sq_diff / no_sum
	SqrtVariance := math.Sqrt(variance)
	return cRaR, variance, SqrtVariance
}

type StatsResult struct {
	Average          float64
	Variance         float64
	StdDeviation     float64
	Median           float64
	Max              float64
	Min              float64
	Range            float64
	Count            int
	RangeDescription string
}

// наименьшее наибольшее размах
type Stats struct {
	Max   float64
	Min   float64
	Range float64
}

func maXmiN(array []float64) Stats {
	if len(array) == 0 {
		return Stats{0, 0, 0}
	}
	Max := array[0]
	Min := array[0]
	for i := 1; i < len(array); i++ {
		if array[i] > Max {
			Max = array[i]
		} else if array[i] < Min {
			Min = array[i]
		}
	}
	rangeVal := Max - Min
	return Stats{Max, Min, rangeVal}
}

func calculateStats(input string) (StatsResult, error) {
	numbersStr := strings.Split(input, ",")
	numbers := make([]float64, 0, len(numbersStr))
	for _, numStr := range numbersStr {
		num, err := strconv.ParseFloat(strings.TrimSpace(numStr), 64)
		if err != nil {
			return StatsResult{}, fmt.Errorf("ошибка преобразования числа: %w", err)
		}
		numbers = append(numbers, num)
	}

	average, variance, stdDeviation := cr_ar(numbers)
	median := Median(numbers)
	stats := maXmiN(numbers)
	rangeDesc := ""
	if stats.Range > stats.Min {
		rangeDesc = "значительный"
	} else if stats.Range == 0 {
		rangeDesc = ""
	} else if stats.Range == stats.Min {
		rangeDesc = "нормальный"
	} else {
		rangeDesc = "незначительный"
	}

	return StatsResult{
		Average:          average,
		Variance:         variance,
		StdDeviation:     stdDeviation,
		Median:           median,
		Max:              stats.Max,
		Min:              stats.Min,
		Range:            stats.Range,
		Count:            len(numbers),
		RangeDescription: rangeDesc,
	}, nil
}

func Median(array []float64) float64 {
	NewArray := make([]float64, len(array))
	copy(NewArray, array) // cоздаем копию, чтобы не менять исходный массив
	sort.Float64s(NewArray)
	if len(NewArray)%2 != 0 {
		return NewArray[len(NewArray)/2]
	} else {
		left_central_element := NewArray[len(NewArray)/2-1]
		right_central_element := NewArray[len(NewArray)/2]
		return (left_central_element + right_central_element) / 2
	}
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

func StatsTelegram(input string) (string, error) {
	result, err := calculateStats(input)
	if err != nil {
		return "", fmt.Errorf("ошибка вычисления статистики: %w", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Количество чисел: %d\n", result.Count))
	sb.WriteString(fmt.Sprintf("Среднее арифметическое: %s\n", formatBigFloat(big.NewFloat(result.Average))))
	sb.WriteString(fmt.Sprintf("Дисперсия: %s\n", formatBigFloat(big.NewFloat(result.Variance))))
	sb.WriteString(fmt.Sprintf("Стандартное отклонение: %s\n", formatBigFloat(big.NewFloat(result.StdDeviation))))
	sb.WriteString(fmt.Sprintf("Медиана: %s\n", formatBigFloat(big.NewFloat(result.Median))))
	sb.WriteString(fmt.Sprintf("Максимальное значение: %s\n", formatBigFloat(big.NewFloat(result.Max))))
	sb.WriteString(fmt.Sprintf("Минимальное значение: %s\n", formatBigFloat(big.NewFloat(result.Min))))
	sb.WriteString(fmt.Sprintf("Размах: %s (%s)\n", formatBigFloat(big.NewFloat(result.Range)), result.RangeDescription))
	return sb.String(), nil
}
