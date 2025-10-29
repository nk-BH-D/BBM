package BBM

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	crar "github.com/nk-BH-D/BBM/bak/cr_ar"

	"github.com/nk-BH-D/BBM/bak/bezu"
	"github.com/nk-BH-D/BBM/bak/calculator"
	"github.com/nk-BH-D/BBM/bak/diskriminant"
	"github.com/nk-BH-D/BBM/bak/frequency"
)

var userStates = make(map[int64]map[string]string)

// вывод документации в текстовом формате
func docx_txt(_ string) (string, error) {
	file_in, err := os.Open("docx.txt")
	if err != nil {
		return "", fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file_in.Close()
	var content bytes.Buffer
	buffer := make([]byte, 2048)
	for {
		file_out, err := file_in.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("ошибка при чтении файла: %w", err)
		}
		content.Write(buffer[:file_out])
	}
	return content.String(), nil
}

// sendDocxFile отправляет файл .docx в Telegram чат.
func sendDocxFile(bot *tgbotapi.BotAPI, chatID int64, filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("файл '%s' не найден: %w", filePath, err)
	}

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл '%s': %w", filePath, err)
	}
	// получаем имя файла из пути
	fileName := filepath.Base(filePath)
	// создаем объект DocumentConfig для отправки в Telegram
	documentConfig := tgbotapi.NewDocument(chatID, tgbotapi.FileBytes{
		Name:  fileName,
		Bytes: fileBytes,
	})

	documentConfig.Caption = "Документация"

	_, err = bot.Send(documentConfig)
	if err != nil {
		return fmt.Errorf("не удалось отправить файл '%s' в чат %d: %w", filePath, chatID, err)
	}
	//log.Printf("Файл '%s' успешно отправлен в чат %d", filePath, chatID)
	return nil
}

func HandleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	text := strings.TrimSpace(message.Text)

	//этот блок НЕОБХОДИМ! он обрабатывает переходы между состояниями
	if state, ok := userStates[chatID]; ok {
		for key := range state {
			switch key {
			case "calc":
				handleCalcInput(bot, chatID, text)
			case "bezu":
				handleBezuInput(bot, chatID, text)
			case "discriminant":
				handleDiscriminantInput(bot, chatID, text)
			case "stats":
				handleStatsInput(bot, chatID, text)
				handleFrequencyInput(bot, chatID, text)
			default:
				delete(userStates, chatID)
				bot.Send(tgbotapi.NewMessage(chatID, "Неизвестная команда"))
			}
			return
		}
	}

	if text == "/docx_docx" {
		userStates[chatID] = map[string]string{"docx_docx": ""}
		docxFilePath := "docx.docx"
		msg := tgbotapi.NewMessage(chatID, "Внимательно изучите содержимое файла")
		sendDocxFile(bot, chatID, docxFilePath)
		msg.ReplyMarkup = createMenuKeyboard()
		bot.Send(msg)
		return
	}

	if text == "/docx_txt" {
		userStates[chatID] = map[string]string{"docx_txt": ""}
		msg := tgbotapi.NewMessage(chatID, "Внимательно прочитайте инструкцию и изучите операторы")
		bot.Send(msg)
		handleTxtDocxInput(bot, chatID, text)
		return
	}

	if text == "/start" {
		showMainMenu(bot, chatID)
		return
	}

	bot.Send(tgbotapi.NewMessage(chatID, "Неизвестная команда."))
}

// функция для отправки сообщения об ошибке помогает избежать дублирования кода
func sendErrorMessage(bot *tgbotapi.BotAPI, chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ReplyMarkup = createMenuKeyboard()
	_, sendErr := bot.Send(msg)
	if sendErr != nil {
		log.Printf("ошибка отправки сообщения: %v, ошибка: %v", message, sendErr)
		return
	}
}

