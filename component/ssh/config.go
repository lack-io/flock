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
	host       string
	port       int
	user       string
	passwd     string
	privatekey string
	passparse  string
}

type section struct {
	name  string
	nodes map[string]*node
}

type ConfigFile struct {
	lock     sync.RWMutex
	filename string

	sections map[string]*section

	BlockMode bool
}

func NewConfigFile() *ConfigFile {
	c := &ConfigFile{}
	c.filename = DefaultConfigFile
	c.sections = make(map[string]*section)
	c.BlockMode = true
	return c
}

// 加载配置
func (c *ConfigFile) read(reader io.Reader) (err error) {
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
				host:       "",
				port:       22,
				user:       DefaultUser,
				passwd:     DefaultPasswd,
				privatekey: DefaultPrivateKey,
				passparse:  DefaultPassParse}
			s := strings.Split(line, ",")
			for _, v := range s {
				item := strings.Split(strings.TrimSpace(v), "=")
				if len(item) == 1 {
					// 解析行 192.168.100.1
					_node.host = item[0]
				} else {
					// 解析行
					// host=192.168.100.1, port=22, user=root, passwd=123456, privatekey=.id_rsa
					switch {
					case item[0] == "host":
						_node.host = item[1]
					case item[0] == "port":
						_node.port, _ = strconv.Atoi(item[1])
					case item[0] == "user" || item[0] == "username":
						_node.user = item[1]
					case item[0] == "passwd" || item[0] == "password":
						_node.passwd = item[1]
					case item[0] == "privatekey":
						_node.privatekey = item[1]
					case item[0] == "passparse":
						_node.passparse = item[1]
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
func (c *ConfigFile) LoadDefaultFile() error {
	return c.LoadFile(DefaultConfigFile)
}

// 加载配置文件
func (c *ConfigFile) LoadFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return c.read(f)
}

// 重新加载配置文件
func (c *ConfigFile) Reload() (err error) {
	c.sections = make(map[string]*section)
	err = c.LoadFile(c.filename)
	return err
}

// 获取所有分组
func (c *ConfigFile) GetGroupList() []string {
	list := make([]string, 0)
	for k, _ := range c.sections {
		list = append(list, k)
	}
	return list
}

// 获取指定分组
func (c *ConfigFile) GetGroup(group string) *section {
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
func (c *ConfigFile) AddGroup(group string) error {
	if c.BlockMode {
		c.lock.Lock()
		defer c.lock.Unlock()
	}
	if _, ok := c.sections[group]; ok {
		return errors.New("exists this group")
	}
	c.sections[group] = &section{group, map[string]*node{}}
	return nil
}

// 删除分组
func (c *ConfigFile) DeleteGroup(group string) error {
	if len(group) == 0 {
		return errors.New("point a empty group")
	}

	if c.BlockMode {
		c.lock.Lock()
		defer c.lock.Unlock()
	}

	if _, ok := c.sections[group]; !ok {
		return errors.New("not exists this group")
	}
	delete(c.sections, group)
	return nil
}

// 添加分组中的主机
func (c *ConfigFile) AddHost(group string, hostMap *node) error {
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
		return errors.New("not exists this group")
	}

	if _, ok := c.sections[group].nodes[hostMap.host]; ok {
		return errors.New("not exists this host")
	}
	c.sections[group].nodes[hostMap.host] = hostMap
	return nil
}

// 删除分组中的主机
func (c *ConfigFile) DeleteHost(group, host string) error {
	if len(group) == 0 {
		group = DEFAULT_GROUP
	}

	if c.BlockMode {
		c.lock.Lock()
		defer c.lock.Unlock()
	}
	if _, ok := c.sections[group]; !ok {
		return errors.New("not exists this group")
	}

	if _, ok := c.sections[group].nodes[host]; ok {
		return errors.New("not exists this host")
	}
	delete(c.sections[group].nodes, host)
	return nil
}

// 获取所有主机
func (c *ConfigFile) GetHostList() []string {
	list := make([]string, 0)
	for _, group := range c.sections {
		for host, _ := range group.nodes {
			list = append(list, host)
		}
	}
	return list
}

func (c *ConfigFile) GetHost(host string) *node {
	if len(host) == 0 {
		return nil
	}

	if c.BlockMode {
		c.lock.Lock()
		defer c.lock.Unlock()
	}

	for _, group := range c.sections {
		for _, node := range group.nodes {
			if node.host == host {
				return node
			}
		}
	}
	return nil
}
