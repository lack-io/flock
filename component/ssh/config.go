package ssh

import (
	"runtime"
	"sync"
)

const (
	DEFAULT_SECTION = "Default"
)

var (
	LineBreak         = "\n"
	DefaultConfigFile = "/etc/flock/ssh/hosts"
	DefaultPrivateKey = "/root/.ssh/id_rsa"
)

func init() {
	if runtime.GOOS == "windows" {
		LineBreak = "\r\n"
	}
}

type ConfigFile struct {
	lock     sync.RWMutex
	filename string

	groupList []string          // Group name list
	groupMap  map[string][]string // Group -> Host name list

	hostList []string                          // Host name list
	hostMap  map[string]map[string]interface{} // Host -> key name list

	BlockMode bool
}

func NewConfigFile() *ConfigFile {
	c := &ConfigFile{}
	c.filename = DefaultConfigFile
	c.groupList = make([]string, 0)
	c.groupMap = make(map[string][]string)
	c.hostList = make([]string, 0)
	c.hostMap = make(map[string]map[string]interface{})
	c.BlockMode = true
	return c
}

func (c *ConfigFile) GetGroupList() []string {
	list := make([]string, len(c.groupList))
	copy(list, c.groupList)
	return list
}

func (c *ConfigFile) GetHostList() []string {
	list := make([]string, len(c.hostList))
	copy(list, c.hostList)
	return list
}

func (c *ConfigFile) GetGroup() []string {

}
