package main
import (
  "github.com/Syfaro/telegram-bot-api"
  "log"
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "path/filepath"
  "strings"
  "time"
  "github.com/stianeikeland/go-rpio"
)

type Config struct {
  Token string `yaml:"token"`
}

var bot *tgbotapi.BotAPI
var config *Config
var OpenDoorPhrases []string
var TurnLedOnPhrases []string
var TurnLedOffPhrases []string
var WhiteChatIds []int
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
  if config, err = readConfig(); err != nil {
    panic(err)
  }

  bot, err = tgbotapi.NewBotAPI(config.Token)
  if err != nil {
    log.Panic(err)
  }

  if err = rpio.Open(); err != nil {
    log.Panic(err)
  }
  defer rpio.Close()
  ledPin.Output()
  doorPin.Output()

  doorOpened = make(chan *tgbotapi.Message)
  ledTurnedOn = make(chan *tgbotapi.Message)
  ledTurnedOff = make(chan *tgbotapi.Message)
  WhiteChatIds = []int{50815686, -33208400}
  OpenDoorPhrases = []string{"open", "open the door", "open door", "door open", "дверь откройся", "открыть дверь", "открыть", "откройся, мразь", "откройся мразь", "открыть", "откройся", "открой", "сим-сим, откройся"}
  TurnLedOnPhrases = []string{"led on", "test on"}
  TurnLedOffPhrases = []string{"led off", "test off"}
  log.Printf("Authorized on account %s", bot.Self.UserName)

  var ucfg tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
  ucfg.Timeout = 60
  err = bot.UpdatesChan(ucfg)
  go Listen()
  ListenUpdates()
}

func OpenDoor() chan<- *tgbotapi.Message {
  go launchDoor()
  return (chan<-*tgbotapi.Message)(doorOpened)
} 

func TurnLedOn() chan<- *tgbotapi.Message {
  ledPin.High()
  return (chan<-*tgbotapi.Message)(ledTurnedOn)
}

func TurnLedOff() chan<- *tgbotapi.Message {
  ledPin.Low()
  return (chan<-*tgbotapi.Message)(ledTurnedOff)
}

func launchDoor() {
  log.Println("door is beeing opened")
  doorPin.High()
  ledPin.High()
  time.Sleep(100*time.Millisecond)
  doorPin.Low()
  ledPin.Low()
}

func tryToDo(text string, phrases []string) bool {
  for i:=0; i<len(phrases); i++ {
    if strings.ToLower(text) == phrases[i] {
      return true
    }
  }
  return false
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

func send(chatId int, msg string) {
  log.Println(msg)
  bot_msg := tgbotapi.NewMessage(chatId, msg)
  bot.SendMessage(bot_msg)
}

func ListenUpdates() {
  for {
    select {
    case update := <-bot.Updates:
      userName := update.Message.From.UserName
      // UserID := update.Message.From.ID
      chatID := update.Message.Chat.ID
      text := update.Message.Text
      if !auth(chatID) {
        reply := "Вам нельзя это делать"
        log.Println(reply)
        bot_msg := tgbotapi.NewMessage(chatID, reply)
        bot.SendMessage(bot_msg)
        continue
      }
      log.Println(chatID)
      log.Printf("[%s] %d %s", userName, chatID, text)
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

func auth(chatId int) bool {
  for i:=0; i<len(WhiteChatIds); i++ {
    if chatId == WhiteChatIds[i] {
      return true
    }
  }
  return false
}