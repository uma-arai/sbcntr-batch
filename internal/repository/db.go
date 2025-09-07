package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	dbInstance *DB
	once       sync.Once
)

type DB struct {
	*sqlx.DB
}

type DBConfig struct {
	Host     string
	Port     int
	UserName string
	Password string
	DBName   string
	SSLMode  string
}

func NewDB(cfg *DBConfig) (*DB, error) {
	once.Do(func() {
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host,
			cfg.Port,
			cfg.UserName,
			cfg.Password,
			cfg.DBName,
			cfg.SSLMode,
		)

		// X-Ray対応のSQLコンテキストを作成
		db, err := xray.SQLContext("postgres", dsn)
		if err != nil {
			log.Fatalf("failed to create X-Ray SQL context: %v", err)
		}

		conn := sqlx.NewDb(db, "postgres")

		// 接続プールの設定
		conn.SetMaxOpenConns(25)
		conn.SetMaxIdleConns(25)
		conn.SetConnMaxLifetime(5 * time.Minute)

		// 接続テスト
		if err := conn.Ping(); err != nil {
			db.Close()
			log.Fatalf("failed to ping database: %v", err)
		}

		// 接続成功をログに記録
		log.Println("DB connected successfully")

		dbInstance = &DB{conn}
	})

	return dbInstance, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	_, seg := xray.BeginSegment(context.Background(), "DB.Close")
	defer seg.Close(nil)

	return db.DB.Close()
}

// BeginTx starts a new transaction
func (db *DB) BeginTx() (*sqlx.Tx, error) {
	ctx, seg := xray.BeginSegment(context.Background(), "DB.BeginTx")
	defer seg.Close(nil)

	return db.DB.BeginTxx(ctx, nil)
}

// QueryContext wraps sqlx.DB.QueryContext with X-Ray tracing
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	ctx, seg := xray.BeginSubsegment(ctx, "DB.Query")
	if seg == nil {
		return db.DB.QueryxContext(ctx, query, args...)
	}
	defer seg.Close(nil)

	// クエリをメタデータとして追加
	if err := seg.AddMetadata("query", query); err != nil {
		log.Printf("Failed to add query metadata: %v", err)
	}

	rows, err := db.DB.QueryxContext(ctx, query, args...)
	if err != nil {
		seg.Close(err)
		return nil, err
	}

	return rows, nil
}

// QueryxContext wraps sqlx.DB.QueryxContext with X-Ray tracing
func (db *DB) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	ctx, seg := xray.BeginSubsegment(ctx, "DB.Queryx")
	if seg == nil {
		return db.DB.QueryxContext(ctx, query, args...)
	}
	defer seg.Close(nil)

	// クエリをメタデータとして追加
	if err := seg.AddMetadata("query", query); err != nil {
		log.Printf("Failed to add query metadata: %v", err)
	}

	rows, err := db.DB.QueryxContext(ctx, query, args...)
	if err != nil {
		seg.Close(err)
		return nil, err
	}

	return rows, nil
}

// ExecContext wraps sqlx.DB.ExecContext with X-Ray tracing
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	ctx, seg := xray.BeginSubsegment(ctx, "DB.Exec")
	if seg == nil {
		return db.DB.ExecContext(ctx, query, args...)
	}
	defer seg.Close(nil)

	// クエリをメタデータとして追加
	if err := seg.AddMetadata("query", query); err != nil {
		log.Printf("Failed to add query metadata: %v", err)
	}

	result, err := db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		seg.Close(err)
		return nil, err
	}

	return result, nil
}
