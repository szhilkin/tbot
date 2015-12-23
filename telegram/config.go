package telegram
type Config struct {
  MainChatId          int `yaml:"main_chat_id"`
  Token               string `yaml:"token"`
  AllowedChatIds      []int `yaml:"allowed_chat_ids"`
  SudoersIds          []int `yaml:"sudoers_ids"`
  BlockedIds          []int `yaml:"blocked_ids"`
}