# Implementation Plan: ETH 估值分析仪表板

## Overview

本实现计划将 ETH 估值分析仪表板拆分为后端 Go (go-zero + GORM + Redis) 和前端 React (TypeScript + Zustand + Recharts) 两大部分，按照基础设施 → 核心计算 → 数据模块 → 前端展示 → 集成联调的顺序递进实现。每个任务构建在前序任务之上，确保无孤立代码。

## Tasks

- [x] 1. 后端项目初始化与基础设施搭建
  - [x] 1.1 初始化 go-zero 项目结构与配置
    - 创建 `eth-valuation-api/` 目录结构（etc/、internal/config、handler、logic、model、svc、types、middleware、fetcher）
    - 编写 `etc/config.yaml` 配置文件（数据库连接、Redis 连接、外部 API Key、缓存 TTL 配置）
    - 编写 `internal/config/config.go` 配置结构体
    - 编写 `main.go` 入口文件，初始化 go-zero HTTP 服务
    - _Requirements: 1.1, 1.2, 14.1_

  - [x] 1.2 定义共享类型与 API 定义文件
    - 创建 `internal/types/common.go`，定义 TimeSeriesPoint、APIResponse[T]、Meta 等共享类型
    - 创建 `api/valuation.api` go-zero API 定义文件，定义所有 REST 端点路由
    - 使用 goctl 生成 handler 和 types 骨架代码
    - _Requirements: 1.1, 1.2_

  - [x] 1.3 实现 GORM 数据模型与数据库迁移
    - 创建 `internal/model/alert_rule.go`（AlertRuleModel）
    - 创建 `internal/model/alert_history.go`（AlertHistoryModel）
    - 创建 `internal/model/share_link.go`（ShareLinkModel）
    - 编写 GORM AutoMigrate 初始化逻辑
    - 在 `internal/svc/servicecontext.go` 中注入 GORM DB 实例和 Redis 客户端
    - _Requirements: 15.1, 15.3, 16.3_

  - [x] 1.4 实现 DataFetcher 数据获取服务与缓存层
    - 创建 `internal/fetcher/fetcher.go`，实现 DataFetcher 接口（Fetch、ForceFetch、InvalidateCache、FetchBatch）
    - 实现 Redis 缓存读写逻辑，按 CacheTTLConfig 设置不同 TTL
    - 实现重试机制（最多 3 次，指数退避 1s/2s/4s）
    - 实现连续失败计数与数据源不可用标记
    - _Requirements: 14.1, 14.2, 14.4, 14.5_

  - [x] 1.5 编写 DataFetcher 缓存往返属性测试
    - **Property 13: 缓存存取往返正确性**
    - 使用 rapid 生成随机数据对象和 TTL，验证存取一致性、TTL 过期行为、失败回退逻辑
    - **Validates: Requirements 14.2, 14.4, 14.5**

  - [x] 1.6 实现 CORS 中间件与通用错误处理中间件
    - 创建 `internal/middleware/cors.go` CORS 中间件
    - 创建通用错误处理中间件，统一 API 错误响应格式
    - _Requirements: 1.1_

- [x] 2. Checkpoint - 确保后端基础设施测试通过
  - 确保所有测试通过，如有问题请询问用户。

