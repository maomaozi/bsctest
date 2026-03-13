# BSC Transaction Pre-Filter 使用指南

## 架构概述

交易过滤器在P2P层接收到交易后、进入txpool之前进行拦截，这是最快的拦截点。

### 交易流程
```
P2P接收 → handleTransactions → backend.Handle → [过滤器] → txFetcher.Enqueue → txpool
```

### 拦截位置
- 文件: `eth/handler_eth.go`
- 方法: `Handle()` 中的 `TransactionsPacket` 和 `PooledTransactionsResponse` case
- 时机: 反序列化后、入队前（延迟最低）

## 核心组件

### 1. 过滤器接口 (`core/txfilter/filter.go`)
```go
type TxFilter interface {
    Filter(tx *types.Transaction) bool
}
```

### 2. FourMeme过滤器
识别FourMeme代币创建交易，提取：
- 代币地址（通过CREATE2预测）
- 代币名称、符号
- Quote地址

### 3. 过滤器管理器 (`core/txfilter/manager.go`)
- 支持注册多个过滤器
- 并发安全
- 批量处理

## 使用方法

### 当前实现
过滤器已集成到 `eth/handler.go`，在初始化时自动创建：
```go
h.txFilterManager = txfilter.NewManager()
fourMemeFilter := txfilter.NewFourMemeFilter(nil)
h.txFilterManager.AddFilter(fourMemeFilter)
```

### 添加自定义处理逻辑

修改 `eth/handler.go` 中的初始化代码：

```go
// 定义处理函数
customHandler := func(info *txfilter.TokenInfo, tx *types.Transaction) {
    // 自定义处理逻辑
    // 例如：发送到消息队列、写入数据库、触发webhook等
    log.Info("Custom processing",
        "token", info.TokenAddress.Hex(),
        "txHash", tx.Hash().Hex())
}

// 创建带处理器的过滤器
fourMemeFilter := txfilter.NewFourMemeFilter(customHandler)
h.txFilterManager.AddFilter(fourMemeFilter)
```

### 添加新的过滤器

在 `core/txfilter/filter.go` 中实现 `TxFilter` 接口：

```go
type MyCustomFilter struct {
    // 自定义字段
}

func (f *MyCustomFilter) Filter(tx *types.Transaction) bool {
    // 实现过滤逻辑
    data := tx.Data()
    if len(data) < 4 {
        return false
    }

    // 检查方法选择器
    selector := hex.EncodeToString(data[:4])
    if selector == "your_selector" {
        // 处理匹配的交易
        log.Info("Matched transaction", "hash", tx.Hash().Hex())
        return true
    }
    return false
}
```

然后在 `eth/handler.go` 中注册：
```go
myFilter := &txfilter.MyCustomFilter{}
h.txFilterManager.AddFilter(myFilter)
```

## 性能特点

1. **零拷贝**: 直接处理已解码的交易对象
2. **同步处理**: 不增加goroutine开销
3. **最早拦截**: 在验证和状态检查之前
4. **模块化**: 易于添加/替换过滤器

## 扩展点

### 1. 异步处理
如果处理逻辑耗时，可以使用channel异步处理：

```go
type AsyncFilter struct {
    txChan chan *types.Transaction
}

func (f *AsyncFilter) Filter(tx *types.Transaction) bool {
    select {
    case f.txChan <- tx:
    default:
        // channel满时丢弃
    }
    return true
}
```

### 2. 条件过滤
可以添加配置来动态启用/禁用过滤器：

```go
type ConfigurableFilter struct {
    enabled atomic.Bool
}

func (f *ConfigurableFilter) Filter(tx *types.Transaction) bool {
    if !f.enabled.Load() {
        return false
    }
    // 过滤逻辑
}
```

### 3. 统计指标
添加metrics来监控过滤器性能：

```go
var (
    filteredTxCounter = metrics.NewRegisteredCounter("txfilter/matched", nil)
    filterDuration    = metrics.NewRegisteredTimer("txfilter/duration", nil)
)

func (f *MyFilter) Filter(tx *types.Transaction) bool {
    defer filterDuration.UpdateSince(time.Now())

    if matched {
        filteredTxCounter.Inc(1)
        return true
    }
    return false
}
```

## 注意事项

1. **性能**: 过滤逻辑应尽可能快，避免阻塞交易处理
2. **错误处理**: Filter方法不应panic，内部错误应优雅处理
3. **日志**: 使用适当的日志级别，避免日志洪水
4. **并发**: Manager已处理并发安全，Filter实现也应注意线程安全
