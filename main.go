package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"net/http"
)

// 配置选项
var addr = flag.String("addr", "0.0.0.0:8080", "http service address")
var dsn = flag.String("dsn", "davinci:4KzyzTL9gyQpycJ9@/DaVinci_code", "database address")
var redisAddr = flag.String("redisAddr", "127.0.0.1:6379", "redis address")
var redisPass = flag.String("redisPass", "", "redis password")

var db *sql.DB
var redisclient *redis.Client

// 忽略Origin检查
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 登录请求
func login(json *simplejson.Json, conn *websocket.Conn) {
	code, err := json.Get("data").Get("code").String()
	if err != nil {
		fmt.Println("get code: ", err)
	}
	fmt.Println(code)
}

func ws(w http.ResponseWriter, r *http.Request) {
	// 建立websocket连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("upgrade: ", err)
		return
	}
	defer conn.Close()

	for {
		// 接收消息
		mt, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("read: ", err)
			break
		}
		if mt != websocket.TextMessage {
			break
		}

		// 转换为json
		json, err := simplejson.NewJson(message)
		if err != nil {
			fmt.Println("newjson: ", err)
			break
		}

		// 获取action
		action, err := json.Get("action").String()
		if err != nil {
			fmt.Println("get action: ", err)
			break
		}

		// 调用对应请求
		if action == "login" {
			go login(json, conn)
		}
	}
}

func main() {
	flag.Parse()

	// 连接mysql数据库
	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		fmt.Println("database: ", err)
		return
	}
	defer db.Close()

	// 连接redis数据库
	redisclient = redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		Password: *redisPass,
		DB:       0,
	})
	_, err = redisclient.Ping().Result()
	if err != nil {
		fmt.Println("redis: ", err)
		return
	}

	// 监听websocket连接
	http.HandleFunc("/websocket", ws)
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		fmt.Println("listen: ", err)
	}
}