- [x] 3. 后端核心计算引擎实现
  - [x] 3.1 实现比率与百分比计算工具函数
    - 创建 `internal/logic/calc/ratios.go`，实现通用比率计算函数（含除零保护）
    - 实现 NVT Ratio、MVRV Ratio、P/F Ratio、TVL/MarketCap、TPS Ratio、Staking Percentage、ETF Holdings Percentage、ETH Dominance、Annualized Burn Rate、ATH Drawdown 等计算
    - _Requirements: 2.3, 3.5, 4.4, 5.4, 5.5, 6.6, 7.1, 7.2, 8.3, 10.1, 11.2, 12.3_

  - [x] 3.2 编写比率计算属性测试
    - **Property 5: 比率与百分比计算正确性**
    - 使用 rapid 生成随机分子/分母，验证所有比率计算公式正确性和除零保护
    - **Validates: Requirements 2.3, 3.5, 4.4, 5.4, 5.5, 6.6, 7.1, 7.2, 8.3, 10.1, 11.2, 12.3**

  - [x] 3.3 实现份额/占比计算工具函数
    - 创建 `internal/logic/calc/shares.go`，实现通用百分比份额计算（组件值 → 百分比数组，保证求和为 100%）
    - 应用于协议 TVL 份额、ETF 市场份额、流动性质押份额、客户端多样性份额、供应量分布
    - _Requirements: 3.4, 5.3, 8.7, 10.5, 11.5, 17.2_

  - [x] 3.4 编写份额计算属性测试
    - **Property 4: 份额/占比计算求和不变量**
    - 使用 rapid 生成随机非负组件值数组，验证百分比求和为 100%（±0.01% 容差）
    - **Validates: Requirements 3.4, 5.3, 8.7, 10.5, 11.5, 17.2**

  - [x] 3.5 实现百分位计算与信号分类
    - 创建 `internal/logic/calc/percentile.go`，实现历史百分位计算函数
    - 实现信号分类逻辑：percentile > 90 → "overvalued"，percentile < 10 → "undervalued"，否则 "neutral"
    - 应用于 NVT Ratio 信号、ETH/BTC 比率信号
    - _Requirements: 4.6, 4.7, 7.6, 12.4, 12.5, 12.6_

  - [x] 3.6 编写百分位信号分类属性测试
    - **Property 6: 百分位信号分类正确性**
    - 使用 rapid 生成随机历史分布和当前值，验证百分位计算和信号分类
    - **Validates: Requirements 4.6, 4.7, 7.6, 12.4, 12.5, 12.6**

  - [x] 3.7 实现移动平均计算
    - 创建 `internal/logic/calc/moving_average.go`，实现 N 日移动平均计算函数
    - _Requirements: 4.1_

  - [x] 3.8 编写移动平均属性测试
    - **Property 9: 移动平均计算正确性**
    - 使用 rapid 生成至少 7 个非负日值数组，验证 7 日 MA 等于最后 7 个值的算术平均
    - **Validates: Requirements 4.1**

  - [x] 3.9 实现滚动相关系数计算
    - 创建 `internal/logic/calc/correlation.go`，实现 Pearson 相关系数计算函数
    - _Requirements: 12.2, 13.3_

  - [x] 3.10 编写滚动相关系数属性测试
    - **Property 8: 滚动相关系数范围不变量**
    - 使用 rapid 生成随机非常数价格序列对，验证相关系数在 [-1, 1] 范围内
    - **Validates: Requirements 12.2, 13.3**

  - [x] 3.11 实现净发行量与通胀/通缩计算
    - 创建 `internal/logic/calc/issuance.go`，实现净发行量、IsDeflationary 标志、年化通胀率计算
    - _Requirements: 2.4, 17.4_

  - [x] 3.12 编写净发行量属性测试
    - **Property 7: 净发行量与通胀/通缩分类**
    - 使用 rapid 生成随机发行量和销毁量，验证净发行量计算和通胀/通缩分类
    - **Validates: Requirements 2.4, 17.4**

  - [x] 3.13 实现交易所价差计算
    - 创建 `internal/logic/calc/spread.go`，实现交易所价差计算（各交易所价格与均价的偏差百分比）
    - _Requirements: 6.5_

  - [x] 3.14 编写交易所价差属性测试
    - **Property 12: 交易所价差计算**
    - 使用 rapid 生成随机交易所价格集（至少 2 个），验证价差计算公式
    - **Validates: Requirements 6.5**

  - [x] 3.15 实现 Grayscale 溢价/折价率计算
    - 创建 `internal/logic/calc/premium.go`，实现溢价/折价率计算
    - _Requirements: 9.1_

  - [x] 3.16 编写 Grayscale 溢价/折价率属性测试
    - **Property 15: Grayscale 溢价/折价率计算**
    - 使用 rapid 生成随机正 NAV 和正市场价格，验证溢价/折价率计算
    - **Validates: Requirements 9.1**

