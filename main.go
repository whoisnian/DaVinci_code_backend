package main

import (
	"database/sql"
	gojson "encoding/json"
	"flag"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
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
var connall map[string]*websocket.Conn
var muxall map[string]*sync.Mutex

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

	// 调用微信api获取openid
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
	// 获取失败时json才有errcode和errmsg，且没有openid和session_key
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
		if !row.Next() {
			_, err = db.Exec("INSERT user SET openid=?,time=?", openid, time.Now().Format("2006-01-02 15:04:05"))
			if err != nil {
				fmt.Println("insert: ", err)
				return
			}
			_, err = db.Exec("INSERT score SET openid=?", openid)
			if err != nil {
				fmt.Println("insert: ", err)
				return
			}
			_, err = db.Exec("INSERT setting SET openid=?", openid)
			if err != nil {
				fmt.Println("insert: ", err)
				return
			}
		}

		connall[openid] = conn
		muxall[openid] = &sync.Mutex{}
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
		muxall[openid].Lock()
		conn.WriteJSON(res.Interface())
		muxall[openid].Unlock()
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
		conn.WriteJSON(res.Interface())
	}
}

// 更新个人信息
func updateuserinfo(json *simplejson.Json, conn *websocket.Conn) {
	openid, err := json.Get("data").Get("openid").String()
	if err != nil {
		fmt.Println("get openid: ", err)
		return
	}
	connall[openid] = conn
	muxall[openid] = &sync.Mutex{}
	nickName, err := json.Get("data").Get("nickName").String()
	if err != nil {
		fmt.Println("get nickName: ", err)
		return
	}
	avatarUrl, err := json.Get("data").Get("avatarUrl").String()
	if err != nil {
		fmt.Println("get avatarUrl: ", err)
		return
	}
	gender, err := json.Get("data").Get("gender").Int()
	if err != nil {
		fmt.Println("get gender: ", err)
		return
	}

	_, err = db.Exec("UPDATE user SET nickname=?,avatarurl=?,gender=? WHERE openid=?", nickName, avatarUrl, gender, openid)
	var res *simplejson.Json
	if err != nil {
		fmt.Println("update: ", err)
		res, err = simplejson.NewJson([]byte(`{
    "action": "updateuserinfores",
    "status": -1,
    "msg": "` + err.Error() + `",
    "data": {
    }
		}`))
		if err != nil {
			fmt.Println("new json: ", err)
			return
		}
	} else {
		res, err = simplejson.NewJson([]byte(`{
    "action": "updateuserinfores",
    "status": 0,
    "msg": "ok",
    "data": {
    }
		}`))
		if err != nil {
			fmt.Println("new json: ", err)
			return
		}
	}
	muxall[openid].Lock()
	conn.WriteJSON(res.Interface())
	muxall[openid].Unlock()
}

// 创建新的游戏房间
func createroom(json *simplejson.Json, conn *websocket.Conn) {
	openid, err := json.Get("data").Get("openid").String()
	if err != nil {
		fmt.Println("get openid: ", err)
		return
	}
	connall[openid] = conn
	muxall[openid] = &sync.Mutex{}
	roomcapacity, err := json.Get("data").Get("roomcapacity").Int()
	if err != nil {
		fmt.Println("get roomcapacity: ", err)
		return
	}

	rand.Seed(time.Now().UnixNano())
	roomid := strconv.Itoa(rand.Intn(899999) + 100000)
	cnt := 0
	for redisclient.Get("roomcap"+roomid).Err() != redis.Nil {
		roomid = strconv.Itoa(rand.Intn(899999) + 100000)
		cnt++
		if cnt > 10 {
			break
		}
	}

	if cnt <= 10 {
		err = redisclient.SetNX("roomcap"+roomid, roomcapacity, time.Hour).Err()
		if err == nil {
			redisclient.Del("room" + roomid)
			redisclient.RPush("room"+roomid, openid)
			redisclient.Expire("room"+roomid, time.Hour)
			redisclient.Set("userroom"+openid, roomid, time.Hour)
		}
	}

	var res *simplejson.Json
	if err != nil || cnt > 10 {
		fmt.Println("create: ", err)
		var errmsg string
		if cnt > 10 {
			errmsg = "服务器繁忙"
		} else {
			errmsg = err.Error()
		}
		res, err = simplejson.NewJson([]byte(`{
    "action": "createroomres",
    "status": -1,
	"msg": "` + errmsg + `",
    "data": {
    }
		}`))
		if err != nil {
			fmt.Println("new json: ", err)
			return
		}
	} else {
		res, err = simplejson.NewJson([]byte(`{
    "action": "createroomres",
    "status": 0,
    "msg": "ok",
    "data": {
	"roomid": "` + roomid + `"
    }
		}`))
		if err != nil {
			fmt.Println("new json: ", err)
			return
		}
	}
	muxall[openid].Lock()
	conn.WriteJSON(res.Interface())
	muxall[openid].Unlock()
}

