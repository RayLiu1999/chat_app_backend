package providers

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"chat_app_backend/config"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	cfg = config.GetConfig()
	// mongoDatabase *mongo.Database
	// mongoOnce     sync.Once
	// postgresPool  *pgxpool.Pool
	// pgOnce        sync.Once
)

// DBConnection 介面定義了不同資料庫連接的通用行為
// 這裡我們假設所有連接都應該有 Close 和 Ping 方法
type DBConnection interface {
	Close()                         // 例如，mongo.Client 有 Close() 方法，pgxpool.Pool 也有 Close() 方法
	Ping(ctx context.Context) error // 例如，mongo.Client 和 pgxpool.Pool 都有 Ping() 方法
	// 根據你的需求，可以添加更多通用方法，例如：
	// Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
	// Execute(ctx context.Context, query string, args ...interface{}) (Result, error)
}

// MongoWrapper 結構體包裝 *mongo.Database，使其實現 DBConnection 介面
type MongoWrapper struct {
	Client *mongo.Client   // 實際的 mongo Client
	DB     *mongo.Database // 實際的 mongo Database
}

func (mw *MongoWrapper) Close() {
	if mw.Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mw.Client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}
}

func (mw *MongoWrapper) Ping(ctx context.Context) error {
	if mw.Client == nil {
		return fmt.Errorf("mongo client is not initialized")
	}
	return mw.Client.Ping(ctx, nil)
}

// PgxPoolWrapper 結構體包裝 *pgxpool.Pool，使其實現 DBConnection 介面
type PgxPoolWrapper struct {
	Pool *pgxpool.Pool
}

func (pw *PgxPoolWrapper) Close() {
	if pw.Pool != nil {
		pw.Pool.Close()
	}
}

func (pw *PgxPoolWrapper) Ping(ctx context.Context) error {
	if pw.Pool == nil {
		return fmt.Errorf("pgxpool is not initialized")
	}
	return pw.Pool.Ping(ctx)
}

// DBConnect 連接數據庫並返回相應類型的數據庫實例
func DBConnect[T DBConnection](dbType string) (T, error) {
	var result T
	switch dbType {
	case "mongodb":
		client, db, err := connectMongoDB()
		if err != nil {
			return result, err
		}
		return any(&MongoWrapper{Client: client, DB: db}).(T), nil
	case "postgresql":
		db, err := connectPostgreSQL()
		if err != nil {
			return result, err
		}
		return any(&PgxPoolWrapper{Pool: db}).(T), nil
	default:
		return result, fmt.Errorf("不支持的資料庫類型: %s", dbType)
	}
}

// 更新 connectMongoDB 函數，使其返回 *mongo.Client 和 *mongo.Database
var (
	mongoOnce     sync.Once
	mongoClient   *mongo.Client   // 全局 mongo Client
	mongoDatabase *mongo.Database // 全局 mongo Database
)

func connectMongoDB() (*mongo.Client, *mongo.Database, error) {
	var mongoErr error

	mongoOnce.Do(func() {
		DBcfg := cfg.Database.MongoDB
		isDebug := cfg.Server.Mode

		var mongoURI string
		if isDebug == "debug" {
			mongoURI = fmt.Sprintf("mongodb://%s:%s",
				DBcfg.Host,
				DBcfg.Port,
			)
		} else {
			mongoURI = fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=%s",
				DBcfg.Username,
				DBcfg.Password,
				DBcfg.Host,
				DBcfg.Port,
				DBcfg.DBName,
				DBcfg.AuthSource,
			)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
		if err != nil {
			mongoErr = err
			return
		}

		err = client.Ping(context.TODO(), nil)
		if err != nil {
			mongoErr = fmt.Errorf("could not connect to MongoDB: %v", err)
			return
		}

		mongoDatabase = client.Database(DBcfg.DBName)
		fmt.Println("Connected to MongoDB!")
	})
	return mongoClient, mongoDatabase, mongoErr
}

// connectPostgreSQL 函數保持不變
var (
	pgOnce sync.Once
	pgPool *pgxpool.Pool // 全局 pgxpool.Pool
)

func connectPostgreSQL() (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	var err error

	pgOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pgCfg := cfg.Database.PostgreSQL
		connectURI := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
			pgCfg.Username,
			pgCfg.Password,
			pgCfg.Host,
			pgCfg.Port,
			pgCfg.DBName,
			pgCfg.SSLMode)

		config, err := pgxpool.ParseConfig(connectURI)
		if err != nil {
			return
		}

		// 設置連接池參數
		config.MaxConns = 10                      // 最大連接數
		config.MaxConnLifetime = 1 * time.Hour    // 連接最大存活時間
		config.MaxConnIdleTime = 30 * time.Minute // 最大閒置時間
		config.MinConns = 2                       // 最小連接數

		pool, err = pgxpool.ConnectConfig(ctx, config)
		if err != nil {
			return
		}

		// 測試連接
		err = pool.Ping(ctx)
		if err != nil {
			return
		}

		log.Println("成功連接到 PostgreSQL!")
	})

	return pool, err
}