- [x] 4. Checkpoint - 确保核心计算引擎测试通过
  - 确保所有测试通过，如有问题请询问用户。

- [x] 5. 后端估值引擎与预警系统
  - [x] 5.1 实现估值引擎 (Valuation_Engine)
    - 创建 `internal/logic/valuation/engine.go`
    - 实现 MVRV Ratio 评分计算（Req 7.1）
    - 实现 P/F Ratio 评分计算（Req 7.2）
    - 实现 DCF 折现现金流模型，输出 fairValueLow/Mid/High（Req 7.3）
    - 实现 Stock-to-Flow 模型计算（Req 7.4）
    - 实现 NVT Ratio 评分计算
    - 实现 ETH/BTC 相对估值评分计算
    - 实现综合评分聚合（0-100 分），根据分数设置 status 标签
    - 实现雷达图数据生成（RadarDataPoint 数组）
    - _Requirements: 1.3, 7.1, 7.2, 7.3, 7.4, 7.5, 7.6_

  - [x] 5.2 编写估值评分范围属性测试
    - **Property 1: 综合估值评分范围不变量**
    - 使用 rapid 生成随机指标集合，验证评分在 [0, 100] 且 status 标签正确
    - **Validates: Requirements 1.3**

  - [x] 5.3 编写 DCF 估值范围有序性属性测试
    - **Property 10: DCF 估值范围有序性**
    - 使用 rapid 生成随机 DCF 参数，验证 fairValueLow ≤ fairValueMid ≤ fairValueHigh 且均非负
    - **Validates: Requirements 7.3**

  - [x] 5.4 编写 Stock-to-Flow 模型属性测试
    - **Property 11: Stock-to-Flow 模型计算**
    - 使用 rapid 生成随机正 stock 和正 flow，验证 S2F 比率和偏差计算
    - **Validates: Requirements 7.4**

  - [x] 5.5 实现预警系统 (Alert_System)
    - 创建 `internal/logic/alert/service.go`，实现 AlertService 接口
    - 实现 EvaluateAlerts：遍历所有启用的规则，根据条件类型（gt/lt/gt_percent_change/lt_percent_change）判断是否触发
    - 实现 AddRule/UpdateRule/RemoveRule/ToggleRule（GORM CRUD）
    - 实现 GetActiveAlerts/GetAlertHistory/AcknowledgeAlert
    - 实现 SortAlertsBySeverity（稳定排序：high > medium > low）
    - 配置内置预警规则：销毁量异常(Req 2.5)、高 Gas(Req 3.6)、TVL 份额下降(Req 5.6)、ETF 异常流入/流出(Req 8.5/8.6)、Grayscale 折价(Req 9.5)、验证者退出(Req 10.7)、missed slots(Req 11.6)、大规模提币(Req 17.5)
    - _Requirements: 2.5, 3.6, 5.6, 8.5, 8.6, 9.5, 10.7, 11.6, 15.1, 15.2, 15.3, 15.4, 15.5, 17.5_

  - [x] 5.6 编写预警规则评估属性测试
    - **Property 2: 预警规则评估正确性**
    - 使用 rapid 生成随机指标值、规则组合和启用/禁用状态，验证触发逻辑正确性
    - **Validates: Requirements 2.5, 3.6, 5.6, 8.5, 8.6, 9.5, 10.7, 11.6, 15.2, 15.4, 17.5**

  - [x] 5.7 编写预警严重程度排序属性测试
    - **Property 3: 预警严重程度排序**
    - 使用 rapid 生成随机严重程度的预警列表，验证排序结果（high > medium > low，稳定排序）
    - **Validates: Requirements 15.5**

- [x] 6. Checkpoint - 确保估值引擎与预警系统测试通过
  - 确保所有测试通过，如有问题请询问用户。


