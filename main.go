package main
import (
  // Дефолтные пакаджи
  "log"
  "io/ioutil"
  "path/filepath"
  "strings"
  "time"

  // Парсер yaml файлов
  "gopkg.in/yaml.v2"

  // Библиетка для работы с rpi
  "github.com/stianeikeland/go-rpio"

  // Библитека для работы с telegram api
  "github.com/Syfaro/telegram-bot-api"
)

type Config struct {
  // Токен телеграм бота
  Token             string `yaml:"token"`
  // Разрешенные айдишники чатов
  AllowedChatIds    []int `yaml:"allowed_chat_ids"`
  MainChatId        int `yaml:"main_chat_id"`
  // Ключевые слова для открытия двери
  OpenDoorPhrases   []string `yaml:"open_door_phrases"`
}

var bot *tgbotapi.BotAPI
var config *Config
var OpenDoorPhrases []string
var TurnLedOnPhrases []string
var TurnLedOffPhrases []string
var AllowedChatIds []int
var doorOpened chan *tgbotapi.Message
var doorOpenedByButton chan struct{}
var doorPin = rpio.Pin(10)
var doorReadPin = rpio.Pin(25)
// var ledPin = rpio.Pin(9)

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

  // Для работы с gpio в rpi
  if err = rpio.Open(); err != nil {
    log.Panic(err)
  }
  defer rpio.Close()
  // Устанавливаем пины на output
  doorPin.Output()
  doorReadPin.Input()
  doorReadPin.PullUp()

  // Инициализируем все остальные переменные 
  doorOpened = make(chan *tgbotapi.Message)
  doorOpenedByButton = make(chan struct{})
  AllowedChatIds = config.AllowedChatIds
  OpenDoorPhrases = config.OpenDoorPhrases
  log.Printf("Authorized on account %s", bot.Self.UserName)

  var ucfg tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
  ucfg.Timeout = 60
  err = bot.UpdatesChan(ucfg)

  // Слушаем события
  go Listen()
  go ListenDoor()
  ListenUpdates()
}

func OpenDoor() chan<- *tgbotapi.Message {
  // Открываем дверь
  go launchDoor()
  return (chan<-*tgbotapi.Message)(doorOpened)
} 

// Открывание двери
func launchDoor() {
  log.Println("door is beeing opened")
  doorPin.High()
  time.Sleep(1000*time.Millisecond)
  doorPin.Low()
}

func ListenDoor() {
  log.Println("Listen door")
  for {
    if doorReadPin.Read() == 0 {
      log.Println("Door has been opened")
      doorOpenedByButton <- struct{}{}
      time.Sleep(time.Second*3)
    }
  }
}

// Проверяет, является ли указанное сообщение ключевым
func tryToDo(text string, phrases []string) bool {
  for i:=0; i<len(phrases); i++ {
    if strings.ToLower(text) == phrases[i] {
      return true
    }
  }
  return false
}

// Проверяет, является ли указанное сообщение разрешенным
func auth(chatId int) bool {
  for i:=0; i<len(AllowedChatIds); i++ {
    if chatId == AllowedChatIds[i] {
      return true
    }
  }
  return false
}

// Отправка сообщения
func send(chatId int, msg string) {
  log.Println(msg)
  bot_msg := tgbotapi.NewMessage(chatId, msg)
  bot.SendMessage(bot_msg)
}

func Listen() {
  for {
    select {
      case msg := <- doorOpened:
        reply := msg.From.FirstName + " открыл(а) дверь"
        send(msg.Chat.ID, reply)
      case <- doorOpenedByButton:
        send(config.MainChatId, "Дверь была открыта")
    }
  }
}

func ListenUpdates() {
  for {
    select {
    case update := <-bot.Updates:
      userName := update.Message.From.UserName
      chatID := update.Message.Chat.ID
      text := update.Message.Text
      // Проверяем является ли этот чат разрешенным
      if !auth(chatID) {
        reply := "Вам нельзя это делать"
        log.Println(reply)
        bot_msg := tgbotapi.NewMessage(chatID, reply)
        bot.SendMessage(bot_msg)
        continue
      }

      log.Printf("[%s] %d %s", userName, chatID, text)
      // По очереди вытаемся выполнить какое-то действие
      if tryToDo(text, OpenDoorPhrases) {
        OpenDoor() <- &update.Message
      }
    }
  }
}