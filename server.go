package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// API 请求和响应结构体
type CreatePortfolioRequest struct {
	Name                  string  `json:"name"`
	Description           string  `json:"description"`
	TotalInvestmentAmount float64 `json:"total_investment_amount"`
	ShortTermRatio        float64 `json:"short_term_ratio"`
	MediumTermRatio       float64 `json:"medium_term_ratio"`
	LongTermRatio         float64 `json:"long_term_ratio"`
}

type UpdatePortfolioRequest struct {
	ID                    int     `json:"id"`
	Name                  string  `json:"name"`
	Description           string  `json:"description"`
	TotalInvestmentAmount float64 `json:"total_investment_amount"`
}

type AddFundRequest struct {
	BucketIndex int     `json:"bucket_index"` // 旧系统兼容
	BucketID    int     `json:"bucket_id"`    // 新投资组合系统
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Current     float64 `json:"current"`
	Weight      float64 `json:"weight"`
}

type UpdateFundRequest struct {
	BucketIndex int    `json:"bucket_index"` // 旧系统兼容
	FundIndex   int    `json:"fund_index"`   // 旧系统兼容
	FundID      int    `json:"fund_id"`      // 新投资组合系统
	Field       string `json:"field"`
	Value       string `json:"value"`
}

type DeleteFundRequest struct {
	BucketIndex int `json:"bucket_index"` // 旧系统兼容
	FundIndex   int `json:"fund_index"`   // 旧系统兼容
	FundID      int `json:"fund_id"`      // 新投资组合系统
}

type RebalanceRequest struct {
	Threshold   float64 `json:"threshold"`
	PortfolioID int     `json:"portfolio_id,omitempty"` // 可选：指定投资组合ID
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 初始化数据
func initData() {
	if err := initDatabase(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
}

// API 处理器

// Portfolio API handlers
func getPortfolios(c *gin.Context) {
	portfolios, err := getAllPortfolios()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "获取投资组合列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    portfolios,
	})
}

func getPortfolioDetail(c *gin.Context) {
	portfolioIDStr := c.Param("id")
	portfolioID, err := strconv.Atoi(portfolioIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的投资组合ID",
		})
		return
	}

	portfolio, err := getPortfolioByID(portfolioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "获取投资组合详情失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    portfolio,
	})
}

func createPortfolioHandler(c *gin.Context) {
	var req CreatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的请求参数",
		})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "投资组合名称不能为空",
		})
		return
	}

	// 验证投资金额
	if req.TotalInvestmentAmount <= 0 {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "初始投资金额必须大于0",
		})
		return
	}

	// 验证桶占比总和为100%
	totalRatio := req.ShortTermRatio + req.MediumTermRatio + req.LongTermRatio
	if totalRatio < 0.99 || totalRatio > 1.01 { // 允许小的浮点误差
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "三桶投资占比总和必须等于100%",
		})
		return
	}

	// 验证每个占比都大于等于0
	if req.ShortTermRatio < 0 || req.MediumTermRatio < 0 || req.LongTermRatio < 0 {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "桶投资占比不能为负数",
		})
		return
	}

	// 准备桶占比数组
	bucketRatios := []float64{req.ShortTermRatio, req.MediumTermRatio, req.LongTermRatio}

	portfolioID, err := createPortfolio(req.Name, req.Description, req.TotalInvestmentAmount, bucketRatios)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "创建投资组合失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "投资组合创建成功",
		Data:    map[string]int{"id": portfolioID},
	})
}

func updatePortfolioHandler(c *gin.Context) {
	var req UpdatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的请求参数",
		})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "投资组合名称不能为空",
		})
		return
	}

	if req.TotalInvestmentAmount < 0 {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "投资金额不能为负数",
		})
		return
	}

	err := updatePortfolio(req.ID, req.Name, req.Description, req.TotalInvestmentAmount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "更新投资组合失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "投资组合更新成功",
	})
}

func deletePortfolioHandler(c *gin.Context) {
	portfolioIDStr := c.Param("id")
	portfolioID, err := strconv.Atoi(portfolioIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的投资组合ID",
		})
		return
	}

	err = deletePortfolio(portfolioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "删除投资组合失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "投资组合删除成功",
	})
}

