package main

import (
	"fmt"
	"math"
	"os"
	"strings"
)

type Fund struct {
	Name    string  `json:"name"`
	Code    string  `json:"code"`
	Current float64 `json:"current"`
	Weight  float64 `json:"weight"` // 在桶内的权重
	Target  float64 `json:"target"`
	Diff    float64 `json:"diff"`
	Advice  string  `json:"advice"`
	Reason  string  `json:"reason"` // 操作原因
}

type Bucket struct {
	Name       string  `json:"name"`
	TargetRate float64 `json:"target_rate"`
	Funds      []Fund  `json:"funds"`
}

func rebalance(buckets []Bucket, threshold float64) []Bucket {
	// 计算总市值
	var total float64
	for _, b := range buckets {
		for _, f := range b.Funds {
			total += f.Current
		}
	}

	// 计算目标值 & 建议
	for bi := range buckets {
		bucket := &buckets[bi]
		bucketTarget := total * bucket.TargetRate
		var bucketCurrent float64
		for _, f := range bucket.Funds {
			bucketCurrent += f.Current
		}

		// 计算桶的偏差
		bucketDeviation := (bucketCurrent / total) - bucket.TargetRate
		bucketDeviationPercent := bucketDeviation * 100

		for fi := range bucket.Funds {
			fund := &bucket.Funds[fi]
			fund.Target = bucketTarget * fund.Weight
			fund.Diff = fund.Target - fund.Current

			// 计算基金的偏差
			fundCurrentPercent := (fund.Current / total) * 100
			fundTargetPercent := (fund.Target / total) * 100
			fundDeviationPercent := fundCurrentPercent - fundTargetPercent

			if math.Abs(bucketDeviation) > threshold {
				if fund.Diff > 0 {
					fund.Advice = "买入"
					fund.Reason = fmt.Sprintf("当前市值%.2f万(占比%.1f%%)低于目标%.2f万(占比%.1f%%)，%s整体偏低%.1f%%，需要买入%.2f万",
						fund.Current, fundCurrentPercent, fund.Target, fundTargetPercent,
						bucket.Name, math.Abs(bucketDeviationPercent), math.Abs(fund.Diff))
				} else if fund.Diff < 0 {
					fund.Advice = "卖出"
					fund.Reason = fmt.Sprintf("当前市值%.2f万(占比%.1f%%)高于目标%.2f万(占比%.1f%%)，%s整体偏高%.1f%%，需要卖出%.2f万",
						fund.Current, fundCurrentPercent, fund.Target, fundTargetPercent,
						bucket.Name, math.Abs(bucketDeviationPercent), math.Abs(fund.Diff))
				} else {
					fund.Advice = "买入"
					fund.Reason = fmt.Sprintf("当前市值%.2f万符合目标配置，但%s整体偏低%.1f%%，需要适量买入",
						fund.Current, bucket.Name, math.Abs(bucketDeviationPercent))
				}
			} else {
				fund.Advice = "保持不动"
				fund.Diff = 0
				if math.Abs(fundDeviationPercent) < 0.5 {
					fund.Reason = fmt.Sprintf("当前市值%.2f万(占比%.1f%%)与目标%.2f万(占比%.1f%%)基本一致，无需调整",
						fund.Current, fundCurrentPercent, fund.Target, fundTargetPercent)
				} else {
					fund.Reason = fmt.Sprintf("虽然偏离目标%.1f%%，但%s整体偏差%.1f%%在阈值范围内，暂不调整",
						math.Abs(fundDeviationPercent), bucket.Name, math.Abs(bucketDeviationPercent))
				}
			}
		}
	}
	return buckets
}

// 显示菜单
func showMenu() {
	fmt.Println("\n🏦 动态基金再平衡系统")
	fmt.Println("============================")
	fmt.Println("1. 查看当前基金配置")
	fmt.Println("2. 执行再平衡分析")
	fmt.Println("3. 添加基金")
	fmt.Println("4. 删除基金")
	fmt.Println("5. 修改基金信息")
	fmt.Println("6. 退出")
	fmt.Print("请选择操作 (1-6): ")
}

// 查看当前基金配置
func listFunds(buckets []Bucket) {
	fmt.Println("\n📊 当前基金配置")
	fmt.Println("=======================================================")
	for _, bucket := range buckets {
		fmt.Printf("\n🗂️  %s (目标占比: %.1f%%)\n", bucket.Name, bucket.TargetRate*100)
		fmt.Println("-------------------------------------------------------")
		for i, fund := range bucket.Funds {
			fmt.Printf("%d. %s (%s) | 当前: %.2f万 | 权重: %.1f%%\n",
				i+1, fund.Name, fund.Code, fund.Current, fund.Weight*100)
		}
	}
}

