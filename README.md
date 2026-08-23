# 履历工坊

履历工坊是一个完全离线运行的求职材料工作台。它把求职目标、模板版本、个人资料、草稿历史、模板切换、实时预览、本地导出和隐私策略放在同一条一致性链路中。用户切换模板时，已填写内容仍保存在草稿；目标模板无法接收的字段会进入明确的未映射清单。导出使用冻结快照和版本化隐私策略，不会写入未选择的敏感字段。

## 模块职责

- `cmd/server`：装配 PostgreSQL 仓储、本地文件/通知适配器、Gin 路由和优雅停机。
- `cmd/migrate`：按文件名顺序执行可重复的 PostgreSQL 迁移与演示数据。
- `internal/domain`：模板状态机、字段可见规则、草稿/快照、映射、隐私、附件、导出和反馈实体。
- `internal/application`：创建/自动保存/恢复/切换/导出的事务边界、乐观并发和幂等用例。
- `internal/repository/postgres`：基于 pgx 的正式持久化实现。
- `internal/repository/memory`：只供领域和应用单元测试使用的本地替身。
- `internal/transport/http`：`/api/v1` HTTP 接口、分页白名单、稳定错误协议。
- `internal/middleware`：request_id、身份、结构化访问日志、panic 恢复、CORS 和安全响应头。
- `internal/platform`：受控本地文件、离线渲染、通知、回调、时钟、ID 与定时器适配器。
- `migrations`：表结构、唯一约束、状态约束、索引和不会覆盖已有数据的种子模板。
- `api/openapi`：OpenAPI 3.0 接口契约。
- `web`：Vue 3、TypeScript、Vite、Pinia 中文工作台。
- `tests`：需要真实 PostgreSQL 的仓储集成测试。

## 本地启动

要求 Go 1.24+、Node.js 22+、npm 10+、Docker 和 Docker Compose。

```bash
cp .env.example .env
docker compose up -d postgres
set -a && . ./.env && set +a
go run ./cmd/migrate ./migrations
go run ./cmd/server
```

Windows PowerShell：

```powershell
Copy-Item .env.example .env
docker compose up -d postgres
$env:DATABASE_URL='postgres://cry076:cry076@localhost:5432/cry076?sslmode=disable'
go run ./cmd/migrate ./migrations
go run ./cmd/server
```

前端在另一个终端启动：

```bash
cd web
npm ci
npm run dev
```

浏览器访问 `http://localhost:5173`。Vite 会把 `/api` 请求转发到 `http://localhost:8080`。服务健康检查为 `GET /healthz`，数据库就绪检查为 `GET /readyz`。

## 配置

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `HTTP_ADDRESS` | `:8080` | HTTP 监听地址 |
| `HTTP_READ_TIMEOUT` | `10s` | 请求读取超时 |
| `HTTP_WRITE_TIMEOUT` | `30s` | 响应写入超时 |
| `HTTP_SHUTDOWN_TIMEOUT` | `10s` | 优雅停机上限 |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173` | 逗号分隔的本地前端来源 |
| `DATABASE_URL` | 本地 `cry076` PostgreSQL | pgx 连接地址 |
| `DATABASE_REQUIRED` | `true` | 数据库不可用时拒绝启动 |
| `DATABASE_TIMEOUT` | `5s` | 连接与就绪探测超时 |
| `STORAGE_ROOT` | `./var/data` | 受控附件和导出文件根目录 |
| `MAX_UPLOAD_BYTES` | `8388608` | 单个附件最大字节数 |
| `ALLOWED_UPLOAD_TYPES` | PDF/PNG/JPEG | 上传 MIME 白名单 |
| `LOG_LEVEL` | `info` | zap 日志级别 |

不要把 `.env`、数据库目录或 `var/` 运行期文件提交到仓库。

## 迁移与演示数据

`cmd/migrate` 首次运行会创建 `schema_migrations`，每个迁移文件只应用一次。所有建表、索引和种子写入都可重复执行：

```bash
go run ./cmd/migrate ./migrations
```

`002_seed.sql` 提供一个已发布的“清晰双栏”模板版本，覆盖校招、社招和技术场景。种子使用 `ON CONFLICT DO NOTHING`，不会覆盖管理员已经修改的数据。

## 主要状态规则

模板版本只允许以下流转：

```text
draft -> review -> published -> deprecated
            |-> draft