- [x] 7. 后端数据模块与外部 API 适配器
  - [x] 7.1 实现外部数据源适配器
    - 创建 `internal/fetcher/etherscan.go`（Etherscan API 适配器：Gas 数据、交易数据、销毁数据）
    - 创建 `internal/fetcher/coingecko.go`（CoinGecko API 适配器：价格、市值、交易量、OHLCV）
    - 创建 `internal/fetcher/defillama.go`（DefiLlama API 适配器：TVL、协议数据）
    - 创建 `internal/fetcher/glassnode.go`（Glassnode API 适配器：链上指标、MVRV、NVT）
    - 创建 `internal/fetcher/beaconchain.go`（Beacon Chain API 适配器：质押、验证者数据）
    - 创建 `internal/fetcher/tradfi.go`（传统金融数据适配器：DXY、国债收益率、纳斯达克）
    - 每个适配器实现请求构造、响应解析、错误处理
    - _Requirements: 14.1, 14.4, 14.5_

  - [x] 7.2 实现链上数据模块 (On_Chain_Module)
    - 创建 `internal/logic/onchain/burn.go`（EIP-1559 销毁数据逻辑：24h/7d/30d/累计销毁量、年化销毁率）
    - 创建 `internal/logic/onchain/gas.go`（Gas 费用逻辑：当前均价、费用收入、市费率、高费用标记）
    - 创建 `internal/logic/onchain/activity.go`（活跃度逻辑：DAA、交易笔数、新增地址、NVT Ratio、L2 对比）
    - 创建 `internal/logic/onchain/tvl.go`（TVL 逻辑：总 TVL、协议分布、TVL/市值比、市场份额）
    - 创建 `internal/logic/onchain/supply.go`（供应量逻辑：总供应、分类分布、交易所余额、净发行量）
    - 编写对应的 handler 和路由注册
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.6, 3.1, 3.2, 3.3, 3.4, 3.5, 4.1, 4.2, 4.3, 4.4, 4.5, 5.1, 5.2, 5.3, 5.4, 5.5, 17.1, 17.2, 17.3, 17.4_

  - [x] 7.3 编写链上数据模块单元测试
    - 使用 Mock 外部 API 测试各子模块的数据获取和计算逻辑
    - 测试 Gas 高费用标记、NVT 信号分类、通胀/通缩标记等边界条件
    - _Requirements: 2.1-2.6, 3.1-3.6, 4.1-4.7, 5.1-5.6, 17.1-17.5_

  - [x] 7.4 实现市场数据模块 (Market_Module)
    - 创建 `internal/logic/market/market.go`（市场数据逻辑：实时价格、交易量、市值、ATH 回撤）
    - 创建 `internal/logic/market/price_history.go`（历史价格逻辑：OHLCV 数据、时间范围过滤）
    - 实现交易所价格差异计算
    - 编写对应的 handler 和路由注册
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

  - [x] 7.5 实现机构数据模块 (Institutional_Module)
    - 创建 `internal/logic/institutional/etf.go`（ETF 数据逻辑：持仓量、净流入/流出、占比、发行商排名）
    - 创建 `internal/logic/institutional/grayscale.go`（灰度信托逻辑：持仓、NAV、溢价/折价率）
    - 创建 `internal/logic/institutional/holdings.go`（机构持仓逻辑：大型机构汇总、CME 期货 OI）
    - 编写对应的 handler 和路由注册
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.7, 9.1, 9.2, 9.3, 9.4_

  - [x] 7.6 实现网络健康模块 (Network_Module)
    - 创建 `internal/logic/network/staking.go`（质押数据逻辑：总质押量、验证者数量、收益率、队列、流动性质押份额）
    - 创建 `internal/logic/network/performance.go`（网络性能逻辑：出块时间、TPS、missed slots、客户端多样性）
    - 编写对应的 handler 和路由注册
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6, 11.1, 11.2, 11.3, 11.4, 11.5_

  - [x] 7.7 实现宏观经济模块 (Macro_Module)
    - 创建 `internal/logic/macro/ethbtc.go`（ETH/BTC 逻辑：价格、相关系数、ETH Dominance、百分位信号）
    - 创建 `internal/logic/macro/indicators.go`（宏观指标逻辑：DXY、国债收益率、纳斯达克相关性、联邦基金利率、恐惧贪婪指数、稳定币市值）
    - 编写对应的 handler 和路由注册
    - _Requirements: 12.1, 12.2, 12.3, 12.4, 13.1, 13.2, 13.3, 13.4, 13.5, 13.6_

  - [x] 7.8 实现估值与预警 API 端点
    - 编写估值相关 handler（GetValuation、GetDCFValuation、GetDistribution）
    - 编写预警相关 handler（GetActiveAlerts、GetAlertHistory、CreateAlertRule、UpdateAlertRule、DeleteAlertRule）
    - 编写 Overview handler（聚合各模块摘要数据）
    - _Requirements: 1.1, 1.3, 7.1-7.6, 15.1-15.5_

  - [x] 7.9 实现定时调度器 (Scheduler)
    - 创建 `internal/scheduler/cron.go`，使用 cron 定时触发各模块数据刷新
    - 配置刷新频率：价格 10s、链上数据 5min、机构数据 1h、网络数据 5min、宏观数据 1h
    - 在数据刷新后自动触发预警评估
    - _Requirements: 14.1_

  - [x] 7.10 实现导出与分享 API 端点
    - 编写 ExportCSV handler（将数据记录导出为 CSV 格式）
    - 编写 ExportChart handler（服务端图表导出支持）
    - 编写 GenerateShareLink handler（生成分享链接，保存仪表板状态到数据库）
    - 编写 ForceRefresh handler（强制刷新所有模块数据）
    - _Requirements: 14.3, 16.1, 16.2, 16.3, 16.4_

