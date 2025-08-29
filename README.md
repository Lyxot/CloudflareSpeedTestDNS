# CloudflareSpeedTestDNS

[![Go Version](https://img.shields.io/github/go-mod/go-version/Lyxot/CloudflareSpeedTestDNS.svg?style=flat-square&label=Go&color=00ADD8&logo=go)](https://github.com/Lyxot/CloudflareSpeedTestDNS/)
[![Release Version](https://img.shields.io/github/v/release/Lyxot/CloudflareSpeedTestDNS.svg?style=flat-square&label=Release&color=00ADD8&logo=github)](https://github.com/Lyxot/CloudflareSpeedTestDNS/releases/latest)
[![GitHub license](https://img.shields.io/github/license/Lyxot/CloudflareSpeedTestDNS.svg?style=flat-square&label=License&color=00ADD8&logo=github)](https://github.com/Lyxot/CloudflareSpeedTestDNS/)
[![GitHub Star](https://img.shields.io/github/stars/Lyxot/CloudflareSpeedTestDNS.svg?style=flat-square&label=Star&color=00ADD8&logo=github)](https://github.com/Lyxot/CloudflareSpeedTestDNS/)

> 🚀 **基于 [XIU2/CloudflareSpeedTest](https://github.com/XIU2/CloudflareSpeedTest) 的增强版本**

一个强大的 CDN 测速工具，支持对 **Cloudflare 等 CDN 服务商** 进行延迟测速、优选 IP 同步到 DNS 解析，并提供持续监控功能。

## ✨ 主要特性

- 🔍 **智能测速**：支持 TCPing 和 HTTPing 两种测速模式
- 📊 **多维度筛选**：基于延迟、丢包率、下载速度等多重条件筛选最优 IP
- 🔄 **自动同步**：支持阿里云 DNS、DNSPod、Cloudflare DNS 服务商
- 📈 **持续监控**：定时检测优选 IP 质量，自动更新解析记录
- 🌐 **多协议支持**：同时支持 IPv4 和 IPv6 地址测速
- 📝 **日志记录**：完整的操作日志，支持文件输出

 **示例站点**：[Best Cloudflare](https://cf.hyli.xyz/)

## 🚀 快速开始

### 📥 下载安装

1. 从 [GitHub Releases](https://github.com/Lyxot/CloudflareSpeedTestDNS/releases) 下载对应系统的可执行文件
2. 解压到任意目录
3. 双击运行 `cfst.exe`（Windows）或 `./cfst`（Linux/macOS），开始测速

<details>
<summary><code><strong>🐧 Linux/macOS 用户点击查看详细安装步骤</strong></code></summary>

> 💡 以下命令仅为示例，请根据实际版本号调整下载链接

``` bash
# 如果是第一次使用，则建议创建新文件夹（后续更新时，跳过该步骤）
mkdir cfstd

# 进入文件夹（后续更新，只需要从这里重复下面的下载、解压命令即可）
cd cfstd

# 下载 CFST 压缩包（自行根据需求替换 URL 中 [版本号] 和 [文件名]）
wget -N https://github.com/Lyxot/CloudflareSpeedTestDNS/releases/download/v3.0.0/cfstd-linux-x86_64-v3.0.0.zip

# 解压（不需要删除旧文件，会直接覆盖，自行根据需求替换 文件名）
unzip -o cfstd-linux-x86_64-v3.0.0.zip

# 赋予执行权限
chmod +x cfstd

# 运行（默认配置）
./cfstd

# 运行（指定配置文件）
./cfstd -c config.toml
```

</details>

<details>
<summary><code><strong>🐳 Docker 用户点击查看详细说明</strong></code></summary>

> ⚠️ 测试 IPv6 需要为 Docker 启用 IPv6 或者使用 host 网络模式
#### Docker

```bash
# 拉取镜像
docker pull lyxot/cfstd:latest

# 运行容器
docker run -it --rm lyxot/cfstd:latest
```

#### Docker Compose（推荐）

```bash
mkdir cfstd && cd cfstd
wget https://raw.githubusercontent.com/Lyxot/CloudflareSpeedTestDNS/master/docker-compose.yml

 # 配置环境变量
vim docker-compose.yml

# 运行容器
docker compose up

# 后续更新
docker compose pull
```

</details>

### 📊 测速结果示例

测速完成后，程序会显示**延迟最低、速度最快的 IP 地址**。以下是典型输出示例：

``` bash
IP 地址           已发送  已接收  丢包率  平均延迟  下载速度(MB/s)  地区码
104.27.200.69     4      4       0.00   146.23    28.64          LAX
172.67.60.78      4      4       0.00   139.82    15.02          SEA
104.25.140.153    4      4       0.00   146.49    14.90          SJC
104.27.192.65     4      4       0.00   140.28    14.07          LAX
172.67.62.214     4      4       0.00   139.29    12.71          LAX
104.27.207.5      4      4       0.00   145.92    11.95          LAX
172.67.54.193     4      4       0.00   146.71    11.55          LAX
104.22.66.8       4      4       0.00   147.42    11.11          SEA
104.27.197.63     4      4       0.00   131.29    10.26          FRA
172.67.58.91      4      4       0.00   140.19    9.14           SJC
...

# ⚠️  注意事项：
# - 如果延迟显示异常低（如 0.xx），请检查是否开启了代理软件
# - 在路由器上运行时，请确保关闭路由器内的代理功能
# - 每次测速结果可能不同，这是正常现象（随机选择 IP 段中的地址）

# 📋 测速流程：
# 1. 延迟测速 → 2. 延迟排序 → 3. 下载测速 → 4. 速度排序 → 5. 输出结果
```

> 🎯 **测速结果第一行即为最优 IP**（延迟最低 + 速度最快）

完整结果将保存为 `result.csv` 文件，可用 Excel、记事本等软件打开查看：

```
IP 地址,已发送,已接收,丢包率,平均延迟,下载速度(MB/s),地区码
104.27.200.69,4,4,0.00,146.23,28.64,LAX
```

## ⚙️ 进阶配置

默认配置适合大多数用户，如需更精确的测速结果，可通过配置文件或环境变量自定义参数。详细配置说明请参考[示例配置文件](conf/config.example.toml)或[环境变量说明](conf/env.md)。

### 🖥️ 命令行参数
```text
参数：
    -c config.toml
        指定TOML配置文件；默认为config.toml，不存在时使用默认参数
    -debug
        调试输出模式；会在一些非预期情况下输出更多日志以便判断原因；(默认 关闭)
    -v
        打印程序版本
    -u
        检查版本更新
    -h
        打印帮助说明
```

### 📖 界面说明

> 💡 为了避免对测速过程中显示的数据产生误解，这里引用 [XIU2/CloudflareSpeedTestDNS](https://github.com/XIU2/CloudflareSpeedTest) 详细解释各个数值的含义

<details>
<summary><code><strong>🔍 点击展开查看详细说明</strong></code></summary>

****

> 该示例把常用参数都给加上了，即为：`-tll 40 -tl 150 -sl 1 -dn 5`，最后输出结果如下：

```bash
# XIU2/CloudflareSpeedTestDNS vX.X.X

开始延迟测速（模式：TCP, 端口：443, 范围：40 ~ 150 ms, 丢包：1.00)
321 / 321 [-----------------------------------------------------------] 可用: 30
开始下载测速（下限：1.00 MB/s, 数量：5, 队列：10）
3 / 5 [-----------------------------------------↗--------------------]
IP 地址           已发送  已接收  丢包率  平均延迟  下载速度(MB/s)  地区码
XXX.XXX.XXX.XXX   4      4       0.00   83.32     3.66           LAX
XXX.XXX.XXX.XXX   4      4       0.00   107.81    2.49           LAX
XXX.XXX.XXX.XXX   4      3       0.25   149.59    1.04           N/A

完整测速结果已写入 result.csv 文件，可使用记事本/表格软件查看。
按下 回车键 或 Ctrl+C 退出。
```

****

> 刚接触 CFST 的人，可能会迷惑**明明延迟测速可用 IP 有 30 个，怎么最后只剩下 3 个了呢？**  
> 下载测速里的队列又是什么意思？难道我下载测速还要排队？

CFST 会先延迟测速，在这过程中进度条右侧会实时显示可用 IP 数量（`可用: 30`），但注意该可用数量指的是**测试通过没有超时的 IP 数量**，和延迟上下限、丢包条件无关。当延迟测速完成后，因为还指定了**延迟上下限、丢包**的条件，所以按照条件过滤后只剩下 `10` 个了（也就是等待下载测速的 `队列：10`）。

即以上示例中，`321` 个 IP 延迟测速完成后，只有 `30` 个 IP 测试通过没有超时，然后根据延迟上下限范围：`40 ~ 150 ms` 及丢包上限条件过滤后，只剩下 `10` 个满足要求的 IP 了。如果你 `-dd` 禁用了下载测速，那么就会直接输出这 `10` 个 IP 了。当然该示例并未禁用，因此接下来软件会继续对这 `10` 个 IP 进行下载测速（`队列：10`）。

> 因为下载测速是单线程一个个 IP 挨着排队测速的，因此等待下载测速的 IP 数量才会叫做 `队列`。

****

> 你可能注意到了，**明明指定了要找到 5 个满足下载速度条件的 IP，怎么才 3 个就 “中断” 了呢？**

下载测速进度条中的 `3 / 5`，前者指的是找到了 `3` 个满足下载速度下限条件的 IP（即下载速度高于 `1 MB/s` ），后者 `5` 指的是你要求找到 `5` 个满足下载速度下限条件的 IP（`-dn 5`）。

> 另外，提醒一下，如果你指定的 `-dn` 大于下载测速队列，比如你延迟测速后只剩下 `4` 个 IP 了，那么下载测速进度条中后面的数字就会和下载测速队列一样都是 `4` 个，而非你 `-dn` 指定的 `5` 个了。

软件在测速完这 `10` 个 IP 后，只找到了 `3` 个下载速度高于 `1 MB/s` 的 IP，剩下的 `7` 个 IP 都是 “不及格” 的。

因此，这不是 `“每次测速都不到 5 就中断了”`，而是所有 IP 都下载测速完了，但却只找到了 `3` 个满足条件的。

****

还有一种情况，那就是当可用 IP 很多时（几百几千），你还设置了下载速度条件，那么可能就会遇到：**怎么下载测速进度条老是卡在 `X / 5` 了呢？**

这其实并不是卡住了，而是只有当找到一个满足条件的 IP 时，进度条才会 +1，因此如果一直找不到，那么 CFST 就会一直下载测速下去，因此在表现为进度条卡住不动，但这也是在提醒你：你设置的下载速度条件对你来说已经高于实际了，你需要适当调低预期。

****

如果不想遇到这种全部测速一遍都没几个满足条件的情况，那么就要**调低下载速度上限参数 `-sl`**，或者移除。

因为只要指定了 `-sl` 参数，那么只要没有凑够 `-dn` 的数量（默认 10 个），就会一直测速下去，直到凑够或全部测速完。移除 `-sl` 并添加 `-dn 20` 参数，这样就是只测速延迟最低的前 20 个 IP，测速完就停止，节省时间。

****

另外，如果全部队列 IP 都测速完了，但一个满足下载速度条件的 IP 都没有，你可能需要调低预期的下载测速下限条件，但你需要知道当前的大概测速速度都在什么范围，那么你就可以加上 `-debug` 参数开启调试模式，这样再遇到这种情况时，就会**忽略条件返回所有测速结果**，你就能看到这些 IP 的下载速度都有多少，心里也就有数了，然后**适当调低 `-sl` 再试试**。

> 注意，如果你**没有指定**下载测速下限 `-sl` 条件，那么无论什么情况下 CFST 都会**输出所有测速结果**。

同样，延迟测速方面，`可用: 30`、`队列：10` 这两个数值也可以让你清楚，你设置的延迟条件对你来说是否过于苛刻。如果可用 IP 一大堆，但条件过滤后只剩下 2、3 个，那不用说就知道需要**调低预期的延迟/丢包条件**了。

这两个机制，一个是告诉你**延迟丢包条件**是否合适的，一个是告诉你**下载速度条件**是否合适的。

</details>

### 🌐 IPv4/IPv6 分离测速

通过以下配置项可分别指定 IPv4 和 IPv6 的测速数据：

- `ipv4_file`：IPv4 段数据文件路径
- `ipv6_file`：IPv6 段数据文件路径

> ⚠️ 当指定了任意一个文件时，`ip_file` 配置将失效

**同时指定两个文件时**：
- 将分别对 IPv4 和 IPv6 进行测速
- 结果文件会自动分离：`result.csv` → `result_ipv4.csv` + `result_ipv6.csv`

### 🔄 DNS 自动同步

#### 阿里云 DNS
修改 config 中的 `alidns` 部分：

- `enable`：是否启用阿里云 DNS (默认 false)
- `accesskey_id`：阿里云 AccessKey ID
- `accesskey_secret`：阿里云 AccessKey Secret
- `domain`：域名
- `subdomain`：子域名
- `ttl`：TTL 值 (默认 600)

> 🔑 **获取 AccessKey**：[阿里云 RAM 控制台](https://ram.console.aliyun.com/profile/access-keys)

#### DNSPod DNS
修改 config 中的 `dnspod` 部分：

- `enable`：是否启用 DNSPod DNS (默认 false)
- `secret_id`：DNSPod Secret ID
- `secret_key`：DNSPod Secret Key
- `domain`：域名
- `subdomain`：子域名
- `ttl`：TTL 值 (默认 600)

> 🔑 **获取 API 密钥**：[DNSPod 控制台](https://console.dnspod.cn/account/token/apikey)
> 
> ⚠️ 需要使用腾讯云 API 密钥

#### Cloudflare DNS
修改 config 中的 `cloudflare` 部分：

- `enable`：是否启用 Cloudflare DNS (默认 false)
- `api_token`：Cloudflare API Token
- `zone_id`：Cloudflare 域名的 Zone ID
- `domain`：域名
- `subdomain`：子域名
- `proxied`：是否开启 Cloudflare 代理 (默认 false)
- `ttl`：TTL 值 (1为自动，默认 1)

> 🔑 **获取 API Token**：[Cloudflare 控制台](https://dash.cloudflare.com/profile/api-tokens)
> 
> 📝 **创建步骤**：创建令牌 → 使用模板 → 编辑区域 DNS → 区域资源：`包括` `账户的所有区域` `xxx's Account`

#### Cloudflare Workers KV
修改 config 中的 `cfkv` 部分：

- `enable`：是否启用 Cloudflare KV (默认 false)
- `api_token`：Cloudflare API Token
- `account_id`：Cloudflare Account ID
- `namespace_id`：Cloudflare KV Namespace ID

> 🔑 **获取 API Token**：[Cloudflare 控制台](https://dash.cloudflare.com/profile/api-tokens)
> 
> 📝 **创建步骤**：创建令牌 → 创建自定义令牌 → 权限：`帐户` `Workers KV 存储` `编辑` → 帐户资源：`包括` `xxx's Account`

同步到 Workers KV 的格式如下：
- `ipv4`：`IP`,`已发送`,`已接收`,`丢包率`,`平均延迟`,`下载速度`,`数据中心`&...
- `ipv4time`：更新时间
- `ipv6`：`IP`,`已发送`,`已接收`,`丢包率`,`平均延迟`,`下载速度`,`数据中心`&...
- `ipv6time`：更新时间

> 💡 **推荐搭配 Cloudflare Workers 使用**
<details>
<summary><code><strong>🚀 点击查看 Cloudflare Workers 部署步骤</strong></code></summary>

1. **访问 [Cloudflare Dashboard](https://dash.cloudflare.com)**

2. **创建 KV 存储**：
   - 进入"存储和数据库" → "KV"
   - 点击"Create Instance"
   - 记录 Namespace ID 并填入配置文件

3. **创建 Workers**：
   - 进入"Workers" → "创建"
   - 选择"从 Hello World! 开始"
   - 部署后编辑代码，复制 [worker.js](worker.js) 内容
   - 修改顶部的配置后重新部署

4. **绑定 Workers KV**：
   - 进入你的 Workers → "绑定"
   - 添加绑定 → "KV 命名空间"
   - 变量名称：`KV_NAMESPACE`
   - KV 命名空间：选择刚创建的存储
   - 添加绑定
</details>


### 📈 持续监控

修改 config 中的 `cron` 部分：

- `enable`：是否启用定时任务 (默认 false)
- `latency_threshold`：延迟阈值 (毫秒，默认 9999)
- `loss_rate_threshold`：丢包率阈值 (默认 1.0)
- `check_interval`：检测间隔 (分钟，默认 30)
- `test_interval`：强制刷新间隔 (小时，默认 24)

> 🔄 **监控机制**：
- 每隔 `test_interval` 重新测速并更新 DNS 记录
- 每隔 `check_interval` 分钟检测优选 IP 的延迟、丢包率，当延迟或丢包率超过阈值时自动重新测速并更新 DNS 记录

## 🙏 致谢

本项目基于以下优秀项目开发：

- **[XIU2/CloudflareSpeedTest](https://github.com/XIU2/CloudflareSpeedTest)** - 核心测速功能
- **[IonRh/Cloudflare-BestIP](https://github.com/IonRh/Cloudflare-BestIP)** - 项目灵感来源

****

## 🔧 从源码编译

如需从源码编译，请使用以下命令：

```bash
go build -ldflags "-s -w -X main.version=v1.0.0 -X main.gitCommit=xxxxxxx"
```

> 💡 编译时会将版本号和 Git 提交哈希写入可执行文件

## 📄 开源协议

本项目采用 GPL-3.0 开源协议。
