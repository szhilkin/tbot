package config
import(
  "gopkg.in/yaml.v2"
  "path/filepath"
  "io/ioutil"
  // "log"
)

func ReadConfig(configPath string, out interface{}) error {
  var( 
    yamlFile []byte
    err error
    filename string
  )

  filename, err = filepath.Abs(configPath)
  yamlFile, err = ioutil.ReadFile(filename)
  if err != nil {
    return err
  }
  if err = yaml.Unmarshal(yamlFile, out); err != nil {
    return err
  }
  return nil
}