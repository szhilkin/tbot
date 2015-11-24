package main
import (
  // Дефолтные пакаджи
  "log"
  "io/ioutil"
  "path/filepath"
  // "time"
  "fmt"
  "io"
  "net/http"
  "os"
  "strings"
  "bytes"
  "os/exec"

  // Парсер yaml файлов
  "gopkg.in/yaml.v2"

  // Библитека для работы с telegram api
  "github.com/Syfaro/telegram-bot-api"
)

type Config struct {
  // Токен телеграм бота
  Token               string `yaml:"token"`
}

var bot *tgbotapi.BotAPI
var key = "AIzaSyA4JqD8Xx6yjkZZRo9W4WZT5XV2ipLoWHw"
var config *Config

func readConfig() (*Config, error) {
  var yamlFile []byte
  var err error
  filename, _ := filepath.Abs("./config.yml")
  yamlFile, err = ioutil.ReadFile(filename)
  if err != nil {
    return nil, err
  }
  var conf Config
  if err := yaml.Unmarshal(yamlFile, &conf); err != nil {
    return nil, err
  }
  return &conf, err
}

func main() {
  var err error
  // Читаем конфиг
  if config, err = readConfig(); err != nil {
    panic(err)
  }

  // Инициализируем бота
  bot, err = tgbotapi.NewBotAPI(config.Token)
  if err != nil {
    log.Panic(err)
  }

  log.Printf("Authorized on account %s", bot.Self.UserName)

  var ucfg tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
  ucfg.Timeout = 60
  err = bot.UpdatesChan(ucfg)
  ListenUpdates()
}

// Отправка сообщения
func send(chatId int, msg string) {
  log.Println(msg)
  bot_msg := tgbotapi.NewMessage(chatId, msg)
  bot.SendMessage(bot_msg)
}

func DownloadFromUrl(url string) (string, error) {
  var err error
  tokens := strings.Split(url, "/")
  fileName := tokens[len(tokens)-1]
  log.Println("Downloading", url, "to", fileName)

  output, err := os.Create(fileName)
  if err != nil {
    log.Println("Error while creating", fileName, "-", err)
    return "", err
  }
  defer output.Close()

  response, err := http.Get(url)
  if err != nil {
    log.Println("Error while downloading", url, "-", err)
    return "", err
  }
  defer response.Body.Close()

  n, err := io.Copy(output, response.Body)
  if err != nil {
    log.Println("Error while downloading", url, "-", err)
    return "", err
  }

  log.Println(n, "bytes downloaded.")
  return fileName, nil
}

func Translate(file string) string {
  var key string = "AIzaSyASaHjzUyLpcm-nmcCH4Q_6M3oflvdNFVc"
  var url string = "https://www.google.com/speech-api/v2/recognize?output=json&lang=ru-ru&key=" + key
  stream, err := ioutil.ReadFile(file)
  req, err := http.NewRequest("POST", url, bytes.NewBuffer(stream))
  if err != nil {
    panic(err)
  }

  req.Header.Set("Content-Type", "audio/x-flac; rate=44100;")
  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    panic(err)
  }
  defer resp.Body.Close()

  var body, _ = ioutil.ReadAll(resp.Body)
  return string(body)
}

func ListenUpdates() {
  for {
    select {
    case update := <-bot.Updates:
      // userName := update.Message.From.UserName
      // userId := update.Message.From.ID
      // chatID := update.Message.Chat.ID
      // text := update.Message.Text

      // fmt.Printf("%#v\n", update)
      if update.Message.Voice.FileID != "" {
        log.Println("voice")
        file,_ := bot.GetFile(tgbotapi.FileConfig{update.Message.Voice.FileID})
        url := file.Link(config.Token)
        log.Println(url)
        fileName, _ := DownloadFromUrl(url)
        _, err := exec.Command("/usr/local/bin/ffmpeg", "-i", fileName, fileName + ".flac").Output()
        if err != nil {
          log.Fatal(err)
        }
        fileName = fileName+".flac"
        res := Translate(fileName)
        // fmt.Printf("%#v\n", file)
        // fmt.Println("%#v\n", res)
        fmt.Println(res)
        // log.Println(res.Fi)
      }
      // log.Println("user id: ", userId)
      // log.Println("chat id: ", chatID)
      // log.Println("")
    }
  }
}