- [x] 8. Checkpoint - 确保后端所有 API 端点测试通过
  - 确保所有测试通过，如有问题请询问用户。

- [ ] 9. 前端项目初始化与基础架构
  - [x] 9.1 初始化 React + TypeScript 项目
    - 使用 Vite 创建 React + TypeScript 项目
    - 安装依赖：zustand（状态管理）、recharts 或 echarts（图表）、axios（HTTP 请求）
    - 配置 Vitest 测试框架和 fast-check 属性测试库
    - 配置 ESLint、Prettier
    - _Requirements: 1.1_

  - [x] 9.2 实现 API 客户端与轮询服务
    - 创建 `src/api/client.ts`，封装 axios 实例和所有 API 端点调用函数
    - 创建 `src/api/polling.ts`，实现 REST 轮询服务（按配置间隔轮询各数据端点）
    - 配置轮询间隔：价格 10s、链上 5min、机构 1h、网络 5min、宏观 1h、预警 30s
    - 实现轮询失败时保留上次数据、标注最后更新时间的逻辑
    - _Requirements: 6.1, 14.1, 14.4_

  - [x] 9.3 实现 Zustand 全局状态管理
    - 创建 `src/store/dashboard.ts`，定义 DashboardStore
    - 实现各模块数据状态（onChainData、marketData、institutionalData、networkData、macroData、valuationScore）
    - 实现 UI 状态（theme、activeAlerts、isLoading、lastUpdated、errors）
    - 实现 refreshAll、refreshModule、toggleTheme、setAlertRule 等操作
    - _Requirements: 1.2, 1.5, 14.3_

  - [x] 9.4 实现主题系统（浅色/深色模式）
    - 创建 `src/context/ThemeContext.tsx`，实现主题上下文
    - 定义浅色和深色主题的 CSS 变量/样式
    - 实现主题切换按钮组件
    - _Requirements: 1.5_

  - [x] 9.5 编写轮询服务和状态管理单元测试
    - 测试轮询间隔配置正确性
    - 测试轮询失败时的数据保留逻辑
    - 测试主题切换功能
    - _Requirements: 1.5, 14.1, 14.4_

