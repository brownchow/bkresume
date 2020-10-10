# Go Redis 实现断点续传功能



 通过循环取固定量的字节，取一次将取得的数据量做叠加，总数（已读取的文件字节数）存入到 `redis`，如果中途中断了，下一次校对文件名或者文件 `hash` , 传输的时候就先偏移到之前传输的字节位置，然后继续上传。 



## 准备工作

依赖包：

1、https://github.com/go-redis/redis

2、https://github.com/gin-gonic/gin.git

工具：

1、Postman 或者 httpie

写完代码之后，在项目根目录下执行 `go mod init` ，此时会出现 `go.mod`  和 `go.sum`  文件，这就是一个类似 `maven` 的包管理工具。

然后执行 `go run main.go` 才能跑起来

2、redis

可以在本地安装，或者用 `docker` 容器



## Go后台代码

第一版的后台代码：

```go
import (
	"fmt"
	"io"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()

	engine.POST("/upload", func(c *gin.Context) {
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
		defer saveFile.Close()

		// 创建缓冲区，2048字节
		buf := make([]byte, 2048)

		// 循环读取
		for {
			read, err := file.Read(buf)
			if err == io.EOF {
				break
			}
			fmt.Println("we have read another", read, " bytes")
			saveFile.Write(buf)
		}
        c.String(http.StatusOK, "upload file success")   
	})
    
	engine.Run(":8080")
}
```

相应的服务端，咱用 `httpie` 执行下面一段命令：

```sh
http --form http://127.0.0.1:8080/upload  upload@./title.png
```

`--form` 可以简写成 `-f`， `upload` 是上传的表单的 `key`，与服务端代码对应，`@` 后跟的是本地路径

`httpie` 官方文档的介绍：

If one or more file fields is present, the serialization and content type is `multipart/form-data`:

```bash
▶ RUN$ http -f POST httpbin.org/post name='John Smith' cv@~/files/data.xml
```

The request above is the same as if the following HTML form were submitted:

```html
<form enctype="multipart/form-data" method="post" action="http://example.com/jobs">
    <input type="text" name="name" />
    <input type="file" name="cv" />
</form>
```



增加断点续传功能主要是写缓冲要记录下之前读取的总字节数，存入` Redis` 中，中途断了重新上传，会去 redis 中获取之前已经写入的字节数，





## TODO:

1、实现下载的断点续传功能

2、将文件上传到 FastDFS，而不是存在本地文件系统

3、多线（协）程的时候，如何统计，是否可以不用 Redis，反正就是存一个数据

4、将文件的 Hash 作为 key 存储到 Redis 中，文件秒传



## 常见问题：

### 1、依赖出错，下载不了依赖包？

```sh
export GOPROXY=https://goproxy.cn
```

然后再执行原来的命令



### 2、http错误：

```sh
Content-Type: multipart/form-data; boundary=217e91dc518615dc8509c8b9f9c4b44a
```

boundary 是什么意思？



### 3、感觉还是一次性全部从内存中写入的，而不是每次只写一部分

































































































