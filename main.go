package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

var Client *redis.Client

func InitRedis() {
	Client = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6389",
		Password: "13579",
		DB:       0,
	})
}

func main() {
	engine := gin.Default()
	InitRedis()

	engine.POST("/upload", func(c *gin.Context) {
		fmt.Println("start upload file")
		// 已上传文件总字节数
		total := 0

		// 获得上传文件，key是 upload
		file, header, err := c.Request.FormFile("upload")
		if err != nil {
			panic(err)
		}

		// 本地创建文件，有就直接打开，没有再创建
		saveFile, err := os.OpenFile(header.Filename, os.O_RDWR, os.ModePerm)
		if err != nil {
			saveFile, _ = os.Create(header.Filename)
		}
		// 创建缓冲区，2048字节
		buf := make([]byte, 8196)

		// 执行client.Get之后，如果key不存在，肯定会返回Redis.Nil错误
		result, err := Client.Get(header.Filename).Int64()
		if err != nil {
			fmt.Println("execute redis#get, err is", err)
		}
		// 覆盖total
		total = int(result)
		// 设置上传文件seek,从文件开头开始偏移
		file.Seek(int64(total), io.SeekStart)
		fmt.Println("this part file has already been saved before", total)

		// 循环读取
		for {
			read, err := file.Read(buf)
			// 大文件突然中断上传，因为发生EOF错误，跳出循环？
			if err == io.EOF {
				break
			}
			// fmt.Println("we have read another", read, " bytes")
			// 跳过total字节，追加保存
			saveFile.WriteAt(buf, int64(total))
			// 将上传的字节数量保存到redis
			total += read
			if err = Client.Set(header.Filename, total, time.Duration(0)).Err(); err != nil {
				fmt.Println("redis save file size failed!")
			}
		}
		// 不用defer
		saveFile.Close()
		fmt.Println("upload file success, total size: ", total, " bytes")
		c.String(http.StatusOK, "upload file success")
	})

	engine.Run(":8080")
}
