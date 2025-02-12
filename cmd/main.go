package main

import (
	"log"
	
	"github.com/your-username/verify-code-tool/internal/config"
	"github.com/your-username/verify-code-tool/internal/service"
	"github.com/your-username/verify-code-tool/internal/ui"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	log.Printf("配置加载成功: %+v", cfg)

	// 初始化API服务
	apiService := service.NewAPIService(cfg)
	
	// 创建并显示主窗口
	window := ui.NewWindow(cfg, apiService)
	window.Show()
} 