// 获取指定投资组合的桶配置
func getBucketsByPortfolio(c *gin.Context) {
	portfolioIDStr := c.Param("id")
	portfolioID, err := strconv.Atoi(portfolioIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的投资组合ID",
		})
		return
	}

	dbBuckets, err := getAllBucketsByPortfolioID(portfolioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "获取基金配置失败: " + err.Error(),
		})
		return
	}

	// 转换为API格式
	buckets := convertDBBucketsToAPIBuckets(dbBuckets)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    buckets,
	})
}

func getBuckets(c *gin.Context) {
	dbBuckets, err := getAllBucketsFromDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "获取基金配置失败: " + err.Error(),
		})
		return
	}

	// 转换为API格式
	buckets := convertDBBucketsToAPIBuckets(dbBuckets)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    buckets,
	})
}

func addFund(c *gin.Context) {
	var req AddFundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的请求参数",
		})
		return
	}

	var bucket DBBucket
	var err error

	// 判断使用新的bucket_id还是旧的bucket_index
	if req.BucketID > 0 {
		// 新系统：直接通过bucket_id获取桶信息
		bucketFound := false

		// 获取所有投资组合来查找正确的桶
		portfolios, err := getAllPortfolios()
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "获取投资组合信息失败: " + err.Error(),
			})
			return
		}

		// 查找指定的桶
		for _, portfolio := range portfolios {
			buckets, err := getAllBucketsByPortfolioID(portfolio.ID)
			if err != nil {
				continue
			}
			for _, b := range buckets {
				if b.ID == req.BucketID {
					bucket = b
					bucketFound = true
					break
				}
			}
			if bucketFound {
				break
			}
		}

		if !bucketFound {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: "无效的桶ID",
			})
			return
		}
	} else {
		// 旧系统：通过bucket_index获取桶信息（向后兼容）
		dbBuckets, err := getAllBucketsFromDB()
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "获取桶信息失败: " + err.Error(),
			})
			return
		}

		if req.BucketIndex < 0 || req.BucketIndex >= len(dbBuckets) {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: "无效的桶索引",
			})
			return
		}

		bucket = dbBuckets[req.BucketIndex]
	}

	// 验证权重
	var totalWeight float64
	for _, f := range bucket.Funds {
		totalWeight += f.Weight
	}

	if totalWeight+req.Weight > 1.0 {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "权重超出限制！当前桶内总权重: " + strconv.FormatFloat(totalWeight, 'f', 2, 64) +
				"，剩余可分配: " + strconv.FormatFloat(1.0-totalWeight, 'f', 2, 64),
		})
		return
	}

	// 添加到数据库
	err = addFundToDB(bucket.ID, req.Name, req.Code, req.Current, req.Weight)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "添加基金失败: " + err.Error(),
		})
		return
	}

	// 返回更新后的数据
	var dbBuckets []DBBucket
	if req.BucketID > 0 {
		// 新系统：返回相应投资组合的数据
		dbBuckets, _ = getAllBucketsByPortfolioID(bucket.PortfolioID)
	} else {
		// 旧系统：返回默认数据
		dbBuckets, _ = getAllBucketsFromDB()
	}
	buckets := convertDBBucketsToAPIBuckets(dbBuckets)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "基金添加成功",
		Data:    buckets,
	})
}

