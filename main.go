package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/baidubce/bce-sdk-go/services/cdn"
	"github.com/baidubce/bce-sdk-go/services/cert"
)

// CertInfo 证书信息
type CertInfo struct {
	CertID         string
	CertName       string
	CertCommonName string
	CertStartTime  string
	CertStopTime   string
	ExpireTime     time.Time
	InUse          bool
}

var (
	version   = "1.0.0"
	buildTime = "unknown"
)

func main() {
	// 命令行参数
	dryRun := flag.Bool("dry-run", false, "模拟运行，仅列出未使用且过期的证书，不执行删除")
	listAll := flag.Bool("list-all", false, "列出所有证书信息")
	auto := flag.Bool("auto", false, "自动模式，无需确认直接删除（用于 crontab）")
	interactive := flag.Bool("interactive", false, "交互模式，逐个确认删除每个证书")
	quiet := flag.Bool("quiet", false, "静默模式，仅输出错误和删除结果")
	logFile := flag.String("log", "", "日志文件路径（用于 crontab）")
	accessKey := flag.String("ak", os.Getenv("BCE_ACCESS_KEY"), "百度云 Access Key")
	secretKey := flag.String("sk", os.Getenv("BCE_SECRET_KEY"), "百度云 Secret Key")
	showVersion := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	// 显示版本
	if *showVersion {
		fmt.Printf("bce-cert-cleaner version %s (built: %s)\n", version, buildTime)
		os.Exit(0)
	}

	// 设置日志输出
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("无法打开日志文件: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
		log.SetFlags(log.Ldate | log.Ltime)
	}

	// 检查参数冲突
	if *auto && *interactive {
		fmt.Println("错误: -auto 和 -interactive 参数不能同时使用")
		fmt.Println("  -auto: 自动删除所有未使用且已过期的证书（用于定时任务）")
		fmt.Println("  -interactive: 逐个确认删除每个证书（交互模式）")
		os.Exit(1)
	}

	if *accessKey == "" || *secretKey == "" {
		msg := `错误: 请设置 BCE_ACCESS_KEY 和 BCE_SECRET_KEY 环境变量，或使用 -ak 和 -sk 参数

使用方法:
  # 设置环境变量
  export BCE_ACCESS_KEY='your-access-key'
  export BCE_SECRET_KEY='your-secret-key'

  # 列出所有证书
  ./bce-cert-cleaner -list-all

  # 模拟运行（仅显示，不删除）
  ./bce-cert-cleaner -dry-run

  # 批量确认删除（需输入 yes 确认）
  ./bce-cert-cleaner

  # 逐个确认删除（交互模式）
  ./bce-cert-cleaner -interactive

  # 自动删除（用于 crontab）
  ./bce-cert-cleaner -auto

  # crontab 定时任务示例（每天凌晨 3 点执行）
  0 3 * * * /path/to/bce-cert-cleaner -ak "YOUR_AK" -sk "YOUR_SK" -auto -quiet -log /var/log/cert-cleaner.log`
		fmt.Println(msg)
		os.Exit(1)
	}

	// 创建 Cert 客户端
	certClient, err := cert.NewClient(*accessKey, *secretKey, "certificate.baidubce.com")
	if err != nil {
		logError("创建证书客户端失败: %v", err)
		os.Exit(1)
	}

	// 创建 CDN 客户端
	cdnClient, err := cdn.NewClient(*accessKey, *secretKey, "")
	if err != nil {
		logError("创建 CDN 客户端失败: %v", err)
		os.Exit(1)
	}

	// 获取所有证书
	certs, err := getAllCerts(certClient)
	if err != nil {
		logError("获取证书列表失败: %v", err)
		os.Exit(1)
	}

	if !*quiet {
		fmt.Printf("共获取到 %d 个证书\n", len(certs))
	}

	// 获取 CDN 域名使用的证书
	usedCertIDs, err := getUsedCertIDs(cdnClient, *quiet)
	if err != nil {
		logError("获取 CDN 域名证书信息失败: %v", err)
		os.Exit(1)
	}

	if !*quiet {
		fmt.Printf("CDN 正在使用 %d 个证书\n\n", len(usedCertIDs))
	}

	// 标记使用状态
	for i := range certs {
		certs[i].InUse = usedCertIDs[certs[i].CertID]
	}

	// 列出所有证书
	if *listAll {
		listAllCerts(certs)
		return
	}

	// 筛选未使用且过期的证书
	now := time.Now()
	var expiredUnusedCerts []CertInfo

	for _, c := range certs {
		if !c.InUse && !c.ExpireTime.IsZero() && c.ExpireTime.Before(now) {
			expiredUnusedCerts = append(expiredUnusedCerts, c)
		}
	}

	if len(expiredUnusedCerts) == 0 {
		if !*quiet {
			fmt.Println("✓ 没有发现未使用且已过期的证书")

			// 如果用户指定了 -interactive 参数，给出额外提示
			if *interactive {
				fmt.Println("\n提示: -interactive 参数仅对「未使用且已过期」的证书有效")
				fmt.Println("当前所有证书均为正常证书（使用中或未过期），无需使用此参数")

				// 显示当前证书统计
				var usedCount, unusedCount, expiredCount int
				for _, c := range certs {
					if c.InUse {
						usedCount++
					} else {
						unusedCount++
					}
					if !c.ExpireTime.IsZero() && c.ExpireTime.Before(now) {
						expiredCount++
					}
				}
				fmt.Printf("\n当前证书状态统计:\n")
				fmt.Printf("  - 总计: %d 个\n", len(certs))
				fmt.Printf("  - 使用中: %d 个\n", usedCount)
				fmt.Printf("  - 未使用: %d 个\n", unusedCount)
				fmt.Printf("  - 已过期: %d 个\n", expiredCount)
				fmt.Printf("  - 未使用且已过期: 0 个\n")
			}
		}
		logInfo("检查完成，没有需要清理的证书")
		return
	}

	// 显示待删除的证书
	if !*quiet {
		fmt.Printf("发现 %d 个未使用且已过期的证书:\n", len(expiredUnusedCerts))

		// 交互模式提示
		if *interactive {
			fmt.Println("\n[交互模式] 您将逐个确认是否删除以下证书")
		}

		fmt.Println("================================================================================")

		for i, c := range expiredUnusedCerts {
			expiredDays := int(now.Sub(c.ExpireTime).Hours() / 24)
			fmt.Printf("[%d] 证书ID: %s\n", i+1, c.CertID)
			fmt.Printf("    证书名称: %s\n", c.CertName)
			fmt.Printf("    通用名称: %s\n", c.CertCommonName)
			fmt.Printf("    过期时间: %s (已过期 %d 天)\n", c.ExpireTime.Format("2006-01-02 15:04:05"), expiredDays)
			fmt.Println("--------------------------------------------------------------------------------")
		}
	}

	// 模拟运行
	if *dryRun {
		fmt.Println("\n[模拟运行] 以上证书将在非模拟模式下被删除")
		fmt.Println("如需执行删除，请去掉 -dry-run 参数运行")
		return
	}

	// 非自动且非交互模式需要一次性确认
	if !*auto && !*interactive {
		fmt.Printf("\n确认要删除以上 %d 个证书吗？(输入 yes 确认): ", len(expiredUnusedCerts))
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "yes" {
			fmt.Println("已取消删除操作")
			return
		}
	}

	// 执行删除
	if !*quiet && !*interactive {
		fmt.Println("\n开始删除过期未使用的证书...")
	}

	logInfo("开始清理 %d 个未使用且已过期的证书", len(expiredUnusedCerts))

	successCount := 0
	failCount := 0
	skipCount := 0

	for i, c := range expiredUnusedCerts {
		// 交互模式：逐个确认
		if *interactive {
			fmt.Printf("\n[%d/%d] 证书ID: %s\n", i+1, len(expiredUnusedCerts), c.CertID)
			fmt.Printf("      证书名称: %s\n", c.CertName)
			fmt.Printf("      通用名称: %s\n", c.CertCommonName)
			expiredDays := int(time.Now().Sub(c.ExpireTime).Hours() / 24)
			fmt.Printf("      过期时间: %s (已过期 %d 天)\n", c.ExpireTime.Format("2006-01-02 15:04:05"), expiredDays)
			fmt.Printf("\n删除此证书？(y=是, n=否, a=全部, q=退出): ")

			var answer string
			fmt.Scanln(&answer)

			switch answer {
			case "q", "quit", "Q":
				fmt.Println("\n已退出删除操作")
				logInfo("用户中止删除操作，已删除 %d 个，跳过 %d 个", successCount, skipCount)
				fmt.Printf("\n================================================================================\n")
				fmt.Printf("删除完成: 成功 %d 个, 跳过 %d 个, 失败 %d 个\n", successCount, skipCount, failCount)
				if failCount > 0 {
					os.Exit(1)
				}
				return
			case "n", "no", "N":
				fmt.Printf("⊘ 已跳过: %s (%s)\n", c.CertName, c.CertID)
				logInfo("跳过证书: %s (%s)", c.CertName, c.CertID)
				skipCount++
				continue
			case "a", "all", "A":
				// 删除当前证书
				err := certClient.DeleteCert(c.CertID)
				if err != nil {
					fmt.Printf("✗ 删除证书失败: %s (%s) - %v\n", c.CertName, c.CertID, err)
					logError("删除证书失败: %s (%s) - %v", c.CertName, c.CertID, err)
					failCount++
				} else {
					fmt.Printf("✓ 删除证书成功: %s (%s)\n", c.CertName, c.CertID)
					logInfo("删除证书成功: %s (%s)", c.CertName, c.CertID)
					successCount++
				}

				// 删除后续所有证书
				fmt.Println("\n开始删除所有剩余证书...")
				for j := i + 1; j < len(expiredUnusedCerts); j++ {
					cert := expiredUnusedCerts[j]
					err := certClient.DeleteCert(cert.CertID)
					if err != nil {
						fmt.Printf("✗ 删除证书失败: %s (%s) - %v\n", cert.CertName, cert.CertID, err)
						logError("删除证书失败: %s (%s) - %v", cert.CertName, cert.CertID, err)
						failCount++
					} else {
						fmt.Printf("✓ 删除证书成功: %s (%s)\n", cert.CertName, cert.CertID)
						logInfo("删除证书成功: %s (%s)", cert.CertName, cert.CertID)
						successCount++
					}
				}

				fmt.Printf("\n================================================================================\n")
				fmt.Printf("删除完成: 成功 %d 个, 跳过 %d 个, 失败 %d 个\n", successCount, skipCount, failCount)
				logInfo("清理完成: 成功 %d 个, 跳过 %d 个, 失败 %d 个", successCount, skipCount, failCount)
				if failCount > 0 {
					os.Exit(1)
				}
				return
			case "y", "yes", "Y":
				// 继续删除当前证书
			default:
				fmt.Printf("⊘ 输入无效，已跳过: %s (%s)\n", c.CertName, c.CertID)
				skipCount++
				continue
			}
		}

		// 执行删除
		err := certClient.DeleteCert(c.CertID)
		if err != nil {
			if !*quiet {
				fmt.Printf("✗ 删除证书失败: %s (%s) - %v\n", c.CertName, c.CertID, err)
			}
			logError("删除证书失败: %s (%s) - %v", c.CertName, c.CertID, err)
			failCount++
		} else {
			if !*quiet {
				fmt.Printf("✓ 删除证书成功: %s (%s)\n", c.CertName, c.CertID)
			}
			logInfo("删除证书成功: %s (%s)", c.CertName, c.CertID)
			successCount++
		}
	}

	if !*quiet {
		fmt.Println("\n================================================================================")
		if *interactive {
			fmt.Printf("删除完成: 成功 %d 个, 跳过 %d 个, 失败 %d 个\n", successCount, skipCount, failCount)
		} else {
			fmt.Printf("删除完成: 成功 %d 个, 失败 %d 个\n", successCount, failCount)
		}
	}

	if *interactive {
		logInfo("清理完成: 成功 %d 个, 跳过 %d 个, 失败 %d 个", successCount, skipCount, failCount)
	} else {
		logInfo("清理完成: 成功 %d 个, 失败 %d 个", successCount, failCount)
	}

	// 如果有失败，返回非零退出码
	if failCount > 0 {
		os.Exit(1)
	}
}

