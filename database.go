package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// 数据库模型
type DBPortfolio struct {
	ID                    int        `json:"id" db:"id"`
	Name                  string     `json:"name" db:"name"`
	Description           string     `json:"description" db:"description"`
	TotalInvestmentAmount float64    `json:"total_investment_amount" db:"total_investment_amount"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
	Buckets               []DBBucket `json:"buckets"`
}

type DBBucket struct {
	ID          int       `json:"id" db:"id"`
	PortfolioID int       `json:"portfolio_id" db:"portfolio_id"`
	Name        string    `json:"name" db:"name"`
	TargetRate  float64   `json:"target_rate" db:"target_rate"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Funds       []DBFund  `json:"funds"`
}

type DBFund struct {
	ID        int       `json:"id" db:"id"`
	BucketID  int       `json:"bucket_id" db:"bucket_id"`
	Name      string    `json:"name" db:"name"`
	Code      string    `json:"code" db:"code"`
	Current   float64   `json:"current" db:"current"`
	Weight    float64   `json:"weight" db:"weight"`
	Target    float64   `json:"target" db:"target"`
	Diff      float64   `json:"diff" db:"diff"`
	Advice    string    `json:"advice" db:"advice"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type RebalanceRecord struct {
	ID         int       `json:"id" db:"id"`
	Threshold  float64   `json:"threshold" db:"threshold"`
	TotalValue float64   `json:"total_value" db:"total_value"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type RebalanceSuggestion struct {
	ID           int       `json:"id" db:"id"`
	RecordID     int       `json:"record_id" db:"record_id"`
	FundID       int       `json:"fund_id" db:"fund_id"`
	FundName     string    `json:"fund_name" db:"fund_name"`
	FundCode     string    `json:"fund_code" db:"fund_code"`
	CurrentValue float64   `json:"current_value" db:"current_value"`
	TargetValue  float64   `json:"target_value" db:"target_value"`
	DiffValue    float64   `json:"diff_value" db:"diff_value"`
	Advice       string    `json:"advice" db:"advice"`
	Reason       string    `json:"reason" db:"reason"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// 初始化数据库
func initDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", "./fund_data.db")
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}

	// 创建表
	if err := createTables(); err != nil {
		return fmt.Errorf("创建表失败: %v", err)
	}

	// 初始化默认数据
	if err := initDefaultData(); err != nil {
		log.Printf("初始化默认数据警告: %v", err)
	}

	log.Println("✅ 数据库初始化完成")
	return nil
}

// 创建数据库表
func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS portfolios (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT DEFAULT '',
			total_investment_amount REAL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS buckets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			portfolio_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			target_rate REAL NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (portfolio_id) REFERENCES portfolios(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS funds (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bucket_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			code TEXT NOT NULL,
			current REAL NOT NULL DEFAULT 0,
			weight REAL NOT NULL DEFAULT 0,
			target REAL NOT NULL DEFAULT 0,
			diff REAL NOT NULL DEFAULT 0,
			advice TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (bucket_id) REFERENCES buckets(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS rebalance_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			threshold REAL NOT NULL,
			total_value REAL NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS rebalance_suggestions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			record_id INTEGER NOT NULL,
			fund_id INTEGER NOT NULL,
			fund_name TEXT NOT NULL,
			fund_code TEXT NOT NULL,
			current_value REAL NOT NULL,
			target_value REAL NOT NULL,
			diff_value REAL NOT NULL,
			advice TEXT NOT NULL,
			reason TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (record_id) REFERENCES rebalance_records(id) ON DELETE CASCADE,
			FOREIGN KEY (fund_id) REFERENCES funds(id) ON DELETE CASCADE
		)`,

		`CREATE INDEX IF NOT EXISTS idx_buckets_portfolio_id ON buckets(portfolio_id)`,
		`CREATE INDEX IF NOT EXISTS idx_funds_bucket_id ON funds(bucket_id)`,
		`CREATE INDEX IF NOT EXISTS idx_suggestions_record_id ON rebalance_suggestions(record_id)`,
		`CREATE INDEX IF NOT EXISTS idx_suggestions_fund_id ON rebalance_suggestions(fund_id)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("执行SQL失败 [%s]: %v", query, err)
		}
	}

	return nil
}

// 初始化默认数据
func initDefaultData() error {
	// 检查是否已有数据
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM portfolios").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("数据库已有数据，跳过初始化")
		return nil
	}

	// 插入默认投资组合
	result, err := db.Exec(
		"INSERT INTO portfolios (name, description, total_investment_amount) VALUES (?, ?, ?)",
		"默认投资组合", "基于三桶投资法的默认投资组合配置", 350.0,
	)
	if err != nil {
		return fmt.Errorf("插入默认投资组合失败: %v", err)
	}

	portfolioID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取投资组合ID失败: %v", err)
	}

	// 插入默认桶
	buckets := []struct {
		name       string
		targetRate float64
	}{
		{"短期桶（现金）", 0.10},
		{"中期桶（债券基金）", 0.30},
		{"长期桶（股票基金）", 0.60},
	}

	for _, bucket := range buckets {
		_, err := db.Exec(
			"INSERT INTO buckets (portfolio_id, name, target_rate) VALUES (?, ?, ?)",
			portfolioID, bucket.name, bucket.targetRate,
		)
		if err != nil {
			return fmt.Errorf("插入桶数据失败: %v", err)
		}
	}

	// 插入默认基金
	funds := []struct {
		bucketID int
		name     string
		code     string
		current  float64
		weight   float64
	}{
		{1, "现金管理", "CASH001", 20.0, 1.0},
		{2, "广发国开债7-10A", "003375", 50.0, 0.5},
		{2, "博时信用债纯债A", "050026", 40.0, 0.5},
		{3, "易方达沪深300ETF联接A", "110020", 100.0, 0.4},
		{3, "南方中证500ETF联接A", "160119", 80.0, 0.3},
		{3, "汇添富海外互联网50ETF", "006327", 60.0, 0.3},
	}

	for _, fund := range funds {
		_, err := db.Exec(`
			INSERT INTO funds (bucket_id, name, code, current, weight) 
			VALUES (?, ?, ?, ?, ?)`,
			fund.bucketID, fund.name, fund.code, fund.current, fund.weight,
		)
		if err != nil {
			return fmt.Errorf("插入基金数据失败: %v", err)
		}
	}

	log.Println("✅ 默认数据初始化完成")
	return nil
}

// 数据库操作函数

// Portfolio 管理函数
func getAllPortfolios() ([]DBPortfolio, error) {
	query := `
		SELECT id, name, description, total_investment_amount, created_at, updated_at 
		FROM portfolios 
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var portfolios []DBPortfolio
	for rows.Next() {
		var portfolio DBPortfolio
		err := rows.Scan(&portfolio.ID, &portfolio.Name, &portfolio.Description,
			&portfolio.TotalInvestmentAmount, &portfolio.CreatedAt, &portfolio.UpdatedAt)
		if err != nil {
			return nil, err
		}
		portfolios = append(portfolios, portfolio)
	}

	return portfolios, nil
}

func getPortfolioByID(portfolioID int) (*DBPortfolio, error) {
	query := `
		SELECT id, name, description, total_investment_amount, created_at, updated_at 
		FROM portfolios 
		WHERE id = ?
	`

	var portfolio DBPortfolio
	err := db.QueryRow(query, portfolioID).Scan(&portfolio.ID, &portfolio.Name,
		&portfolio.Description, &portfolio.TotalInvestmentAmount, &portfolio.CreatedAt, &portfolio.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// 获取投资组合内的桶
	buckets, err := getAllBucketsByPortfolioID(portfolioID)
	if err != nil {
		return nil, err
	}
	portfolio.Buckets = buckets

	return &portfolio, nil
}

func createPortfolio(name, description string, totalAmount float64, bucketRatios []float64) (int, error) {
	result, err := db.Exec(
		"INSERT INTO portfolios (name, description, total_investment_amount) VALUES (?, ?, ?)",
		name, description, totalAmount,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// 创建三个桶使用用户指定的占比
	bucketNames := []string{"短期桶（现金）", "中期桶（债券基金）", "长期桶（股票基金）"}

	// 如果没有提供占比，使用默认值
	if len(bucketRatios) != 3 {
		bucketRatios = []float64{0.10, 0.30, 0.60}
	}

	for i, bucketName := range bucketNames {
		_, err := db.Exec(
			"INSERT INTO buckets (portfolio_id, name, target_rate) VALUES (?, ?, ?)",
			id, bucketName, bucketRatios[i],
		)
		if err != nil {
			return 0, err
		}
	}

	return int(id), nil
}

func updatePortfolio(portfolioID int, name, description string, totalAmount float64) error {
	_, err := db.Exec(
		"UPDATE portfolios SET name = ?, description = ?, total_investment_amount = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		name, description, totalAmount, portfolioID,
	)
	return err
}

func deletePortfolio(portfolioID int) error {
	_, err := db.Exec("DELETE FROM portfolios WHERE id = ?", portfolioID)
	return err
}

func getAllBucketsByPortfolioID(portfolioID int) ([]DBBucket, error) {
	query := `
		SELECT id, portfolio_id, name, target_rate, created_at, updated_at 
		FROM buckets 
		WHERE portfolio_id = ?
		ORDER BY id
	`

	rows, err := db.Query(query, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buckets []DBBucket
	for rows.Next() {
		var bucket DBBucket
		err := rows.Scan(&bucket.ID, &bucket.PortfolioID, &bucket.Name, &bucket.TargetRate,
			&bucket.CreatedAt, &bucket.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// 获取桶内基金
		funds, err := getFundsByBucketID(bucket.ID)
		if err != nil {
			return nil, err
		}
		bucket.Funds = funds

		buckets = append(buckets, bucket)
	}

	return buckets, nil
}

// 保持对旧版本的兼容，默认获取第一个投资组合的数据
func getAllBucketsFromDB() ([]DBBucket, error) {
	// 获取第一个投资组合的ID
	var portfolioID int
	err := db.QueryRow("SELECT id FROM portfolios ORDER BY created_at LIMIT 1").Scan(&portfolioID)
	if err != nil {
		return nil, err
	}

	return getAllBucketsByPortfolioID(portfolioID)
}

func getFundsByBucketID(bucketID int) ([]DBFund, error) {
	query := `
		SELECT id, bucket_id, name, code, current, weight, target, diff, advice, created_at, updated_at
		FROM funds 
		WHERE bucket_id = ?
		ORDER BY id
	`

	rows, err := db.Query(query, bucketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var funds []DBFund
	for rows.Next() {
		var fund DBFund
		err := rows.Scan(&fund.ID, &fund.BucketID, &fund.Name, &fund.Code,
			&fund.Current, &fund.Weight, &fund.Target, &fund.Diff, &fund.Advice,
			&fund.CreatedAt, &fund.UpdatedAt)
		if err != nil {
			return nil, err
		}
		funds = append(funds, fund)
	}

	return funds, nil
}

func addFundToDB(bucketID int, name, code string, current, weight float64) error {
	query := `
		INSERT INTO funds (bucket_id, name, code, current, weight) 
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query, bucketID, name, code, current, weight)
	return err
}

// 添加新的函数：通过fund_id获取基金信息
func getFundByID(fundID int) (*DBFund, error) {
	query := `
		SELECT id, bucket_id, name, code, current, weight, target, diff, advice, created_at, updated_at
		FROM funds 
		WHERE id = ?
	`

	var fund DBFund
	err := db.QueryRow(query, fundID).Scan(&fund.ID, &fund.BucketID, &fund.Name, &fund.Code,
		&fund.Current, &fund.Weight, &fund.Target, &fund.Diff, &fund.Advice,
		&fund.CreatedAt, &fund.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &fund, nil
}

func updateFundInDB(fundID int, field, value string) error {
	query := fmt.Sprintf("UPDATE funds SET %s = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", field)
	_, err := db.Exec(query, value, fundID)
	return err
}

func deleteFundFromDB(fundID int) error {
	_, err := db.Exec("DELETE FROM funds WHERE id = ?", fundID)
	return err
}

func saveRebalanceRecord(threshold, totalValue float64, suggestions []RebalanceSuggestion) (int, error) {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 插入再平衡记录
	result, err := tx.Exec(
		"INSERT INTO rebalance_records (threshold, total_value) VALUES (?, ?)",
		threshold, totalValue,
	)
	if err != nil {
		return 0, err
	}

	recordID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// 插入建议
	for _, suggestion := range suggestions {
		_, err := tx.Exec(`
			INSERT INTO rebalance_suggestions 
			(record_id, fund_id, fund_name, fund_code, current_value, target_value, diff_value, advice, reason)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			recordID, suggestion.FundID, suggestion.FundName, suggestion.FundCode,
			suggestion.CurrentValue, suggestion.TargetValue, suggestion.DiffValue, suggestion.Advice, suggestion.Reason,
		)
		if err != nil {
			return 0, err
		}
	}

	return int(recordID), tx.Commit()
}

func getRebalanceHistory(limit int) ([]RebalanceRecord, error) {
	query := `
		SELECT id, threshold, total_value, created_at 
		FROM rebalance_records 
		ORDER BY created_at DESC 
		LIMIT ?
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []RebalanceRecord
	for rows.Next() {
		var record RebalanceRecord
		err := rows.Scan(&record.ID, &record.Threshold, &record.TotalValue, &record.CreatedAt)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

func getRebalanceRecordByID(recordID int) (*RebalanceRecord, error) {
	query := `
		SELECT id, threshold, total_value, created_at 
		FROM rebalance_records 
		WHERE id = ?
	`

	var record RebalanceRecord
	err := db.QueryRow(query, recordID).Scan(&record.ID, &record.Threshold, &record.TotalValue, &record.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func getRebalanceSuggestionsByRecordID(recordID int) ([]RebalanceSuggestion, error) {
	query := `
		SELECT id, record_id, fund_id, fund_name, fund_code, 
		       current_value, target_value, diff_value, advice, 
		       COALESCE(reason, '') as reason, created_at
		FROM rebalance_suggestions 
		WHERE record_id = ?
		ORDER BY fund_name
	`

	rows, err := db.Query(query, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suggestions []RebalanceSuggestion
	for rows.Next() {
		var suggestion RebalanceSuggestion
		err := rows.Scan(&suggestion.ID, &suggestion.RecordID, &suggestion.FundID,
			&suggestion.FundName, &suggestion.FundCode, &suggestion.CurrentValue,
			&suggestion.TargetValue, &suggestion.DiffValue, &suggestion.Advice,
			&suggestion.Reason, &suggestion.CreatedAt)
		if err != nil {
			return nil, err
		}
		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

// 转换函数：DB模型 -> API模型
func convertDBBucketsToAPIBuckets(dbBuckets []DBBucket) []Bucket {
	var buckets []Bucket
	for _, dbBucket := range dbBuckets {
		bucket := Bucket{
			ID:         dbBucket.ID,
			Name:       dbBucket.Name,
			TargetRate: dbBucket.TargetRate,
			Funds:      make([]Fund, len(dbBucket.Funds)),
		}

		for i, dbFund := range dbBucket.Funds {
			bucket.Funds[i] = Fund{
				ID:      dbFund.ID,
				Name:    dbFund.Name,
				Code:    dbFund.Code,
				Current: dbFund.Current,
				Weight:  dbFund.Weight,
				Target:  dbFund.Target,
				Diff:    dbFund.Diff,
				Advice:  dbFund.Advice,
			}
		}

		buckets = append(buckets, bucket)
	}
	return buckets
}

// 更新基金的再平衡结果到数据库
func updateFundRebalanceResults(dbBuckets []DBBucket, rebalancedBuckets []Bucket) error {
	// 构建基金ID到再平衡结果的映射
	fundResultMap := make(map[string]Fund) // 使用 "name-code" 作为key

	for _, bucket := range rebalancedBuckets {
		for _, fund := range bucket.Funds {
			key := fmt.Sprintf("%s-%s", fund.Name, fund.Code)
			fundResultMap[key] = fund
		}
	}

	// 更新数据库中的基金信息
	for _, dbBucket := range dbBuckets {
		for _, dbFund := range dbBucket.Funds {
			key := fmt.Sprintf("%s-%s", dbFund.Name, dbFund.Code)
			if result, exists := fundResultMap[key]; exists {
				query := `
					UPDATE funds 
					SET target = ?, diff = ?, advice = ?, updated_at = CURRENT_TIMESTAMP 
					WHERE id = ?
				`
				_, err := db.Exec(query, result.Target, result.Diff, result.Advice, dbFund.ID)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// 关闭数据库连接
func closeDatabase() {
	if db != nil {
		db.Close()
	}
}

// 投资组合收益计算相关函数
type PortfolioPerformance struct {
	TotalInvestment  float64 `json:"total_investment"`  // 初始投资金额
	CurrentValue     float64 `json:"current_value"`     // 当前市值
	TotalReturn      float64 `json:"total_return"`      // 总收益金额
	ReturnRate       float64 `json:"return_rate"`       // 收益率 (%)
	AnnualizedReturn float64 `json:"annualized_return"` // 年化收益率 (%)
	DaysHeld         int     `json:"days_held"`         // 持有天数
	YearsHeld        float64 `json:"years_held"`        // 持有年数
}

// 计算投资组合收益表现
func calculatePortfolioPerformance(portfolioID int) (*PortfolioPerformance, error) {
	// 获取投资组合基本信息
	portfolio, err := getPortfolioByID(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("获取投资组合信息失败: %v", err)
	}

	// 计算当前总市值
	var currentValue float64
	for _, bucket := range portfolio.Buckets {
		for _, fund := range bucket.Funds {
			currentValue += fund.Current
		}
	}

	// 计算持有时间
	now := time.Now()
	daysHeld := int(now.Sub(portfolio.CreatedAt).Hours() / 24)
	yearsHeld := now.Sub(portfolio.CreatedAt).Hours() / (24 * 365.25) // 考虑闰年

	// 计算收益
	totalReturn := currentValue - portfolio.TotalInvestmentAmount
	var returnRate, annualizedReturn float64

	if portfolio.TotalInvestmentAmount > 0 {
		returnRate = (totalReturn / portfolio.TotalInvestmentAmount) * 100

		// 计算年化收益率
		if yearsHeld > 0.001 { // 至少持有约8.76小时才计算年化收益率
			// 年化收益率 = ((当前价值/初始投资)^(1/年数) - 1) * 100
			if currentValue > 0 && portfolio.TotalInvestmentAmount > 0 {
				ratio := currentValue / portfolio.TotalInvestmentAmount
				if ratio > 0 {
					annualizedReturn = (math.Pow(ratio, 1/yearsHeld) - 1) * 100
					// 限制年化收益率在合理范围内 (-1000%, +1000%)
					if math.IsInf(annualizedReturn, 0) || math.IsNaN(annualizedReturn) {
						annualizedReturn = 0
					} else if annualizedReturn > 1000 {
						annualizedReturn = 1000
					} else if annualizedReturn < -1000 {
						annualizedReturn = -1000
					}
				}
			}
		} else {
			// 如果持有时间少于约8.76小时，年化收益率设为0
			annualizedReturn = 0
		}
	}

	return &PortfolioPerformance{
		TotalInvestment:  portfolio.TotalInvestmentAmount * 10000, // 转换为元
		CurrentValue:     currentValue * 10000,                    // 转换为元
		TotalReturn:      totalReturn * 10000,                     // 转换为元
		ReturnRate:       returnRate,
		AnnualizedReturn: annualizedReturn,
		DaysHeld:         daysHeld,
		YearsHeld:        yearsHeld,
	}, nil
}
