package main

import (
	"log"
	"strings"

	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
)

var modifierMap = map[string]hotkey.Modifier{
	"Ctrl":  hotkey.ModCtrl,
	"Shift": hotkey.ModShift,
	"Alt":   hotkey.ModAlt,
}

var keyMap = map[string]hotkey.Key{
	"Space": hotkey.KeySpace,
	"A":     hotkey.KeyA,
	"B":     hotkey.KeyB,
	"C":     hotkey.KeyC,
	"D":     hotkey.KeyD,
	// Add more keys as needed
}

func setupHotkey(hotkeyStr string) {
	if hotkeyStr == "" {
		log.Println("未配置快捷键")
		return
	}

	keys := strings.Split(hotkeyStr, " ")
	if len(keys) < 2 {
		log.Printf("无效的快捷键格式: %s", hotkeyStr)
		return
	}

	// Parse modifiers
	var mods []hotkey.Modifier
	for _, key := range keys[:len(keys)-1] {
		if mod, ok := modifierMap[key]; ok {
			mods = append(mods, mod)
		} else {
			log.Printf("未知的修饰键: %s", key)
			return
		}
	}

	// Parse main key
	mainKey, ok := keyMap[keys[len(keys)-1]]
	if !ok {
		log.Printf("未知的按键: %s", keys[len(keys)-1])
		return
	}

	mainthread.Call(func() {
		hk := hotkey.New(mods, mainKey)
		if err := hk.Register(); err != nil {
			log.Printf("注册快捷键失败: %v", err)
			return
		}

		log.Printf("成功注册快捷键: %s", hotkeyStr)

		go func() {
			for {
				<-hk.Keydown()
				log.Printf("触发快捷键: %s", hotkeyStr)
				if code, err := getAndCopyCode(); err != nil {
					log.Printf("获取验证码失败: %v", err)
				} else {
					log.Printf("成功获取验证码: %s", code)
				}
			}
		}()
	})
}