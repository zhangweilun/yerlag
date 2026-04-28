# Requirements Document

## Introduction

ETH 估值分析仪表板是一个全面的以太坊估值分析工具，旨在为投资者和分析师提供多维度的 ETH 估值数据和可视化分析。该工具整合链上数据指标、市场数据指标、机构持仓数据、网络健康指标以及宏观经济关联指标，帮助用户从多个角度全面评估 ETH 的当前估值水平，做出更明智的投资决策。

## Glossary

- **Dashboard**: ETH 估值分析仪表板的主界面，用于展示所有估值相关数据和图表
- **On_Chain_Module**: 链上数据模块，负责获取和展示以太坊区块链上的原生数据指标
- **Market_Module**: 市场数据模块，负责获取和展示 ETH 的市场交易相关数据
- **Institutional_Module**: 机构数据模块，负责获取和展示机构投资者（如 ETF、信托基金）的持仓和资金流向数据
- **Network_Module**: 网络健康模块，负责获取和展示以太坊网络运行状态和安全性相关指标
- **Macro_Module**: 宏观经济模块，负责获取和展示 ETH 与宏观经济环境的关联指标
- **Valuation_Engine**: 估值引擎，负责综合各模块数据计算 ETH 的综合估值评分
- **Data_Fetcher**: 数据获取服务，负责从外部 API 和数据源拉取原始数据
- **Chart_Renderer**: 图表渲染器，负责将数据可视化为交互式图表
- **Alert_System**: 预警系统，负责在关键指标达到阈值时通知用户
- **EIP_1559**: 以太坊改进提案 1559，引入了 base fee 销毁机制，使 ETH 具有通缩属性
- **Base_Fee_Burn**: 基础费用销毁，指 EIP-1559 机制下每笔交易的基础费用被永久销毁
- **TVL**: Total Value Locked，锁定总价值，指 DeFi 协议中锁定的资产总价值
- **NVT_Ratio**: Network Value to Transactions Ratio，网络价值与交易量比率，类似于传统金融的市盈率
- **MVRV_Ratio**: Market Value to Realized Value Ratio，市场价值与已实现价值比率
- **Staking_Yield**: 质押收益率，指验证者通过质押 ETH 获得的年化收益率
- **ETF**: Exchange-Traded Fund，交易所交易基金，此处特指以太坊现货 ETF
- **Grayscale_Trust**: 灰度以太坊信托（ETHE），机构级以太坊投资工具
- **Gas_Fee**: 以太坊网络上执行交易或智能合约所需支付的费用
- **Validator**: 以太坊 PoS 共识机制中的验证者节点
- **DeFi**: Decentralized Finance，去中心化金融
- **L2**: Layer 2，以太坊二层扩展网络

## Requirements

### Requirement 1: 仪表板总览与布局

**User Story:** 作为投资分析师，我希望有一个结构清晰的仪表板总览页面，以便快速了解 ETH 的整体估值状态。

#### Acceptance Criteria

1. THE Dashboard SHALL 展示一个包含 ETH 当前价格、24小时涨跌幅、市值排名的顶部摘要栏
2. THE Dashboard SHALL 将所有指标按照链上数据、市场数据、机构数据、网络健康、宏观经济五个维度分区展示
3. THE Dashboard SHALL 展示由 Valuation_Engine 计算的 ETH 综合估值评分（0-100 分），并标注"低估"、"合理"或"高估"状态
4. WHEN 用户首次加载 Dashboard 时，THE Data_Fetcher SHALL 在 5 秒内完成所有模块数据的初始加载
5. THE Dashboard SHALL 支持浅色和深色两种主题模式切换

---

### Requirement 2: EIP-1559 Base Fee 销毁数据

**User Story:** 作为 ETH 持有者，我希望查看 EIP-1559 机制下的 base fee 销毁数据，以便评估 ETH 的通缩程度和内在价值。

#### Acceptance Criteria

1. THE On_Chain_Module SHALL 展示过去 24 小时、7 天、30 天和自 EIP-1559 上线以来的累计 ETH 销毁总量
2. THE On_Chain_Module SHALL 以折线图形式展示每日 Base_Fee_Burn 数量的历史趋势
3. THE On_Chain_Module SHALL 计算并展示当前的年化销毁率（年化销毁量占总供应量的百分比）
4. THE On_Chain_Module SHALL 展示当前 ETH 的净发行量（新发行量减去销毁量），并标注 ETH 当前处于通胀还是通缩状态
5. WHEN 单日 Base_Fee_Burn 数量超过过去 30 天平均值的 200% 时，THE Alert_System SHALL 生成一条异常销毁量预警通知
6. THE Chart_Renderer SHALL 以堆叠面积图形式展示 ETH 供应量变化（新发行 vs 销毁）的历史对比