func deleteFund(c *gin.Context) {
	var req DeleteFundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的请求参数",
		})
		return
	}

	var fund DBFund
	var portfolioID int
	var fundName string

	// 判断使用新的fund_id还是旧的bucket_index/fund_index
	if req.FundID > 0 {
		// 新系统：直接通过fund_id获取基金信息
		fundPtr, err := getFundByID(req.FundID)
		if err != nil {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: "无效的基金ID",
			})
			return
		}
		fund = *fundPtr
		fundName = fund.Name

		// 获取对应的投资组合ID
		portfolios, err := getAllPortfolios()
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "获取投资组合信息失败: " + err.Error(),
			})
			return
		}

		// 查找对应的投资组合
		portfolioFound := false
		for _, portfolio := range portfolios {
			buckets, err := getAllBucketsByPortfolioID(portfolio.ID)
			if err != nil {
				continue
			}
			for _, bucket := range buckets {
				for _, f := range bucket.Funds {
					if f.ID == req.FundID {
						portfolioID = portfolio.ID
						portfolioFound = true
						break
					}
				}
				if portfolioFound {
					break
				}
			}
			if portfolioFound {
				break
			}
		}

		if !portfolioFound {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "找不到对应的投资组合",
			})
			return
		}
	} else {
		// 旧系统：通过bucket_index/fund_index获取基金信息（向后兼容）
		dbBuckets, err := getAllBucketsFromDB()
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "获取桶信息失败: " + err.Error(),
			})
			return
		}

		if req.BucketIndex < 0 || req.BucketIndex >= len(dbBuckets) {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: "无效的桶索引",
			})
			return
		}

		bucket := dbBuckets[req.BucketIndex]

		if req.FundIndex < 0 || req.FundIndex >= len(bucket.Funds) {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: "无效的基金索引",
			})
			return
		}

		fund = bucket.Funds[req.FundIndex]
		fundName = fund.Name
		portfolioID = bucket.PortfolioID
	}

	// 从数据库删除
	err := deleteFundFromDB(fund.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "删除基金失败: " + err.Error(),
		})
		return
	}

	// 返回更新后的数据
	var dbBuckets []DBBucket
	if req.FundID > 0 {
		// 新系统：返回相应投资组合的数据
		dbBuckets, _ = getAllBucketsByPortfolioID(portfolioID)
	} else {
		// 旧系统：返回默认数据
		dbBuckets, _ = getAllBucketsFromDB()
	}
	buckets := convertDBBucketsToAPIBuckets(dbBuckets)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "已删除基金: " + fundName,
		Data:    buckets,
	})
}

func updateFund(c *gin.Context) {
	var req UpdateFundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的请求参数",
		})
		return
	}

	var fund DBFund
	var bucket DBBucket
	var err error
	var portfolioID int

	// 判断使用新的fund_id还是旧的bucket_index/fund_index
	if req.FundID > 0 {
		// 新系统：直接通过fund_id获取基金信息
		fundPtr, err := getFundByID(req.FundID)
		if err != nil {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: "无效的基金ID",
			})
			return
		}
		fund = *fundPtr

		// 获取对应的桶信息
		portfolios, err := getAllPortfolios()
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "获取投资组合信息失败: " + err.Error(),
			})
			return
		}

		// 查找对应的桶和投资组合
		bucketFound := false
		for _, portfolio := range portfolios {
			buckets, err := getAllBucketsByPortfolioID(portfolio.ID)
			if err != nil {
				continue
			}
			for _, b := range buckets {
				if b.ID == fund.BucketID {
					bucket = b
					portfolioID = portfolio.ID
					bucketFound = true
					break
				}
			}
			if bucketFound {
				break
			}
		}

		if !bucketFound {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "找不到对应的桶信息",
			})
			return
		}
	} else {
		// 旧系统：通过bucket_index/fund_index获取基金信息（向后兼容）
		dbBuckets, err := getAllBucketsFromDB()
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "获取桶信息失败: " + err.Error(),
			})
			return
		}

		if req.BucketIndex < 0 || req.BucketIndex >= len(dbBuckets) {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: "无效的桶索引",
			})
			return
		}

		bucket = dbBuckets[req.BucketIndex]

		if req.FundIndex < 0 || req.FundIndex >= len(bucket.Funds) {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: "无效的基金索引",
			})
			return
		}

		fund = bucket.Funds[req.FundIndex]
		portfolioID = bucket.PortfolioID
	}

	// 验证字段
	switch req.Field {
	case "name", "code":
		// 字符串字段直接更新
	case "current":
		if _, err := strconv.ParseFloat(req.Value, 64); err != nil {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: "无效的数值",
			})
			return
		}
	case "weight":
		if val, err := strconv.ParseFloat(req.Value, 64); err == nil {
			// 验证权重
			var totalWeight float64
			for _, f := range bucket.Funds {
				if f.ID != fund.ID {
					totalWeight += f.Weight
				}
			}

			if totalWeight+val > 1.0 {
				c.JSON(http.StatusBadRequest, Response{
					Success: false,
					Message: "权重超出限制！其他基金总权重: " + strconv.FormatFloat(totalWeight, 'f', 2, 64) +
						"，剩余可分配: " + strconv.FormatFloat(1.0-totalWeight, 'f', 2, 64),
				})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, Response{
				Success: false,
				Message: "无效的数值",
			})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的字段",
		})
		return
	}

	// 更新数据库
	err = updateFundInDB(fund.ID, req.Field, req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "更新基金失败: " + err.Error(),
		})
		return
	}

	// 返回更新后的数据
	var dbBuckets []DBBucket
	if req.FundID > 0 {
		// 新系统：返回相应投资组合的数据
		dbBuckets, _ = getAllBucketsByPortfolioID(portfolioID)
	} else {
		// 旧系统：返回默认数据
		dbBuckets, _ = getAllBucketsFromDB()
	}
	buckets := convertDBBucketsToAPIBuckets(dbBuckets)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "基金信息更新成功",
		Data:    buckets,
	})
}

