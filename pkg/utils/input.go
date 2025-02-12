package utils

import (
    "syscall"
    "time"
    "unsafe"
)

var (
    user32                  = syscall.NewLazyDLL("user32.dll")
    procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
    procGetFocus           = user32.NewProc("GetFocus")
    procSendInput          = user32.NewProc("SendInput")
)

const (
    INPUT_KEYBOARD = 1
    KEYEVENTF_UNICODE = 0x0004
)

type KEYBDINPUT struct {
    Vk         uint16
    Scan       uint16
    Flags      uint32
    Time       uint32
    ExtraInfo  uintptr
}

type INPUT struct {
    Type uint32
    Ki   KEYBDINPUT
    Padding uint64
}

// SimulateInput 模拟键盘输入
func SimulateInput(text string) error {
    time.Sleep(100 * time.Millisecond) // 给用户切换窗口的时间

    var inputs []INPUT
    for _, char := range text {
        input := INPUT{
            Type: INPUT_KEYBOARD,
            Ki: KEYBDINPUT{
                Scan:  uint16(char),
                Flags: KEYEVENTF_UNICODE,
            },
        }
        inputs = append(inputs, input)
    }

    size := unsafe.Sizeof(INPUT{})
    ret, _, err := procSendInput.Call(
        uintptr(len(inputs)),
        uintptr(unsafe.Pointer(&inputs[0])),
        uintptr(size),
    )

    if ret == 0 {
        return err
    }
    return nil
}

// IsFocusedWindowActive 检查当前窗口是否处于活动状态
func IsFocusedWindowActive() bool {
    hwnd, _, _ := procGetForegroundWindow.Call()
    focus, _, _ := procGetFocus.Call()
    return hwnd != 0 && focus != 0
} 