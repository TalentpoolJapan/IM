# IM Server Version 1.4.2

## 业务支撑

- [x] 4倍速度于ES的全文搜索引擎数据库
- [x] 本地高速KV数据库
- [x] HashMap纯内存级数据交互


## 开发进度

- [x] 单插入记录的聊天消息
- [x] 重复聊天消息过滤
  - [x] 根据亚马逊FIFO消息去重方法，设置消息唯一ID **[5]** 分钟生命周期，即重复消息ID将被接受，但不会传输
- [x] 获取最近和以前的聊天记录
- [x] 设置/获取用户读取的最后一条聊天记录ID
- [x] 联系人列表
- [x] 联系人黑名单
  - [x] 内存数据库业务
  - [x] 全文数据库业务
- [x] 只能在对方回复消息后才能继续发送 **[2]** 条消息给对方
  - [x] 内存数据库业务
  - [x] 全文数据库业务
- [x] 搜索历史聊天记录
  - [x] 点对点聊天记录
  - [x] 所有聊天记录
- [x] 聊天通讯
  - [x] 基本通讯(接收/发送)框架
  - [x] 长连接Websocket业务逻辑
    - [x] 设置/删除通讯连接
    - [x] 注册登录到聊天系统 
    - [x] 针对每个连接都有唯一通讯~~授权密钥~~标识
    - [x] 消息收发对接后端业务
      - [x] 消息收发按照规则过滤
        - [x] 收发消息传送英文日文昵称公司名称和头像
      - [x] 记录最后已读消息ID
      - [x] 获取当前用户的联系人列表（含黑名单）
      - [x] 设置到黑名单
      - [x] 根据最近已读消息ID获取最近聊天消息和历史消息
    - [ ] ~~全局黑名单屏蔽恶意群发SPAM用户~~
    - [ ] ~~恶意词汇过滤 [可选]~~
  - [ ] ~~用于网络质量差的短链接通讯业务逻辑~~ 不适合多用户同步聊天场景
- [x] 对接用户信息主表（头像，名称，职位等）
  - [x] 获取当前用户的个人基本信息
  - [x] 获取单个联系人的个人基本信息
  - [ ] ~~获取数据库所有用户的基本信息~~

## 开发文档 [多线程需要对每条消息进行回复告知状态]

### Manticore 数据库创建语句
```
CREATE TABLE im_friend_list (
  id bigint,
  touser text,
  fromuser text,
  isblack integer,
  status integer,
  count integer,
  created bigint,
  nexttime bigint
)

CREATE TABLE im_message (
  id bigint,
  touser text,
  fromuser text,
  msg text,
  sessionid text,
  msgtype integer,
  totype integer,
  fromtype integer,
  created bigint,
  msgid string attribute
) ngram_len='1' ngram_chars='cjk'
```


### 连接服务器 

```
//初次握手没有升级前还是http协议，请携带原网站的 Header["Authorization"]
ws://IP:8888/ws

//服务重启后内存中是没有该用户的所以会有一个初始化这个用户的过程 如果存在跳过这步
//除错误信息以外首先会告知正在初始化用户
{"action": "SystemMsg", "code": 20000, "msg": "INITIALIZING_USER"}

//初始化完毕以后会提示，并告知当前用户的唯一用户ID即UUID
{"action": "SystemMsg", "code": 20001, "msg": "IS_READY", "uuid": user.Uuid}
```

### 收发送消息 

```
//发送普通消息
{"action":"SendP2PMsg","msgid":"前端唯一消息ID","touser":"接收用户的UUID","msg":"some msgs"}

//群发正常处理后群发给发送人以及接收方的 所有接入IM的连接Conn
{"action":"RecvP2PMsg","readid":"服务器插入的后返回的消息ID","msgid":"原样前端唯一消息ID","fromuser":"sender_uuid","touser":"接收用户的UUID","msg":"some msgs"}
```

### 更新最后阅读过的服务器消息ID [阅读过指的是点开过这个聊天对话框，接收到的不算已阅读]
```
//多连接情况下，不能保证每个客户端看到的和阅读过的内容是一样的
//发送已读最后消息ID 它发给我的消息我已经读过的消息ID是last_read_msg_id 系统处理
{"action":"PutRecvMsgId","msgid":"前端唯一消息ID","touser":"它","msg":"last_read_msg_id"}
//回复
{"action":"RecvMsgId","msgid":"前端唯一消息ID","fromuser":"sender_uuid","touser":"它","msg":"last_read_msg_id"}
```

### 有条件的获取聊天记录 默认一次返回10条
```
//通过服务器的消息ID获取新的或者旧的消息

//比当前消息ID新
{"action":"GetP2PMsgsNew","msgid":"前端唯一消息ID","touser":"它","msg":"服务器的消息ID"}

//比当前消息ID旧
{"action":"GetP2PMsgsOld","msgid":"前端唯一消息ID","touser":"它","msg":"服务器的消息ID"}

//最近的消息
{"action":"GetP2PMsgsRecent","msgid":"前端唯一消息ID","touser":"它","msg":"0"}
//回复
{"action":"RecvP2PMsgsNew","msgid":"前端唯一消息ID","fromuser":"sender_uuid(我)","touser":"它","msg":[{ImMessage}]}

{"action":"RecvP2PMsgsOld","msgid":"前端唯一消息ID","fromuser":"sender_uuid(我)","touser":"它","msg":[{ImMessage}]}

{"action":"RecvP2PMsgsRecent","msgid":"前端唯一消息ID","fromuser":"sender_uuid(我)","touser":"它","msg":[{ImMessage}]}

//因为是一条数据记录的聊天记录信息，所以这里的touser和fromuser需要自行判断哪个是 **[我]** 
ImMessage = {"id":123,"touser":"","fromuser":"","msg":"","totype":1,"fromtype":0,"created":123,"msgid":""}
```

### 黑名单操作
```
//添加到黑名单 我要把它放到黑名单
{"action":"PutBlack","msgid":"前端唯一消息ID","touser":"它","msg":"空"}
//群发给当前用户回复
{"action":"RecvBlack","msgid":"前端唯一消息ID","fromuser":"sender_uuid(我)","touser":"它","msg":"空"}
//移除黑名单 我要把它移出黑名单
{"action":"DelBlack","msgid":"前端唯一消息ID","touser":"它","msg":"空"}
//群发给当前用户回复
{"action":"RecvDelBlack","msgid":"前端唯一消息ID","fromuser":"sender_uuid(我)","touser":"它","msg":"空"}
```


### 获取自己或者其他人的基本信息
```
//获取个人基本信息 我要获取它的基本信息
//获取我的个人基本信息 touser 必须要和sender_uuid(我)一致
{"action":"GetMyProfile","msgid":"","touser":"我","msg":"空"}
{"action":"RecvMyProfile","msgid":"","fromuser":"sender_uuid(我)","touser":"我","msg":"空"}
//获取其他人的基本信息
{"action":"GetUserProfile","msgid":"","touser":"它","msg":UserProfile}
{"action":"RecvUserProfile","msgid":"","fromuser":"sender_uuid(我)","touser":"它","msg":UserProfile}
UserProfile = {"uuid":"","full_name_en":"","full_name_ja","avatar":""}
```

### 获取我的联系人列表
```
//touser 必须要和sender_uuid(我)一致
{"action":"GetMyContacts","msgid":"","touser":"我","msg":"空"}
//回复
{"action":"RecvyContacts","msgid":"","fromuser":"sender_uuid(我)","touser":"我","msg":[{MyContacts}]}
MyContacts={"uuid":"","full_name_en":"","full_name_ja":"","avatar":""}
```

