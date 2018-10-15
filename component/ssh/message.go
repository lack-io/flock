// ssh 请求返回信息

package ssh

import (
	"errors"
	"fmt"
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

func NewMessage(host, status string, change bool, msg interface{}, time time.Duration) *Message {
	return &Message{host, status, change, msg, time}
}

func (m *Message) String() string {
	format := fmt.Sprintf("%8s |%8s |%13v =>\n%s\n", m.Host, m.Status, m.Duration, m.Msg)
	switch {
	case m.Chanage == true:
		return ChangeColor(format)
	case m.Status == Success:
		return SuccessColor(format)
	case m.Status == Failed:
		return FailColor(format)
	case m.Status == Warning:
		return WarnColor(format)
	default:
		return fmt.Sprintf(format)
	}
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
		return errors.New("message exists")
	}

	m.Stack[msg.Host] = msg
	return nil
}

func (m *messageStack) DeleteMsg(host string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.Stack[host]; !ok {
		return errors.New("message not exists")
	}
	delete(m.Stack, host)
	return nil
}

func (m *messageStack) SetMsg(msg *Message) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.Stack[msg.Host]; !ok {
		return errors.New("message not exists")
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
