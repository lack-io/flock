package main

import (
	"xingyys/flock/component/ssh"
)

func main() {
	engine := ssh.NewEngine()
	engine.LoadHostFile("./config/ssh.conf")
	engine.LoadClient("192.168.3.111")
	engine.Run("w")
}
