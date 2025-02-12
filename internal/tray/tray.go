package tray

import (
    "fmt"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/driver/desktop"
    "fyne.io/systray"
)

type TrayManager struct {
    app        fyne.App
    mainWindow fyne.Window
    trayIcon   *systray.MenuItem
    onCopy     func() // 复制验证码的回调函数
}

func NewTrayManager(app fyne.App, mainWindow fyne.Window, onCopy func()) (*TrayManager, error) {
    if _, ok := app.(desktop.App); ok {
        systray.SetIcon(app.Icon().Content())
        
        manager := &TrayManager{
            app:        app,
            mainWindow: mainWindow,
            onCopy:     onCopy,
        }
        
        manager.setupTrayMenu()
        return manager, nil
    }
    
    return nil, fmt.Errorf("当前系统不支持系统托盘功能")
}

func (t *TrayManager) setupTrayMenu() {
    // 直接使用 systray 包的方法创建菜单项
    systray.SetTitle("验证码工具")
    
    showItem := systray.AddMenuItem("显示主界面", "显示主界面")
    copyItem := systray.AddMenuItem("复制验证码", "复制验证码")
    systray.AddSeparator()
    quitItem := systray.AddMenuItem("退出", "退出程序")
    
    // 在 goroutine 中处理菜单点击事件
    go func() {
        for {
            select {
            case <-showItem.ClickedCh:
                t.mainWindow.Show()
                t.mainWindow.RequestFocus()
            case <-copyItem.ClickedCh:
                t.onCopy()
            case <-quitItem.ClickedCh:
                t.app.Quit()
            }
        }
    }()
}

func (t *TrayManager) SetIcon(icon fyne.Resource) {
    systray.SetIcon(icon.Content())
} 