// функция для отправки сообщений с ответом, тоже для избежания дублирования кода в коскаде функций ниже
func sendSuccessMessage(bot *tgbotapi.BotAPI, chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ReplyMarkup = createMenuKeyboard()
	_, sendErr := bot.Send(msg)
	if sendErr != nil {
		log.Printf("ошибка отправки сообщения: %v, ошибка: %v", message, sendErr)
		sendSuccessMessage(bot, chatID, "произошла ошибка при отправке сообщения, повторите попытку")
		return
	}
}

// каскад функций отвечающие за отправку сообщений для каждого состояния
func handleCalcInput(bot *tgbotapi.BotAPI, chatID int64, expression string) {
	result, err := calculator.CalculatorTelegram(expression)
	if err != nil {
		sendErrorMessage(bot, chatID, fmt.Sprintf("Ошибка вычисления: \n%v", err))
	} else {
		sendSuccessMessage(bot, chatID, fmt.Sprintln(result))
	}
}
func handleBezuInput(bot *tgbotapi.BotAPI, chatID int64, expression string) {
	result, err := bezu.BezuTelegram(expression)
	if err != nil {
		sendErrorMessage(bot, chatID, fmt.Sprintf("Ошибка вычисления: \n%v", err))
	} else {
		sendSuccessMessage(bot, chatID, fmt.Sprintf("Результат: \n%s", result))
	}

}
func handleDiscriminantInput(bot *tgbotapi.BotAPI, chatID int64, expression string) {
	result, err := diskriminant.DiscriminantFromString(expression)
	if err != nil {
		sendErrorMessage(bot, chatID, fmt.Sprintf("Ошибка вычисления: \n%v", err))
	} else {
		sendSuccessMessage(bot, chatID, fmt.Sprintf("Результат: \n%s", result))
	}
}
func handleStatsInput(bot *tgbotapi.BotAPI, chatID int64, expression string) {
	result, err := crar.StatsTelegram(expression)
	if err != nil {
		sendErrorMessage(bot, chatID, fmt.Sprintf("Ошибка вычисления: \n%v", err))
	} else {
		sendSuccessMessage(bot, chatID, fmt.Sprintf("Результат: \n%s", result))
	}
}
func handleFrequencyInput(bot *tgbotapi.BotAPI, chatID int64, expression string) {
	result, err := frequency.CalculateFrequency(expression)
	if err != nil {
		sendErrorMessage(bot, chatID, fmt.Sprintf("Ошибка вычисления: \n%v", err))
	} else {
		sendSuccessMessage(bot, chatID, fmt.Sprintf("Результат: \n%s", result))
	}
}
func handleTxtDocxInput(bot *tgbotapi.BotAPI, chatID int64, expression string) {
	result, err := docx_txt(expression)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Ошибка:\n%v", err))
		msg.ReplyMarkup = createMenuKeyboard()
		_, sendErr := bot.Send(msg)
		if sendErr != nil {
			log.Printf("ошибка отправки документации: %v", sendErr)
		}
	} else {
		msg := tgbotapi.NewMessage(chatID, result)
		msg.ReplyMarkup = createMenuKeyboard()
		_, sendErr := bot.Send(msg)
		if sendErr != nil {
			log.Printf("ошибка отправки документации: %v", sendErr)
		}
	}
}

