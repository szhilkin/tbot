package telegram
import(
  "log"
  "strings"
  "github.com/Syfaro/telegram-bot-api"
)
type Phrases map[string][]string

func (phrases Phrases) CheckUpdate(message *tgbotapi.Message) {
  var(
    err error
    telegramService *TelegramService
  )
  if telegramService, err = NewTelegramService("", ""); err != nil {
    panic(err)
  }
  log.Println(message.Text)
  for action, phrase := range phrases {
    for i:=0; i<len(phrase); i++ {
      if proc(message, action, phrase[i]) {
        subact := action[0:len(action)-8]
        go telegramService.RunAction(subact, message)
        break
      }
    }
  }
}

func proc(message *tgbotapi.Message, action string, phrase string) bool {
  var(
    err error
    telegramService *TelegramService
  )
  if telegramService, err = NewTelegramService("", ""); err != nil {
    panic(err)
  }
  if strings.ToLower(message.Text) == phrase {
    if action[0:4] == "sudo" {
      if telegramService.IsUserAdmin(message.From.ID) {
        log.Println("SUDO")
        log.Println(action)
      } else {
        log.Println("you're not allowed")
        telegramService.onSudo <- message
        return false
      }
    } else {
      log.Println("not sudo")
      log.Println(action)
    }
    return true
  }
  return false
}