// 加入到已有的房间
func enterroom(json *simplejson.Json, conn *websocket.Conn) {
	openid, err := json.Get("data").Get("openid").String()
	if err != nil {
		fmt.Println("get openid: ", err)
		return
	}
	connall[openid] = conn
	muxall[openid] = &sync.Mutex{}
	roomid, err := json.Get("data").Get("roomid").String()
	if err != nil {
		fmt.Println("get roomid: ", err)
		return
	}

	roomcap, err := redisclient.Get("roomcap" + roomid).Int()
	roomnow := int(redisclient.LLen("room" + roomid).Val())

	var openids []string
	if roomnow > 0 && roomnow < roomcap {
		openids = redisclient.LRange("room"+roomid, 0, -1).Val()
		redisclient.RPush("room"+roomid, openid)
		redisclient.Set("userroom"+openid, roomid, time.Hour)
	}

	var res *simplejson.Json
	if err != nil || roomnow <= 0 || roomnow >= roomcap {
		var errmsg string
		if roomnow <= 0 {
			errmsg = "房间不存在"
		} else if roomnow >= roomcap {
			errmsg = "房间已满"
		} else {
			errmsg = err.Error()
		}
		res, err = simplejson.NewJson([]byte(`{
    "action": "enterroomres",
    "status": -1,
	"msg": "` + errmsg + `",
    "data": {
    }
		}`))
		if err != nil {
			fmt.Println("new json: ", err)
			return
		}
	} else {
		row := db.QueryRow("SELECT nickname, avatarurl FROM user WHERE openid=?", openid)
		var othernickName string
		var otheravatarUrl string
		row.Scan(&othernickName, &otheravatarUrl)
		members := "["
		for key, memberid := range openids {
			if key != 0 {
				members = members + ","
			}
			row = db.QueryRow("SELECT nickname, avatarurl FROM user WHERE openid=?", memberid)
			var nickName string
			var avatarUrl string
			row.Scan(&nickName, &avatarUrl)
			members = members + `{
				"openid":"` + memberid + `",
				"nickName":"` + nickName + `",
				"avatarUrl":"` + avatarUrl + `"
			}`
			go otherenterroom(roomid, memberid, openid, othernickName, otheravatarUrl)
		}
		members = members + "]"
		res, err = simplejson.NewJson([]byte(`{
    "action": "enterroomres",
    "status": 0,
    "msg": "ok",
    "data": {
	"roomcapacity": ` + strconv.Itoa(roomcap) + `,
	"members": ` + members + `
    }
		}`))
		if err != nil {
			fmt.Println("new json: ", err)
			return
		}
	}
	muxall[openid].Lock()
	conn.WriteJSON(res.Interface())
	muxall[openid].Unlock()
}

// 其他人进入房间
func otherenterroom(roomid string, memberid string, openid string, nickName string, avatarUrl string) {
	var conn *websocket.Conn
	conn = connall[memberid]
	if conn == nil {
		fmt.Println("get user conn: ")
		return
	}

	res, err := simplejson.NewJson([]byte(`{
    "action": "otherenterroom",
    "data": {
	"openid":"` + openid + `",
	"nickName":"` + nickName + `",
	"avatarUrl":"` + avatarUrl + `",
	"roomid":"` + roomid + `"
    }
		}`))
	if err != nil {
		fmt.Println("new json: ", err)
		return
	}
	muxall[memberid].Lock()
	conn.WriteJSON(res.Interface())
	muxall[memberid].Unlock()
}

