# 百度云 CDN 证书清理工具

![Go Version](https://img.shields.io/badge/Go-%3E%3D1.18-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)

基于 [bce-sdk-go](https://github.com/baidubce/bce-sdk-go) 实现的 SSL 证书自动化清理工具，用于清理**未使用且已过期**的证书。

## 快速开始

```bash
# 1. 设置环境变量
export BCE_ACCESS_KEY="your-access-key"
export BCE_SECRET_KEY="your-secret-key"

# 2. 查看所有证书
./bce-cert-cleaner -list-all

# 3. 模拟运行（安全预览）
./bce-cert-cleaner -dry-run

# 4. 交互式删除（推荐首次使用）
./bce-cert-cleaner -interactive
```

## 功能特性

- ✅ 列出所有 SSL 证书及其状态
- ✅ 自动检测 CDN 域名中正在使用的证书
- ✅ 筛选「未使用」且「已过期」的证书进行清理
- ✅ 支持 `-dry-run` 模拟运行模式
- ✅ 支持 `-interactive` 逐个确认删除模式
- ✅ 支持 `-auto` 自动删除模式（用于 crontab）
- ✅ 支持 `-quiet` 静默模式
- ✅ 支持 `-log` 日志文件输出
- ✅ 多种删除模式：批量确认、逐个确认、自动删除

## 安全说明

脚本只会删除**同时满足以下两个条件**的证书:

| 证书状态            | 是否删除   |
| ------------------- | ---------- |
| 使用中 + 未过期     | ❌ 不删除  |
| 使用中 + 已过期     | ❌ 不删除  |
| 未使用 + 未过期     | ❌ 不删除  |
| **未使用 + 已过期** | ✅ 删除    |

**不会误删正常证书**，只清理那些既没有被 CDN 域名使用、又已经过期的无效证书。

## 安装

### 方式一：源码编译

```bash
git clone https://github.com/alex1528/bce-cert-cleaner.git
cd bce-cert-cleaner
go mod tidy
go build -o bce-cert-cleaner
```

### 方式二：直接下载二进制

```bash
# 将编译好的二进制放到系统路径
sudo cp bce-cert-cleaner /usr/local/bin/
sudo chmod +x /usr/local/bin/bce-cert-cleaner
```

## 参数说明

| 参数           | 说明                                       | 默认值                       |
| -------------- | ------------------------------------------ | ---------------------------- |
| `-ak`          | 百度云 Access Key                          | 环境变量 `BCE_ACCESS_KEY`    |
| `-sk`          | 百度云 Secret Key                          | 环境变量 `BCE_SECRET_KEY`    |
| `-list-all`    | 列出所有证书详情                           | `false`                      |
| `-dry-run`     | 模拟运行，仅显示待删除证书，不执行删除     | `false`                      |
| `-interactive` | 交互模式，逐个确认删除每个证书             | `false`                      |
| `-auto`        | 自动模式，无需确认直接删除（用于 crontab） | `false`                      |
| `-quiet`       | 静默模式，仅输出错误和删除结果             | `false`                      |
| `-log`         | 日志文件路径                               | 空（输出到 stdout）          |
| `-version`     | 显示版本信息                               | -                            |

## 删除模式对比

| 模式         | 参数组合       | 行为说明                                                   | 适用场景           |
| ------------ | -------------- | ---------------------------------------------------------- | ------------------ |
| **批量确认** | 无特殊参数     | 显示所有待删除证书，输入 `yes` 一次性删除全部             | 快速批量清理       |
| **逐个确认** | `-interactive` | 对每个证书单独确认（y/n/a/q），可选择性删除               | 谨慎筛选删除       |
| **自动删除** | `-auto`        | 无需确认，直接删除所有未使用且过期的证书（用于定时任务）   | Crontab 定时任务   |
| **模拟运行** | `-dry-run`     | 仅显示待删除证书，不执行任何删除操作                       | 预览确认删除范围   |

### 交互模式操作说明

**重要提示**：

- `-interactive` 参数**仅对「未使用且已过期」的证书有效**
- 如果所有证书都是正常证书（使用中或未过期），该参数不会生效，工具会提示当前证书状态
- 不能与 `-auto` 参数同时使用

在 `-interactive` 模式下，对每个证书可以选择：

- `y` / `yes` - 删除当前证书，继续下一个
- `n` / `no` - 跳过当前证书，继续下一个
- `a` / `all` - 删除当前证书及后续所有证书
- `q` / `quit` - 退出，不再处理后续证书

## 使用方法

### 1. 设置凭证

#### 方式一：环境变量（推荐）

```bash
export BCE_ACCESS_KEY="your-access-key"
export BCE_SECRET_KEY="your-secret-key"
```

#### 方式二：命令行参数

```bash
./bce-cert-cleaner -ak "your-access-key" -sk "your-secret-key" [其他参数]
```

### 2. 常用命令

```bash
# 查看帮助
./bce-cert-cleaner -help

# 查看版本
./bce-cert-cleaner -version

# 列出所有证书及状态统计
./bce-cert-cleaner -list-all

# 模拟运行 - 仅显示将被删除的证书，不实际删除
./bce-cert-cleaner -dry-run

# 批量确认删除（需输入 yes 确认）
./bce-cert-cleaner

# 逐个确认删除（交互模式，可选择性删除）
./bce-cert-cleaner -interactive

# 自动删除（无需确认，用于脚本/crontab）
./bce-cert-cleaner -auto

# 静默自动删除（仅输出错误）
./bce-cert-cleaner -auto -quiet

# 自动删除并记录日志
./bce-cert-cleaner -auto -quiet -log /var/log/bce-cert-cleaner.log
```

## Crontab 定时任务配置

### 方式一：直接使用命令行参数

```bash
# 编辑 crontab
crontab -e
```

添加定时任务：

```bash
# 每天凌晨 3 点执行清理
0 3 * * * /usr/local/bin/bce-cert-cleaner -ak "YOUR_AK" -sk "YOUR_SK" -auto -quiet -log /var/log/bce-cert-cleaner.log

# 每周一凌晨 2 点执行清理
0 2 * * 1 /usr/local/bin/bce-cert-cleaner -ak "YOUR_AK" -sk "YOUR_SK" -auto -quiet -log /var/log/bce-cert-cleaner.log

# 每月 1 号凌晨 4 点执行清理
0 4 1 * * /usr/local/bin/bce-cert-cleaner -ak "YOUR_AK" -sk "YOUR_SK" -auto -quiet -log /var/log/bce-cert-cleaner.log
```

### 方式二：使用环境变量

```bash
crontab -e
```

在 crontab 文件中添加：

```bash
# 在 crontab 中设置环境变量
BCE_ACCESS_KEY=your-access-key
BCE_SECRET_KEY=your-secret-key

# 每天凌晨 3 点执行
0 3 * * * /usr/local/bin/bce-cert-cleaner -auto -quiet -log /var/log/bce-cert-cleaner.log
```

### 方式三：使用 Wrapper 脚本（推荐）

创建脚本 `/usr/local/bin/clean-bce-certs.sh`:

```bash
#!/bin/bash
# 百度云证书清理脚本

export BCE_ACCESS_KEY="your-access-key"
export BCE_SECRET_KEY="your-secret-key"

LOG_FILE="/var/log/bce-cert-cleaner.log"

echo "========== $(date '+%Y-%m-%d %H:%M:%S') 开始执行 ==========" >> "$LOG_FILE"
/usr/local/bin/bce-cert-cleaner -auto -log "$LOG_FILE"
echo "========== $(date '+%Y-%m-%d %H:%M:%S') 执行结束 ==========" >> "$LOG_FILE"
```

设置执行权限：

```bash
sudo chmod +x /usr/local/bin/clean-bce-certs.sh
```

配置 crontab：

```bash
# 每天凌晨 3 点执行
0 3 * * * /usr/local/bin/clean-bce-certs.sh
```

## 日志轮转配置（可选）

创建 `/etc/logrotate.d/bce-cert-cleaner`:

```text
/var/log/bce-cert-cleaner.log {
    weekly
    rotate 4
    compress
    missingok
    notifempty
    create 644 root root
}
```

## 输出示例

### 列出所有证书 (`-list-all`)

```text
共获取到 23 个证书
正在检查 12 个 CDN 域名的证书使用情况...
CDN 正在使用 0 个证书

所有证书列表:
================================================================================
证书ID             证书名称                       使用中   过期时间
--------------------------------------------------------------------------------
cert-xj4k6gp9hzwn  yiming68_com                   否       2026-04-30 (剩余80天)
cert-a9w6v40dsmjh  magicrepokit_com               否       2026-05-06 (剩余86天)
cert-d5bsxfvpawgr  example_cn                     否       2026-05-10 (剩余90天)
cert-pxhmmb2n6azs  expired_example_com            否       2025-11-15 (已过期86天)
...
================================================================================
统计: 总计 23 个证书
  - 使用中: 0 个
  - 未使用: 23 个
  - 已过期: 2 个
  - 未使用且已过期: 2 个 (可清理)
```

### 模拟运行 (`-dry-run`)

```text
共获取到 23 个证书
正在检查 12 个 CDN 域名的证书使用情况...
CDN 正在使用 0 个证书

发现 2 个未使用且已过期的证书:
================================================================================
[1] 证书ID: cert-pxhmmb2n6azs
    证书名称: expired_example_com
    通用名称: *.example.com
    过期时间: 2025-11-15 00:00:00 (已过期 86 天)
--------------------------------------------------------------------------------
[2] 证书ID: cert-abc123xyz
    证书名称: old_domain_cert
    通用名称: old.domain.com
    过期时间: 2025-12-01 00:00:00 (已过期 70 天)
--------------------------------------------------------------------------------
[模拟运行] 以上证书将在非模拟模式下被删除
如需执行删除，请去掉 -dry-run 参数运行
```

### 无过期证书时

```text
共获取到 23 个证书
正在检查 12 个 CDN 域名的证书使用情况...
CDN 正在使用 0 个证书
✓ 没有发现未使用且已过期的证书
```

### 使用交互模式但无符合条件证书时

```text
共获取到 23 个证书
正在检查 12 个 CDN 域名的证书使用情况...
CDN 正在使用 0 个证书
✓ 没有发现未使用且已过期的证书

提示: -interactive 参数仅对「未使用且已过期」的证书有效
当前所有证书均为正常证书（使用中或未过期），无需使用此参数

当前证书状态统计:
  - 总计: 23 个
  - 使用中: 5 个
  - 未使用: 18 个
  - 已过期: 0 个
  - 未使用且已过期: 0 个
```

### 交互模式删除 (`-interactive`)

```text
共获取到 23 个证书
正在检查 12 个 CDN 域名的证书使用情况...
CDN 正在使用 0 个证书

发现 3 个未使用且已过期的证书:
================================================================================
[1] 证书ID: cert-pxhmmb2n6azs
    证书名称: expired_example_com
    通用名称: *.example.com
    过期时间: 2025-11-15 00:00:00 (已过期 86 天)
--------------------------------------------------------------------------------
[2] 证书ID: cert-abc123xyz
    证书名称: old_domain_cert
    通用名称: old.domain.com
    过期时间: 2025-12-01 00:00:00 (已过期 70 天)
--------------------------------------------------------------------------------
[3] 证书ID: cert-def456uvw
    证书名称: another_old_cert
    通用名称: another.domain.com
    过期时间: 2025-11-25 00:00:00 (已过期 76 天)
--------------------------------------------------------------------------------

[1/3] 证书ID: cert-pxhmmb2n6azs
      证书名称: expired_example_com
      通用名称: *.example.com
      过期时间: 2025-11-15 00:00:00 (已过期 86 天)

删除此证书？(y=是, n=否, a=全部, q=退出): y
✓ 删除证书成功: expired_example_com (cert-pxhmmb2n6azs)

[2/3] 证书ID: cert-abc123xyz
      证书名称: old_domain_cert
      通用名称: old.domain.com
      过期时间: 2025-12-01 00:00:00 (已过期 70 天)

删除此证书？(y=是, n=否, a=全部, q=退出): n
⊘ 已跳过: old_domain_cert (cert-abc123xyz)

[3/3] 证书ID: cert-def456uvw
      证书名称: another_old_cert
      通用名称: another.domain.com
      过期时间: 2025-11-25 00:00:00 (已过期 76 天)

删除此证书？(y=是, n=否, a=全部, q=退出): y
✓ 删除证书成功: another_old_cert (cert-def456uvw)

================================================================================
删除完成: 成功 2 个, 跳过 1 个, 失败 0 个
```

### 批量确认删除

```text
共获取到 23 个证书
正在检查 12 个 CDN 域名的证书使用情况...
CDN 正在使用 0 个证书

发现 2 个未使用且已过期的证书:
================================================================================
[1] 证书ID: cert-pxhmmb2n6azs
    证书名称: expired_example_com
    ...

确认要删除以上 2 个证书吗？(输入 yes 确认): yes

开始删除过期未使用的证书...
✓ 删除证书成功: expired_example_com (cert-pxhmmb2n6azs)
✓ 删除证书成功: old_domain_cert (cert-abc123xyz)
================================================================================
删除完成: 成功 2 个, 失败 0 个
```

### 日志文件示例

```text
2026/02/09 03:00:01 [INFO] 开始清理 2 个未使用且已过期的证书
2026/02/09 03:00:02 [INFO] 删除证书成功: expired_example_com (cert-pxhmmb2n6azs)
2026/02/09 03:00:02 [INFO] 删除证书成功: old_domain_cert (cert-abc123xyz)
2026/02/09 03:00:02 [INFO] 清理完成: 成功 2 个, 失败 0 个
```

## 退出码

| 退出码 | 说明                               |
| ------ | ---------------------------------- |
| `0`    | 成功（无证书需清理或全部删除成功） |
| `1`    | 失败（有证书删除失败或配置错误）   |

## 最佳实践

1. **首次使用**：先用 `-list-all` 查看所有证书状态
2. **确认范围**：用 `-dry-run` 确认将被删除的证书
3. **谨慎删除**：使用 `-interactive` 逐个确认，选择性删除证书
4. **批量删除**：确认无误后，不带 `-auto` 参数批量删除（需输入 `yes`）
5. **自动化**：配置 crontab 使用 `-auto -quiet -log` 参数定时清理

## 注意事项

⚠️ **删除操作不可逆，请先用 `-dry-run` 确认**

⚠️ **本工具仅检查 CDN 域名的证书使用情况**

⚠️ **如证书用于其他百度云服务（BLB、WAF、GAAP 等），请自行确认后再删除**

⚠️ **建议定时任务配合日志文件，便于问题排查**

⚠️ **参数使用限制**：

- `-interactive` 参数仅对未使用且已过期的证书有效，不适用于正常证书
- `-auto` 和 `-interactive` 不能同时使用（一个是自动化，一个是交互式）
- `-quiet` 在交互模式下会被忽略（交互模式需要显示提示信息）

## 常见问题

### Q: 工具会删除正在使用的证书吗？

**A:** 不会。工具会自动检查 CDN 域名的证书使用情况，只删除**同时满足「未使用」且「已过期」**两个条件的证书。

### Q: 如何确保不会误删证书？

**A:** 推荐使用三步确认流程：

1. 使用 `-list-all` 查看所有证书状态
2. 使用 `-dry-run` 预览将被删除的证书
3. 使用 `-interactive` 逐个确认删除

### Q: 证书用于其他百度云服务怎么办？

**A:** 本工具仅检查 CDN 域名的使用情况。如果证书还用于 BLB、WAF、GAAP 等其他服务，请在删除前自行确认。

### Q: 删除操作可以撤销吗？

**A:** 不可以。删除操作不可逆，请务必谨慎操作。

### Q: 可以在 Windows 上使用吗？

**A:** 可以。工具支持 Linux、macOS 和 Windows 平台。

## 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 更新日志

### v1.0.0 (2026-02-09)

- ✨ 初始版本发布
- ✅ 支持列出所有 SSL 证书
- ✅ 支持检测 CDN 域名证书使用情况
- ✅ 支持批量确认删除模式
- ✅ 支持逐个确认删除模式（-interactive）
- ✅ 支持自动删除模式（-auto）
- ✅ 支持模拟运行模式（-dry-run）
- ✅ 支持静默模式和日志输出

## 依赖

- Go 1.18+
- [bce-sdk-go](https://github.com/baidubce/bce-sdk-go) v0.9.188+

## License

MIT

## 鸣谢

感谢 [bce-sdk-go](https://github.com/baidubce/bce-sdk-go) 项目提供的 SDK 支持。

---

**⚠️ 免责声明**：本工具仅供学习和自动化管理使用，删除操作不可逆，使用前请务必确认。作者不对因使用本工具导致的任何损失负责。
