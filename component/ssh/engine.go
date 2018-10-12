// Copyright 2018 xingyys, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// flock ssh 组件的主体，整合各个部分，提供服务

package ssh

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

func init() {

}

type Engine struct {
	// 配置文件
	Cnofig map[string]string

	// ssh 服务器配置信息
	HostFile *hostFile

	// 连接ssh服务器的客户端实例
	Clis map[string]*Client

	// 请求返回的结果
	Messages *messageStack
}

func NewEngine() *Engine {
	engine := &Engine{}

	engine.Cnofig = make(map[string]string)

	engine.HostFile = NewHostFile()

	engine.Clis = make(map[string]*Client, 0)

	engine.Messages = NewMessageStack()

	return engine
}

// 读取hosts文件
func (e *Engine) LoadHostFile(conf string) error {
	return e.HostFile.LoadFile(conf)
}

// 获取hosts文件解析后的分组
func (e *Engine) GetGroupList() []string {
	return e.HostFile.GetGroupList()
}

// 获取hosts文件解析后的所有ssh参数
func (e *Engine) GetHostList() map[string]*node {
	return e.HostFile.GetHostList()
}

// 获取ssh客户端
func (e *Engine) GetCli(host string) *Client {
	if cli, ok := e.Clis[host]; ok {
		return cli
	}
	return nil
}

// 添加ssh客户端
func (e *Engine) AddCli(cli *Client) error {
	if cli == nil {
		return errors.New("client is nil")
	}
	if _, ok := e.Clis[cli.Host]; ok {
		return errors.New("client exists")
	}
	e.Clis[cli.Host] = cli
	return nil
}

// 删除ssh客户端
func (e *Engine) DeleteCli(host string) error {
	if !e.ExistedCli(host) {
		return errors.New("client not exists")
	}
	delete(e.Clis, host)
	return nil
}

// 是否存在客户端
func (e *Engine) ExistedCli(host string) bool {
	_, ok := e.Clis[host]
	return ok
}

// 加载想要请求的ssh服务器
func (e *Engine) LoadClient(clis string) error {
	if len(clis) == 0 {
		return errors.New("\"hosts\" cannot have empty value")
	}
	hostList := strings.Split(clis, ",")
	for _, v := range hostList {
		v = strings.TrimSpace(v)
		switch {
		case v == "*": // 匹配所有主机
			for k, _ := range e.HostFile.GetHostList() {
				e.AddCli(&Client{Host: k})
			}
		case strings.Contains(v, "*"): // 配置带*的所有主机
			// 10* -> `^10.*`
			v = fmt.Sprintf(`^%s`, strings.Replace(v, "*", ".*", -1))
			pattern := regexp.MustCompile(v)
			for k, _ := range e.HostFile.GetHostList() {
				if pattern.MatchString(k) {
					e.AddCli(&Client{Host: k})
				}
			}
		case e.HostFile.ExistGroup(v): // 配置主机组
			group := e.HostFile.GetGroup(v)
			for _, v := range group.Nodes {
				e.AddCli(&Client{Host: v.Host})
			}
		default: // 默认配置主机IP
			e.AddCli(&Client{Host: v})
		}
	}
	return nil
}

func (e *Engine) parseCmd(module, params string) error {
	return nil
}

func (e *Engine) exec(cmd string) error {
	return nil
}

func (e *Engine) Run(cmd string) error {
	now := time.Now()

	for k, v := range e.Clis {
		msg := &Message{Host: k}

		if !e.HostFile.ExistHost(k) {
			msg.Status = Warning
			msg.Msg = errors.New("Could not match supplied host pattern ")
			msg.Chanage = false
			msg.Duration = time.Now().Sub(now)
			//fmt.Println(msg)
			e.Messages.AddMsg(msg)
			continue
		}
		host := e.HostFile.GetHost(k)
		v = NewClient(host.Host, host.User, host.Port, time.Second*3)
		wg.Add(1)

		go func(cli *Client) {
			defer wg.Done()
			if host.Passwd != "" {
				cli.AddPassword(host.Passwd)
			}
			if host.Privatekey != "" {
				err := cli.AddPrivateKey(host.Privatekey, host.Passparse)
				if err != nil {
					msg.Status = Failed
					msg.Msg = err.Error()
					msg.Chanage = false
					msg.Duration = time.Now().Sub(now)
					//fmt.Println(msg)
					e.Messages.AddMsg(msg)
					return
				}
			}
			err := cli.Conn()
			if err != nil {
				msg.Status = Failed
				msg.Msg = err.Error()
				msg.Chanage = false
				msg.Duration = time.Now().Sub(now)
				//fmt.Println(msg)
				e.Messages.AddMsg(msg)
				return
			}
			defer cli.Close()
			buf, err := cli.Exec(cmd)
			if err != nil {
				msg.Status = Failed
				msg.Msg = err.Error()
				msg.Chanage = false
				msg.Duration = time.Now().Sub(now)
				//fmt.Println(msg)
				e.Messages.AddMsg(msg)
				return
			}
			msg.Status = Success
			msg.Msg = string(buf)
			msg.Chanage = false
			msg.Duration = time.Now().Sub(now)
			e.Messages.AddMsg(msg)
			//fmt.Println(msg)
		}(v)
	}
	wg.Wait()
	e.Messages.Duration = time.Now().Sub(now)
	return nil
}
