## 数据库存储

### 玩家信息
openid
昵称
头像
性别

### 玩家战绩
积分
胜场数
总场数

### 玩家个人设置
？？？

## 游戏流程

### 好友对战
房主创建房间
其它玩家输入房间号加入房间
房主点击开始游戏

### 初始分配
按0，1，2，3为玩家分配编号

### 游戏过程
转发房间内某玩家请求到其它各玩家

### 胜利
结束游戏，积分计入排行榜
四：1+6，2+3，3+1，4+0
三：1+4，2+2，3+0
二：1+2，2+0

## json格式
```json
{
    "action": "",
    "data": {

    }
}
```
```json
{
    "action": "",
    "status": 0,
    "msg": "",
    "data": {

    }
}
```
status:
-1或其它     失败
0           成功

### 登录
* 发送
```json
{
    "action": "login",
    "data": {
        "code": ""      
    }
}
```
code: wx.login() 得到的code  

* 响应
```json
{
    "action": "loginres",
    "status": 0,
    "msg": "",
    "data": {
        "openid": "",
        "session_key": ""
    }
}
```
status:
-1     系统繁忙  
0      请求成功  
40029  code无效  
45011  频率限制  
openid: 用户唯一标识  
session_key: 会话密钥  

### 个人信息
* 发送
```json
{
    "action": "updateuserinfo",
    "data": {
        "openid": ""
        "nickName": "",
        "avatarUrl": "",
        "gender": 0
    }
}
```
openid: 用户唯一标识  
nickName: wx.getUserInfo() 得到的nickName  
avatarUrl: wx.getUserInfo() 得到的avatarUrl  
gender: wx.getUserInfo() 得到的gender  

* 响应
```json
{
    "action": "updateuserinfores",
    "status": 0,
    "msg": "",
    "data": {
    }
}
```

### 创建房间
* 发送
```json
{
    "action": "createroom",
    "data": {
        "openid": "",
        "roomcapacity": 4
    }
}
```
roomcapacity: 房间容量，即4人房，3人房，2人房

* 响应
```json
{
    "action": "createroomres",
    "status": 0,
    "msg": "",
    "data": {
        "roomid": ""
    }
}
```
roomid: 其它玩家加入时要使用的房间id

### 进入房间
* 发送
```json
{
    "action": "enterroom",
    "data": {
        "openid": "",
        "roomid": ""
    }
}
```

* 响应
```json
{
    "action": "enterroomres",
    "status": 0,
    "msg": "",
    "data": {
        "roomcapacity": 4
        "members": [
        {
            "openid": "",
            "nickName": "",
            "avatarUrl": ""
        },
        {
            "openid": "",
            "nickName": "",
            "avatarUrl": ""
        },
        {
            "openid": "",
            "nickName": "",
            "avatarUrl": ""
        }
        ]
    }
}
```
members: 房间其它成员信息

### 其他人进入房间（服务器发送给玩家）
* 接收
```json
{
    "action": "otherenterroom",
    "data": {
        "openid": "",
        "nickName": "",
        "avatarUrl": "",
        "roomid": ""
    }
}
```

* 响应
```json
{
    "action": "otherenterroomres",
    "status": 0,
    "msg": "",
    "data": {
    }
}
```

### 房主开始房间内游戏
* 发送
```json
{
    "action": "startroomgame",
    "data": {
        "openid": "",
        "roomid": ""
    }
}
```

* 响应
```json
{
    "action": "startroomgameres",
    "status": 0,
    "msg": "",
    "data": {
        "members": [
        {
            "openid": "",
            "nickName": "",
            "avatarUrl": ""
        },
        {
            "openid": "",
            "nickName": "",
            "avatarUrl": ""
        },
        {
            "openid": "",
            "nickName": "",
            "avatarUrl": ""
        },
        {
            "openid": "",
            "nickName": "",
            "avatarUrl": ""
        }
        ]
    }
}
```

### 房间内游戏开始（服务器发送给玩家）
* 接收
```json
{
    "action": "roomgamestarted",
    "data": {
        "openid": "",
        "roomid": ""
    }
}
```

* 响应
```json
{
    "action": "roomgamestartedres",
    "status": 0,
    "msg": "",
    "data": {
        "members": [
        {
            "openid": "",
            "nickName": "",
            "avatarUrl": ""
        },
        {
            "openid": "",
            "nickName": "",
            "avatarUrl": ""
        },
        {
            "openid": "",
            "nickName": "",
            "avatarUrl": ""
        },
        {
            "openid": "",
            "nickName": "",
            "avatarUrl": ""
        }
        ]
    }
}
```

### 请求服务器转发
* 发送
```json
{
    "action": "broadcast",
    "data": {
        "openid": "",
        "roomid": "",
        "content": {
        
        }
    }
}
```

* 响应
```json
{
    "action": "broadcastres",
    "status": 0,
    "msg": "",
    "data": {

    }
}
```

### 服务器转发（服务器发送给玩家）
* 接收
```json
{
    "action": "otherbroadcast",
    "data": {
        "openid": "",
        "roomid": "",
        "content": {
        
        }
    }
}
```

* 响应
```json
{
    "action": "otherbroadcastres",
    "status": 0,
    "msg": "",
    "data": {

    }
}
```

### 提交成绩
* 发送
```json
{
    "action": "uploadscores",
    "data": {
        "openid": "",
        "roomid": "",
        "members": [
        {
            "openid": "",
            "score":
        },
        {
            "openid": "",
            "score":
        },
        {
            "openid": "",
            "score":
        },
        {
            "openid": "",
            "score":
        }
        ]
    }
}
```

* 响应
```json
{
    "action": "uploadscores",
    "status": 0,
    "msg": "",
    "data": {

    }
}
```
