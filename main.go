package main
import (
  // Дефолтные пакаджи
  "log"
  "io/ioutil"
  "path/filepath"
  "strings"
  "time"
  "fmt"

  // Парсер yaml файлов
  "gopkg.in/yaml.v2"

  // Библиетка для работы с rpi
  "github.com/stianeikeland/go-rpio"

  // Библитека для работы с telegram api
  "github.com/Syfaro/telegram-bot-api"

  "github.com/d2r2/go-dht"
)

type Config struct {
  // Токен телеграм бота
  Token               string `yaml:"token"`
  // Разрешенные айдишники чатов
  AllowedChatIds      []int `yaml:"allowed_chat_ids"`
  MainChatId          int `yaml:"main_chat_id"`
  SudoersIds          []int `yaml:"sudoers_ids"`
  BlockDoorPhrases    []string `yaml:"block_door_phrases"`
  UnblockDoorPhrases  []string `yaml:"unblock_door_phrases"`
  // Ключевые слова для открытия двери
  OpenDoorPhrases     []string `yaml:"open_door_phrases"`
  GetTempPhrases      []string `yaml:"get_temp_phrases"`
  GetHumPhrases       []string `yaml:"get_hum_phrases"`
  BlockedIds          []int `yaml:"blocked_ids"`
}

var (
  bot *tgbotapi.BotAPI
  config *Config
  OpenDoorPhrases []string
  BlockDoorPhrases []string
  GetTempPhrases []string
  GetHumPhrases []string
  UnblockDoorPhrases []string
  AllowedChatIds []int
  SudoersIds []int
  MainChatId int
  BlockedIds []int
  doorOpened chan *tgbotapi.Message
  userBlocked chan *tgbotapi.Message
  doorOpenedByButton chan struct{}
  doorBlocked chan *tgbotapi.Message
  doorUnblocked chan *tgbotapi.Message
  getTemp chan *tgbotapi.Message
  getHum chan *tgbotapi.Message
  doorPin = rpio.Pin(10)
  doorReadPin = rpio.Pin(25)
  lockPin = rpio.Pin(9)
  blocked bool
  dhtSensor = dht.DHT11
  dhtPin = 4
  temperature float32
  humidity float32
)

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
  lockPin.Output()
  doorReadPin.Input()
  doorReadPin.PullUp()

  // Инициализируем все остальные переменные 
  blocked = false
  doorOpened = make(chan *tgbotapi.Message)
  userBlocked = make(chan *tgbotapi.Message)
  doorOpenedByButton = make(chan struct{})
  doorBlocked = make(chan *tgbotapi.Message)
  doorUnblocked = make(chan *tgbotapi.Message)
  getTemp = make(chan *tgbotapi.Message)
  getHum = make(chan *tgbotapi.Message)
  temperature = 0.0
  humidity = 0.0
  BlockedIds = config.BlockedIds
  AllowedChatIds = config.AllowedChatIds
  OpenDoorPhrases = config.OpenDoorPhrases
  BlockDoorPhrases = config.BlockDoorPhrases
  GetTempPhrases = config.GetTempPhrases
  GetHumPhrases = config.GetHumPhrases
  UnblockDoorPhrases = config.UnblockDoorPhrases
  SudoersIds = config.SudoersIds
  MainChatId = config.MainChatId
  log.Printf("Authorized on account %s", bot.Self.UserName)

  var ucfg tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
  ucfg.Timeout = 60
  err = bot.UpdatesChan(ucfg)

  // Слушаем события
  go Listen()
  go ListenDoor()
  go ListenDHTsensor()
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
    time.Sleep(time.Millisecond*10)
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
func authIds(chatId int, AllowedIds []int) bool {
  for i:=0; i<len(AllowedIds); i++ {
    if chatId == AllowedIds[i] {
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
        send(MainChatId, "Дверь была открыта")
      case msg := <- doorBlocked:
        send(msg.Chat.ID, "Дверь заблокирована")
      case msg := <- doorUnblocked:
        send(msg.Chat.ID, "Дверь разблокирована")
      case msg := <- getTemp:
        send(msg.Chat.ID, fmt.Sprintf("Температура на борту: %v °C", temperature))
      case msg := <- getHum:
        send(msg.Chat.ID, fmt.Sprintf("Влажность на борту: %v%%", humidity))
      case msg := <- userBlocked:
        reply := msg.From.FirstName + ", извините, но вы заблокированы :("
        send(msg.Chat.ID, reply)
    }
  }
}

func ListenDHTsensor() {
  var err error
  for {
    temperature, humidity, _, err =
      dht.ReadDHTxxWithRetry(dhtSensor, dhtPin, false, 10)
    if err != nil {
      log.Fatal(err)
    }
    time.Sleep(time.Second*10)
  }
}

func ListenUpdates() {
  for {
    select {
    case update := <-bot.Updates:
      userName := update.Message.From.UserName
      userId := update.Message.From.ID
      chatID := update.Message.Chat.ID
      text := update.Message.Text

      log.Println(text)
      log.Println(userId)

      if authIds(userId, BlockedIds) {
        userBlocked <- &update.Message
        continue
      }
      
      // Проверяем является ли этот чат разрешенным
      if !authIds(chatID, AllowedChatIds) {
        reply := "Вам нельзя это делать"
        log.Println(reply)
        bot_msg := tgbotapi.NewMessage(chatID, reply)
        bot.SendMessage(bot_msg)
        continue
      }

      log.Printf("[%s] %d %s", userName, chatID, text)
      // По очереди вытаемся выполнить какое-то действие
      if tryToDo(text, OpenDoorPhrases) && (blocked == true) {
        doorBlocked <- &update.Message
      } else if tryToDo(text, OpenDoorPhrases) {
        log.Println("door open")
        if (text == "откройся, мразь" || text == "откройся мразь") && (string(userName) != "Twizty") {
          reply := "Ошибка! Только Максиму можно со мной так обращаться!"
          log.Println(reply)
          bot_msg := tgbotapi.NewMessage(chatID, reply)
          bot.SendMessage(bot_msg)
          continue
        } else {
          OpenDoor() <- &update.Message
        }
      }

      if tryToDo(text, GetTempPhrases) {
        log.Println("get temperature")
        getTemp <- &update.Message
      }

      if tryToDo(text, GetHumPhrases) {
        log.Println("get humidity")
        getHum <- &update.Message
      }

      if authIds(userId, SudoersIds) && tryToDo(text, BlockDoorPhrases) {
        log.Println("door blocked")
        blocked = true
        lockPin.High()
        doorBlocked <- &update.Message
      }

      if authIds(userId, SudoersIds) && tryToDo(text, UnblockDoorPhrases) {
        log.Println("door unblocked")
        blocked = false
        lockPin.Low()
        doorUnblocked <- &update.Message
      }
    }
  }
}
