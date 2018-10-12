package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
	"xingyys/flock/component/ssh"
)

var wg sync.WaitGroup

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	conf := ssh.NewHostFile()
	conf.LoadFile("./config/ssh.conf")
	hostList := conf.GetHostList()
	stack := ssh.NewMessageStack()
	now := time.Now()
	wg.Add(len(hostList))
	for _, item := range hostList {
		host := conf.GetHost(item.Host)
		client := ssh.NewClient(host.Host, host.User, host.Port, time.Second*3)
		go func(cli *ssh.Client) {
			defer wg.Done()
			cli.AddPassword(host.Passwd)
			cli.AddPrivateKey(host.Privatekey, host.Passparse)
			err := cli.Conn()
			msg := &ssh.Message{Host: host.Host}
			if err != nil {
				msg.Status = ssh.Failed
				msg.Msg = err.Error()
				msg.Chanage = false
				msg.Duration = time.Now().Sub(now)
				err = stack.AddMsg(msg)
				fmt.Println(msg)
				return
			}
			defer cli.Close()
			buf, err := cli.Exec("hostname")
			if err != nil {
				msg.Status = ssh.Failed
				msg.Msg = err.Error()
				msg.Chanage = false
				stack.AddMsg(msg)
				msg.Duration = time.Now().Sub(now)
				err = stack.AddMsg(msg)
				fmt.Println(msg)
				return
			}
			msg.Status = ssh.Success
			msg.Msg = string(buf)
			msg.Chanage = true
			msg.Duration = time.Now().Sub(now)
			err = stack.AddMsg(msg)
			fmt.Println(msg)
			return
		}(client)
	}
	wg.Wait()
	stack.Duration = time.Now().Sub(now)
	fmt.Println(stack.Duration)
}
