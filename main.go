package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// 配置选项
var addr = flag.String("addr", "0.0.0.0:8080", "http service address")
var dsn = flag.String("dsn", "davinci:4KzyzTL9gyQpycJ9@/DaVinci_code", "database address")
var redisAddr = flag.String("redisAddr", "127.0.0.1:6379", "redis address")
var redisPass = flag.String("redisPass", "", "redis password")
var appid = flag.String("appid", "", "wechat appid")
var secret = flag.String("secret", "", "wechat secret")

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
		return
	}
	fmt.Println(code)

	var httpclient http.Client
	req, err := http.NewRequest(http.MethodGet, "https://api.weixin.qq.com/sns/jscode2session?appid="+*appid+"&secret="+*secret+"&js_code="+code+"&grant_type=authorization_code", nil)
	if err != nil {
		fmt.Println("new request: ", err)
		return
	}
	resp, err := httpclient.Do(req)
	if err != nil {
		fmt.Println("do req: ", err)
		return
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("read all: ", err)
		return
	}
	respjson, err := simplejson.NewJson(content)
	if err != nil {
		fmt.Println("newjson: ", err)
		return
	}
	var errcode int
	_, ok := respjson.CheckGet("errcode")
	if ok {
		errcode, err = respjson.Get("errcode").Int()
		if err != nil {
			fmt.Println("get errcode: ", err)
			return
		}
	} else {
		errcode = 0
	}
	var res *simplejson.Json
	if errcode == 0 {
		openid, err := respjson.Get("openid").String()
		session_key, err := respjson.Get("session_key").String()

		row, err := db.Query("SELECT 1 from user where openid=? limit 1", openid)
		defer row.Close()
		if err != nil {
			fmt.Println("select: ", err)
			return
		}
		if row.Next() {
			return
		}
		_, err = db.Exec("INSERT user SET openid=?,time=?", openid, time.Now().Format("2006-01-02 15:04:05"))
		_, err = db.Exec("INSERT score SET openid=?", openid)
		_, err = db.Exec("INSERT setting SET openid=?", openid)
		if err != nil {
			fmt.Println("insert: ", err)
			return
		}

		res, err = simplejson.NewJson([]byte(`{
    "action": "loginres",
    "status": 0,
    "msg": "ok",
    "data": {
        "openid": "` + openid + `",
        "session_key": "` + session_key + `"
    }
		}`))
		if err != nil {
			fmt.Println("new json: ", err)
			return
		}
	} else {
		errmsg, err := respjson.Get("errmsg").String()
		res, err = simplejson.NewJson([]byte(`{
    "action": "loginres",
    "status": ` + strconv.Itoa(errcode) + `,
    "msg": "` + errmsg + `",
    "data": {
        "openid": "",
        "session_key": ""
    }
		}`))
		if err != nil {
			fmt.Println("new json: ", err)
			return
		}
	}
	conn.WriteJSON(res.Interface())
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
	var err error
	db, err = sql.Open("mysql", *dsn)
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
