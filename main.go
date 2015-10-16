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
  // Ключевые слова для открытия двери
  OpenDoorPhrases   []string `yaml:"open_door_phrases"`
  TurnLedOnPhrases  []string `yaml:"turn_led_on_phrases"`
  TurnLedOffPhrases []string `yaml:"turn_led_off_phrases"`
}

var bot *tgbotapi.BotAPI
var config *Config
var OpenDoorPhrases []string
var TurnLedOnPhrases []string
var TurnLedOffPhrases []string
var AllowedChatIds []int
var doorOpened chan *tgbotapi.Message
var ledTurnedOn chan *tgbotapi.Message
var ledTurnedOff chan *tgbotapi.Message
var doorPin = rpio.Pin(10)
var ledPin = rpio.Pin(9)

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
  ledPin.Output()
  doorPin.Output()

  // Инициализируем все остальные переменные 
  doorOpened = make(chan *tgbotapi.Message)
  ledTurnedOn = make(chan *tgbotapi.Message)
  ledTurnedOff = make(chan *tgbotapi.Message)
  AllowedChatIds = config.AllowedChatIds
  OpenDoorPhrases = config.OpenDoorPhrases
  TurnLedOnPhrases = config.TurnLedOnPhrases
  TurnLedOffPhrases = config.TurnLedOffPhrases
  log.Printf("Authorized on account %s", bot.Self.UserName)

  var ucfg tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
  ucfg.Timeout = 60
  err = bot.UpdatesChan(ucfg)

  // Слушаем события
  go Listen()
  ListenUpdates()
}

func OpenDoor() chan<- *tgbotapi.Message {
  // Открываем дверь
  go launchDoor()
  return (chan<-*tgbotapi.Message)(doorOpened)
} 

func TurnLedOn() chan<- *tgbotapi.Message {
  // Включаем светодиод
  ledPin.High()
  return (chan<-*tgbotapi.Message)(ledTurnedOn)
}

func TurnLedOff() chan<- *tgbotapi.Message {
  // Выключаем светодиод
  ledPin.Low()
  return (chan<-*tgbotapi.Message)(ledTurnedOff)
}

// Открывание двери
func launchDoor() {
  log.Println("door is beeing opened")
  doorPin.High()
  ledPin.High()
  time.Sleep(100*time.Millisecond)
  doorPin.Low()
  ledPin.Low()
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
      case msg := <- ledTurnedOn:
        reply := msg.From.FirstName + " включил(а) светодиод"
        send(msg.Chat.ID, reply)
      case msg := <- ledTurnedOff:
        reply := msg.From.FirstName + " выключил(а) светодиод"
        send(msg.Chat.ID, reply)
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
      if tryToDo(text, TurnLedOnPhrases) {
        TurnLedOn() <- &update.Message
      }
      if tryToDo(text, TurnLedOffPhrases) {
        TurnLedOff() <- &update.Message
      }
    }
  }
}