
package telegram
import(
  "log"
  "fmt"
  "time"
  "github.com/Syfaro/telegram-bot-api"
  "bitbucket.org/kaikash/headmade_bot/gpio"
)
func (self *TelegramService) RunAction(action string, message *tgbotapi.Message) {
  log.Println(action)
  switch action {
    case "open_door":
      self.openDoor(message, self.gpioService)
    case "get_temp":
      self.getTemp(message, self.gpioService)
    case "get_hum":
      self.getHum(message, self.gpioService)
    case "sudo_block_door":
      self.lockDoor(message, self.gpioService)
    case "sudo_unblock_door":
      self.unlockDoor(message, self.gpioService)
  }  
}

func (self *TelegramService) lockDoor(message *tgbotapi.Message, gpioService *gpio.GpioService) {
  chatId := message.Chat.ID
  if err := gpioService.LockDoor(); err != nil {
    self.Send(chatId, "Хьюстон! У нас проблемы! Не получилось запереть дверь")
    log.Println(err)
    return
  }
  self.Send(chatId, "Cэр, дверь заперта!")
}

func (self *TelegramService) unlockDoor(message *tgbotapi.Message, gpioService *gpio.GpioService) {
  chatId := message.Chat.ID
  if err := gpioService.UnlockDoor(); err != nil {
    self.Send(chatId, "Хьюстон! У нас проблемы! Не получилось отпереть дверь")
    log.Println(err)
    return
  }
  self.Send(chatId, "Cэр, дверь больше не заперта!")
}

func (self *TelegramService) getTemp(message *tgbotapi.Message, gpioService *gpio.GpioService) {
  chatId := message.Chat.ID
  var(
    temp float32
    err error
  )
  if temp, err = gpioService.GetTemp(); err != nil {
    self.Send(chatId, "Хьюстон! У нас проблемы! Не получилось узнать температуру")
    log.Println(err)
    return
  }
  self.Send(chatId, fmt.Sprintf("Капитан! На борту %v °C", temp))
}

func (self *TelegramService) getHum(message *tgbotapi.Message, gpioService *gpio.GpioService) {
  chatId := message.Chat.ID
  var(
    hum float32
    err error
  )
  if hum, err = gpioService.GetHum(); err != nil {
    self.Send(chatId, "Хьюстон! У нас проблемы! Не получилось узнать влажность")
    log.Println(err)
    return
  }
  self.Send(chatId, fmt.Sprintf("Капитан! Влажность на борту %v%%", hum))
}

func (self *TelegramService) openDoor(message *tgbotapi.Message, gpioService *gpio.GpioService) {
  chatId := message.Chat.ID
  
  if gpioService.IsBlocked() == true {
    self.Send(chatId, "Дверь заперта")
    return
  }
  go gpioService.OpenDoor()//; err != nil {
    //self.Send(chatId, "К сожалению не получилось открыть дверь :(")
     // log.Println(err)
     // return
  //}
  if chatId != self.config.MainChatId {
    go self.Send(self.config.MainChatId, message.From.FirstName + " открыл(а) дверь")
  }
  msg := self.getOpenReply(message.From.FirstName)
  self.Send(chatId, msg)
}

func (self *TelegramService) getOpenReply(name string) string {
  log.Println(time.Now().Hour())
  if time.Now().Hour()<6 {
    return "Антон, это опять ты шарахаешься?"
  } else if time.Now().Hour()<10 {
    return name + ", доброе утро!"
  } else if time.Now().Hour()<17 {
    return name + ", хорошего тебе дня!"
  } else if time.Now().Hour()<22 {
    return name + ", хорошего тебе вечера!"
  }
  return name + ", доброй тебе ночи!"
}
