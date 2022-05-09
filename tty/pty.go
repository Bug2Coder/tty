package tty

import (
	"errors"
	ptyDevice "github.com/elisescu/pty"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// PTYHandler 处理函数
type PTYHandler interface {
	Write(data []byte) (int, error) // 消息写入本地pty
	Run()                           // 开启本地pty服务
	ID() int                        // 获取本地ptyID
	Close()                         // 关闭本地pty服务
}

type ptyManager struct {
	fd                int             // 文件描述符
	w                 io.WriteCloser  // 远程写入
	ptyFile           *os.File        // 文件
	command           *exec.Cmd       // 命令
	terminalInitState *terminal.State // 终端初始状态
}

// NewPTY 创建pty会话
func NewPTY(w io.WriteCloser) (*ptyManager, error) {
	if w == nil {
		return nil, errors.New("writer is nil")
	}
	commandName := os.Getenv("SHELL")
	if commandName == "" {
		commandName = "bash"
	}
	envVars := os.Environ()
	ptyMaster := &ptyManager{}
	err := ptyMaster.initEnv(commandName, envVars)
	if err != nil {
		return nil, err
	}
	ptyMaster.w = w
	return ptyMaster, nil
}

// Close 从远程主动关闭虚拟终端
func (pty *ptyManager) Close() {
	pty.StopPtyAndRestore()
}

// StopPtyAndRestore 停止并恢复pty
func (pty *ptyManager) StopPtyAndRestore() {
	pty.Stop()
	pty.Restore()
}

// Run 启动pty服务
func (pty *ptyManager) Run() {
	if pty.w == nil {
		return
	}
	defer pty.StopPtyAndRestore()
	err := pty.MakeRaw()
	if err != nil {
		return
	}
	go func() {
		_, err := io.Copy(pty.w, pty)
		if err != nil {
			pty.StopPtyAndRestore()
			// 关闭远程链接
			pty.w.Close()
		}
	}()
	// 等待pty结束
	err = pty.Wait()
	if err != nil {
		return
	}
}

// 检查是否为标准终端
func isStdinTerminal() bool {
	return terminal.IsTerminal(0)
}

// ID 获取文件的描述符
func (pty *ptyManager) ID() int {
	return pty.fd
}

// initEnv 初始化终端环境
func (pty *ptyManager) initEnv(command string, envVars []string) error {
	var err error
	pty.command = exec.Command(command)
	pty.command.Env = envVars
	pty.ptyFile, err = ptyDevice.Start(pty.command)
	if err != nil {
		return err
	}
	cols, rows, err := terminal.GetSize(0)
	if err != nil {
		return err
	}
	// 给定一个初始大小
	pty.SetWinSize(rows, cols)
	pty.fd = int(pty.ptyFile.Fd())
	return nil
}

// MakeRaw 设置终端为Raw模式、原样输出显示
func (pty *ptyManager) MakeRaw() (err error) {
	pty.terminalInitState, err = terminal.MakeRaw(int(os.Stdin.Fd()))
	return
}

// 将远程数据写入虚拟终端中
func (pty *ptyManager) Write(b []byte) (int, error) {
	return pty.ptyFile.Write(b)
}

// 从虚拟终端中读取数据
func (pty *ptyManager) Read(b []byte) (int, error) {
	return pty.ptyFile.Read(b)
}

// SetWinSize 设置终端窗口大小
func (pty *ptyManager) SetWinSize(rows, cols int) {
	ptyDevice.Setsize(pty.ptyFile, rows, cols)
}

// Wait 等待结束
func (pty *ptyManager) Wait() (err error) {
	err = pty.command.Wait()
	return
}

func (pty *ptyManager) Restore() {
	terminal.Restore(0, pty.terminalInitState)
	return
}

// Stop 关闭终端
func (pty *ptyManager) Stop() (err error) {
	signal.Ignore(syscall.SIGWINCH)

	pty.command.Process.Signal(syscall.SIGTERM)

	pty.command.Process.Signal(syscall.SIGKILL)
	return
}
