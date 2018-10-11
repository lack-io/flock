package main

import (
	"fmt"
	"xingyys/flock/component/ssh"
)

func main() {
	engine := ssh.NewEngine()
	engine.LoadHostFile("./config/ssh.conf")
	engine.LoadClient(".*3.2**")
	fmt.Println(engine.Clis)
}