- [ ] 10. 前端 Dashboard 布局与总览组件
  - [x] 10.1 实现 Dashboard 主布局
    - 创建 `src/pages/Dashboard.tsx` 主页面组件
    - 实现顶部摘要栏（ETH 价格、24h 涨跌幅、市值排名、综合估值评分）
    - 实现五大维度分区布局（链上数据、市场数据、机构数据、网络健康、宏观经济）
    - 实现 React Error Boundary 包裹每个模块
    - _Requirements: 1.1, 1.2, 1.3_

  - [x] 10.2 实现通用图表渲染器组件
    - 创建 `src/components/charts/` 目录
    - 实现 LineChart（折线图）、AreaChart（面积图）、StackedAreaChart（堆叠面积图）组件
    - 实现 BarChart（柱状图）、CandlestickChart（K 线图）组件
    - 实现 PieChart（饼图）、RadarChart（雷达图）组件
    - 所有图表支持 responsive 自适应和 tooltip 交互
    - _Requirements: 2.2, 2.6, 3.2, 4.2, 4.3, 5.2, 5.3, 6.4, 7.5, 7.6_

  - [x] 10.3 编写图表组件单元测试
    - 使用 React Testing Library 验证各图表组件正确渲染
    - 测试图表接收正确的数据格式
    - _Requirements: 2.2, 6.4, 7.5_

- [ ] 11. 前端数据展示模块
  - [x] 11.1 实现链上数据展示组件
    - 创建 `src/components/onchain/BurnDataPanel.tsx`（销毁数据面板：24h/7d/30d/累计销毁量、年化销毁率、每日销毁趋势图、供应量堆叠面积图）
    - 创建 `src/components/onchain/GasDataPanel.tsx`（Gas 费用面板：当前均价、费用收入、市费率、历史趋势图、高费用标记）
    - 创建 `src/components/onchain/ActivityPanel.tsx`（活跃度面板：DAA、交易笔数、新增地址、NVT Ratio、L2 对比柱状图）
    - 创建 `src/components/onchain/TVLPanel.tsx`（TVL 面板：总 TVL、协议分布饼图、TVL/市值比、市场份额趋势）
    - 创建 `src/components/onchain/SupplyPanel.tsx`（供应量面板：总供应、分类分布、交易所余额趋势、通胀/通缩状态）
    - _Requirements: 2.1-2.6, 3.1-3.5, 4.1-4.5, 5.1-5.5, 17.1-17.4_

  - [x] 11.2 实现市场数据展示组件
    - 创建 `src/components/market/MarketOverview.tsx`（市场总览：实时价格、涨跌幅、交易量、市值、ATH 回撤）
    - 创建 `src/components/market/PriceChart.tsx`（K 线图：支持 1d/1w/1m/3m/1y/all 时间范围切换）
    - 创建 `src/components/market/ExchangeSpread.tsx`（交易所价差表格）
    - _Requirements: 6.1-6.6_

  - [x] 11.3 实现估值模型展示组件
    - 创建 `src/components/valuation/ValuationScore.tsx`（综合评分展示：分数、状态标签、雷达图）
    - 创建 `src/components/valuation/ModelDetails.tsx`（各模型详情：MVRV、P/F、DCF、S2F、NVT、ETH/BTC）
    - 创建 `src/components/valuation/DistributionChart.tsx`（历史分布图，标注当前值位置）
    - _Requirements: 1.3, 7.1-7.6_

  - [x] 11.4 实现机构数据展示组件
    - 创建 `src/components/institutional/ETFPanel.tsx`（ETF 面板：持仓量、净流入/流出柱状图、占比、发行商排名）
    - 创建 `src/components/institutional/GrayscalePanel.tsx`（灰度面板：持仓、NAV、溢价/折价率趋势）
    - 创建 `src/components/institutional/HoldingsPanel.tsx`（机构持仓面板：大型机构汇总、CME 期货 OI 趋势）
    - _Requirements: 8.1-8.4, 8.7, 9.1-9.4_

  - [x] 11.5 实现网络健康展示组件
    - 创建 `src/components/network/StakingPanel.tsx`（质押面板：总质押量、验证者数量、收益率、队列、流动性质押份额饼图）
    - 创建 `src/components/network/PerformancePanel.tsx`（性能面板：出块时间、TPS、missed slots、客户端多样性饼图）
    - _Requirements: 10.1-10.6, 11.1-11.5_

  - [x] 11.6 实现宏观经济展示组件
    - 创建 `src/components/macro/ETHBTCPanel.tsx`（ETH/BTC 面板：价格趋势、相关系数、ETH Dominance、百分位信号）
    - 创建 `src/components/macro/MacroIndicatorsPanel.tsx`（宏观指标面板：DXY 叠加图、国债收益率叠加图、纳斯达克相关性、利率预期、恐惧贪婪指数、稳定币市值）
    - _Requirements: 12.1-12.4, 13.1-13.6_

  - [x] 11.7 编写数据展示组件单元测试
    - 使用 React Testing Library 验证各面板组件正确渲染所需数据字段
    - 测试加载状态和错误状态的展示
    - 测试时间范围切换逻辑
    - _Requirements: 1.1, 1.2, 6.4_

