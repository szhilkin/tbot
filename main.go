package main
import (
  "github.com/Syfaro/telegram-bot-api"
  "log"
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "path/filepath"
  "strings"

  // "github.com/stianeikeland/go-rpio"
)

type Config struct {
  Token string `yaml:"token"`
}

var bot *tgbotapi.BotAPI
var config *Config
var OpenDoorPhrases []string
var TurnLedOnPhrases []string
var TurnLedOffPhrases []string
var doorOpened chan *tgbotapi.Message
var ledTurnedOn chan *tgbotapi.Message
var ledTurnedOff chan *tgbotapi.Message

// var pin = rpio.Pin(10)

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

  // if err = rpio.Open(); err != nil {
    // log.Panic(err)
  // }
  // defer rpio.Close()
  // pin.Output()

  doorOpened = make(chan *tgbotapi.Message)
  ledTurnedOn = make(chan *tgbotapi.Message)
  ledTurnedOff = make(chan *tgbotapi.Message)
  OpenDoorPhrases = []string{"open", "open the door", "open door", "door open", "дверь откройся", "открыть дверь", "открыть"}
  TurnLedOnPhrases = []string{"led on"}
  TurnLedOffPhrases = []string{"led off"}

  // bot.Debug = true
  log.Printf("Authorized on account %s", bot.Self.UserName)

  var ucfg tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
  ucfg.Timeout = 60
  err = bot.UpdatesChan(ucfg)
  go Listen()
  ListenUpdates()
}

func OpenDoor() chan<- *tgbotapi.Message {
  log.Println("door is beeing opened")
  return (chan<-*tgbotapi.Message)(doorOpened)
} 

func TurnLedOn() chan<- *tgbotapi.Message {
  // pin.High()
  return (chan<-*tgbotapi.Message)(ledTurnedOn)
}

func TurnLedOff() chan<- *tgbotapi.Message {
  // pin.Low()
  return (chan<-*tgbotapi.Message)(ledTurnedOff)
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
        reply := msg.From.UserName + " открыл(а) дверь"
        log.Println(reply)
        bot_msg := tgbotapi.NewMessage(msg.Chat.ID, reply)
        bot.SendMessage(bot_msg)
      case msg := <- ledTurnedOn:
        reply := msg.From.UserName + "  включил(а) светодиод"
        log.Println(reply)
        bot_msg := tgbotapi.NewMessage(msg.Chat.ID, reply)
        bot.SendMessage(bot_msg)
      case msg := <- ledTurnedOff:
        reply := msg.From.UserName + " выключил(а) светодиод"
        log.Println(reply)
        bot_msg := tgbotapi.NewMessage(msg.Chat.ID, reply)
        bot.SendMessage(bot_msg)
    }
  }
}

func ListenUpdates() {
  for {
    select {
    case update := <-bot.Updates:
      userName := update.Message.From.UserName
      // UserID := update.Message.From.ID
      chatID := update.Message.Chat.ID
      text := update.Message.Text
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
      // log.Println(openDoorPhrases[0])
      // reply := Text
      // msg := tgbotapi.NewMessage(ChatID, reply)
      // bot.SendMessage(msg)
    }
  }
}