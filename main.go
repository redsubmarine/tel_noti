package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/redsubmarine/tel_noti/model"
)

func envDirectory() string {
	home, _ := os.UserHomeDir()
	path := "/.config/tel_noti"
	return home + path
}

func envFilePath() string {
	home, _ := os.UserHomeDir()
	path := "/.config/tel_noti/config.json"
	return home + path

}

func main() {

	log.SetFlags(0)

	didSetup := GetBoolFromFile(envFilePath())

	if !didSetup {
		SetupConfig()
		os.Exit(0)
	}

	config := GetConfig()
	text := GetMessage()

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", config.BotToken)
	m := model.Message{
		ChatID: config.ChatId,
		Text:   text,
	}
	SendMessage(url, &m)
}

func SetupConfig() {
	log.Print("telegram bot father로부터 발급받은 bot Token을 입력하세요:")
	token, err := InputText()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("알림을 받길 원하는 chat id 를 입력하세요:(참고 url: https://api.telegram.org/bot%s/getUpdates)", token)
	cid, err := InputText()
	if err != nil {
		log.Fatal(err)
	}

	chatID, err := strconv.Atoi(cid)
	if err != nil {
		log.Fatal(fmt.Errorf("chat id must be integer: %w", err))
	}

	config := &model.Config{
		BotToken: token,
		ChatId:   int64(chatID),
	}
	c, err := json.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}

	_ = os.MkdirAll(envDirectory(), os.ModePerm)

	f, err := os.Create(envFilePath())
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.Write(c)

	if err != nil {
		log.Fatal(err)
	}

	log.Print("설정이 완료되었습니다. $ tel_noti hello world 를 입력해보세요.")
}

func GetMessage() string {
	args := os.Args[1:]
	text := strings.Join(args, " ")
	if text == "" {
		text = "Hello, tel_noti!"
	}
	return text
}

func GetConfig() model.Config {
	file, err := os.Open(envFilePath())
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	b := make([]byte, 1024)
	n, err := file.Read(b)
	if err != nil {
		log.Fatal(err)
	}

	var config model.Config
	err = json.Unmarshal(b[:n], &config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

func GetBoolFromFile(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func SendMessage(url string, message *model.Message) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	response, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			log.Println("failed to close response body")
		}
	}(response.Body)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send successful request. Status was %q", response.Status)
	}
	return nil
}

func InputText() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", fmt.Errorf("failed to read input text")
}