// фронтенд часть там надписи на кнопочках и всё такое можно ещё смайлики добавить анимированные и вобще будут хорошо
func HandleCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	chatID := callbackQuery.Message.Chat.ID
	data := callbackQuery.Data

	switch data {
	case "show_menu":
		delete(userStates, chatID) // выход из всех режимов
		msg := tgbotapi.NewMessage(chatID, "Выберите режим:")
		msg.ReplyMarkup = createCommandsKeyboard()
		bot.Send(msg)
	case "two_variable_inequalities":
		msg := tgbotapi.NewMessage(chatID, "Находиться в процессе разработки:")
		msg.ReplyMarkup = createMenuKeyboard()
		bot.Send(msg)
	case "one_variable_inequalities":
		msg := tgbotapi.NewMessage(chatID, "Находиться в процессе разработки:")
		msg.ReplyMarkup = createMenuKeyboard()
		bot.Send(msg)
	case "two_variable_equations":
		msg := tgbotapi.NewMessage(chatID, "Находиться в процессе разработки:")
		msg.ReplyMarkup = createMenuKeyboard()
		bot.Send(msg)
	case "one_variable_equations":
		//userStates[chatID] = map[string]string{"one_variable_equations": ""}
		msg := tgbotapi.NewMessage(chatID, "Находиться в процессе разработки:")
		msg.ReplyMarkup = createMenuKeyboard()
		bot.Send(msg)
	case "calc":
		userStates[chatID] = map[string]string{"calc": ""}
		msg := tgbotapi.NewMessage(chatID, "Введите математическое выражение:")
		msg.ReplyMarkup = createMenuKeyboard()
		bot.Send(msg)
	case "bezu":
		userStates[chatID] = map[string]string{"bezu": ""}
		msg := tgbotapi.NewMessage(chatID, "Введите коэффициенты кубического уравнения (a b c d):")
		msg.ReplyMarkup = createMenuKeyboard()
		bot.Send(msg)
	case "discriminant":
		userStates[chatID] = map[string]string{"discriminant": ""}
		msg := tgbotapi.NewMessage(chatID, "Введите коэфиценты квадратного уравнения (a b c):")
		msg.ReplyMarkup = createMenuKeyboard()
		bot.Send(msg)
	case "stats":
		userStates[chatID] = map[string]string{"stats": ""}
		msg := tgbotapi.NewMessage(chatID, "Введите список чисел (через запятую):")
		msg.ReplyMarkup = createMenuKeyboard()
		bot.Send(msg)
	case "inequalities":
		userStates[chatID] = map[string]string{"inequalities": ""}
		msg := tgbotapi.NewMessage(chatID, "Выберите количество переменных в вашем неравенстве:")

		//объядинение клавиатур
		buttoms := append(createInequalitiesComandKeybord().InlineKeyboard, createMenuKeyboard().InlineKeyboard...)

		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttoms...)
		bot.Send(msg)
	case "equations":
		userStates[chatID] = map[string]string{"equations": ""}
		msg := tgbotapi.NewMessage(chatID, "Выберите количество переменных в вашем уравнении:")
		//объядинение клавиатур
		buttoms := append(createEquationsComandKeybord().InlineKeyboard, createMenuKeyboard().InlineKeyboard...)

		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttoms...)
		bot.Send(msg)
	default:
		bot.Send(tgbotapi.NewMessage(chatID, "Неизвестная команда"))
	}
}

func createMenuKeyboard() *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Меню", "show_menu"),
		),
	)
	return &keyboard
}

func createInequalitiesComandKeybord() *tgbotapi.InlineKeyboardMarkup {
	keybord := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("x, y переменные", "two_variable_inequalities"),
			tgbotapi.NewInlineKeyboardButtonData("x переменная", "one_variable_inequalities"),
		),
	)
	return &keybord
}

func createEquationsComandKeybord() *tgbotapi.InlineKeyboardMarkup {
	keybord := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("x, y переменные", "two_variable_equations"),
			tgbotapi.NewInlineKeyboardButtonData("x переменная", "one_variable_equations"),
		),
	)
	return &keybord
}

func createCommandsKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Калькулятор", "calc"),
			tgbotapi.NewInlineKeyboardButtonData("Кубические уравнения", "bezu"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Квадратные уравнения", "discriminant"),
			tgbotapi.NewInlineKeyboardButtonData("Статистика", "stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Неравенства", "inequalities"),
			tgbotapi.NewInlineKeyboardButtonData("Уравнения", "equations"),
		),
	)
}

func showMainMenu(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Привет! Я математический бот. Нажмите /docx_docx что бы ознакомиться с инструкцией в формате файла (.docx) или\n/docx_txt что бы получить инструкшию в текстовом формате:")
	bot.Send(msg)
}
