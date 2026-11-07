//go:build windows

package selfupdate

// busyErr 在 Windows 上恒为 false：单文件 bind-mount 的 EBUSY 场景不存在，
// 占位文件类冲突以其它错误形态出现，走 swap 的常规错误路径即可。
func busyErr(err error) bool {
	return false
}
