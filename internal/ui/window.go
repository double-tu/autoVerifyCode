package ui

import (
    "fmt"
    "log"
    "time"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "fyne.io/fyne/v2/theme"
    
    "github.com/your-username/verify-code-tool/internal/config"
    "github.com/your-username/verify-code-tool/internal/service"
    "github.com/your-username/verify-code-tool/internal/tray"
    "github.com/your-username/verify-code-tool/pkg/utils"
)

type Window struct {
    app        fyne.App
    mainWindow fyne.Window
    apiService *service.APIService
    config     *config.Config
    trayManager *tray.TrayManager
    currentCode string // 添加当前验证码缓存
    lastCode string // 用于比较验证码是否变化

    // UI 组件
    providerSelect *widget.Select
    codeLabel      *widget.Label
    timeLabel      *widget.Label
    refreshBtn     *widget.Button
    copyBtn        *widget.Button
    autoFillCheck  *widget.Check
}

func NewWindow(cfg *config.Config, apiService *service.APIService) *Window {
    w := &Window{
        app:        app.New(),
        config:     cfg,
        apiService: apiService,
    }
    
    w.mainWindow = w.app.NewWindow("验证码自动填充工具")
    w.setupUI()
    
    // 初始化系统托盘
    trayManager, err := tray.NewTrayManager(w.app, w.mainWindow, w.copyVerifyCode)
    if err != nil {
        log.Printf("初始化系统托盘失败: %v", err)
    } else {
        w.trayManager = trayManager
        // 设置托盘图标
        w.trayManager.SetIcon(theme.ConfirmIcon())
    }
    
    // 设置窗口关闭行为
    w.mainWindow.SetCloseIntercept(func() {
        w.mainWindow.Hide()
    })
    
    return w
}

func (w *Window) setupUI() {
    // 配置管理区
    w.providerSelect = widget.NewSelect(
        []string{"github", "gitee"},
        func(value string) {
            if err := w.config.UpdateServiceProvider(value); err != nil {
                w.showError("更新服务商失败", err)
                // 还原选择
                w.providerSelect.SetSelected(w.config.ServiceProvider)
                return
            }
            // 立即刷新验证码
            w.refreshVerifyCode()
        },
    )
    w.providerSelect.SetSelected(w.config.ServiceProvider)
    
    // 添加API配置区域
    apiTokenEntry := widget.NewPasswordEntry()
    apiTokenEntry.SetText(w.config.APIToken)
    apiTokenEntry.OnSubmitted = func(token string) {
        if err := w.config.UpdateAPIToken(token); err != nil {
            w.showError("更新API令牌失败", err)
            apiTokenEntry.SetText(w.config.APIToken)
        }
    }

    repoPathEntry := widget.NewEntry()
    repoPathEntry.SetText(w.config.RepositoryPath)
    repoPathEntry.OnSubmitted = func(path string) {
        if err := w.config.UpdateRepositoryPath(path); err != nil {
            w.showError("更新仓库路径失败", err)
            repoPathEntry.SetText(w.config.RepositoryPath)
        }
    }

    pollingSlider := widget.NewSlider(1, 300)
    pollingSlider.Value = float64(w.config.PollingInterval)
    pollingSlider.OnChanged = func(v float64) {
        if err := w.config.UpdatePollingInterval(int(v)); err != nil {
            w.showError("更新轮询间隔失败", err)
            pollingSlider.Value = float64(w.config.PollingInterval)
            pollingSlider.Refresh()
        }
    }

    configBox := container.NewVBox(
        widget.NewLabel("服务商:"),
        w.providerSelect,
        widget.NewLabel("API令牌:"),
        apiTokenEntry,
        widget.NewLabel("仓库路径:"),
        repoPathEntry,
        widget.NewLabel(fmt.Sprintf("轮询间隔(秒): %d", w.config.PollingInterval)),
        pollingSlider,
    )

    // 信息展示区
    w.codeLabel = widget.NewLabel("等待获取验证码...")
    w.timeLabel = widget.NewLabel("")
    
    infoBox := container.NewVBox(
        widget.NewLabel("验证码信息:"),
        w.codeLabel,
        w.timeLabel,
    )

    // 操作功能区
    w.refreshBtn = widget.NewButton("获取", func() {
        w.refreshVerifyCode()
    })
    
    w.copyBtn = widget.NewButton("复制", func() {
        w.copyVerifyCode()
    })
    
    buttonBox := container.NewHBox(
        w.refreshBtn,
        w.copyBtn,
    )

    // 修改自动填充开关的声明
    w.autoFillCheck = widget.NewCheck("自动填充", func(checked bool) {
        if err := w.config.UpdateAutoFill(checked); err != nil {
            w.showError("更新自动填充设置失败", err)
            // 还原选择
            w.autoFillCheck.SetChecked(!checked)
            return
        }
    })
    w.autoFillCheck.SetChecked(w.config.AutoFill)
    
    // 主布局
    content := container.NewVBox(
        configBox,
        widget.NewSeparator(),
        infoBox,
        widget.NewSeparator(),
        w.autoFillCheck,
        buttonBox,
    )

    w.mainWindow.SetContent(content)
}

