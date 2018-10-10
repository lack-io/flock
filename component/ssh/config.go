package ssh

import (
	"bufio"
	"errors"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const (
	DEFAULT_GROUP = "default"
)

var (
	DefaultConfigFile = "/etc/flock/ssh/hosts" // 默认hosts文件路径
	LineBreak         = "\n"                   // 默认换行符
	DefaultUser       = "root"                 // ssh默认登录用户
	DefaultPort       = 22                     // ssh默认端口
	DefaultPasswd     = ""                     // ssh默认登录密码，默认使用密钥登录
	DefaultPrivateKey = ""                     // ssh默认私钥路径
	DefaultPassParse  = ""                     // ssh默认私钥验证密码
)

func init() {
	if runtime.GOOS == "windows" {
		LineBreak = "\r\n"
	}
}

type node struct {
	Host       string
	Port       int
	User       string
	Passwd     string
	Privatekey string
	Passparse  string
}

type section struct {
	Name  string
	Nodes map[string]*node
}

type configFile struct {
	lock     sync.RWMutex
	filename string

	sections map[string]*section

	BlockMode bool
}

func NewConfigFile() *configFile {
	c := &configFile{}
	c.filename = DefaultConfigFile
	c.sections = make(map[string]*section)
	c.BlockMode = true
	return c
}

// 加载配置
func (c *configFile) read(reader io.Reader) (err error) {
	buf := bufio.NewReader(reader)

	mask, err := buf.Peek(3)
	if err == nil && len(mask) >= 3 &&
		mask[0] == 239 && mask[1] == 187 && mask[2] == 191 {
		buf.Read(mask)
	}
	row := 1 // 当前行
	group := DEFAULT_GROUP

	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		lineLength := len(line)
		if err != nil {
			if err != io.EOF {
				return err
			}
			if lineLength == 0 {
				break
			}
		}
		row += 1
		switch {
		case lineLength == 0: // 空行
			continue
		case line[0] == '#' || line[0] == ';': // 注释
			continue
		case line[0] == '[' && line[lineLength-1] == ']': // 组
			group = strings.TrimSpace(line[1 : lineLength-1])
			err = c.AddGroup(group)
			if err != nil {
				return err
			}
			continue
		default: // 默认解析为主机
			_node := &node{
				Host:       "",
				Port:       22,
				User:       DefaultUser,
				Passwd:     DefaultPasswd,
				Privatekey: DefaultPrivateKey,
				Passparse:  DefaultPassParse}
			s := strings.Split(line, ",")
			for _, v := range s {
				item := strings.Split(strings.TrimSpace(v), "=")
				if len(item) == 1 {
					// 解析行 192.168.100.1
					_node.Host = item[0]
				} else {
					// 解析行
					// host=192.168.100.1, port=22, user=root, passwd=123456, privatekey=.id_rsa
					switch {
					case item[0] == "host":
						_node.Host = item[1]
					case item[0] == "port":
						_node.Port, _ = strconv.Atoi(item[1])
					case item[0] == "user" || item[0] == "username":
						_node.User = item[1]
					case item[0] == "passwd" || item[0] == "password":
						_node.Passwd = item[1]
					case item[0] == "privatekey":
						_node.Privatekey = item[1]
					case item[0] == "passparse":
						_node.Passparse = item[1]
					}
				}
			}
			c.AddHost(group, _node)
		}
		if err == io.EOF {
			break
		}
	}
	return nil
}

// 加载默认配置文件
func (c *configFile) LoadDefaultFile() error {
	return c.LoadFile(DefaultConfigFile)
}

// 加载配置文件
func (c *configFile) LoadFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return c.read(f)
}

// 重新加载配置文件
func (c *configFile) Reload() (err error) {
	c.sections = make(map[string]*section)
	err = c.LoadFile(c.filename)
	return err
}

// 获取所有分组
func (c *configFile) GetGroupList() []string {
	list := make([]string, 0)
	for k, _ := range c.sections {
		list = append(list, k)
	}
	return list
}

// 获取指定分组
func (c *configFile) GetGroup(group string) *section {
	// 没有分组的主机默认为 [default] 组
	if len(group) == 0 {
		group = DEFAULT_GROUP
	}

	if c.BlockMode {
		c.lock.Lock()
		defer c.lock.Unlock()
	}

	if _, ok := c.sections[group]; !ok {
		return nil
	}

	return c.sections[group]
}

// 添加分组
func (c *configFile) AddGroup(group string) error {
	if c.BlockMode {
		c.lock.Lock()
		defer c.lock.Unlock()
	}
	if _, ok := c.sections[group]; ok {
		return errors.New("add an existed group")
	}
	c.sections[group] = &section{group, map[string]*node{}}
	return nil
}

// 删除分组
func (c *configFile) DeleteGroup(group string) error {
	if len(group) == 0 {
		return errors.New("point an empty group")
	}

	if c.BlockMode {
		c.lock.Lock()
		defer c.lock.Unlock()
	}

	if _, ok := c.sections[group]; !ok {
		return errors.New("no such this group")
	}
	delete(c.sections, group)
	return nil
}

// 添加分组中的主机
func (c *configFile) AddHost(group string, hostMap *node) error {
	if len(group) == 0 {
		group = DEFAULT_GROUP
	}

	if group == DEFAULT_GROUP {
		c.AddGroup("default")
	}

	if c.BlockMode {
		c.lock.Lock()
		defer c.lock.Unlock()
	}

	if _, ok := c.sections[group]; !ok {
		return errors.New("no such this group")
	}

	if _, ok := c.sections[group].Nodes[hostMap.Host]; ok {
		return errors.New("no such this host")
	}
	c.sections[group].Nodes[hostMap.Host] = hostMap
	return nil
}

// 删除分组中的主机
func (c *configFile) DeleteHost(group, host string) error {
	if len(group) == 0 {
		group = DEFAULT_GROUP
	}

	if c.BlockMode {
		c.lock.Lock()
		defer c.lock.Unlock()
	}
	if _, ok := c.sections[group]; !ok {
		return errors.New("no such this group")
	}

	if _, ok := c.sections[group].Nodes[host]; ok {
		return errors.New("not such this host")
	}
	delete(c.sections[group].Nodes, host)
	return nil
}

// 获取所有主机
func (c *configFile) GetHostList() []string {
	list := make([]string, 0)
	for _, group := range c.sections {
		for host, _ := range group.Nodes {
			list = append(list, host)
		}
	}
	return list
}

func (c *configFile) GetHost(host string) *node {
	if len(host) == 0 {
		return nil
	}

	if c.BlockMode {
		c.lock.Lock()
		defer c.lock.Unlock()
	}

	for _, group := range c.sections {
		for _, node := range group.Nodes {
			if node.Host == host {
				return node
			}
		}
	}
	return nil
}
