# 验证码自动获取工具

## 项目简介
这是一个基于 PyQt6 开发的验证码自动获取工具，支持从 GitHub 或 Gitee 仓库实时获取验证码。该工具提供了图形界面和系统托盘功能，可以方便地获取、复制和自动填写验证码。

## 主要功能
- 支持 GitHub 和 Gitee 平台
- 自动定时获取验证码
- 系统托盘常驻
- 一键复制验证码
- 自动粘贴验证码
- 平台快速切换

## 安装要求
- Python 3.6+
- 必要的依赖包：
  ```bash
  pip install PyQt6 pyperclip keyboard pyyaml requests  ```

## 使用方法

### 1. 配置文件设置
1. 将 `config.example.yaml` 复制并重命名为 `config.yaml`
2. 编辑 `config.yaml`，填入相关配置： 