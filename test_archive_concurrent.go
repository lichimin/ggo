package main

import (
	"context"
	"encoding/json"
	"fmt"
	"ggo/database"
	"ggo/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// 测试存档的并发处理能力
func TestArchiveConcurrent(t *testing.T) {
	// 配置数据库连接
	dsn := "host=localhost user=postgres password=postgres dbname=gamego port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移模型
	db.AutoMigrate(&models.Archive{})

	// 初始化Redis
	database.RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// 测试数据
	userID := uint(1)
	testData := map[string]interface{}{
		"level":    10,
		"score":    10000,
		"inventory": []string{"sword", "shield", "potion"},
		"position": map[string]int{"x": 100, "y": 200},
	}

	// 将测试数据转换为JSON字符串
	jsonData, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	// 并发测试参数
	concurrentWriters := 10
	concurrentReaders := 10
	writesPerWriter := 5
	readsPerReader := 10

	// 等待组
	var wg sync.WaitGroup
	wg.Add(concurrentWriters + concurrentReaders)

	// 记录开始时间
	startTime := time.Now()

	// 并发写入测试
	for i := 0; i < concurrentWriters; i++ {
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < writesPerWriter; j++ {
				// 为每个写入操作生成略有不同的数据
			writeData := make(map[string]interface{})
			for k, v := range testData {
				writeData[k] = v
			}
			writeData["writer_id"] = writerID
			writeData["write_count"] = j
			writeData["timestamp"] = time.Now().UnixNano()

			// 转换为JSON字符串
			dataBytes, err := json.Marshal(writeData)
			if err != nil {
				t.Errorf("Writer %d, write %d: failed to marshal data: %v", writerID, j, err)
				continue
			}

			// 使用Redis分布式锁确保并发安全
			lockKey := fmt.Sprintf("lock:archive:%d", userID)
			ctx := context.Background()
			lock := database.RedisClient.SetNX(ctx, lockKey, "locked", 5*time.Second)
			if lock.Val() == false {
				// 锁获取失败，重试
				j--
				continue
			}

			// 保存到数据库
			var archive models.Archive
			db.Transaction(func(tx *gorm.DB) error {
				result := tx.Where("user_id = ?", userID).First(&archive)
				if result.Error != nil {
					if result.Error == gorm.ErrRecordNotFound {
						// 创建新存档
						archive = models.Archive{
							UserID:   userID,
							JSONData: string(dataBytes),
						}
						return tx.Create(&archive).Error
					}
					return result.Error
				}

				// 更新现有存档
				archive.JSONData = string(dataBytes)
				return tx.Save(&archive).Error
			})

			// 释放锁
			database.RedisClient.Del(ctx, lockKey)

			// 更新Redis缓存
			redisKey := fmt.Sprintf("archive:%d", userID)
			database.RedisClient.Set(ctx, redisKey, string(dataBytes), 24*time.Hour)

			fmt.Printf("Writer %d: Write %d completed\n", writerID, j)
			}
		}(i)
	}

	// 并发读取测试
	for i := 0; i < concurrentReaders; i++ {
		go func(readerID int) {
			defer wg.Done()
			for j := 0; j < readsPerReader; j++ {
				// 从Redis缓存读取
				redisKey := fmt.Sprintf("archive:%d", userID)
				ctx := context.Background()
				jsonData, err := database.RedisClient.Get(ctx, redisKey).Result()
				if err != nil {
					// 缓存未命中，从数据库读取
					var archive models.Archive
					result := db.Where("user_id = ?", userID).First(&archive)
					if result.Error != nil {
						t.Errorf("Reader %d, read %d: failed to read from database: %v", readerID, j, result.Error)
						continue
					}
					jsonData = archive.JSONData

					// 更新缓存
					database.RedisClient.Set(ctx, redisKey, jsonData, 24*time.Hour)
				}

				// 解析JSON数据
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
					t.Errorf("Reader %d, read %d: failed to unmarshal data: %v", readerID, j, err)
					continue
				}

				fmt.Printf("Reader %d: Read %d completed, data: %v\n", readerID, j, data)
				}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 计算测试时间
	testTime := time.Since(startTime)
	fmt.Printf("\nTest completed in %v\n", testTime)
	fmt.Printf("Concurrent writers: %d, writes per writer: %d, total writes: %d\n", concurrentWriters, writesPerWriter, concurrentWriters*writesPerWriter)
	fmt.Printf("Concurrent readers: %d, reads per reader: %d, total reads: %d\n", concurrentReaders, readsPerReader, concurrentReaders*readsPerReader)
	fmt.Printf("Total operations: %d\n", concurrentWriters*writesPerWriter+concurrentReaders*readsPerReader)
	fmt.Printf("Operations per second: %.2f\n", float64(concurrentWriters*writesPerWriter+concurrentReaders*readsPerReader)/testTime.Seconds())
}

func main() {
	TestArchiveConcurrent(nil)
}
