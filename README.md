# Y2R Streaming Tool

A tool for streaming YouTube videos to other live streaming platforms.

[![English](https://img.shields.io/badge/lang-English-blue)](README.md)
[![中文](https://img.shields.io/badge/lang-中文-red)](README_Zh-CN.md)

## Features

- Real-time YouTube video streaming
- Multi-stream management
- Real-time status monitoring
- WebSocket real-time communication
- Support for both Python and Go backend implementations

## Tech Stack

### Backend
- Python Implementation:
  - FastAPI
  - python-socketio
  - python-dotenv
  
- Go Implementation:
  - Gin
  - go-socket.io
  - godotenv

### Frontend
- Vue.js
- Socket.IO Client
- Tailwind CSS

## Project Structure
```tree
.
├── .env                    # Backend environment variables
├── nginx.conf             # Nginx configuration
├── backend/               # Python backend
│   ├── main.py           # Main program
│   └── requirements.txt   # Dependencies
├── backend-go/            # Go backend
│   ├── main.go           # Main program
│   └── go.mod            # Go dependencies
└── frontend/             # Vue.js frontend
    ├── .env              # Frontend environment variables
    ├── src/              # Source code
    │   ├── App.vue       # Main application component
    │   ├── main.js       # Entry file
    │   └── components/   # Components directory
    └── package.json      # Project configuration
```

## Requirements

- Node.js >= 14
- Python >= 3.8 or Go >= 1.16
- FFmpeg
- yt-dlp

## Environment Variables

### Backend Configuration (root .env)
```env
DOMAIN_URL=https://your-domain.com
COOKIES_PATH=/path/to/cookies.txt
```

### Frontend Configuration (frontend/.env)
```env
REACT_APP_DOMAIN_URL=https://your-domain.com
```

## Installation

1. Clone the repository
```bash
git clone <repository-url>
cd <project-directory>
```

2. Install backend dependencies
```bash
# Python backend
cd backend
pip install -r requirements.txt

# Or Go backend
cd backend-go
go mod download
```

3. Install frontend dependencies
```bash
cd frontend
npm install
```

4. Configure environment variables
- Copy `.env.example` to `.env` and modify the configuration
- Copy `frontend/.env.example` to `frontend/.env` and modify the configuration

5. Start the services
```bash
# Start backend (choose one)
cd backend
python main.py
# or
cd backend-go
go run main.go

# Start frontend
cd frontend
npm run serve
```

## Usage

1. Access the frontend page
2. Enter the YouTube video URL
3. Enter the target platform's RTMP URL and stream key
4. Click start streaming

## Notes

- Ensure cookies.txt file is properly configured
- Ensure sufficient network bandwidth
- Follow the streaming rules of relevant platforms

## License

[Apache License Version 2.0](LICENSE) 
