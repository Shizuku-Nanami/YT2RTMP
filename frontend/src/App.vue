<template>
  <div class="min-h-screen bg-gray-100">
    <nav class="bg-white shadow-lg">
      <div class="max-w-7xl mx-auto px-4">
        <div class="flex justify-between h-16">
          <div class="flex">
            <div class="flex-shrink-0 flex items-center">
              <h1 class="text-xl font-bold text-gray-800">YT2RTMP</h1>
            </div>
          </div>
        </div>
      </div>
    </nav>

    <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
      <div class="px-4 py-6 sm:px-0">
        <div class="bg-white overflow-hidden shadow rounded-lg">
          <div class="px-4 py-5 sm:p-6">
            <h2 class="text-lg font-medium text-gray-900 mb-4">新建推流任务</h2>
            <form @submit.prevent="startStream" class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-gray-700">YouTube直播URL</label>
                <input
                  v-model="form.url"
                  type="text"
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                  required
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">RTMP服务器地址</label>
                <input
                  v-model="form.rtmp_url"
                  type="text"
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                  required
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">推流密钥</label>
                <input
                  v-model="form.stream_key"
                  type="text"
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"
                  required
                />
              </div>
              <button
                type="submit"
                class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              >
                开始推流
              </button>
            </form>
          </div>
        </div>

        <div class="mt-8 bg-white overflow-hidden shadow rounded-lg">
          <div class="px-4 py-5 sm:p-6">
            <h2 class="text-lg font-medium text-gray-900 mb-4">当前推流任务</h2>
            <div class="space-y-4">
              <div v-for="stream in streams" :key="stream.url" class="border rounded-lg p-4">
                <div class="flex justify-between items-center">
                  <div>
                    <p class="text-sm font-medium text-gray-900">{{ stream.url }}</p>
                    <p class="text-sm text-gray-500">PID: {{ stream.pid }}</p>
                  </div>
                  <div class="flex items-center space-x-2">
                    <span
                      :class="[
                        'px-2 py-1 text-xs font-medium rounded-full',
                        stream.status === 'running'
                          ? 'bg-green-100 text-green-800'
                          : 'bg-red-100 text-red-800'
                      ]"
                    >
                      {{ stream.status === 'running' ? '运行中' : '已停止' }}
                    </span>
                    <button
                      @click="stopStream(stream.url)"
                      class="inline-flex items-center px-3 py-1 border border-transparent text-sm leading-4 font-medium rounded-md text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                    >
                      停止
                    </button>
                  </div>
                </div>
                <div class="mt-4">
                  <div class="bg-gray-900 text-white p-4 rounded-lg font-mono text-sm overflow-y-auto max-h-64">
                    <div v-for="(line, index) in stream.output" :key="index" 
                         :class="['py-1', {
                           'text-white': line.type === 'stdout',
                           'text-red-400': line.type === 'stderr',
                           'text-red-600': line.type === 'error'
                         }]">
                      {{ line.text }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script>
import axios from "axios";
import { io } from "socket.io-client";

// 配置 axios 基础 URL
axios.defaults.baseURL = "/";

export default {
  name: 'App',
  data() {
    return {
      form: {
        url: '',
        rtmp_url: '',
        stream_key: ''
      },
      streams: [],
      socket: null
    };
  },
  methods: {
    async startStream() {
      try {
        await axios.post('/api/stream/start', this.form);
        this.form = { url: '', rtmp_url: '', stream_key: '' };
      } catch (error) {
        console.error('启动推流失败:', error);
      }
    },
    async stopStream(url) {
      try {
        await axios.post(`/api/stream/stop?url=${encodeURIComponent(url)}`);
      } catch (error) {
        console.error('停止推流失败:', error);
      }
    },
    async loadStreams() {
      try {
        const response = await axios.get('/api/streams');
        this.streams = response.data.streams;
      } catch (error) {
        console.error('加载推流列表失败:', error);
      }
    },
    initSocket() {
      // 连接到 WebSocket 服务器
      this.socket = io({
        path: '/socket.io',
        transports: ['websocket'],
        reconnection: true,
        reconnectionAttempts: 5,
        reconnectionDelay: 1000
      });

      // 监听连接事件
      this.socket.on('connect', () => {
        console.log('WebSocket 连接成功');
      });

      // 监听断开连接事件
      this.socket.on('disconnect', () => {
        console.log('WebSocket 连接断开');
      });

      // 监听流状态更新事件
      this.socket.on('streams_update', (data) => {
        this.streams = data.streams;
      });

      // 监听 ffmpeg 输出事件
      this.socket.on('ffmpeg_output', (data) => {
        console.log('FFmpeg output:', data);
        this.streams = this.streams.map(stream => {
          if (stream.url === data.url) {
            return {
              ...stream,
              output: [...(stream.output || []), {
                text: data.output,
                type: data.type,
                timestamp: new Date().toISOString()
              }]
            };
          }
          return stream;
        });
      });
    }
  },
  mounted() {
    this.initSocket();
    this.loadStreams();
  },
  beforeUnmount() {
    // 组件销毁前断开 WebSocket 连接
    if (this.socket) {
      this.socket.disconnect();
    }
  }
};
</script>

<style>
.stream-output {
  background-color: #1e1e1e;
  color: #ffffff;
  padding: 10px;
  border-radius: 4px;
  margin-top: 10px;
  max-height: 300px;
  overflow-y: auto;
  font-family: monospace;
}

.output-line {
  margin: 2px 0;
  padding: 2px 4px;
  border-radius: 2px;
}

.output-line.stdout {
  color: #ffffff;
}

.output-line.stderr {
  color: #ff6b6b;
}

.output-line.error {
  color: #ff0000;
  background-color: rgba(255, 0, 0, 0.1);
}
</style> 