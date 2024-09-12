package database

import (
	"chat_app_backend/config"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	cfg           = config.GetConfig()
	mongoDatabase *mongo.Database
	mongoOnce     sync.Once
)

func DBConnect() interface{} {
	db, err := ConnectDatabase()
	if err != nil {
		panic(err)
	}

	return db
}

func ConnectDatabase() (interface{}, error) {
	dbType := cfg.Database.Type

	switch dbType {
	case "mongodb":
		return connectMongoDB()
	// case "mysql":
	// 	return connectMySQL()
	// case "postgresql":
	// 	return connectPostgreSQL()
	default:
		log.Fatalf("Unsupported database type: %s", dbType)
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

func connectMongoDB() (*mongo.Database, error) {
	var mongoErr error

	mongoOnce.Do(func() {
		DBcfg := cfg.Database.MongoDB

		mongoURI := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=%s",
			DBcfg.Username,
			DBcfg.Password,
			DBcfg.Host,
			DBcfg.Port,
			DBcfg.DBName,
			DBcfg.AuthSource,
		)

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
	return mongoDatabase, mongoErr
}

// func connectMySQL() (*gorm.DB, error) {
// 	cfg := config.AppConfig.Database.MySQL
// 	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
// 		cfg.Username,
// 		cfg.Password,
// 		cfg.Host,
// 		cfg.Port,
// 		cfg.DBName,
// 	)

// 	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		return nil, err
// 	}

// 	fmt.Println("Connected to MySQL!")
// 	return db, nil
// }

// func connectPostgreSQL() (*gorm.DB, error) {
// 	cfg := config.AppConfig.Database.PostgreSQL
// 	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
// 		cfg.Host,
// 		cfg.Username,
// 		cfg.Password,
// 		cfg.DBName,
// 		cfg.Port,
// 	)

// 	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		return nil, err
// 	}

// 	fmt.Println("Connected to PostgreSQL!")
// 	return db, nil
// }