```

草稿为 `active` 或 `archived`。自动保存、模板切换、恢复和归档都要求调用方提供当前版本；过期版本返回 `VERSION_CONFLICT`。恢复快照前会在同一事务内创建 `before_restore` 快照。模板切换前会创建 `before_template_switch` 快照；无法映射的字段进入 `unmapped` 并随草稿继续保存。

模板反馈按 `open -> triaged -> resolved|dismissed` 流转，解决和驳回必须带处理说明。模板发布、反馈审核、附件读取和导出都会记录 actor、resource、request_id 与业务元数据。

## API 示例

所有业务接口位于 `/api/v1`。本地演示使用请求头传入身份：

```bash
curl -s http://localhost:8080/api/v1/templates?page=1&page_size=20\&sort=name\&order=asc \
  -H 'X-Actor-ID: demo-user' \
  -H 'X-Actor-Role: owner'
```

自动保存需要乐观版本和幂等键：

```bash
curl -X PUT http://localhost:8080/api/v1/drafts/draft_123/autosave \
  -H 'Content-Type: application/json' \
  -H 'X-Actor-ID: demo-user' \
  -H 'X-Actor-Role: owner' \
  -H 'If-Match: 4' \
  -H 'Idempotency-Key: autosave-20260822-001' \
  -d '{"values":{"full_name":"林未","email":"lin@example.test"}}'
```

错误响应始终包含稳定错误码、可读消息、可选字段错误和 `request_id`：

```json
{
  "error": {
    "code": "VERSION_CONFLICT",
    "message": "资源版本已变化",
    "request_id": "e17cc3d962cc4b4dbd6f18b7"
  }
}
```

完整请求与响应结构见 `api/openapi/openapi.yaml`。

## 测试与构建

```bash
go build ./...
go test ./...
go test -race ./...
go vet ./...

cd web
npm ci
npm audit
npm test -- --run
npm run build
```

真实 PostgreSQL 仓储测试会应用迁移并验证 profile 的乐观并发：

```bash
TEST_DATABASE_URL='postgres://cry076:cry076@localhost:5432/cry076?sslmode=disable' \
  go test ./tests -run '^TestPostgresProfileRepositoryOptimisticConcurrency$' -count=1
```

## 实际验证结果

2026-08-22 在 Go 1.24+、Windows amd64、Node.js 25.9.0 和 PostgreSQL 17 容器上完成：

- `go build ./...`：通过。
- `go test ./...`：通过；无数据库配置时集成测试按约定跳过。
- `go test -race ./...`：通过。
- `go vet ./...`：通过。
- PostgreSQL 迁移与 `TestPostgresProfileRepositoryOptimisticConcurrency`：通过。
- `npm audit`：0 个已知漏洞。
- `npm test -- --run`：2 个前端测试通过。
- `npm run build`：类型检查与 Vite 生产构建通过。

## 隐私与本地文件

应用不调用第三方接口、CDN、云存储、在线模型或真实消息服务。附件和导出文件只能写到 `STORAGE_ROOT` 下的安全相对路径；上传同时校验 MIME、大小和 SHA-256。附件读取需要所有者权限或隐私策略中的明确授权，并写入审计事件。预览可遮罩敏感字段，导出只包含 `included_fields` 中明确选择的字段；导出记录保存字段清单、文件哈希、模板版本、草稿快照和隐私策略版本，可随时在本机重新校验。
