package ssh

import (
	"errors"
	"sync"
	"time"
)

const (
	Success = "SUCCESS"
	Failed  = "FAILED"
	Warning = "WARNING"
)

// ssh 请求的输出结果
type Message struct {
	Host     string        // 对应的主机
	Status   string        // 执行结果状态
	Chanage  bool          // 文件变化
	Msg      interface{}   // 输出信息
	Duration time.Duration // 花费时间
}

// 输出结果的集合
type messageStack struct {
	lock     sync.RWMutex
	Stack    map[string]*Message
	Duration time.Duration
}

func NewMessageStack() *messageStack {
	return &messageStack{Stack: make(map[string]*Message)}
}

func (m *messageStack) AddMsg(msg *Message) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.Stack[msg.Host]; ok {
		return errors.New("add an existed message")
	}

	m.Stack[msg.Host] = msg
	return nil
}

func (m *messageStack) DeleteMsg(host string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.Stack[host]; !ok {
		return errors.New("no such this message")
	}
	delete(m.Stack, host)
	return nil
}

func (m *messageStack) SetMsg(msg *Message) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.Stack[msg.Host]; !ok {
		return errors.New("no such this message")
	}
	m.Stack[msg.Host] = msg
	return nil
}

func (m *messageStack) Clean() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Stack = make(map[string]*Message)
	return nil
}
