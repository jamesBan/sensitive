#敏感词过滤服务


##原理
传统的trie-tree过滤系统容易误判一些容易引起歧义的词语。
利用结巴分词技术，添加敏感词为自定义字库。每次文本分词后检查是否为自定义词组，符合自定义词组的即为敏感词。


##web服务
```go
package main

import (
	"time"
	"sensitive"
	filter2 "sensitive/filter"
	store2 "sensitive/store"
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/gookit/config"
	"github.com/gookit/config/yaml"
	"os"
	"github.com/pkg/errors"
)

type wordForm struct {
	Word string `form:"word" binding:"required"`
}

type contentForm struct {
	Content string `form:"content" binding:"required"`
}

type replaceForm struct {
	Content string `form:"content" binding:"required"`
	Replace string `form:"replace" binding:"required"`
}

func init() {
	loadConfig()
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	manager := createManager()
	r.POST("/sensitive/web/add", func(c *gin.Context) {
		var form wordForm
		if c.ShouldBind(&form) != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "missing word param",})
			return
		}

		err := manager.GetStore().Write(form.Word)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	r.DELETE("/sensitive/web/remove", func(c *gin.Context) {
		var form wordForm
		if c.ShouldBind(&form) != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "missing word param",})
			return
		}

		err := manager.GetStore().Remove(form.Word)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	r.GET("/sensitive/web/check", func(c *gin.Context) {
		var form contentForm
		if c.BindQuery(&form) != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "missing word param",})
			return
		}

		words := manager.GetFilter().Find(form.Content)
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"words": words,
		})
	})

	r.PUT("/sensitive/web/replace", func(c *gin.Context) {
		var form replaceForm
		if err := c.ShouldBind(&form); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		content := manager.GetFilter().Replace(form.Content, form.Replace)

		c.JSON(http.StatusOK, gin.H{
			"content": content,
		})
	})

	r.Run(":9512")
}

func loadConfig() {
	config.AddDriver(yaml.Driver)
	err := config.LoadFiles("config.yaml")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func createManager() *sensitive.Manager {
	store, err := createStore()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	filter, err := createFilter()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	interval := config.Int64("interval", 60)
	return sensitive.NewManager(store, filter, time.Duration(interval)*time.Second)
}

func createStore() (store2.Store, error) {
	driver := config.String("store_dirver")
	if driver == "mongo" {
		mongoConfig := config.StringMap("mongo_dirver")
		dsn := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s", mongoConfig["user"], mongoConfig["password"], mongoConfig["host"], mongoConfig["port"], mongoConfig["name"])
		store, err := store2.NewMongoStore(dsn, mongoConfig["name"], mongoConfig["table"], time.Second)
		if err != nil {
			return nil, errors.Wrapf(err, "connect mongo %s", err)
		}

		return store, nil
	}

	return nil, errors.New("error store driver")
}

func createFilter() (filter2.Filter, error) {
	filerName := config.String("filter")
	if filerName == "jieba" {
		pwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		jiebaConfig := config.StringMap("jieba_filter")
		dictPath := jiebaConfig["jieba_filter"]

		return filter2.NewJiebaFilter(
			pwd+"/"+dictPath+"/jieba.dict.utf8",
			pwd+"/"+dictPath+"/hmm_model.utf8",
			pwd+"/"+dictPath+"/user.dict.utf8",
			pwd+"/"+dictPath+"/idf.utf8",
			pwd+"/"+dictPath+"/stop_words.utf8",
		), nil
	}

	return nil, errors.New("error filter name")
}
```

###RPC服务

```go
package main

import (
	"time"
	"sensitive"
	filter2 "sensitive/filter"
	store2 "sensitive/store"
	"fmt"
)

const (
	DB_USER       = "root"
	DB_PASSWORD   = "root"
	DB_HOST       = "127.0.0.1"
	DB_PORET      = "27017"
	DB_NAME       = "test"
	DB_COOLECTION = "filter"
	DB_INTERVAL   = time.Second * 10 //定时更新数据库
)

func main() {
	dsn := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s", DB_USER, DB_PASSWORD, DB_HOST, DB_PORET, DB_NAME)
	store, err := store2.NewStore(dsn, DB_NAME, DB_COOLECTION, time.Second)
	if err != nil {
		panic(err)
	}

	filter := filter2.NewFilter(store)
	manager := sensitive.NewManager(store, filter, DB_INTERVAL)
	sensitive.RpcServer(manager, "9510")
}
```

参考http://blog.41ms.com/post/40.html