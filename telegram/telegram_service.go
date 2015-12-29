package telegram
import(
  "log"
  "github.com/Syfaro/telegram-bot-api"
  "bitbucket.org/kaikash/headmade_bot/config"
  "bitbucket.org/kaikash/headmade_bot/gpio"
)

var instantiated *TelegramService = nil

type TelegramService struct {
  config      *Config
  bot         *tgbotapi.BotAPI
  gpioService *gpio.GpioService
  phrases     Phrases
  onSudo      chan *tgbotapi.Message
  OnOpen      chan struct{}
}

// Input: configPath - path for config
//        phrasesPath - path for phrases
// Output: new TelegramService
// Creates TelegramService
func NewTelegramService(configPath, phrasesPath string) (*TelegramService, error) {
  var(
    err error
  )
  if instantiated == nil {
    telegramService := &TelegramService{}
    telegramService.onSudo = make(chan *tgbotapi.Message)
    telegramService.OnOpen = make(chan struct{})
    if telegramService.gpioService, err = gpio.NewGpioService(configPath); err != nil {
      return nil, err
    }
    if err = config.ReadConfig(configPath, &telegramService.config); err != nil {
      return nil, err 
    }
    if telegramService.bot, err = tgbotapi.NewBotAPI(telegramService.config.Token); err != nil {
      return nil, err
    }
    if err = config.ReadConfig(phrasesPath, &telegramService.phrases); err != nil {
      return nil, err
    }
    instantiated = telegramService
    return telegramService, nil
  }
  return instantiated, nil
}


// Input: chatId - Id of chat
//        message - an actuall message which will be sent
// Output: nothing
// Sends a message to bot
func (self *TelegramService) Send(chatId int, message string) {
  log.Println(message)
  bot_msg := tgbotapi.NewMessage(chatId, message)
  self.bot.Send(bot_msg)
}

func (self *TelegramService) Listen() {
  go self.gpioService.Listen(self.OnOpen)
  go self.ListenUpdates()
  for {
    select {
      case message := <- self.onSudo:
        chatId := message.Chat.ID
        log.Println("Only admins can use this function")
        self.Send(chatId, "Только администраторы могут использовать данную команду")
      case <- self.OnOpen:
        chatId := self.config.MainChatId
        log.Println("The door has been opened by button")
        self.Send(chatId, "Дверь была открыта кнопочкой")
    }
  }
}

func (self *TelegramService) ListenUpdates() {
  var u tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
  u.Timeout = 60
  updates, _ := self.bot.GetUpdatesChan(u)
  for update := range updates {
    userName, userId, chatId := update.Message.From.UserName, update.Message.From.ID, update.Message.Chat.ID
    if self.IsUserBlocked(userId) {
      go self.Send(chatId, userName + ", извините, но вы заблокированы :(")
      continue
    }

    if self.IsChatNotAllowed(chatId) {
      go self.Send(chatId, userName + ", вам нельзя писать боту в лс, пишите в общий чат")
    }

    self.phrases.CheckUpdate(&update.Message)
  }
}

func (self *TelegramService) IsUserBlocked(userId int) bool {
  return self.authId(userId, self.config.BlockedIds)
}

func (self *TelegramService) IsChatAllowed(chatId int) bool {
  return self.authId(chatId, self.config.AllowedChatIds)
}

func (self *TelegramService) IsChatNotAllowed(chatId int) bool {
  return !self.authId(chatId, self.config.AllowedChatIds)
}

func (self *TelegramService) IsUserAdmin(userId int) bool {
  return self.authId(userId, self.config.SudoersIds)
}

func (self *TelegramService) IsChatMain(chatId int) bool {
  return chatId == self.config.MainChatId
}

func (self *TelegramService) authId(id int, ids []int) bool {
  for i:=0; i<len(ids); i++ {
    if id == ids[i] {
      return true
    }
  }
  return false
}


