//go:build !windows

package selfupdate

import (
	"errors"
	"syscall"
)

// busyErr 判断 rename 失败是否为"设备或资源忙"（EBUSY）。挂载点被 rename 时
// 走到这里 —— 见 swap 里对单文件 bind-mount 部署的处理。
func busyErr(err error) bool {
	return errors.Is(err, syscall.EBUSY)
}
