package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Client struct {
	Host string `yaml:"host"`
}


func main() {
	file, _ := ioutil.ReadFile("/tmp/my.yml")
	fmt.Println(string(file))
	s := Client{}
	yaml.Unmarshal(file, &s)
	fmt.Println(s)
}

