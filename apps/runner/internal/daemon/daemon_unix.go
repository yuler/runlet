//go:build !windows

package daemon

import "syscall"

func detachProcessAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
