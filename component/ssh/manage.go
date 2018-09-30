package ssh

import (
	"errors"
)


type Manage struct {
	cli []*Client
}

func NewManage() *Manage {
	return &Manage{cli: make([]*Client, 0)}
}

func (m *Manage) GetClient(host string) (*Client, error) {
	for _, client := range m.cli {
		if client.Host == host {
			return client, nil
		}
	}
	return nil, errors.New("not found connection")
}

func (m *Manage) AddClient(conn *Client) error {
	m.cli = append(m.cli, conn)
	return nil
}

func (m *Manage) RmClient(host string) (*Client, error) {
	for i, client := range m.cli {
		if client.Host == host {
			m.cli = append(m.cli[0:i], m.cli[i+1:]...)
			return client, nil
		}
	}
	return nil, errors.New("not found connection")
}

func (m *Manage) Start() error {
	for _, client := range m.cli {
		client.Conn()
	}
	return nil
}

func (m *Manage) Exec(cmd string) error {
	for _, client := range m.cli {
		client.Exec(cmd)
	}
	return nil
}