// logInfo 记录信息日志
func logInfo(format string, v ...interface{}) {
	log.Printf("[INFO] "+format, v...)
}

// logError 记录错误日志
func logError(format string, v ...interface{}) {
	log.Printf("[ERROR] "+format, v...)
}

// getAllCerts 获取所有证书
func getAllCerts(client *cert.Client) ([]CertInfo, error) {
	result, err := client.ListCerts()
	if err != nil {
		return nil, fmt.Errorf("获取证书列表失败: %w", err)
	}

	var allCerts []CertInfo
	for _, c := range result.Certs {
		expireTime, _ := parseTime(c.CertStopTime)

		allCerts = append(allCerts, CertInfo{
			CertID:         c.CertId,
			CertName:       c.CertName,
			CertCommonName: c.CertCommonName,
			CertStartTime:  c.CertStartTime,
			CertStopTime:   c.CertStopTime,
			ExpireTime:     expireTime,
		})
	}

	return allCerts, nil
}

// parseTime 尝试多种格式解析时间
func parseTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, fmt.Errorf("空时间字符串")
	}

	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05+08:00",
		"2006-01-02T15:04:05.000Z",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析时间: %s", timeStr)
}

// getUsedCertIDs 获取 CDN 中正在使用的证书 ID
func getUsedCertIDs(client *cdn.Client, quiet bool) (map[string]bool, error) {
	usedCerts := make(map[string]bool)

	domains, _, err := client.ListDomains("")
	if err != nil {
		return nil, fmt.Errorf("获取域名列表失败: %w", err)
	}

	if !quiet {
		fmt.Printf("正在检查 %d 个 CDN 域名的证书使用情况...\n", len(domains))
	}

	for _, domain := range domains {
		httpsConfig, err := client.GetDomainHttps(domain)
		if err != nil {
			continue
		}

		if httpsConfig != nil && httpsConfig.CertId != "" {
			usedCerts[httpsConfig.CertId] = true
		}
	}

	return usedCerts, nil
}