// 房主开始游戏
func startroomgame(json *simplejson.Json, conn *websocket.Conn) {
	openid, err := json.Get("data").Get("openid").String()
	if err != nil {
		fmt.Println("get openid: ", err)
		return
	}
	connall[openid] = conn
	muxall[openid] = &sync.Mutex{}
	roomid, err := json.Get("data").Get("roomid").String()
	if err != nil {
		fmt.Println("get roomid: ", err)
		return
	}

	roomcap, err := redisclient.Get("roomcap" + roomid).Int()
	roomnow := int(redisclient.LLen("room" + roomid).Val())

	var openids []string
	if err == nil && roomnow == roomcap {
		openids = redisclient.LRange("room"+roomid, 0, -1).Val()
	}

	var res *simplejson.Json
	if err != nil || roomnow != roomcap || openid != openids[0] {
		var errmsg string
		if err != nil {
			errmsg = "房间获取失败"
		} else if roomnow != roomcap {
			errmsg = "房间人数不足"
		} else if openid != openids[0] {
			errmsg = "只有房主才能开始游戏"
		} else {
			errmsg = err.Error()
		}
		res, err = simplejson.NewJson([]byte(`{
    "action": "startroomgameres",
    "status": -1,
	"msg": "` + errmsg + `",
    "data": {
    }
		}`))
		if err != nil {
			fmt.Println("new json: ", err)
			return
		}
	} else {
		members := "["
		for key, memberid := range openids {
			if key != 0 {
				members = members + ","
			}
			row := db.QueryRow("SELECT nickname, avatarurl FROM user WHERE openid=?", memberid)
			var nickName string
			var avatarUrl string
			row.Scan(&nickName, &avatarUrl)
			members = members + `{
				"openid":"` + memberid + `",
				"nickName":"` + nickName + `",
				"avatarUrl":"` + avatarUrl + `"
			}`
		}
		members = members + "]"
		for key, memberid := range openids {
			if key != 0 {
				go roomgamestarted(roomid, memberid, openid, members)
			}
		}
		res, err = simplejson.NewJson([]byte(`{
    "action": "startroomgameres",
    "status": 0,
    "msg": "ok",
    "data": {
	"openid":"` + openid + `",
	"roomid":"` + roomid + `",
	"members": ` + members + `
    }
		}`))
		if err != nil {
			fmt.Println("new json: ", err)
			return
		}
	}
	muxall[openid].Lock()
	conn.WriteJSON(res.Interface())
	muxall[openid].Unlock()
}

// 向房主之外的玩家发送开始游戏信号
func roomgamestarted(roomid string, memberid string, openid string, members string) {
	var conn *websocket.Conn
	conn = connall[memberid]
	if conn == nil {
		fmt.Println("get user conn: ")
		return
	}

	res, err := simplejson.NewJson([]byte(`{
    "action": "roomgamestarted",
    "data": {
	"openid":"` + openid + `",
	"roomid":"` + roomid + `",
	"members": ` + members + `
    }
		}`))
	if err != nil {
		fmt.Println("new json: ", err)
		return
	}
	muxall[memberid].Lock()
	conn.WriteJSON(res.Interface())
	muxall[memberid].Unlock()
}

// 请求转发消息
func broadcast(json *simplejson.Json, conn *websocket.Conn) {
	openid, err := json.Get("data").Get("openid").String()
	if err != nil {
		fmt.Println("get openid: ", err)
		return
	}
	connall[openid] = conn
	muxall[openid] = &sync.Mutex{}
	roomid, err := json.Get("data").Get("roomid").String()
	if err != nil {
		fmt.Println("get roomid: ", err)
		return
	}

	_, err = redisclient.Get("roomcap" + roomid).Int()
	roomnow := int(redisclient.LLen("room" + roomid).Val())

	var openids []string
	if err == nil && roomnow > 0 {
		openids = redisclient.LRange("room"+roomid, 0, -1).Val()
	}

	var res *simplejson.Json
	if err != nil || roomnow <= 0 {
		var errmsg string
		if err != nil {
			errmsg = "房间获取失败"
		} else if roomnow <= 0 {
			errmsg = "空房间"
		}
		res, err = simplejson.NewJson([]byte(`{
    "action": "broadcastres",
    "status": -1,
	"msg": "` + errmsg + `",
    "data": {
    }
		}`))
		if err != nil {
			fmt.Println("new json: ", err)
			return
		}
	} else {
		for _, memberid := range openids {
			if openid != memberid {
				go otherbroadcast(memberid, json)
			}
		}
		res, err = simplejson.NewJson([]byte(`{
    "action": "broadcastres",
    "status": 0,
    "msg": "ok",
    "data": {
    }
		}`))
		if err != nil {
			fmt.Println("new json: ", err)
			return
		}
	}
	muxall[openid].Lock()
	conn.WriteJSON(res.Interface())
	muxall[openid].Unlock()
}