func performRebalance(c *gin.Context) {
	var req RebalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的请求参数",
		})
		return
	}

	if req.Threshold <= 0 {
		req.Threshold = 0.05
	}

	// 根据是否指定投资组合ID获取数据
	var dbBuckets []DBBucket
	var err error

	if req.PortfolioID > 0 {
		// 新系统：获取指定投资组合的数据
		dbBuckets, err = getAllBucketsByPortfolioID(req.PortfolioID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "获取投资组合基金配置失败: " + err.Error(),
			})
			return
		}

		// 验证投资组合是否存在
		if len(dbBuckets) == 0 {
			c.JSON(http.StatusNotFound, Response{
				Success: false,
				Message: "没有找到指定的投资组合或组合中没有基金",
			})
			return
		}
	} else {
		// 旧系统：获取默认数据（向后兼容）
		dbBuckets, err = getAllBucketsFromDB()
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Message: "获取基金配置失败: " + err.Error(),
			})
			return
		}
	}

	// 转换为API格式进行再平衡计算
	buckets := convertDBBucketsToAPIBuckets(dbBuckets)
	results := rebalance(buckets, req.Threshold)

	// 更新数据库中的再平衡结果
	err = updateFundRebalanceResults(dbBuckets, results)
	if err != nil {
		log.Printf("更新再平衡结果失败: %v", err)
	}

	// 保存再平衡记录
	var totalValue float64
	for _, bucket := range results {
		for _, fund := range bucket.Funds {
			totalValue += fund.Current
		}
	}

	// 构建建议记录
	var suggestions []RebalanceSuggestion
	for _, dbBucket := range dbBuckets {
		for _, dbFund := range dbBucket.Funds {
			// 找到对应的再平衡结果
			for _, bucket := range results {
				for _, fund := range bucket.Funds {
					if fund.Name == dbFund.Name && fund.Code == dbFund.Code {
						suggestions = append(suggestions, RebalanceSuggestion{
							FundID:       dbFund.ID,
							FundName:     fund.Name,
							FundCode:     fund.Code,
							CurrentValue: fund.Current,
							TargetValue:  fund.Target,
							DiffValue:    fund.Diff,
							Advice:       fund.Advice,
							Reason:       fund.Reason,
						})
						break
					}
				}
			}
		}
	}

	// 保存到历史记录
	recordID, err := saveRebalanceRecord(req.Threshold, totalValue, suggestions)
	if err != nil {
		log.Printf("保存再平衡记录失败: %v", err)
	} else {
		log.Printf("✅ 再平衡记录已保存，ID: %d", recordID)
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "再平衡分析完成",
		Data:    results,
	})
}