func (w *Window) refreshVerifyCode() {
    w.refreshBtn.Disable()
    defer w.refreshBtn.Enable()

    code, err := w.apiService.GetVerifyCode()
    if err != nil {
        w.showError("获取验证码失败", err)
        return
    }

    w.updateVerifyCode(code)
}

func (w *Window) updateVerifyCode(code *service.VerifyCode) {
    // 如果验证码没有变化，直接返回
    if w.lastCode == code.VerifyCode {
        return
    }

    w.currentCode = code.VerifyCode
    w.lastCode = code.VerifyCode
    w.codeLabel.SetText(code.VerifyCode)
    w.timeLabel.SetText(fmt.Sprintf("更新时间: %s", code.Date))

    // 自动填充验证码
    go w.autoFillVerifyCode()
}

func (w *Window) autoFillVerifyCode() {
    // 检查是否启用了自动填充
    if !w.config.AutoFill {
        return
    }

    // 检查是否有活动窗口
    if !utils.IsFocusedWindowActive() {
        return
    }

    // 模拟键盘输入
    if err := utils.SimulateInput(w.currentCode); err != nil {
        w.showError("自动填充失败", err)
    }
}

func (w *Window) copyVerifyCode() {
    if w.currentCode != "" {
        w.mainWindow.Clipboard().SetContent(w.currentCode)
    }
}

func (w *Window) showError(title string, err error) {
    log.Printf("错误: %s - %v", title, err) // 添加日志记录
    
    dialog := widget.NewLabel(fmt.Sprintf("%s: %v", title, err))
    popup := widget.NewModalPopUp(dialog, w.mainWindow.Canvas())
    popup.Show()
    
    go func() {
        time.Sleep(3 * time.Second)
        w.mainWindow.Canvas().Refresh(popup)
    }()
}

func (w *Window) StartPolling() {
    go func() {
        ticker := time.NewTicker(time.Duration(w.config.PollingInterval) * time.Second)
        defer ticker.Stop()

        for range ticker.C {
            code, err := w.apiService.GetVerifyCode()
            if err != nil {
                continue
            }
            
            // 在主线程中更新UI
            w.mainWindow.Canvas().Refresh(w.mainWindow.Content())
            w.updateVerifyCode(code)
        }
    }()
}

func (w *Window) Show() {
    w.mainWindow.Resize(fyne.NewSize(300, 400))
    w.mainWindow.CenterOnScreen()
    w.mainWindow.Show()
    
    // 启动时立即获取一次验证码
    go w.refreshVerifyCode()
    
    // 启动定时轮询
    w.StartPolling()
    
    w.app.Run()
} 