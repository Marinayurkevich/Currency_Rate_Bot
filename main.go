package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// ArrayCurrencyAbbreviation переменная для хранения массива допустимых валют
var ArrayCurrencyAbbreviation []string

// точка входа программы
func main() {
	//читаем json файл (в файле записанный Вами TOKEN, полученный от @BotFather)
	TOKEN, err := os.ReadFile("./TOKEN.json")
	if err != nil {
		log.Fatal(err)
	}
	// создаем переменную, где будет храниться Ваш TOKEN
	var botToken string
	err = json.Unmarshal(TOKEN, &botToken)
	if err != nil {
		log.Fatal(err)
	}

	// https://api.telegram.org/bot<token>/METHOD_NAME
	botApi := "https://api.telegram.org/bot"
	botUrl := botApi + botToken
	//bankApi := "https://www.nbrb.by/api/exrates/.\rates/usd?parammode=2"

	//читаем json файл (в файле список допустимых валют)
	by, err := os.ReadFile("./currency.json")
	if err != nil {
		log.Fatal(err)
	}
	// создаем переменную, где будет храниться список допустимых валют, список берем из json файла
	var JsonResponse []string
	err = json.Unmarshal(by, &JsonResponse)
	if err != nil {
		//завершит здесь программу и выведет ошибку, если она найдена
		log.Fatal(err)
	}
	// выводим список валют
	fmt.Println(JsonResponse)

	//переменная для хранения массива допустимых валют принимает значения из файла
	ArrayCurrencyAbbreviation = JsonResponse

	offset := 0
	for {
		updates, err := getUpdates(botUrl, offset)
		if err != nil {
			log.Println("smth went wrong: ", err.Error())
		}
		for _, update := range updates {
			err = respond(botUrl, update)
			offset = update.UpdateId + 1
		}
		fmt.Println(updates)
	}
}

// запрос обновлений
func getUpdates(botUrl string, offset int) ([]Update, error) {
	resp, err := http.Get(botUrl + "/getUpdates" + "?offset=" + strconv.Itoa(offset))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var restResponse RestResponse
	err = json.Unmarshal(body, &restResponse)
	if err != nil {
		return nil, err
	}
	return restResponse.Result, nil
}

// ответ на обновления
func respond(botUrl string, update Update) error {
	var botMessage BotMessage
	botMessage.ChatId = update.Message.Chat.ChatId
	botMessage.Text = update.Message.Text
	var CurrencyAbbreviation string
	var data CurrencyInfo

	//перебираем массив допустимых валют
	for _, ValueCurrency := range ArrayCurrencyAbbreviation {
		//если введенный пользователем текст соответсвует одной из допустимых валют
		if strings.ToUpper(update.Message.Text) == ValueCurrency {
			CurrencyAbbreviation = ValueCurrency
		} else {
			botMessage.Text = "Enter alphabetic code of currency unit"
		}
	}
	// выполняется, если пользователь ввел корректную валюту из списка
	if CurrencyAbbreviation != "" {
		// делаем запрос в банк о получении инфо о введеной пользователем валюте
		bankApi := "https://www.nbrb.by/api/exrates/rates/" + CurrencyAbbreviation + "?parammode=2"
		resp, err := http.Get(bankApi)
		if err != nil {
			panic(err.Error())
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err.Error())
		}
		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Results: %v\n", data)
		//fmt.Println(data.Cur_OfficialRate)
		botMessage.Text = fmt.Sprintf("%.4f", data.Cur_OfficialRate)
		_, err = http.Get(botUrl + "/sendMessage" + botMessage.Text)
		if err != nil {
			return err
		}

		// Отправляем ответ в Telegram при вводе допустимой валюты
		buf, err := json.Marshal(botMessage)
		if err != nil {
			return err
		}
		_, err = http.Post(botUrl+"/sendMessage", "application/json", bytes.NewBuffer(buf))
		if err != nil {
			return err
		}
	}

	// Отправляем ответ в Telegram если валюта введена НЕкорректно
	if CurrencyAbbreviation == "" {
		buf, err := json.Marshal(botMessage)
		if err != nil {
			return err
		}
		_, err = http.Post(botUrl+"/sendMessage", "application/json", bytes.NewBuffer(buf))
		if err != nil {
			return err
		}
	}
	return nil
}
