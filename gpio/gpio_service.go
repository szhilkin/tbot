package gpio
import(
  "log"
  "time"
  "github.com/stianeikeland/go-rpio"
  "bitbucket.com/kaikash/headmade_bot/config"
  "github.com/d2r2/go-dht"
)

type Config struct {
  DoorPin int `yaml:"door_pin"`
  DoorReadPin int `yaml:"door_read_pin"`
  LockPin int `yaml:"lock_pin"`
  dhtPin int `yaml:"dht_pin"`
}

type GpioService struct {
  // onAction chan
  config  *Config
  Pins map[string]rpio.Pin
  dhtSensor dht.SensorType
  temperature float32
  humidity float32
  blocked bool
}

func NewGpioService(configPath string) (*GpioService, error) {
  var(
    err error
  )

  gpioService := &GpioService{}
  if err = rpio.Open(); err != nil {
    return nil, err
  }
  if err = config.ReadConfig(configPath, &gpioService.config); err != nil {
    return nil, err 
  }
  gpioService.Pins = map[string]rpio.Pin{
    "door": rpio.Pin(gpioService.config.DoorPin),
    "doorRead": rpio.Pin(gpioService.config.DoorReadPin),
    "lock": rpio.Pin(gpioService.config.LockPin),
  }
  gpioService.blocked = false
  gpioService.dhtSensor = dht.DHT11
  gpioService.Pins["door"].Output()
  gpioService.Pins["lock"].Output()
  gpioService.Pins["doorRead"].Input()
  gpioService.Pins["doorRead"].PullUp()
  return gpioService, nil
} 

func (self *GpioService) ListenDHTsensor() {
  var err error
  for {
    self.temperature, self.humidity, _, err = dht.ReadDHTxxWithRetry(self.dhtSensor, self.config.dhtPin, false, 10)
    if err != nil {
      log.Fatal(err)
    }
    log.Println(self.temperature)
    log.Println(self.humidity)
    time.Sleep(time.Second*10)
  }
}

func (self *GpioService) ListenDoor(onOpen chan<- struct{}) {
  log.Println("Listen door")
  for {
    if self.Pins["doorRead"].Read() == 0 {
      log.Println("Door has been opened")
      onOpen <- struct{}{}
      time.Sleep(time.Second*3)
    }
    time.Sleep(time.Millisecond*10)
  }
}

func (self *GpioService) Listen(onOpen chan<- struct{}) {
  go self.ListenDHTsensor()
  self.ListenDoor(onOpen)
  defer rpio.Close()
}

func (self *GpioService) OpenDoor() error {
  log.Println("door is beeing opened")
  self.Pins["door"].High()
  time.Sleep(1000*time.Millisecond)
  self.Pins["door"].Low()
  return nil
}

func (self *GpioService) LockDoor() error {
  log.Println("door blocked")
  self.blocked = true
  self.Pins["lock"].High()
  return nil
}

func (self *GpioService) UnlockDoor() error {
  log.Println("door unblocked")
  self.blocked = false
  self.Pins["lock"].Low()
  return nil
}

func (self *GpioService) GetTemp() (float32, error) {
  return self.temperature, nil
}

func (self *GpioService) GetHum() (float32, error) {
  return self.humidity, nil
}

func (self *GpioService) IsBlocked() bool {
  return self.blocked
}












