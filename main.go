package main
import (
  "github.com/Syfaro/telegram-bot-api"
  "log"
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "path/filepath"
)

type Config struct {
  Token string `yaml:"token"`
}

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
  if config, err = readConfig(); err != nil {
    panic(err)
  }

  bot, err := tgbotapi.NewBotAPI(config.Token)
  if err != nil {
    log.Panic(err)
  }

  bot.Debug = true
  log.Printf("Authorized on account %s", bot.Self.UserName)

  var ucfg tgbotapi.UpdateConfig = tgbotapi.NewUpdate(0)
  ucfg.Timeout = 60
  err = bot.UpdatesChan(ucfg)
  for {
    select {
    case update := <-bot.Updates:
      UserName := update.Message.From.UserName
      // UserID := update.Message.From.ID
      ChatID := update.Message.Chat.ID
      Text := update.Message.Text

      log.Printf("[%s] %d %s", UserName, ChatID, Text)
      reply := Text
      msg := tgbotapi.NewMessage(ChatID, reply)
      bot.SendMessage(msg)
    }
  }
}