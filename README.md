# Y2R 推流工具

一个用于将 YouTube 视频转推到其他直播平台的工具。

## 功能特性

- YouTube 视频实时转推
- 多路流同时管理
- 实时状态监控
- WebSocket 实时通信
- 支持 Python 和 Go 两种后端实现

## 技术栈

### 后端
- Python 实现:
  - FastAPI
  - python-socketio
  - python-dotenv
  
- Go 实现:
  - Gin
  - go-socket.io
  - godotenv

### 前端
- Vue.js
- Socket.IO Client
- Tailwind CSS

## 项目结构
```tree
.
├── .env                    # 后端环境变量配置
├── nginx.conf             # Nginx配置文件
├── backend/               # Python后端
│   ├── main.py           # 主程序
│   └── requirements.txt   # 依赖配置
├── backend-go/            # Go后端
│   ├── main.go           # 主程序
│   └── go.mod            # Go依赖配置
└── frontend/             # Vue.js前端
    ├── .env              # 前端环境变量配置
    ├── src/              # 源代码
    │   ├── App.vue       # 主应用组件
    │   ├── main.js       # 入口文件
    │   └── components/   # 组件目录
    └── package.json      # 项目配置
```

## 环境要求

- Node.js >= 14
- Python >= 3.8 或 Go >= 1.16
- FFmpeg
- yt-dlp

## 环境变量配置

### 后端配置（根目录 .env）
```env
DOMAIN_URL=https://your-domain.com
COOKIES_PATH=/path/to/cookies.txt
```

### 前端配置（frontend/.env）
```env
REACT_APP_DOMAIN_URL=https://your-domain.com
```

## 安装部署

1. 克隆项目
```bash
git clone <repository-url>
cd <project-directory>
```

2. 安装后端依赖
```bash
# Python后端
cd backend
pip install -r requirements.txt

# 或 Go后端
cd backend-go
go mod download
```

3. 安装前端依赖
```bash
cd frontend
npm install
```

4. 配置环境变量
- 复制 `.env.example` 到 `.env` 并修改配置
- 复制 `frontend/.env.example` 到 `frontend/.env` 并修改配置

5. 启动服务
```bash
# 启动后端（选择其一）
cd backend
python main.py
# 或
cd backend-go
go run main.go

# 启动前端
cd frontend
npm run serve
```

## 使用说明

1. 访问前端页面
2. 输入 YouTube 视频链接
3. 输入目标平台的 RTMP 地址和推流密钥
4. 点击开始推流

## 注意事项

- 确保已正确配置 cookies.txt 文件
- 确保有足够的网络带宽
- 注意遵守相关平台的推流规则

## License

[MIT](LICENSE) 
>>>>>>> Stashed changes