---

### Requirement 3: Gas 费用与交易费收入

**User Story:** 作为网络分析师，我希望查看以太坊网络的 Gas 费用和交易费收入数据，以便评估网络的经济活动水平和收入能力。

#### Acceptance Criteria

1. THE On_Chain_Module SHALL 展示当前平均 Gas_Fee（以 Gwei 和 USD 计价）
2. THE On_Chain_Module SHALL 以折线图形式展示过去 30 天、90 天和 1 年的每日平均 Gas_Fee 历史趋势
3. THE On_Chain_Module SHALL 计算并展示以太坊网络的日交易费总收入（以 ETH 和 USD 计价）
4. THE On_Chain_Module SHALL 展示交易费收入在验证者小费（priority fee）和 Base_Fee_Burn 之间的分配比例
5. THE On_Chain_Module SHALL 计算并展示基于交易费收入的年化收入，以及对应的市费率（市值 / 年化交易费收入）
6. WHEN 平均 Gas_Fee 超过 50 Gwei 时，THE Alert_System SHALL 标注当前网络处于"高费用"状态

---

### Requirement 4: 链上活跃度指标

**User Story:** 作为链上分析师，我希望查看以太坊网络的活跃度指标，以便评估网络的使用率和增长趋势。

#### Acceptance Criteria

1. THE On_Chain_Module SHALL 展示每日活跃地址数（DAA）及其 7 天移动平均值
2. THE On_Chain_Module SHALL 展示每日交易笔数及其历史趋势图
3. THE On_Chain_Module SHALL 展示每日新增地址数及其历史趋势图
4. THE On_Chain_Module SHALL 计算并展示 NVT_Ratio（网络市值 / 每日链上交易量的 USD 价值），并与历史中位数进行对比
5. THE On_Chain_Module SHALL 展示以太坊主网与主要 L2 网络（Arbitrum、Optimism、Base、zkSync）的每日交易笔数对比
6. WHEN NVT_Ratio 超过历史 90 百分位数时，THE Valuation_Engine SHALL 将该指标标注为"高估信号"
7. WHEN NVT_Ratio 低于历史 10 百分位数时，THE Valuation_Engine SHALL 将该指标标注为"低估信号"

---

### Requirement 5: DeFi 与 TVL 指标

**User Story:** 作为 DeFi 投资者，我希望查看以太坊生态的 TVL 和 DeFi 相关指标，以便评估以太坊作为 DeFi 基础设施的价值。

#### Acceptance Criteria

1. THE On_Chain_Module SHALL 展示以太坊生态（含主网和 L2）的总 TVL（以 USD 和 ETH 计价）
2. THE On_Chain_Module SHALL 以折线图形式展示 TVL 的历史趋势（至少 1 年数据）
3. THE On_Chain_Module SHALL 展示 TVL 在前 10 大 DeFi 协议中的分布（饼图或柱状图）
4. THE On_Chain_Module SHALL 计算并展示 TVL/市值比率，用于评估 ETH 市值相对于其锁定价值的溢价程度
5. THE On_Chain_Module SHALL 展示以太坊 TVL 在所有公链 TVL 中的市场份额占比及其历史变化趋势
6. WHEN 以太坊 TVL 市场份额单周下降超过 3 个百分点时，THE Alert_System SHALL 生成一条市场份额下降预警

---

### Requirement 6: 市场价格与交易数据

**User Story:** 作为交易员，我希望查看 ETH 的全面市场数据，以便了解当前的市场状况和交易活跃度。

#### Acceptance Criteria

1. THE Market_Module SHALL 展示 ETH 的实时价格（USD 计价），并每 10 秒自动刷新一次
2. THE Market_Module SHALL 展示 ETH 的 24 小时交易量、7 天交易量（以 USD 计价）
3. THE Market_Module SHALL 展示 ETH 的流通市值和完全稀释市值
4. THE Market_Module SHALL 以 K 线图形式展示 ETH 的历史价格，支持 1 天、1 周、1 月、3 月、1 年和全部时间范围切换
5. THE Market_Module SHALL 展示 ETH 在主要交易所（Binance、Coinbase、OKX）的价格差异
6. THE Market_Module SHALL 展示 ETH 的历史最高价（ATH）及当前价格距 ATH 的回撤百分比

---

### Requirement 7: 估值模型指标

**User Story:** 作为量化分析师，我希望查看多种估值模型的计算结果，以便从不同角度评估 ETH 的估值水平。

#### Acceptance Criteria

