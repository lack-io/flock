package main

import (
	"fmt"
	"xingyys/flock/component/ssh"
)

func main() {
	engine := ssh.NewEngine()
	engine.LoadHostFile("./config/ssh.conf")
	engine.LoadClient("192.168.3.111, 192")
	engine.Run("w")
	fmt.Println(engine.Messages)
}
