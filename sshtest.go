package main

import (
	"fmt"
	"sync"
	"time"
	"xingyys/flock/component/ssh"
)

var wg sync.WaitGroup

func main() {

	client1 := ssh.NewClient("192.168.3.111", "root", 22, time.Second*3)
	client2 := ssh.NewClient("192.168.3.161", "root", 22, time.Second*3)
	client3 := ssh.NewClient("192.168.3.188", "root", 22, time.Second*3)

	client1.AddPrivateKey("/Users/xingyys/.ssh/id_rsa", "")
	client2.AddPrivateKey("/Users/xingyys/.ssh/id_rsa", "")
	client3.AddPrivateKey("/Users/xingyys/.ssh/id_rsa", "")

	wg.Add(3)
	stdout := make(chan []byte, 10)
	stderr := make(chan error, 10)

	go func() {
		defer wg.Done()
		stderr <- client1.Conn()
		buf, err := client1.Exec("hostname")
		if err != nil {
			stdout <- nil
			stderr <- err
		} else {
			stdout <- buf
			stderr <- nil
		}
		client1.Close()
	}()

	go func() {
		defer wg.Done()
		stderr <- client2.Conn()
		buf, err := client2.Exec("free -h")
		if err != nil {
			stdout <- nil
			stderr <- err
		} else {
			stdout <- buf
			stderr <- nil
		}
		client2.Close()
	}()

	go func() {
		defer wg.Done()
		stderr <- client3.Conn()
		buf, err := client3.Exec("hostname")
		if err != nil {
			stdout <- nil
			stderr <- err
		} else {
			stdout <- buf
			stderr <- nil
		}
		client3.Close()
	}()

	wg.Wait()
	close(stdout)
	close(stderr)

	fmt.Println(<-stderr)
	fmt.Println(string(<-stdout))
	fmt.Println(<-stderr)
	fmt.Println(string(<-stdout))
	fmt.Println(<-stderr)
	fmt.Println(string(<-stdout))

}
