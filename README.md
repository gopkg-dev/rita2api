# Rati AI Studio

一个基于 Go + Vue 3 + SQLite 的在线生图应用，包含：

- 匿名访客会话
- 文本生图工作台
- 任务提交与 SSE 实时结果同步
- 生成历史记录
- 前端 IndexedDB 任务缓存与恢复
- 自动公开作品画廊

## 本地运行

### 1. 启动前端构建

```bash
cd web
npm install
npm run build
```

### 2. 启动 Go 服务

```bash
go run ./cmd/server
```

默认地址是 `http://localhost:8080`。

## 关键环境变量

- `RATI_ADDR`：服务监听地址，默认 `:8080`
- `RATI_DB_PATH`：SQLite 文件路径，默认 `data/rati.db`
- `RITA_BASE_URL`：上游接口地址，默认 `https://api_v2.rita.ai`
- `RITA_ORIGIN`：上游站点来源，默认 `https://www.rita.ai`
- `RITA_VISITOR_SECRET`：`VisitorId` 签名密钥
- `RITA_MODEL_TYPE_ID`：默认模型类型，默认 `1032`
- `RITA_MODEL_ID`：默认模型 ID，默认 `1121`
- `RATI_COOKIE_NAME`：匿名会话 Cookie 名称，默认 `rati_session`

## Docker

```bash
docker build -t rati-ai-studio .
docker run --rm -p 8080:8080 -v $(pwd)/data:/app/data rati-ai-studio
```
