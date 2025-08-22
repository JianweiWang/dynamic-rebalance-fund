package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// 数据库模型
type DBBucket struct {
	ID         int       `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	TargetRate float64   `json:"target_rate" db:"target_rate"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
	Funds      []DBFund  `json:"funds"`
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
		`CREATE TABLE IF NOT EXISTS buckets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			target_rate REAL NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
	err := db.QueryRow("SELECT COUNT(*) FROM buckets").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("数据库已有数据，跳过初始化")
		return nil
	}

	// 插入默认桶
	buckets := []struct {
		name       string
		targetRate float64
	}{
		{"短期桶（货币基金）", 0.10},
		{"中期桶（债券基金）", 0.30},
		{"长期桶（股票基金）", 0.60},
	}

	for _, bucket := range buckets {
		_, err := db.Exec(
			"INSERT INTO buckets (name, target_rate) VALUES (?, ?)",
			bucket.name, bucket.targetRate,
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
		{1, "易方达货币A", "000009", 20.0, 1.0},
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
func getAllBucketsFromDB() ([]DBBucket, error) {
	query := `
		SELECT id, name, target_rate, created_at, updated_at 
		FROM buckets 
		ORDER BY id
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buckets []DBBucket
	for rows.Next() {
		var bucket DBBucket
		err := rows.Scan(&bucket.ID, &bucket.Name, &bucket.TargetRate,
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
			Name:       dbBucket.Name,
			TargetRate: dbBucket.TargetRate,
			Funds:      make([]Fund, len(dbBucket.Funds)),
		}

		for i, dbFund := range dbBucket.Funds {
			bucket.Funds[i] = Fund{
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
