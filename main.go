package main
import (
  "fmt"
  "log"
  "github.com/mitchellh/go-homedir"
  "os"
  "gopkg.in/ini.v1"
)

// Config is the main ini config
var Config *ini.File

func loadConfig(src string) {
  fn, err := homedir.Expand(src)
  if err != nil {
    log.Println("Unable to parse filename " + src)
    os.Exit(1)
  }
  Config, err = ini.Load(fn)
  if err != nil {
    fmt.Printf("Fail to read file: %v", err)
  }
}

func main() {
  fmt.Println("Hi wtf")
  loadConfig("~/.sotd.ini")
  controller := Controller{}
  fmt.Println("Running controller now")

  controller.start()
  fmt.Println("Controller started")
}

