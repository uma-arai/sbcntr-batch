package database

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DB struct {
	*sqlx.DB
}

type Config struct {
	Host     string
	Port     int
	UserName string
	Password string
	DBName   string
}

type SQLHandler struct {
	Conn *sqlx.DB
}

func NewDB(cfg Config) (*DB, error) {
	// localhostのDBの場合はSSLを無効化
	var sslModeValue string
	if cfg.Host == "localhost" || os.Getenv("DB_HOST") == "localhost" {
		sslModeValue = "disable"
	} else {
		sslModeValue = "require" // 本番環境ではSSLを有効にする
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.UserName,
		cfg.Password,
		cfg.DBName,
		sslModeValue,
	)

	// X-Ray対応のSQLコンテキストを作成
	db, err := xray.SQLContext("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database with X-Ray: %w", err)
	}

	// コネクションプールの設定
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 接続テスト
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{sqlx.NewDb(db, "postgres")}, nil
}
