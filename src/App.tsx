import React, { useState } from 'react';
import { FileJson, Settings, Shield, Server, FileCode2 } from 'lucide-react';

export default function App() {
  const [activeTab, setActiveTab] = useState('readme');
  const [configContent, setConfigContent] = useState(`{
  "service_name": "pc-edge-gateway",
  "log_path": "logs/pc-edge-gateway.log",
  "rules": [
    {
      "name": "web-tcp",
      "enabled": true,
      "protocol": "tcp",
      "listen": "0.0.0.0:8080",
      "target": "192.168.1.100:80",
      "timeout_seconds": 300,
      "max_connections": 1000
    },
    {
      "name": "dns-udp",
      "enabled": true,
      "protocol": "udp",
      "listen": "0.0.0.0:5353",
      "target": "192.168.1.101:5353",
      "timeout_seconds": 60,
      "max_connections": 1000
    }
  ]
}`);

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-50 flex flex-col font-sans">
      {/* Header */}
      <header className="border-b border-zinc-800 bg-zinc-900/50 p-4 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-emerald-500/10 flex items-center justify-center border border-emerald-500/20">
            <Server className="w-5 h-5 text-emerald-400" />
          </div>
          <div>
            <h1 className="font-semibold text-lg tracking-tight">Windows 端口映射服务</h1>
            <p className="text-xs text-zinc-400 font-mono">基于 Go 语言的服务实现</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <span className="px-2.5 py-1 rounded-md bg-zinc-800 text-xs font-medium text-zinc-300 border border-zinc-700">
            v1.0.0
          </span>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex-1 flex overflow-hidden">
        {/* Sidebar */}
        <aside className="w-64 border-r border-zinc-800 bg-zinc-900/30 p-4 flex flex-col gap-2">
          <div className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-2">项目文件</div>
          
          <button 
            onClick={() => setActiveTab('readme')}
            className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${activeTab === 'readme' ? 'bg-zinc-800 text-zinc-100' : 'text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200'}`}
          >
            <FileCode2 className="w-4 h-4" />
            README.md
          </button>
          
          <button 
            onClick={() => setActiveTab('config')}
            className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors ${activeTab === 'config' ? 'bg-zinc-800 text-zinc-100' : 'text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200'}`}
          >
            <FileJson className="w-4 h-4" />
            config.json
          </button>

          <div className="mt-6 text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-2">服务命令</div>
          
          <div className="space-y-1">
            <div className="flex items-center gap-2 px-3 py-2 text-sm text-zinc-400 font-mono bg-zinc-900/50 rounded border border-zinc-800">
              <span className="text-emerald-500">$</span> pc-edge-gateway.exe install
            </div>
            <div className="flex items-center gap-2 px-3 py-2 text-sm text-zinc-400 font-mono bg-zinc-900/50 rounded border border-zinc-800">
              <span className="text-emerald-500">$</span> pc-edge-gateway.exe start
            </div>
          </div>
        </aside>

        {/* Editor Area */}
        <section className="flex-1 bg-zinc-950 overflow-y-auto">
          {activeTab === 'readme' && (
            <div className="p-8 max-w-4xl mx-auto prose prose-invert prose-zinc">
              <h1 className="text-3xl font-semibold tracking-tight mb-6 flex items-center gap-3">
                Windows 端口映射服务
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  Go
                </span>
              </h1>
              <p className="text-zinc-400 text-lg leading-relaxed mb-8">
                一个基于 Go 语言开发的轻量、稳定且易于部署的 Windows TCP/UDP 端口映射服务。
              </p>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-12">
                <div className="p-4 rounded-xl bg-zinc-900 border border-zinc-800">
                  <div className="flex items-center gap-2 mb-3 text-emerald-400">
                    <Shield className="w-5 h-5" />
                    <h3 className="font-medium text-zinc-200 m-0">安全与稳定</h3>
                  </div>
                  <ul className="text-sm text-zinc-400 space-y-2 m-0 list-none p-0">
                    <li>• 支持优雅退出与资源回收</li>
                    <li>• 规则执行相互隔离</li>
                    <li>• 最小权限运行需求</li>
                  </ul>
                </div>
                <div className="p-4 rounded-xl bg-zinc-900 border border-zinc-800">
                  <div className="flex items-center gap-2 mb-3 text-blue-400">
                    <Settings className="w-5 h-5" />
                    <h3 className="font-medium text-zinc-200 m-0">配置简单</h3>
                  </div>
                  <ul className="text-sm text-zinc-400 space-y-2 m-0 list-none p-0">
                    <li>• 简单的 JSON 格式</li>
                    <li>• 同时支持 TCP & UDP</li>
                    <li>• 启动前配置合法性校验</li>
                  </ul>
                </div>
              </div>

              <h2 className="text-xl font-medium text-zinc-200 mt-8 mb-4">编译指令</h2>
              <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 font-mono text-sm text-zinc-300">
                go build -o pc-edge-gateway.exe ./cmd/pc-edge-gateway
              </div>

              <h2 className="text-xl font-medium text-zinc-200 mt-8 mb-4">项目结构</h2>
              <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 font-mono text-sm text-zinc-400 whitespace-pre">
{`pc-edge-gateway/
├── cmd/
│  └── pc-edge-gateway/
│     └── main.go
├── internal/
│  ├── config/
│  │  └── config.go
│  ├── service/
│  │  └── service.go
│  ├── manager/
│  │  └── manager.go
│  ├── forward/
│  │  ├── tcp.go
│  │  └── udp.go
│  ├── logging/
│  │  └── logger.go
│  └── util/
│     └── validate.go
└── configs/
   └── config.json`}
              </div>
            </div>
          )}

          {activeTab === 'config' && (
            <div className="h-full flex flex-col">
              <div className="border-b border-zinc-800 bg-zinc-900/50 p-3 flex items-center gap-2 text-sm text-zinc-400">
                <FileJson className="w-4 h-4" />
                configs/config.json
              </div>
              <textarea
                value={configContent}
                onChange={(e) => setConfigContent(e.target.value)}
                className="flex-1 w-full bg-zinc-950 text-zinc-300 font-mono text-sm p-6 focus:outline-none resize-none"
                spellCheck={false}
              />
            </div>
          )}
        </section>
      </main>
    </div>
  );
}
