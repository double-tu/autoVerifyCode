package utils

import (
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	procGetSystemPowerStatus = kernel32.NewProc("GetSystemPowerStatus")
)

type SYSTEM_POWER_STATUS struct {
	ACLineStatus byte
	BatteryFlag byte
	BatteryLifePercent byte
	Reserved1 byte
	BatteryLifeTime uint32
	BatteryFullLifeTime uint32
}

func IsSystemActive() bool {
	var status SYSTEM_POWER_STATUS
	ret, _, _ := procGetSystemPowerStatus.Call(uintptr(unsafe.Pointer(&status)))
	return ret != 0 && status.ACLineStatus != 0
} 