# IM Server Version 1.2

## 业务支撑

- [x] 4倍速度于ES的全文搜索引擎数据库
- [x] 本地高速KV数据库
- [x] HashMap纯内存级数据交互


## 开发进度

- [x] 单插入记录的聊天消息
- [ ] 重复聊天消息过滤
  - [ ] 根据亚马逊FIFO消息去重方法，设置消息唯一ID **[5]** 分钟生命周期，即重复消息ID将被接受，但不会传输
- [x] 获取最近和以前的聊天记录
- [x] 设置/获取用户读取的最后一条聊天记录ID
- [x] 联系人列表
- [x] 联系人黑名单
  - [x] 内存数据库业务
  - [x] 全文数据库业务
- [ ] 只能在对方回复消息后才能继续发送 **[2]** 条消息给对方
  - [x] 内存数据库业务
  - [ ] 全文数据库业务
- [x] 搜索历史聊天记录
  - [x] 点对点聊天记录
  - [x] 所有聊天记录
- [ ] 聊天通讯
  - [ ] 长连接Websocket业务逻辑
  - [ ] 用于网络质量差的短链接通讯业务逻辑
  - [ ] 唯一通讯授权密钥
- [ ] 对接用户信息主表（头像，名称，职位等）
  - [ ] 获取当前用户的个人基本信息
  - [ ] 获取数据库所有用户的基本信息

## 开发文档