// 投资组合特定的再平衡函数
func performPortfolioRebalance(c *gin.Context) {
	// 从 URL 获取投资组合 ID
	portfolioIDStr := c.Param("id")
	portfolioID, err := strconv.Atoi(portfolioIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的投资组合ID",
		})
		return
	}

	var req RebalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的请求参数",
		})
		return
	}

	if req.Threshold <= 0 {
		req.Threshold = 0.05
	}

	// 强制设置投资组合ID
	req.PortfolioID = portfolioID

	// 获取指定投资组合的数据
	dbBuckets, err := getAllBucketsByPortfolioID(portfolioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "获取投资组合基金配置失败: " + err.Error(),
		})
		return
	}

	// 验证投资组合是否存在及是否有基金
	if len(dbBuckets) == 0 {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Message: "没有找到指定的投资组合或组合中没有基金",
		})
		return
	}

	// 检查是否有基金数据
	var hasFunds bool
	for _, bucket := range dbBuckets {
		if len(bucket.Funds) > 0 {
			hasFunds = true
			break
		}
	}

	if !hasFunds {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "该投资组合中没有任何基金，无法执行再平衡",
		})
		return
	}

	// 转换为API格式进行再平衡计算
	buckets := convertDBBucketsToAPIBuckets(dbBuckets)
	results := rebalance(buckets, req.Threshold)

	// 更新数据库中的再平衡结果
	err = updateFundRebalanceResults(dbBuckets, results)
	if err != nil {
		log.Printf("更新再平衡结果失败: %v", err)
	}

	// 保存再平衡记录
	var totalValue float64
	for _, bucket := range results {
		for _, fund := range bucket.Funds {
			totalValue += fund.Current
		}
	}

	// 构建建议记录
	var suggestions []RebalanceSuggestion
	for _, dbBucket := range dbBuckets {
		for _, dbFund := range dbBucket.Funds {
			// 找到对应的再平衡结果
			for _, bucket := range results {
				for _, fund := range bucket.Funds {
					if fund.Name == dbFund.Name && fund.Code == dbFund.Code {
						suggestions = append(suggestions, RebalanceSuggestion{
							FundID:       dbFund.ID,
							FundName:     fund.Name,
							FundCode:     fund.Code,
							CurrentValue: fund.Current,
							TargetValue:  fund.Target,
							DiffValue:    fund.Diff,
							Advice:       fund.Advice,
							Reason:       fund.Reason,
						})
						break
					}
				}
			}
		}
	}

	// 保存到历史记录
	recordID, err := saveRebalanceRecord(req.Threshold, totalValue, suggestions)
	if err != nil {
		log.Printf("保存再平衡记录失败: %v", err)
	} else {
		log.Printf("✅ 投资组合%d再平衡记录已保存，ID: %d", portfolioID, recordID)
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "投资组合再平衡分析完成",
		Data:    results,
	})
}

// 获取再平衡历史记录
func getRebalanceHistoryHandler(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	records, err := getRebalanceHistory(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "获取历史记录失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    records,
	})
}

// 获取投资组合收益表现
func getPortfolioPerformance(c *gin.Context) {
	portfolioIDStr := c.Param("id")
	portfolioID, err := strconv.Atoi(portfolioIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的投资组合ID",
		})
		return
	}

	// 计算投资组合收益表现
	performance, err := calculatePortfolioPerformance(portfolioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "计算投资组合收益失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "获取投资组合收益表现成功",
		Data:    performance,
	})
}

// 获取再平衡历史详情
func getRebalanceDetailHandler(c *gin.Context) {
	recordIDStr := c.Param("id")
	recordID, err := strconv.Atoi(recordIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "无效的记录ID",
		})
		return
	}

	// 获取记录基本信息
	record, err := getRebalanceRecordByID(recordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "获取记录失败: " + err.Error(),
		})
		return
	}

	// 获取详细建议
	suggestions, err := getRebalanceSuggestionsByRecordID(recordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "获取建议详情失败: " + err.Error(),
		})
		return
	}

	// 组合返回数据
	detail := struct {
		Record      RebalanceRecord       `json:"record"`
		Suggestions []RebalanceSuggestion `json:"suggestions"`
	}{
		Record:      *record,
		Suggestions: suggestions,
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    detail,
	})
}

func setupRoutes() *gin.Engine {
	r := gin.Default()

	// 启用CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"*"}
	r.Use(cors.New(config))

	// 静态文件服务
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// 主页和投资组合页面
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "portfolios.html", nil)
	})

	r.GET("/portfolios/:id", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// API 路由
	api := r.Group("/api")
	{
		// Portfolio endpoints
		api.GET("/portfolios", getPortfolios)
		api.POST("/portfolios", createPortfolioHandler)
		api.PUT("/portfolios", updatePortfolioHandler)
		api.DELETE("/portfolios/:id", deletePortfolioHandler)
		api.GET("/portfolios/:id", getPortfolioDetail)
		api.GET("/portfolios/:id/buckets", getBucketsByPortfolio)
		api.GET("/portfolios/:id/performance", getPortfolioPerformance)
		api.POST("/portfolios/:id/rebalance", performPortfolioRebalance)

		// Existing endpoints (maintained for backward compatibility)
		api.GET("/buckets", getBuckets)
		api.POST("/funds", addFund)
		api.DELETE("/funds", deleteFund)
		api.PUT("/funds", updateFund)
		api.POST("/rebalance", performRebalance)
		api.GET("/rebalance/history", getRebalanceHistoryHandler)
		api.GET("/rebalance/history/:id", getRebalanceDetailHandler)
	}

	return r
}
