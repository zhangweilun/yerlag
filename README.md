# ETH 估值分析仪表板

多维度以太坊估值分析工具，整合链上数据、市场数据、机构持仓、网络健康和宏观经济指标。

## 项目结构

```
├── eth-valuation-api/        # Go 后端 (go-zero + GORM + Redis)
├── eth-valuation-dashboard/   # React 前端 (TypeScript + Zustand + Recharts)
└── .kiro/specs/               # 需求、设计和任务文档
```

## 环境要求

- Go 1.21+
- Node.js 18+
- MySQL 8.0+
- Redis 6.0+

## 1. 启动后端

### 1.1 准备数据库

启动 MySQL，创建数据库（GORM 会自动建表）：

```sql
CREATE DATABASE eth_valuation CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 1.2 启动 Redis

确保 Redis 在默认端口运行：

```bash
redis-server
```

### 1.3 配置 API Key 和连接信息

编辑 `eth-valuation-api/etc/config.yaml`，填入你的实际配置：

```yaml
Database:
  DataSource: "root:root123@tcp(127.0.0.1:3306)/eth_valuation?charset=utf8mb4&parseTime=True&loc=Local"

Redis:
  Host: 127.0.0.1:6379
  Pass: ""

DataSources:
  Etherscan:
    APIKey: "你的 Etherscan API Key"    # 在 etherscan.io/apidashboard 创建，V2 一个 Key 支持所有链
  CoinGecko:
    APIKey: "你的 CoinGecko API Key"
  Glassnode:
    APIKey: "你的 Glassnode API Key"
```

> 不填 API Key 也能启动，但对应的数据模块会返回空值。CoinGecko 免费版不需要 Key。

### 1.4 安装依赖并启动

```bash
cd eth-valuation-api
go mod tidy
go run main.go
```

看到以下输出表示启动成功：

```
Starting valuation-api server at 0.0.0.0:8888...
```

API 地址：`http://localhost:8888/api/v1`

### 1.5 运行后端测试

```bash
cd eth-valuation-api
go test ./...
```

## 2. 启动前端

### 2.1 安装依赖

```bash
cd eth-valuation-dashboard
npm install
```

### 2.2 配置 API 地址（可选）

默认连接 `http://localhost:8888/api/v1`。如需修改，在 `eth-valuation-dashboard/` 目录下创建 `.env` 文件：

```
VITE_API_BASE_URL=http://localhost:8888/api/v1
```

### 2.3 启动开发服务器

```bash
npm run dev
```

浏览器打开 `http://localhost:5173` 即可看到仪表板。

### 2.4 构建生产版本

```bash
npm run build
npm run preview   # 预览构建结果
```

### 2.5 运行前端测试

```bash
npm test
```

## 3. 完整启动流程（快速参考）

打开三个终端窗口：

```bash
# 终端 1：启动 Redis（如果还没运行）
redis-server

# 终端 2：启动后端
cd eth-valuation-api
go run main.go

# 终端 3：启动前端
cd eth-valuation-dashboard
npm install
npm run dev
```

然后访问 http://localhost:5173

## 4. API 端点一览

| 端点 | 说明 |
|------|------|
| GET /api/v1/overview | 仪表板总览 |
| GET /api/v1/onchain/burn | EIP-1559 销毁数据 |
| GET /api/v1/onchain/gas | Gas 费用数据 |
| GET /api/v1/onchain/activity | 链上活跃度 |
| GET /api/v1/onchain/tvl | TVL 数据 |
| GET /api/v1/onchain/supply | 供应量数据 |
| GET /api/v1/market | 市场数据 |
| GET /api/v1/market/price-history | 历史价格 |
| GET /api/v1/valuation | 综合估值评分 |
| GET /api/v1/valuation/dcf | DCF 模型 |
| GET /api/v1/institutional/etf | ETF 数据 |
| GET /api/v1/institutional/grayscale | 灰度信托 |
| GET /api/v1/institutional/holdings | 机构持仓 |
| GET /api/v1/network/staking | 质押数据 |
| GET /api/v1/network/performance | 网络性能 |
| GET /api/v1/macro/ethbtc | ETH/BTC 相关性 |
| GET /api/v1/macro/indicators | 宏观经济指标 |
| GET /api/v1/alerts | 活跃预警 |
| POST /api/v1/alerts/rules | 创建预警规则 |
| POST /api/v1/refresh | 强制刷新数据 |
| POST /api/v1/export/csv | 导出 CSV |
| POST /api/v1/share | 生成分享链接 |