// 查找桶的索引
func findBucketIndex(buckets []Bucket) int {
	fmt.Println("\n选择桶:")
	for i, bucket := range buckets {
		fmt.Printf("%d. %s\n", i+1, bucket.Name)
	}

	var choice int
	fmt.Print("请输入桶编号: ")
	fmt.Scan(&choice)

	if choice < 1 || choice > len(buckets) {
		fmt.Println("❌ 无效的桶编号")
		return -1
	}
	return choice - 1
}

// CLI版本的添加基金
func addFundCLI(buckets []Bucket) []Bucket {
	bucketIndex := findBucketIndex(buckets)
	if bucketIndex == -1 {
		return buckets
	}

	bucket := &buckets[bucketIndex]

	var name, code string
	var current, weight float64

	fmt.Print("基金名称: ")
	fmt.Scan(&name)
	name = strings.ReplaceAll(name, "_", " ") // 处理空格

	fmt.Print("基金代码: ")
	fmt.Scan(&code)

	fmt.Print("当前市值(万元): ")
	fmt.Scan(&current)

	fmt.Print("在桶内的权重(0-1): ")
	fmt.Scan(&weight)

	// 验证权重
	var totalWeight float64
	for _, f := range bucket.Funds {
		totalWeight += f.Weight
	}

	if totalWeight+weight > 1.0 {
		fmt.Printf("❌ 权重超出限制！当前桶内总权重: %.2f，剩余可分配: %.2f\n",
			totalWeight, 1.0-totalWeight)
		return buckets
	}

	newFund := Fund{
		Name:    name,
		Code:    code,
		Current: current,
		Weight:  weight,
		Target:  0,
		Diff:    0,
		Advice:  "",
	}

	bucket.Funds = append(bucket.Funds, newFund)
	fmt.Printf("✅ 已添加基金: %s\n", name)

	return buckets
}

// CLI版本的删除基金
func deleteFundCLI(buckets []Bucket) []Bucket {
	bucketIndex := findBucketIndex(buckets)
	if bucketIndex == -1 {
		return buckets
	}

	bucket := &buckets[bucketIndex]

	if len(bucket.Funds) == 0 {
		fmt.Println("❌ 该桶内没有基金")
		return buckets
	}

	fmt.Printf("\n%s 内的基金:\n", bucket.Name)
	for i, fund := range bucket.Funds {
		fmt.Printf("%d. %s (%s)\n", i+1, fund.Name, fund.Code)
	}

	var choice int
	fmt.Print("请选择要删除的基金编号: ")
	fmt.Scan(&choice)

	if choice < 1 || choice > len(bucket.Funds) {
		fmt.Println("❌ 无效的基金编号")
		return buckets
	}

	fundIndex := choice - 1
	fundName := bucket.Funds[fundIndex].Name

	// 删除基金
	bucket.Funds = append(bucket.Funds[:fundIndex], bucket.Funds[fundIndex+1:]...)
	fmt.Printf("✅ 已删除基金: %s\n", fundName)

	return buckets
}

