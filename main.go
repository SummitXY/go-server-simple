package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

// 声明一个全局的rdb变量
var rdb *redis.Client

// 初始化连接
func initClient() (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s",os.Getenv("REDIS_URL")),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err = rdb.Ping().Result()
	if err != nil {
		return err
	}
	return nil
}

const port = ":8888"

func main() {
	conn, err := net.Dial("tcp", "localhost:9001")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	msg := "ADD 50 2"
	header := make([]byte, 4)
	body := []byte(msg)
	binary.BigEndian.PutUint32(header, uint32(len(body)))

	_, err = conn.Write(header)
	if err != nil {
		log.Fatal(err)
	}
	_, err = conn.Write(body)
	if err != nil {
		log.Fatal(err)
	}

	buffer := make([]byte, 4)
	_, err = conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}

	res := binary.BigEndian.Uint32(buffer)
	log.Printf("ans :%d\n",res)
}

func main2() {
	// init redis
	err := initClient()
	if err != nil {
		panic(fmt.Sprintf("init client fail:%v",err))
	}
	r := gin.Default()

	r.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello, World!")
	})

	api := r.Group("/api")

	api.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	api.GET("/user/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.String(http.StatusOK, "Good Night %s", name)
	})

	redisRouter := r.Group("/redis")
	redisRouter.GET("/set/:key/:val", func(c *gin.Context) {
		k := c.Param("key")
		v := c.Param("val")
		err := rdb.Set(k, v, time.Minute * 30).Err()
		if err != nil {
			c.String(http.StatusOK, "set key:%s val:%s fail:%v",k,v,err)
		} else {
			c.String(http.StatusOK, "set key:%s val:%s ok",k,v)
		}
	})

	redisRouter.GET("/get/:key", func(c *gin.Context) {
		k := c.Param("key")
		v, err := rdb.Get(k).Result()
		if err != nil {
			c.String(http.StatusOK, "get key:%s fail err:%v", k, err)
		} else {
			c.String(http.StatusOK, "get key:%s success val:%s", k, v)
		}
	})

	fmt.Printf("Server running at port%s",port)
	r.Run(port)
}