1. THE Valuation_Engine SHALL 计算并展示 MVRV_Ratio（市场价值 / 已实现价值），并标注当前处于历史百分位的位置
2. THE Valuation_Engine SHALL 计算并展示 ETH 的市费率（P/F Ratio = 市值 / 年化费用收入），类似于传统股票的市盈率
3. THE Valuation_Engine SHALL 基于 DCF（折现现金流）模型，使用年化费用收入和销毁量计算 ETH 的理论估值范围
4. THE Valuation_Engine SHALL 计算并展示 ETH 的 Stock-to-Flow 比率及对应的模型预测价格
5. THE Valuation_Engine SHALL 以雷达图形式展示各估值指标的综合评分，直观呈现 ETH 在各维度的估值状态
6. THE Valuation_Engine SHALL 展示每个估值指标的历史分布图，并标注当前值在历史分布中的位置

---

### Requirement 8: ETF 持仓与资金流向

**User Story:** 作为机构投资者，我希望追踪以太坊 ETF 的持仓变化和资金流向，以便评估机构资金对 ETH 价格的影响。

#### Acceptance Criteria

1. THE Institutional_Module SHALL 展示所有已批准的以太坊现货 ETF 的每日持仓量（以 ETH 和 USD 计价）
2. THE Institutional_Module SHALL 展示每只 ETF 的每日净流入/流出量，并以柱状图形式展示历史趋势
3. THE Institutional_Module SHALL 计算并展示所有 ETF 的合计持仓量占 ETH 总流通量的百分比
4. THE Institutional_Module SHALL 以折线图形式展示 ETF 累计净流入量与 ETH 价格的叠加对比图
5. WHEN 单日 ETF 净流入量超过过去 30 天平均值的 300% 时，THE Alert_System SHALL 生成一条异常资金流入预警
6. WHEN 单日 ETF 净流出量超过过去 30 天平均值的 300% 时，THE Alert_System SHALL 生成一条异常资金流出预警
7. THE Institutional_Module SHALL 展示各 ETF 发行商（BlackRock、Fidelity、Grayscale 等）的持仓量排名和市场份额

---

### Requirement 9: 灰度信托与机构持仓

**User Story:** 作为市场观察者，我希望追踪灰度以太坊信托和其他机构级产品的数据，以便了解机构投资者的整体持仓动态。

#### Acceptance Criteria

1. THE Institutional_Module SHALL 展示 Grayscale_Trust（ETHE）的当前持仓量、资产净值（NAV）和市场溢价/折价率
2. THE Institutional_Module SHALL 以折线图形式展示 Grayscale_Trust 溢价/折价率的历史趋势
3. THE Institutional_Module SHALL 展示已知的大型机构（如 MicroStrategy 等）的 ETH 持仓量汇总
4. THE Institutional_Module SHALL 展示 CME 以太坊期货的未平仓合约量及其历史趋势
5. WHEN Grayscale_Trust 的折价率超过 20% 时，THE Alert_System SHALL 生成一条异常折价预警

---

### Requirement 10: 网络质押与验证者数据

**User Story:** 作为质押参与者，我希望查看以太坊网络的质押和验证者数据，以便评估网络安全性和质押收益。

#### Acceptance Criteria

1. THE Network_Module SHALL 展示当前 ETH 总质押量及其占总供应量的百分比
2. THE Network_Module SHALL 展示当前活跃 Validator 数量及其历史增长趋势
3. THE Network_Module SHALL 计算并展示当前的 Staking_Yield（年化质押收益率）
4. THE Network_Module SHALL 展示质押 ETH 的进入和退出队列长度及预计等待时间
5. THE Network_Module SHALL 展示主要流动性质押协议（Lido、Rocket Pool、Coinbase）的市场份额分布
6. THE Network_Module SHALL 以折线图形式展示 Staking_Yield 的历史趋势
7. WHEN 单日验证者退出数量超过 500 个时，THE Alert_System SHALL 生成一条验证者大规模退出预警

---

### Requirement 11: 网络性能与安全指标

**User Story:** 作为技术分析师，我希望查看以太坊网络的性能和安全指标，以便评估网络的健康状态。

#### Acceptance Criteria

1. THE Network_Module SHALL 展示当前的平均出块时间和区块利用率
2. THE Network_Module SHALL 展示网络的每秒交易处理量（TPS）及其与理论最大值的比率
3. THE Network_Module SHALL 展示最近 24 小时的 missed slots 数量和比率
4. THE Network_Module SHALL 展示网络参与率（attestation participation rate）
5. THE Network_Module SHALL 展示客户端多样性分布（Geth、Prysm、Lighthouse 等的占比）
6. WHEN missed slots 比率超过 5% 时，THE Alert_System SHALL 生成一条网络健康预警

---

### Requirement 12: ETH/BTC 相关性与相对估值

**User Story:** 作为加密货币投资者，我希望查看 ETH 与 BTC 的相关性和相对估值数据，以便做出 ETH/BTC 配置决策。

#### Acceptance Criteria

