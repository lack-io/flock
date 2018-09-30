package main

import (
	"bytes"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	file, err := ioutil.ReadFile("/Users/xingyys/.ssh/id_rsa")
	if err != nil {
		log.Fatal("Failed to read file: " + err.Error())
	}
	signer, err := ssh.ParsePrivateKeyWithPassphrase(file, []byte("123456"))
	if err != nil {
		log.Fatal("Failed to parse private key: " + err.Error())
	}
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			//ssh.Password("123456"),
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", "192.168.3.111:22", config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	var b bytes.Buffer
	var be bytes.Buffer

	session.Stderr = &be
	session.Stdout = &b

	if err := session.Run("w1"); err != nil {
		fmt.Println(be.String())
		fmt.Println("execute command failed: ", err.Error())
	} else {
		fmt.Println(b.String())
	}


	sftpClient, _ := sftp.NewClient(client)
	var localFilePath = "/home/ping/gopath"
	var remoteFilePath = "/tmp/gopath"
	srcFile, _ := os.Open(localFilePath)
	if err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	defer srcFile.Close()

	dstFile, _ := sftpClient.Create(remoteFilePath)
	defer dstFile.Close()

	buf := make([]byte, 1024)
	for {
		n, _ := srcFile.Read(buf)
		if n == 0 {
			break
		}
		dstFile.Write(buf)
	}
	//fmt.Println("Copy file Finish!")
}
