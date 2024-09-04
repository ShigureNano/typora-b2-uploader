package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kurin/blazer/b2"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <file-to-upload> [more-files-to-upload...]\n", os.Args[0])
	}

	filePaths := os.Args[1:]

	// 配置B2凭证和自定义域名
	applicationKeyId := "" // 替换为你的B2应用程序密钥ID
	applicationKey := ""   // 替换为你的B2应用程序密钥
	customUrl := ""        // 替换为你的自定义URL
	bucketName := ""       // 替换为你的B2桶名称

	// 检查配置是否完整
	if applicationKeyId == "" || applicationKey == "" || customUrl == "" || bucketName == "" {
		log.Fatal("B2_APPLICATION_KEY_ID, B2_APPLICATION_KEY, B2_CUSTOM_URL, and B2_BUCKET_NAME environment variables must be set")
	}
	// 确保 customUrl 以 / 结尾
	if !strings.HasSuffix(customUrl, "/") {
		customUrl += "/"
	}
	ctx := context.Background()
	client, err := b2.NewClient(ctx, applicationKeyId, applicationKey)
	if err != nil {
		log.Fatalf("Failed to create B2 client: %v", err)
	}

	bucket, err := client.Bucket(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to get B2 bucket: %v", err)
	}

	var wg sync.WaitGroup
	for _, filePath := range filePaths {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			if err := uploadFile(ctx, bucket, filePath, customUrl); err != nil {
				log.Printf("Failed to upload file %s: %v", filePath, err)
			}
		}(filePath)
	}
	wg.Wait()
}

func uploadFile(ctx context.Context, bucket *b2.Bucket, filePath, customUrl string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 生成文件名
	fileName := generateFileName(filePath)

	// 获取MIME类型
	mimeType := getMimeType(filePath)

	obj := bucket.Object(fileName)
	writer := obj.NewWriter(ctx)

	// 使用 WithAttrs 设置 contentType
	writer.WithAttrs(&b2.Attrs{ContentType: mimeType})

	if _, err := io.Copy(writer, file); err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	url := fmt.Sprintf("%s%s", customUrl, fileName)
	fmt.Println(url)
	return nil
}

// 生成文件名
func generateFileName(filePath string) string {
	now := time.Now().Format("060102")
	hash := md5.New()
	hash.Write([]byte(filePath + now))
	hashValue := hex.EncodeToString(hash.Sum(nil))[:8]
	randomValue := rand.Intn(1000)
	ext := filepath.Ext(filePath)
	fileName := fmt.Sprintf("%s-%s%03d%s", now, hashValue, randomValue, ext)
	return fileName
}

// 获取MIME类型
func getMimeType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return mimeType
}