// CLI版本的修改基金信息
func updateFundCLI(buckets []Bucket) []Bucket {
	bucketIndex := findBucketIndex(buckets)
	if bucketIndex == -1 {
		return buckets
	}

	bucket := &buckets[bucketIndex]

	if len(bucket.Funds) == 0 {
		fmt.Println("❌ 该桶内没有基金")
		return buckets
	}

	fmt.Printf("\n%s 内的基金:\n", bucket.Name)
	for i, fund := range bucket.Funds {
		fmt.Printf("%d. %s (%s) | 当前: %.2f万 | 权重: %.1f%%\n",
			i+1, fund.Name, fund.Code, fund.Current, fund.Weight*100)
	}

	var choice int
	fmt.Print("请选择要修改的基金编号: ")
	fmt.Scan(&choice)

	if choice < 1 || choice > len(bucket.Funds) {
		fmt.Println("❌ 无效的基金编号")
		return buckets
	}

	fund := &bucket.Funds[choice-1]

	fmt.Println("\n选择要修改的属性:")
	fmt.Println("1. 基金名称")
	fmt.Println("2. 基金代码")
	fmt.Println("3. 当前市值")
	fmt.Println("4. 权重")

	var attr int
	fmt.Print("请选择 (1-4): ")
	fmt.Scan(&attr)

	switch attr {
	case 1:
		var newName string
		fmt.Print("新的基金名称: ")
		fmt.Scan(&newName)
		fund.Name = strings.ReplaceAll(newName, "_", " ")
		fmt.Println("✅ 基金名称已更新")
	case 2:
		var newCode string
		fmt.Print("新的基金代码: ")
		fmt.Scan(&newCode)
		fund.Code = newCode
		fmt.Println("✅ 基金代码已更新")
	case 3:
		var newCurrent float64
		fmt.Print("新的当前市值(万元): ")
		fmt.Scan(&newCurrent)
		fund.Current = newCurrent
		fmt.Println("✅ 当前市值已更新")
	case 4:
		var newWeight float64
		fmt.Print("新的权重(0-1): ")
		fmt.Scan(&newWeight)

		// 验证权重
		var totalWeight float64
		for i, f := range bucket.Funds {
			if i != choice-1 { // 排除当前基金
				totalWeight += f.Weight
			}
		}

		if totalWeight+newWeight > 1.0 {
			fmt.Printf("❌ 权重超出限制！其他基金总权重: %.2f，剩余可分配: %.2f\n",
				totalWeight, 1.0-totalWeight)
			return buckets
		}

		fund.Weight = newWeight
		fmt.Println("✅ 权重已更新")
	default:
		fmt.Println("❌ 无效选择")
	}

	return buckets
}

// CLI版本的函数
func performRebalanceCLI(buckets []Bucket) {
	var threshold float64
	fmt.Print("请输入再平衡触发阈值 (例如 0.05 表示 ±5%)，按回车默认 0.05：")
	_, err := fmt.Scan(&threshold)
	if err != nil || threshold <= 0 {
		threshold = 0.05
	}

	// 执行再平衡
	results := rebalance(buckets, threshold)

	// 输出调仓清单
	fmt.Println("\n📋 调仓清单（单位：万元）")
	fmt.Println("-------------------------------------------------------------")
	for _, b := range results {
		for _, f := range b.Funds {
			fmt.Printf("%s (%s) | 当前市值: %.2f | 目标: %.2f | 建议: %s | 调整金额: %.2f\n",
				f.Name, f.Code, f.Current, f.Target, f.Advice, f.Diff)
		}
	}
}

func runCLI() {
	// 初始化默认组合
	clieBuckets := []Bucket{
		{"短期桶（货币基金）", 0.10, []Fund{
			{"易方达货币A", "000009", 20, 1.0, 0, 0, "", ""},
		}},
		{"中期桶（债券基金）", 0.30, []Fund{
			{"广发国开债7-10A", "003375", 50, 0.5, 0, 0, "", ""},
			{"博时信用债纯债A", "050026", 40, 0.5, 0, 0, "", ""},
		}},
		{"长期桶（股票基金）", 0.60, []Fund{
			{"易方达沪深300ETF联接A", "110020", 100, 0.4, 0, 0, "", ""},
			{"南方中证500ETF联接A", "160119", 80, 0.3, 0, 0, "", ""},
			{"汇添富海外互联网50ETF", "006327", 60, 0.3, 0, 0, "", ""},
		}},
	}

	for {
		showMenu()

		var choice int
		_, err := fmt.Scan(&choice)
		if err != nil {
			fmt.Println("❌ 请输入有效数字")
			continue
		}

		switch choice {
		case 1:
			listFunds(clieBuckets)
		case 2:
			performRebalanceCLI(clieBuckets)
		case 3:
			clieBuckets = addFundCLI(clieBuckets)
		case 4:
			clieBuckets = deleteFundCLI(clieBuckets)
		case 5:
			clieBuckets = updateFundCLI(clieBuckets)
		case 6:
			fmt.Println("👋 感谢使用，再见！")
			return
		default:
			fmt.Println("❌ 无效选择，请输入 1-6")
		}

		// 暂停一下，让用户看到结果
		fmt.Print("\n按回车键继续...")
		fmt.Scanln()
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "cli" {
		// 命令行模式
		// 初始化数据库（CLI也需要数据库支持）
		initData()
		defer closeDatabase()
		runCLI()
	} else {
		// Web服务器模式
		fmt.Println("🚀 启动Web服务器模式...")
		fmt.Println("📱 访问 http://localhost:8080 打开Web界面")
		fmt.Println("💻 或使用 'go run . cli' 启动命令行模式")
		fmt.Println("📊 数据将持久化存储到 SQLite 数据库")

		initData()
		defer closeDatabase()

		r := setupRoutes()
		r.Run(":8080")
	}
}