// listAllCerts 列出所有证书
func listAllCerts(certs []CertInfo) {
	now := time.Now()

	fmt.Println("所有证书列表:")
	fmt.Println("================================================================================")
	fmt.Printf("%-18s %-30s %-8s %s\n", "证书ID", "证书名称", "使用中", "过期时间")
	fmt.Println("--------------------------------------------------------------------------------")

	for _, c := range certs {
		inUseStr := "否"
		if c.InUse {
			inUseStr = "是"
		}

		expireStr := "未知"
		if !c.ExpireTime.IsZero() {
			expireStr = c.ExpireTime.Format("2006-01-02")
			if c.ExpireTime.Before(now) {
				days := int(now.Sub(c.ExpireTime).Hours() / 24)
				expireStr = fmt.Sprintf("%s (已过期%d天)", expireStr, days)
			} else {
				days := int(c.ExpireTime.Sub(now).Hours() / 24)
				expireStr = fmt.Sprintf("%s (剩余%d天)", expireStr, days)
			}
		}

		certName := c.CertName
		if len(certName) > 28 {
			certName = certName[:25] + "..."
		}

		fmt.Printf("%-18s %-30s %-8s %s\n",
			c.CertID, certName, inUseStr, expireStr)
	}
	fmt.Println("================================================================================")

	var usedCount, unusedCount, expiredCount, expiredUnusedCount int
	for _, c := range certs {
		if c.InUse {
			usedCount++
		} else {
			unusedCount++
		}
		if !c.ExpireTime.IsZero() && c.ExpireTime.Before(now) {
			expiredCount++
			if !c.InUse {
				expiredUnusedCount++
			}
		}
	}

	fmt.Printf("\n统计: 总计 %d 个证书\n", len(certs))
	fmt.Printf("  - 使用中: %d 个\n", usedCount)
	fmt.Printf("  - 未使用: %d 个\n", unusedCount)
	fmt.Printf("  - 已过期: %d 个\n", expiredCount)
	fmt.Printf("  - 未使用且已过期: %d 个 (可清理)\n", expiredUnusedCount)
}
