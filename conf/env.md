# 环境变量

您可以使用环境变量来覆盖配置文件中的部分或全部设置。
环境变量名称派生自配置文件中的键，前缀为 `CFSTD_`，并采用 `大写蛇形命名法`。

| 环境变量 | 默认值 | 说明 |
| --- | --- | --- |
| `CFSTD_ROUTINES` | `200` | 延迟测速线程数 |
| `CFSTD_PING_TIMES` | `4` | 延迟测速次数 |
| `CFSTD_TCP_PORT` | `443` | 指定测速端口 |
| `CFSTD_MAX_DELAY` | `9999` | 平均延迟上限，单位毫秒 |
| `CFSTD_MIN_DELAY` | `0` | 平均延迟下限，单位毫秒 |
| `CFSTD_MAX_LOSS_RATE` | `1.0` | 丢包几率上限，范围 0.00~1.00 |
| `CFSTD_HTTPING` | `false` | 切换测速模式为 HTTP |
| `CFSTD_HTTPING_CODE` | `0` | 有效状态代码 (0 表示 200, 301, 302) |
| `CFSTD_CFCOLO` | `""` | 匹配指定地区，IATA 机场地区码或国家/城市码 |
| `CFSTD_TEST_COUNT` | `10` | 下载测速数量 |
| `CFSTD_DOWNLOAD_TIME` | `10` | 下载测速时间，单位秒 |
| `CFSTD_URL` | `"https://cf.xiu2.xyz/url"` | 指定测速地址 |
| `CFSTD_MIN_SPEED` | `0.0` | 下载速度下限，单位 MB/s |
| `CFSTD_DISABLE_DOWNLOAD` | `false` | 禁用下载测速 |
| `CFSTD_PRINT_NUM` | `10` | 显示结果数量 |
| `CFSTD_IPV4_FILE` | `""` | IPv4段数据文件路径或 URL |
| `CFSTD_IPV6_FILE` | `""` | IPv6段数据文件路径或 URL |
| `CFSTD_IP_FILE` | `"ip.txt"` | IP段数据文件路径或 URL |
| `CFSTD_IP_TEXT` | `""` | 指定IP段数据，英文逗号分隔 |
| `CFSTD_OUTPUT` | `"result.csv"` | 输出结果文件 |
| `CFSTD_LOG_FILE` | `""` | 日志文件 |
| `CFSTD_TEST_ALL` | `false` | 测速全部IP |
| `CFSTD_DEBUG` | `false` | 调试输出模式 |
| | | |
| **[alidns]** | | |
| `CFSTD_ALIDNS_ENABLE` | `false` | 是否启用阿里云DNS |
| `CFSTD_ALIDNS_ACCESS_KEY_ID` | `""` | 阿里云AccessKeyID |
| `CFSTD_ALIDNS_ACCESS_KEY_SECRET` | `""` | 阿里云AccessKeySecret |
| `CFSTD_ALIDNS_DOMAIN` | `""` | 域名 |
| `CFSTD_ALIDNS_SUBDOMAIN` | `""` | 子域名 |
| `CFSTD_ALIDNS_TTL` | `600` | TTL |
| | | |
| **[dnspod]** | | |
| `CFSTD_DNSPOD_ENABLE` | `false` | 是否启用DNSPod DNS |
| `CFSTD_DNSPOD_SECRET_ID` | `""` | DNSPod Secret ID |
| `CFSTD_DNSPOD_SECRET_KEY` | `""` | DNSPod Secret Key |
| `CFSTD_DNSPOD_DOMAIN` | `""` | 域名 |
| `CFSTD_DNSPOD_SUBDOMAIN` | `""` | 子域名 |
| `CFSTD_DNSPOD_TTL` | `600` | TTL |
| | | |
| **[cloudflare]** | | |
| `CFSTD_CLOUDFLARE_ENABLE` | `false` | 是否启用Cloudflare DNS |
| `CFSTD_CLOUDFLARE_API_TOKEN` | `""` | Cloudflare API Token |
| `CFSTD_CLOUDFLARE_ZONE_ID` | `""` | Cloudflare 域名 Zone ID |
| `CFSTD_CLOUDFLARE_DOMAIN` | `""` | 域名 |
| `CFSTD_CLOUDFLARE_SUBDOMAIN` | `""` | 子域名 |
| `CFSTD_CLOUDFLARE_PROXIED` | `false` | 是否开启Cloudflare代理 |
| `CFSTD_CLOUDFLARE_TTL` | `1` | TTL (1为自动) |
| | | |
| **[cfkv]** | | |
| `CFSTD_CFKV_ENABLE` | `false` | 是否启用Cloudflare KV |
| `CFSTD_CFKV_API_TOKEN` | `""` | Cloudflare API Token |
| `CFSTD_CFKV_ACCOUNT_ID` | `""` | Cloudflare Account ID |
| `CFSTD_CFKV_NAMESPACE_ID` | `""` | Cloudflare KV Namespace ID |
| | | |
| **[cron]** | | |
| `CFSTD_CRON_ENABLE` | `false` | 是否启用定时任务 |
| `CFSTD_CRON_LATENCY_THRESHOLD` | `9999` | 延迟阈值(毫秒) |
| `CFSTD_CRON_LOSS_RATE_THRESHOLD` | `1.0` | 丢包率阈值 |
| `CFSTD_CRON_CHECK_INTERVAL` | `30` | 检测间隔(分钟) |
| `CFSTD_CRON_TEST_INTERVAL` | `24` | 强制刷新间隔(小时) |