1. THE Macro_Module SHALL 展示 ETH/BTC 交易对的当前价格及历史趋势图
2. THE Macro_Module SHALL 计算并展示 ETH 与 BTC 的 30 天、90 天滚动相关系数
3. THE Macro_Module SHALL 展示 ETH 市值占加密货币总市值的百分比（ETH Dominance）及其历史趋势
4. THE Macro_Module SHALL 展示 ETH/BTC 比率的历史百分位位置
5. WHEN ETH/BTC 比率低于历史 10 百分位数时，THE Valuation_Engine SHALL 将该指标标注为"ETH 相对低估信号"
6. WHEN ETH/BTC 比率高于历史 90 百分位数时，THE Valuation_Engine SHALL 将该指标标注为"ETH 相对高估信号"

---

### Requirement 13: 宏观经济关联指标

**User Story:** 作为宏观策略分析师，我希望查看 ETH 与传统宏观经济指标的关联数据，以便在更广泛的经济背景下评估 ETH 的估值。

#### Acceptance Criteria

1. THE Macro_Module SHALL 展示 ETH 价格与美元指数（DXY）的叠加对比图
2. THE Macro_Module SHALL 展示 ETH 价格与美国 10 年期国债收益率的叠加对比图
3. THE Macro_Module SHALL 展示 ETH 价格与纳斯达克指数的 30 天、90 天滚动相关系数
4. THE Macro_Module SHALL 展示当前美联储基准利率及市场对未来利率路径的预期
5. THE Macro_Module SHALL 展示加密货币恐惧与贪婪指数的当前值及历史趋势
6. THE Macro_Module SHALL 展示稳定币（USDT、USDC）总市值的历史趋势，作为加密市场资金量的代理指标

---

### Requirement 14: 数据刷新与缓存策略

**User Story:** 作为用户，我希望数据能够及时更新且加载迅速，以便获取最新的估值分析数据。

#### Acceptance Criteria

1. THE Data_Fetcher SHALL 按以下频率自动刷新数据：价格数据每 10 秒、链上数据每 5 分钟、机构数据每 1 小时
2. THE Data_Fetcher SHALL 对所有已获取的数据进行本地缓存，缓存有效期与对应数据的刷新频率一致
3. WHEN 用户手动点击刷新按钮时，THE Data_Fetcher SHALL 强制重新获取所有模块的最新数据
4. IF 外部数据源 API 请求失败，THEN THE Data_Fetcher SHALL 展示最近一次缓存的数据，并在界面上标注数据的最后更新时间
5. IF 外部数据源连续 3 次请求失败，THEN THE Alert_System SHALL 在界面上展示数据源不可用的警告信息

---

### Requirement 15: 预警与通知系统

**User Story:** 作为投资者，我希望在关键指标出现异常变化时收到通知，以便及时做出投资决策。

#### Acceptance Criteria

1. THE Alert_System SHALL 支持用户为任意指标自定义预警阈值
2. WHEN 任一指标触发预警条件时，THE Alert_System SHALL 在 Dashboard 顶部展示预警通知横幅
3. THE Alert_System SHALL 维护一个预警历史记录列表，展示过去 30 天内所有触发的预警
4. THE Alert_System SHALL 支持用户启用或禁用特定类别的预警通知
5. WHEN 同时触发 3 个以上预警时，THE Alert_System SHALL 按严重程度（高、中、低）排序展示

---

### Requirement 16: 数据导出与分享

**User Story:** 作为研究分析师，我希望能够导出数据和图表，以便在报告中使用或与团队分享。

#### Acceptance Criteria

1. THE Dashboard SHALL 支持将任意图表导出为 PNG 或 SVG 格式的图片
2. THE Dashboard SHALL 支持将任意数据表格导出为 CSV 格式的文件
3. THE Dashboard SHALL 支持生成当前仪表板状态的可分享链接
4. WHEN 用户点击导出按钮时，THE Dashboard SHALL 在 3 秒内完成文件生成并触发下载

---

### Requirement 17: ETH 供应量动态分析

**User Story:** 作为长期投资者，我希望深入了解 ETH 的供应量动态变化，以便评估 ETH 的稀缺性和长期价值。

#### Acceptance Criteria

1. THE On_Chain_Module SHALL 展示 ETH 的当前总供应量及其历史变化趋势图
2. THE On_Chain_Module SHALL 展示 ETH 供应量的分类分布：质押锁定量、DeFi 锁定量、交易所余额、其他
3. THE On_Chain_Module SHALL 展示交易所 ETH 余额的历史趋势（交易所净流入/流出）
4. THE On_Chain_Module SHALL 计算并展示 ETH 的年化通胀/通缩率
5. WHEN 交易所 ETH 余额单周净流出超过总余额的 5% 时，THE Alert_System SHALL 生成一条大规模提币预警
