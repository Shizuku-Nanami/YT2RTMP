from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
import subprocess
import os
from typing import Optional
import socketio
from fastapi.middleware.cors import CORSMiddleware
import asyncio
import signal
from datetime import datetime
import select
from dotenv import load_dotenv

# 加载环境变量
load_dotenv()

# 获取环境变量
DOMAIN_URL = os.getenv("DOMAIN_URL")
COOKIES_PATH = os.getenv("COOKIES_PATH")

if not DOMAIN_URL or not COOKIES_PATH:
    raise ValueError("必须设置 DOMAIN_URL 和 COOKIES_PATH 环境变量")

# 创建 FastAPI 应用
app = FastAPI()

# 配置CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=[DOMAIN_URL],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 创建 Socket.IO 服务器
sio = socketio.AsyncServer(
    async_mode="asgi",
    cors_allowed_origins=[DOMAIN_URL],
)
socket_app = socketio.ASGIApp(sio, app)

# 存储推流进程信息
stream_processes = {}


# Socket.IO 事件处理
@sio.event
async def connect(sid, environ):
    print(f"Client connected: {sid}")
    # 发送当前所有流的状态
    await sio.emit(
        "streams_update",
        {
            "streams": [
                {
                    "url": url,
                    "status": (
                        "running" if process["process"].poll() is None else "stopped"
                    ),
                    "pid": process["process"].pid,
                }
                for url, process in stream_processes.items()
            ]
        },
        to=sid,
    )


@sio.event
async def disconnect(sid):
    print(f"Client disconnected: {sid}")


# 广播流状态更新
async def broadcast_streams_update():
    await sio.emit(
        "streams_update",
        {
            "streams": [
                {
                    "url": url,
                    "status": (
                        "running" if process["process"].poll() is None else "stopped"
                    ),
                    "pid": process["process"].pid,
                    "output": process.get("output", []),  # 添加输出信息
                }
                for url, process in stream_processes.items()
            ]
        },
    )


class StreamRequest(BaseModel):
    url: str
    rtmp_url: str
    stream_key: str


class StreamStatus(BaseModel):
    is_running: bool
    process_id: Optional[int]
    error: Optional[str]


@app.post("/api/stream/start")
async def start_stream(request: StreamRequest):
    try:
        # 构建完整的 ffmpeg 命令，保持双引号
        rtmp_url = f"{request.rtmp_url}/{request.stream_key}"
        ffmpeg_cmd = f'ffmpeg -re -i "$(yt-dlp -g --cookies "{COOKIES_PATH}" {request.url})" -c:v copy -c:a copy -f flv "{rtmp_url}"'

        print(f"Starting stream for URL: {request.url}")

        # 启动 ffmpeg 进程，使用 shell=True 来支持命令替换
        process = subprocess.Popen(
            ffmpeg_cmd,  # 使用字符串形式的命令
            shell=True,  # 启用 shell 模式以支持命令替换
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            preexec_fn=os.setsid,  # 使用进程组
            bufsize=1,  # 行缓冲
            universal_newlines=True,  # 文本模式
        )

        # 存储进程信息
        stream_processes[request.url] = {
            "process": process,
            "rtmp_url": rtmp_url,
            "start_time": datetime.now(),
        }

        # 启动异步任务监控进程状态
        asyncio.create_task(monitor_process(request.url, process))

        # 立即广播更新
        await broadcast_streams_update()

        return {"status": "success", "message": "推流已启动", "pid": process.pid}

    except Exception as e:
        print(f"Error starting stream: {str(e)}")
        if request.url in stream_processes:
            try:
                process = stream_processes[request.url]["process"]
                os.killpg(os.getpgid(process.pid), signal.SIGTERM)
            except Exception as kill_error:
                print(f"Error killing process: {str(kill_error)}")
            del stream_processes[request.url]
        raise HTTPException(status_code=500, detail=str(e))


# 修改进程监控函数
async def monitor_process(url: str, process: subprocess.Popen):
    try:
        while True:
            if url not in stream_processes:
                break

            # 检查进程状态
            return_code = process.poll()
            if return_code is not None:
                # 进程已结束
                print(f"Process ended with code {return_code} for {url}")
                if url in stream_processes:
                    del stream_processes[url]
                    await broadcast_streams_update()
                break

            await asyncio.sleep(5)  # 每5秒检查一次
    except Exception as e:
        print(f"Monitor error for {url}: {str(e)}")
        if url in stream_processes:
            del stream_processes[url]
            await broadcast_streams_update()


@app.post("/api/stream/stop")
async def stop_stream(url: str):
    if url in stream_processes:
        try:
            process = stream_processes[url]["process"]
            # 先尝试正常终止
            try:
                os.killpg(os.getpgid(process.pid), signal.SIGTERM)
            except Exception as e:
                print(f"Error sending SIGTERM: {str(e)}")

            # 等待一小段时间
            await asyncio.sleep(1)

            # 检查进程是否还在运行
            if process.poll() is None:
                # 如果还在运行，强制终止
                try:
                    os.killpg(os.getpgid(process.pid), signal.SIGKILL)
                except Exception as e:
                    print(f"Error sending SIGKILL: {str(e)}")

            # 清理进程信息
            if url in stream_processes:
                del stream_processes[url]
                await broadcast_streams_update()

        except Exception as e:
            print(f"Error stopping process: {str(e)}")
            # 确保清理进程信息
            if url in stream_processes:
                del stream_processes[url]
                await broadcast_streams_update()

        return {"status": "success", "message": "推流已停止"}
    return {"status": "error", "message": "未找到对应的推流进程"}


@app.get("/api/stream/status/{url}")
async def get_stream_status(url: str):
    if url in stream_processes:
        process = stream_processes[url]["process"]
        return StreamStatus(
            is_running=process.poll() is None, process_id=process.pid, error=None
        )
    return StreamStatus(is_running=False, process_id=None, error="未找到对应的推流进程")


@app.get("/api/streams")
async def list_streams():
    return {
        "streams": [
            {
                "url": url,
                "status": "running" if process["process"].poll() is None else "stopped",
                "pid": process["process"].pid,
            }
            for url, process in stream_processes.items()
        ]
    }


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(socket_app, host="0.0.0.0", port=8000)
