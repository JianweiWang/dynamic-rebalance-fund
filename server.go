package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// API 请求和响应结构体
type AddFundRequest struct {
	BucketIndex int     `json:"bucket_index"`
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Current     float64 `json:"current"`
	Weight      float64 `json:"weight"`
}

type UpdateFundRequest struct {
	BucketIndex int    `json:"bucket_index"`
	FundIndex   int    `json:"fund_index"`
	Field       string `json:"field"`
	Value       string `json:"value"`
}

type DeleteFundRequest struct {
	BucketIndex int `json:"bucket_index"`
	FundIndex   int `json:"fund_index"`
}

type RebalanceRequest struct {
	Threshold float64 `json:"threshold"`
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

	// 获取所有桶以验证索引
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
	dbBuckets, _ = getAllBucketsFromDB()
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

	// 获取所有桶以验证索引
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

	fund := bucket.Funds[req.FundIndex]

	// 从数据库删除
	err = deleteFundFromDB(fund.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "删除基金失败: " + err.Error(),
		})
		return
	}

	// 返回更新后的数据
	dbBuckets, _ = getAllBucketsFromDB()
	buckets := convertDBBucketsToAPIBuckets(dbBuckets)

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "已删除基金: " + fund.Name,
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

	// 获取所有桶以验证索引
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

	fund := bucket.Funds[req.FundIndex]

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
			for i, f := range bucket.Funds {
				if i != req.FundIndex {
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
	dbBuckets, _ = getAllBucketsFromDB()
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

	// 从数据库获取当前数据
	dbBuckets, err := getAllBucketsFromDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "获取基金配置失败: " + err.Error(),
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
		log.Printf("✅ 再平衡记录已保存，ID: %d", recordID)
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "再平衡分析完成",
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

	// 主页
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// API 路由
	api := r.Group("/api")
	{
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