- [x] 12. Checkpoint - 确保前端组件渲染测试通过
  - 确保所有测试通过，如有问题请询问用户。

- [ ] 13. 前端预警、导出与分享功能
  - [x] 13.1 实现预警通知组件
    - 创建 `src/components/alert/AlertBanner.tsx`（顶部预警通知横幅，按严重程度排序展示）
    - 创建 `src/components/alert/AlertHistory.tsx`（预警历史记录列表，展示过去 30 天预警）
    - 创建 `src/components/alert/AlertRuleManager.tsx`（预警规则管理：创建、编辑、启用/禁用、删除）
    - _Requirements: 15.1-15.5_

  - [x] 13.2 实现数据导出功能
    - 创建 `src/utils/export.ts`，实现图表导出为 PNG/SVG 功能
    - 实现数据表格导出为 CSV 功能（前端生成）
    - 在各图表和数据面板添加导出按钮
    - _Requirements: 16.1, 16.2, 16.4_

  - [x] 13.3 编写 CSV 导出往返属性测试
    - **Property 14: CSV 导出往返正确性**
    - 使用 fast-check 生成随机数据记录（含字符串和数字字段），验证导出 CSV 后解析回来与原始数据一致
    - **Validates: Requirements 16.2**

  - [x] 13.4 实现分享链接功能
    - 创建 `src/components/share/ShareDialog.tsx`（分享对话框：生成链接、复制链接）
    - 实现从分享链接恢复仪表板状态的逻辑
    - _Requirements: 16.3_

  - [x] 13.5 编写预警组件和导出功能单元测试
    - 测试预警横幅按严重程度排序展示
    - 测试预警规则 CRUD 操作
    - 测试导出按钮触发下载
    - _Requirements: 15.1-15.5, 16.1-16.4_

- [ ] 14. 前端集成与全局联调
  - [x] 14.1 集成所有模块到 Dashboard 主页面
    - 在 Dashboard.tsx 中组装所有数据展示模块
    - 连接 Zustand store 与各组件
    - 启动轮询服务，确保各模块数据自动刷新
    - 实现手动刷新按钮（ForceRefresh）
    - _Requirements: 1.1, 1.2, 1.4, 14.3_

  - [x] 14.2 实现前端错误边界与降级展示
    - 为每个模块添加 Error Boundary
    - 实现各模块独立的加载状态（skeleton/spinner）和错误状态展示
    - 实现数据最后更新时间标注
    - 实现数据源不可用警告展示
    - _Requirements: 14.4, 14.5_

  - [x] 14.3 编写前端集成测试
    - 测试 Dashboard 初始加载流程
    - 测试手动刷新功能
    - 测试单模块错误不影响其他模块
    - _Requirements: 1.4, 14.3, 14.4_

- [x] 15. Final Checkpoint - 确保所有测试通过
  - 确保所有测试通过，如有问题请询问用户。

## Notes

- 标记 `*` 的任务为可选任务，可跳过以加速 MVP 交付
- 每个任务引用了具体的需求编号，确保需求可追溯
- Checkpoint 任务用于阶段性验证，确保增量开发的正确性
- 属性测试覆盖设计文档中全部 15 个 Correctness Properties
- 后端属性测试使用 Go rapid 库，前端属性测试使用 fast-check 库
- 单元测试和属性测试互补：属性测试验证通用正确性，单元测试验证具体场景和边界条件