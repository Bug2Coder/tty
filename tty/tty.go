package tty

import (
	"errors"
	"io"
	"sync"
)

/*
本模块功能：
1、支持本地虚拟终端的构建
2、为web终端提供后台解析支持、（全功能支持终端功能、会话启动目录为项目目录）
3、一个链接对应一个终端
4、链接和会话均可手动关闭、即web端输入`exit`退出和调用端主动关闭会话
使用说明：
1、创建虚拟终端会话需要传入满足`io.ReadWriteCloser`接口的类型、需要自行封装满足的类型
2、读取接口读取到的数据必须为最终终端数据，不能包含其他数据、
   如需封装其他业务、请自行封装类型并解析拆分后传入本模块中
*/

type Manager struct {
	sync.RWMutex
	sessions map[int]SessionIFace
}

func NewManager() *Manager {
	m := Manager{
		sessions: make(map[int]SessionIFace),
	}
	return &m
}

// CreateNewSession 根据rw来创建一个新的tty会话并记录
func (m *Manager) CreateNewSession(rw io.ReadWriteCloser) (int, error) {
	s, err := NewSession(rw)
	if err != nil {
		return 0, err
	}
	m.Lock()
	m.sessions[s.ID()] = s
	m.Unlock()
	return s.ID(), nil
}

// StartSession 根据终端ID开启会话
func (m *Manager) StartSession(id int) error {
	s, ok := m.sessions[id]
	if ok {
		go s.HandleConnection()
		return nil
	}
	return errors.New("id is not found")
}

// CloseSession 关闭会话终端
func (m *Manager) CloseSession(id int) error {
	s, ok := m.sessions[id]
	if ok {
		s.Close()
		m.Lock()
		delete(m.sessions, id)
		m.Unlock()
		return nil
	}
	return errors.New("id is not found")
}
