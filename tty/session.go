package tty

import (
	"errors"
	"io"
	"sync"
)

type SessionIFace interface {
	ID() int           // 获取本地ptyID
	HandleConnection() // 处理链接
	Close()            // 关闭会话
}

type ttySession struct {
	lock       sync.RWMutex       // 锁
	id         int                // pty 会话ID
	rw         io.ReadWriteCloser // 远程读写接口
	ptyHandler PTYHandler         // pty处理接口
}

func NewSession(rw io.ReadWriteCloser) (SessionIFace, error) {
	if rw == nil {
		return nil, errors.New("ReadWriteCloser is nil")
	}
	handle, err := NewPTY(rw)
	if err != nil {
		return nil, err
	}
	s := &ttySession{
		rw:         rw,
		ptyHandler: handle,
		id:         handle.ID(),
	}
	return s, err
}

func (session *ttySession) ID() int {
	return session.id
}

// 将本地pty数据发送到远程
func (session *ttySession) Write(data []byte) (int, error) {
	l, err := session.rw.Write(data)
	return l, err
}

// HandleConnection 读取消息并进行处理
func (session *ttySession) HandleConnection() {
	// 开启pty
	go session.ptyHandler.Run()
	// 等待远程消息，然后写入本地pty
	for {
		err := session.ReadAndHandle(
			func(data []byte) {
				_, err := session.ptyHandler.Write(data)
				if err != nil {
					return
				}
			},
		)
		if err != nil {
			break
		}
	}
	session.lock.Lock()
	session.rw = nil
	session.lock.Unlock()
}

type OnMsgWrite func(data []byte)

// ReadAndHandle 读取远程数据并写入本地pty中
func (session *ttySession) ReadAndHandle(onWrite OnMsgWrite) (err error) {
	var buf = make([]byte, 128)
	n, err := session.rw.Read(buf)
	if err != nil {
		return err
	}
	onWrite(buf[:n])
	return
}

func (session *ttySession) Close() {
	session.ptyHandler.Close()
}
