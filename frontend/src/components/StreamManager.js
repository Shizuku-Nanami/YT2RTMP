import React, { useEffect, useState } from "react";
import io from "socket.io-client";

const StreamManager = () => {
  const [streams, setStreams] = useState([]);

  useEffect(() => {
    // 连接 Socket.IO
    const socket = io(process.env.REACT_APP_DOMAIN_URL, {
      path: "/socket.io/",
      transports: ["websocket"],
      secure: true,
    });

    // 监听 ffmpeg 输出
    socket.on("ffmpeg_output", (data) => {
      console.log("FFmpeg output:", data);
      // 更新对应流的输出
      setStreams((prevStreams) => {
        return prevStreams.map((stream) => {
          if (stream.url === data.url) {
            return {
              ...stream,
              output: [
                ...(stream.output || []),
                {
                  text: data.output,
                  type: data.type,
                  timestamp: new Date().toISOString(),
                },
              ],
            };
          }
          return stream;
        });
      });
    });

    // ... 其他 socket 事件监听 ...

    return () => {
      socket.disconnect();
    };
  }, []);

  // 在渲染部分添加输出显示
  return (
    <div>
      {/* ... 其他组件 ... */}
      {streams.map((stream) => (
        <div key={stream.url}>
          {/* ... 其他流信息 ... */}
          <div className="stream-output">
            {stream.output?.map((line, index) => (
              <div key={index} className={`output-line ${line.type}`}>
                {line.text}
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
};

export default StreamManager;
