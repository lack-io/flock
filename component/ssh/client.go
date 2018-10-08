package ssh

import (
	"bytes"
	"errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

// ssh 连接对象，包含IP、用户名、端口、认证方式、超时时间
type Client struct {
	// ssh 连接地址，可以IP，可以是hosts文件下定义的别名
	Host string

	// ssh 连接用户名
	User string

	// ssh 连接端口，默认为22
	Port int

	// ssh 认证方式，支持密码认证和密钥认证
	// 密码为空时改为密钥认证
	// 两种同时设置时为密码认证，但其中密码出错时会时认证速度减慢
	auth []ssh.AuthMethod

	// 封装了ssh的client
	cli *ssh.Client

	// 超时时间
	Timeout time.Duration

	// 错误信息
	stderr error

	// 输出信息
	stdout []byte
}

// 新建 ssh client
func NewClient(host, user string, port int, timeout time.Duration) *Client {
	return &Client{
		Host:    host,
		User:    user,
		Port:    port,
		auth:    []ssh.AuthMethod{},
		cli:     nil,
		Timeout: timeout,
		stderr:  nil,
		stdout:  nil,
	}
}

// 添加密码认证
// @passwd : ssh 登录密码
func (c *Client) AddPassword(passwd string) error {
	c.auth = append(c.auth, ssh.Password(passwd))
	return nil
}

// 添加密钥认证
// @key : ssh 密钥路径
// @passParse : 加密密钥的密码
func (c *Client) AddPrivateKey(key, passParse string) error {
	privateKey, err := ioutil.ReadFile(key)
	if err != nil {
		c.stderr = err
		return err
	}
	var signer ssh.Signer
	if passParse == "" {
		signer, err = ssh.ParsePrivateKey(privateKey)
		if err != nil {
			c.stderr = err
			return err
		}
	} else {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(privateKey, []byte(passParse))
		if err != nil {
			c.stderr = err
			return err
		}
	}
	c.auth = append(c.auth, ssh.PublicKeys(signer))
	return nil
}

// 连接client
// ssh 的连接首先加载配置信息，包含ip地址，用户名等
// 然后加载验证方式，包含密码和密钥
func (c *Client) Conn() error {
	config := &ssh.ClientConfig{
		User:            c.User,
		Auth:            c.auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         c.Timeout,
	}
	msg := c.Host + ":" + strconv.Itoa(c.Port)
	client, err := ssh.Dial("tcp", msg, config)
	if err != nil {
		c.stderr = err
		return err
	}
	c.cli = client
	return nil
}

// 关闭client
func (c *Client) Close() {
	c.cli.Close()
}

// 远程执行命令
func (c *Client) Exec(cmd string) ([]byte, error) {
	if c.cli == nil {
		c.stderr = errors.New("no point to ssh client")
		return nil, errors.New("no point to ssh client")
	}

	session, err := c.cli.NewSession()
	if err != nil {
		c.stderr = err
		return nil, err
	}

	defer session.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	session.Stderr = &stderr
	session.Stdout = &stdout

	// ssh 的输出会存储到 session.Stdout
	// ssh 的错误会存储到 session.Stderr
	// err 为远程命令的错误码
	err = session.Run(cmd)

	if err != nil {
		c.stderr = errors.New(stderr.String())
		return nil, c.stderr
	}
	c.stdout = stdout.Bytes()
	return c.stdout, nil
}

// ssh 文件传输
// @src: 本地文件路径
// @dest: 远程文件路径
// @buffer: 打开缓存文件的大小
func (c *Client) Transfer(src, dest string, buffer int) ([]byte, error) {
	if c.cli == nil {
		c.stderr = errors.New("no point to ssh client")
		return nil, c.stderr
	}
	sftpClient, _ := sftp.NewClient(c.cli)

	srcFile, err := os.Open(src)
	if err != nil {
		c.stderr = err
		return nil, err
	}
	defer srcFile.Close()

	dstFile, _ := sftpClient.Create(dest)

	defer dstFile.Close()

	buf := make([]byte, buffer)
	for {
		block, err := srcFile.Read(buf)
		if err != nil {
			c.stderr = err
			return nil, c.stderr
		}
		if block == 0 {
			break
		}
		dstFile.Write(buf)
	}
	c.stdout = []byte("Transfer Successful!")
	return c.stdout, nil
}
