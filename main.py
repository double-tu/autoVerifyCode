import sys
import yaml
import json
import requests
import pyperclip
import keyboard
from datetime import datetime
from PyQt6.QtWidgets import (QApplication, QMainWindow, QWidget, QVBoxLayout, 
                           QComboBox, QTextEdit, QPushButton, QSystemTrayIcon, 
                           QMenu, QMessageBox)
from PyQt6.QtCore import Qt, QTimer
from PyQt6.QtGui import QIcon, QAction

class VerifyCodeWindow(QMainWindow):
    def __init__(self):
        super().__init__()
        self.setWindowTitle("验证码工具")
        self.setFixedSize(400, 300)
        
        # 初始化变量
        self.cached_code = {"date": "", "verifyCode": ""}
        self.config = self.load_config()
        
        # 创建主窗口部件
        central_widget = QWidget()
        self.setCentralWidget(central_widget)
        layout = QVBoxLayout(central_widget)
        
        # 创建下拉框
        self.platform_combo = QComboBox()
        self.platform_combo.addItems(["github", "gitee"])
        layout.addWidget(self.platform_combo)
        
        # 创建文本显示框
        self.text_display = QTextEdit()
        self.text_display.setReadOnly(True)
        layout.addWidget(self.text_display)
        
        # 创建按钮
        get_button = QPushButton("获取")
        get_button.clicked.connect(self.manual_fetch)
        layout.addWidget(get_button)
        
        copy_button = QPushButton("复制")
        copy_button.clicked.connect(self.copy_code)
        layout.addWidget(copy_button)
        
        # 创建系统托盘
        self.setup_tray()
        
        # 设置定时器
        self.timer = QTimer()
        self.timer.timeout.connect(self.fetch_code)
        
        # 首次运行立即获取验证码
        self.fetch_code(paste=False)
        
        # 启动定时器
        self.start_timer()

    def load_config(self):
        try:
            with open('config.yaml', 'r', encoding='utf-8') as f:
                return yaml.safe_load(f)
        except Exception as e:
            QMessageBox.critical(self, "错误", f"无法加载配置文件: {str(e)}")
            sys.exit(1)

    def setup_tray(self):
        self.tray = QSystemTrayIcon(self)
        self.tray.setIcon(QIcon("app.ico"))
        
        # 创建托盘菜单
        tray_menu = QMenu()
        
        show_action = QAction("显示界面", self)
        show_action.triggered.connect(self.show)
        
        copy_action = QAction("复制验证码", self)
        copy_action.triggered.connect(self.copy_code)
        
        quit_action = QAction("退出", self)
        quit_action.triggered.connect(self.quit_app)
        
        tray_menu.addAction(show_action)
        tray_menu.addAction(copy_action)
        tray_menu.addAction(quit_action)
        
        self.tray.setContextMenu(tray_menu)
        self.tray.show()

    def fetch_code(self, paste=True):
        platform = self.platform_combo.currentText()
        if platform not in self.config:
            return
        
        cfg = self.config[platform]
        headers = {"Authorization": f"token {cfg['token']}"}
        
        try:
            if platform == "github":
                url = f"https://api.github.com/repos/{cfg['repository']}/contents/{cfg['path']}"
                response = requests.get(url, headers=headers)
                if response.status_code != 200:
                    raise Exception(f"GitHub API 请求失败: HTTP {response.status_code}\n{response.text}")
                
                # GitHub API 返回 base64 编码的内容
                import base64
                content_data = response.json()
                decoded_content = base64.b64decode(content_data["content"]).decode('utf-8')
                content = json.loads(decoded_content)
            else:  # gitee
                url = f"https://gitee.com/api/v5/repos/{cfg['repository']}/contents/{cfg['path']}"
                response = requests.get(url, headers=headers)
                if response.status_code != 200:
                    raise Exception(f"Gitee API 请求失败: HTTP {response.status_code}\n{response.text}")
                
                # Gitee API 也返回 base64 编码的内容
                import base64
                content_data = response.json()
                decoded_content = base64.b64decode(content_data["content"]).decode('utf-8')
                content = json.loads(decoded_content)
            
            # 检查验证码是否更新
            if content["verifyCode"] != self.cached_code["verifyCode"]:
                self.cached_code = content
                self.update_display()
                pyperclip.copy(content["verifyCode"])
                
                if paste:
                    keyboard.write(content["verifyCode"])
                
        except Exception as e:
            error_msg = f"获取验证码失败: {str(e)}"
            self.text_display.setText(error_msg)
            print(f"错误详情: {e}")  # 添加详细错误日志

    def manual_fetch(self):
        self.fetch_code(paste=False)

    def copy_code(self):
        if self.cached_code["verifyCode"]:
            pyperclip.copy(self.cached_code["verifyCode"])

    def update_display(self):
        self.text_display.setText(
            f"验证码: {self.cached_code['verifyCode']}\n"
            f"时间: {self.cached_code['date']}"
        )

    def start_timer(self):
        platform = self.platform_combo.currentText()
        interval = self.config[platform]["interval"] * 1000  # 转换为毫秒
        self.timer.start(interval)

    def closeEvent(self, event):
        event.ignore()
        self.hide()

    def quit_app(self):
        self.tray.hide()
        QApplication.quit()

if __name__ == "__main__":
    app = QApplication(sys.argv)
    window = VerifyCodeWindow()
    window.show()
    sys.exit(app.exec()) 