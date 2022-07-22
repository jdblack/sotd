package main

import (
	"fmt"
	//  "log"
)

var Cfg Config

func main() {
	var err error
	err = Cfg.load("~/.sotd.ini")
	if err != nil {
		panic(err)
	}

	controller := Controller{}
	fmt.Println("Running controller now")

	controller.start()
	fmt.Println("Controller started")
}
