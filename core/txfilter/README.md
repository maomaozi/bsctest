# TxFilter Bundle Trading 集成说明

## 配置文件

在geth数据目录下创建 `txfilter.json`:

```json
{
  "private_key": "your_private_key_without_0x",
  "buy_amount_bnb": 0.1,
  "bribe_amount_bnb": 0.00007,
  "sell_delay_seconds": 0,
  "http_rpc": "http://localhost:8545"
}
```

## 集成方式

### 方式1: 在节点启动时初始化

在 `eth/backend.go` 的 `New()` 函数中添加:

```go
import "github.com/ethereum/go-ethereum/core/txfilter"

// 在 New() 函数末尾添加
configPath := txfilter.GetDefaultConfigPath(config.DataDir)
if err := txfilter.InitFromConfigFile(configPath); err != nil {
    log.Error("Failed to init txfilter", "err", err)
}
```

### 方式2: 使用命令行参数

在 `cmd/geth/config.go` 添加flag:

```go
var txfilterConfigFlag = &cli.StringFlag{
    Name:  "txfilter.config",
    Usage: "Path to txfilter config file",
    Value: "txfilter.json",
}
```

然后在启动时调用 `txfilter.InitFromConfigFile(ctx.String("txfilter.config"))`

## 注册Filter

在需要使用的地方注册filter:

```go
import "github.com/ethereum/go-ethereum/core/txfilter"

manager := txfilter.NewManager()
filter := txfilter.NewFourMemeFilter(txfilter.FourMemeHandler)
manager.AddFilter(filter)

// 处理交易
manager.ProcessTransaction(tx)
```

## 配置参数说明

- `private_key`: 交易私钥（不含0x前缀）
- `buy_amount_bnb`: 每次购买的BNB数量
- `bribe_amount_bnb`: 给builder的贿赂金额
- `sell_delay_seconds`: 购买后延迟卖出的秒数
- `http_rpc`: BSC RPC节点地址
