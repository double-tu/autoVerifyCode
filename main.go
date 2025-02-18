package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/getlantern/systray"
	"golang.design/x/clipboard"
	"golang.design/x/hotkey/mainthread"
	"golang.org/x/sys/windows"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Github struct {
		Token      string `yaml:"token"`
		Repository string `yaml:"repository"`
		Path       string `yaml:"path"`
	} `yaml:"github"`
	Gitee struct {
		Token      string `yaml:"token"`
		Repository string `yaml:"repository"`
		Path       string `yaml:"path"`
	} `yaml:"gitee"`
	Setting struct {
		Interval                int    `yaml:"interval"`
		Hotkey                  string `yaml:"hotkey"`
		MaxDateDiffSeconds      int    `yaml:"max_date_diff_seconds"`
		PollingIntervalMillis   int    `yaml:"polling_interval_milliseconds"`
		MaxPollingAttempts      int    `yaml:"max_polling_attempts"`
	} `yaml:"setting"`
}

type VerifyCode struct {
	Date       string `json:"date"`
	VerifyCode string `json:"verifyCode"`
}

// 全局变量，用于在不同文件间共享
var (
	globalConfig *Config
	logFile      *os.File
)

func main() {
	// 初始化日志
	var err error
	logFile, err = os.OpenFile("log.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("无法打开日志文件:", err)
	}
	defer logFile.Close()

	// 设置日志输出到文件和控制台
	log.SetOutput(io.MultiWriter(logFile, os.Stdout))

	// Initialize main thread handling for macOS
	mainthread.Init(func() {
		// 初始化配置
		var err error
		globalConfig, err = loadConfig()
		if err != nil {
			log.Fatal("加载配置文件失败:", err)
		}

		// 检查程序是否已经在运行
		mutex, err := checkSingleton()
		if err != nil {
			log.Fatal("创建互斥锁失败:", err)
		}
		if mutex == 0 {
			log.Fatal("程序已经在运行")
		}
		defer windows.CloseHandle(mutex)

		// 初始化剪贴板
		if err := clipboard.Init(); err != nil {
			log.Fatal("初始化剪贴板失败:", err)
		}

		// 启动系统托盘
		systray.Run(onReady, onExit)
	})
}

func checkSingleton() (windows.Handle, error) {
	path, err := os.Executable()
	if err != nil {
		return 0, err
	}
	hashName := md5.Sum([]byte(path))
	name, err := syscall.UTF16PtrFromString("Local\\" + hex.EncodeToString(hashName[:]))
	if err != nil {
		return 0, err
	}
	return windows.CreateMutex(nil, false, name)
}

func onReady() {
	// 设置快捷键
	setupHotkey(globalConfig.Setting.Hotkey)

	icon, err := os.ReadFile("app.ico")
	if err != nil {
		log.Fatal("读取图标文件失败:", err)
	}

	systray.SetIcon(icon)
	systray.SetTitle("验证码助手")
	systray.SetTooltip("快速获取验证码")

	mGetCode := systray.AddMenuItem("获取并复制", "获取最新验证码并复制到剪贴板")
	mQuit := systray.AddMenuItem("退出", "退出程序")

	systray.AddSeparator()

	go func() {
		for {
			select {
			case <-mGetCode.ClickedCh:
				if code, err := getAndCopyCode(); err != nil {
					log.Println("获取验证码失败:", err)
				} else {
					log.Printf("成功获取验证码: %s", code)
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	os.Exit(0)
}

func isCodeRecent(dateStr string) bool {
	log.Printf("开始解析日期: %s", dateStr)

	// 使用本地时区解析日期
	parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", dateStr, time.Local)
	if err != nil {
		log.Printf("解析日期失败: %v", err)
		return false
	}

	// 获取当前时间
	now := time.Now()

	// 计算时间差
	timeDiff := now.Sub(parsedTime)

	// 获取最大允许时间差
	maxDateDiff := time.Duration(globalConfig.Setting.MaxDateDiffSeconds) * time.Second

	// 判断时间差的绝对值是否在允许范围内
	isRecent := timeDiff.Abs() <= maxDateDiff

	// 打印关键信息
	log.Printf(
		"解析日期: %s, 解析成功（本地时区）: %v, 当前时间: %v, 时间差: %v, 最大允许时间差: %v, 是否在允许范围内: %v",
		dateStr, parsedTime, now, timeDiff, maxDateDiff, isRecent,
	)

	return isRecent
}


func getAndCopyCode() (string, error) {
	var lastCode string
	var lastErr error

	for attempt := 0; attempt < globalConfig.Setting.MaxPollingAttempts; attempt++ {
		// 优先尝试 GitHub
		if globalConfig.Github.Token != "" {
			code, err := getGithubCode(globalConfig.Github.Token, globalConfig.Github.Repository, globalConfig.Github.Path)
			if err == nil {
				if isCodeRecent(code.Date) {
					clipboard.Write(clipboard.FmtText, []byte(code.VerifyCode))
					robotgo.TypeStr(code.VerifyCode)
					return code.VerifyCode, nil
				}
				lastCode = code.VerifyCode
				lastErr = nil
			} else {
				lastErr = err
				log.Printf("从 GitHub 获取验证码失败: %v", err)
			}
		}

		// 尝试 Gitee
		if globalConfig.Gitee.Token != "" {
			code, err := getGiteeCode(globalConfig.Gitee.Token, globalConfig.Gitee.Repository, globalConfig.Gitee.Path)
			if err == nil {
				if isCodeRecent(code.Date) {
					clipboard.Write(clipboard.FmtText, []byte(code.VerifyCode))
					robotgo.TypeStr(code.VerifyCode)
					return code.VerifyCode, nil
				}
				lastCode = code.VerifyCode
				lastErr = nil
			} else {
				lastErr = err
				log.Printf("从 Gitee 获取验证码失败: %v", err)
			}
		}

		// 如果还有重试机会，等待指定时间
		if attempt < globalConfig.Setting.MaxPollingAttempts-1 {
			time.Sleep(time.Duration(globalConfig.Setting.PollingIntervalMillis) * time.Millisecond)
		}
	}

	if lastCode != "" {
		clipboard.Write(clipboard.FmtText, []byte(lastCode))
		robotgo.TypeStr(lastCode)
		return lastCode, nil
	}

	if lastErr != nil {
		return "", fmt.Errorf("无法从任何源获取验证码: %v", lastErr)
	}

	return "", fmt.Errorf("无法获取有效的验证码")
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