// 向房间内其他人转发消息
func otherbroadcast(memberid string, json *simplejson.Json) {
	var conn *websocket.Conn
	conn = connall[memberid]
	if conn == nil {
		fmt.Println("get user conn: ")
		return
	}
	json.Set("action", "otherbroadcast")
	muxall[memberid].Lock()
	conn.WriteJSON(json.Interface())
	muxall[memberid].Unlock()
}

// 提交成绩
func uploadscores(json *simplejson.Json, conn *websocket.Conn) {
	openid, err := json.Get("data").Get("openid").String()
	if err != nil {
		fmt.Println("get openid: ", err)
		return
	}
	connall[openid] = conn
	muxall[openid] = &sync.Mutex{}
	roomid, err := json.Get("data").Get("roomid").String()
	if err != nil {
		fmt.Println("get roomid: ", err)
		return
	}
	store, err := json.Get("data").Get("store").Bool()
	if err != nil {
		fmt.Println("get store: ", err)
		return
	}
	members, err := json.Get("data").Get("members").Array()
	if err != nil {
		fmt.Println("get members: ", err)
		return
	}

	roomcap, err := redisclient.Get("roomcap" + roomid).Int()
	roomnow := int(redisclient.LLen("room" + roomid).Val())

	var openids []string
	if err == nil && roomnow == roomcap {
		openids = redisclient.LRange("room"+roomid, 0, -1).Val()
	}

	var res *simplejson.Json
	if err != nil || roomnow != roomcap || openid != openids[0] {
		var errmsg string
		if err != nil {
			errmsg = "房间获取失败"
		} else if roomnow != roomcap {
			errmsg = "房间状态异常"
		} else if openid != openids[0] {
			errmsg = "只有房主才能提交成绩"
		} else {
			errmsg = err.Error()
		}
		res, err = simplejson.NewJson([]byte(`{
    "action": "uploadscores",
    "status": -1,
	"msg": "` + errmsg + `",
    "data": {
    }
		}`))
		if err != nil {
			fmt.Println("new json: ", err)
			return
		}
	} else {
		if store {
			for _, member := range members {
				m, _ := member.(map[string]interface{})
				mopenid, _ := m["openid"].(string)
				mscore, _ := m["score"].(gojson.Number).Int64()
				if roomcap == 4 {
					_, err = db.Exec("UPDATE score SET scoreall=scoreall+?,num=num+1,num4=num4+1 WHERE openid=?", mscore, mopenid)
				} else if roomcap == 3 {
					_, err = db.Exec("UPDATE score SET scoreall=scoreall+?,num=num+1,num3=num4+1 WHERE openid=?", mscore, mopenid)
				} else if roomcap == 2 {
					_, err = db.Exec("UPDATE score SET scoreall=scoreall+?,num=num+1,num2=num4+1 WHERE openid=?", mscore, mopenid)
				}
				if err != nil {
					fmt.Println("update: ", err.Error())
					break
				}
			}
		}
		res, err = simplejson.NewJson([]byte(`{
    "action": "uploadscores",
    "status": 0,
    "msg": "ok",
    "data": {
    }
		}`))
		if err != nil {
			fmt.Println("new json: ", err)
			return
		}
	}
	muxall[openid].Lock()
	conn.WriteJSON(res.Interface())
	muxall[openid].Unlock()
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
			continue
		}
		fmt.Printf("%s", message)

		// 转换为json
		json, err := simplejson.NewJson(message)
		if err != nil {
			fmt.Println("newjson: ", err)
			continue
		}

		// 获取action
		action, err := json.Get("action").String()
		if err != nil {
			fmt.Println("get action: ", err)
			continue
		}

		// 调用对应请求
		if action == "login" {
			go login(json, conn)
		} else if action == "updateuserinfo" {
			go updateuserinfo(json, conn)
		} else if action == "createroom" {
			go createroom(json, conn)
		} else if action == "enterroom" {
			go enterroom(json, conn)
		} else if action == "startroomgame" {
			go startroomgame(json, conn)
		} else if action == "broadcast" {
			go broadcast(json, conn)
		} else if action == "uploadscores" {
			go uploadscores(json, conn)
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

	connall = make(map[string]*websocket.Conn)
	muxall = make(map[string]*sync.Mutex)

	// 监听websocket连接
	http.HandleFunc("/websocket", ws)
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		fmt.Println("listen: ", err)
	}
}
