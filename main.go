package main
import(
  "bitbucket.org/kaikash/headmade_bot/telegram"
  "log"
)

func main() {
  var(
    telegram_service *telegram.TelegramService
    err error
  )
  if telegram_service, err = telegram.NewTelegramService("./config.yml", "./phrases.yml"); err != nil {
    log.Println(err)
  }
  telegram_service.Listen()
}