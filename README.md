# IM

## IM_MESSAGE VERSION 1.1

### IMPROVEMENT (TODO)
```
//NOSQL check isset friend uniqueID = touser + fromuser
ok,_ := doubleCheck(uuid,uuid2)
if ok {
	isfriend
} else{
	//isblack: -1 in blacklist | 0 fromuser can't send msg to touser | 1-2 fromuser can send 1 or 2 msg to touser
	//fulltextDB double insert
	insert into im_friend_list (touser,fromuser,isblack,created) values (uuid,uuid2,friendstatus,createdtime)
	insert into im_friend_list (touser,fromuser,isblack,created) values (uuid2,uuid,friendstatus,createdtime)
	//add to memory mapping
}

//fulltextDB GET friends list
res = select * from im_friend_list where macth('@touser uuid') order by id desc
//encrypt each uuid2 for prevent brodcast attack
for range {
	AESCBC(uuid2)
}

```



## IM_MESSSAGE VERSION 1.0 
```
CREATE TABLE im_message (
id bigint,
touser text,
fromuser text,
msg text,
msgtype integer,
totype integer,
fromtype integer,
createdunix bigint,
msgid string
) ngram_len='1' ngram_chars='cjk'
```
### Comment
1. work with both cjk and non-cjk languages
2. message from user1[UUID] To user2[UUID]
3. msg contents
4. user type: 1 normal user | 2 enterprise user 
5. unixtime for order by
6. ~~timestamp for readable~~
7. msgid for filter repeat msgid

### SQL
```
select * from im_message where match('@touser t @fromuser f') and id > {last_read_id};
```

### QUESTION
1. Need store msgid? its NOT a good idea!
```
Any message published with the same deduplication ID, within the five-minute deduplication interval, is accepted but not delivered.[AMAZON FIFO]
```

### SOLUTION
```
RoseDB With TTL


```


## IM_MESSAGE_LAST_READ
```
CREATE TABLE im_message_last_read (
id bigint,
touser text,
fromuser text,
lastid bigint
)
```
### Comment
1. REPLACE works similarly to INSERT, but it marks the previous document with the same ID as deleted before inserting a new one.
2. Transactions are supported for the following commands:
```
INSERT
REPLACE
DELETE
```

### Question
1. How to generate an unique ID ? its NOT a good idea! 

## IM_MESSAGE_LAST_READ MYSQL
```
CREATE TABLE im_message_last_read (
     id BIGINT NOT NULL AUTO_INCREMENT,
     touser CHAR(32) NOT NULL,
     fromuser CHAR(32) NOT NULL,
     lastid bignt NOT NULL DEFAULT 0
     PRIMARY KEY (id)
);
ALTER TABLE im_message_last_read ADD CONSTRAINT unique_index UNIQUE (touser,frouser);
```

## IM_MESSAGE_LAST_READ MEMORY_MAPPING
```
lastreadid := "GET_FROM_CLIENT_FRONT_END_MSG_ID_ACK"

lastreadlist := gmap.New(true)
//TODO load data from mysql database

lastreadlist.SetIfNotExistFuncLock("TOUSER_UUID",func() interface{} {
     return gmap.New(true)
})

lastreadlist.Get("TOUSER_UUID").(*gmap.AnyAnyMap).Set("FROM_UUID",lastreadid)

//insert mysql last_read_id ASYNC
```


## IM_FRIEND_LIST MYSQL
```
CREATE TABLE im_friend_list (
     id BIGINT NOT NULL AUTO_INCREMENT,
     touser CHAR(32) NOT NULL,
     fromuser CHAR(32) NOT NULL,
     isblack tinyint(1) NOT NULL DEFAULT 2
     PRIMARY KEY (id)
);
ALTER TABLE im_friend_list ADD CONSTRAINT unique_index UNIQUE (touser,frouser);
```
### Comment
1. isblack: -1 in blacklist | 0 fromuser can't send msg to touser | 1-2 fromuser can send 1 or 2 msg to touser
2. use MYSQL for datastore and MEMORY Mapping for realtime

## IM_FRIEND_LIST MEMORY_MAPPING
```
friendlist := gmap.New(true)
//TODO load data from mysql database

//func SET TOUSER_UUID
needinserttomysql := friendlist.SetIfNotExistFuncLock("TOUSER_UUID",func() interface{} {
     return gmap.New(true)
})

if needinserttomysql {
	//insert to mysql ASYNC
}

//func GET SET BLACKLIST
isblack := friendlist.Get("TOUSER_UUID").(*gmap.AnyAnyMap).GetOrSetFuncLock("FROM_UUID", func() interface{} {
		return 2
	}).(int)

	switch isblack {
	case -1:
		// can't send to user for in its blacklist
		break
	case 0:
		//can't send to user for no replied
		break
	default:
		friendlist.Get("TOUSER_UUID").(*gmap.AnyAnyMap).GetOrSetFuncLock("FROM_UUID", func() interface{} {
			val := isblack - 1
			if val >= 0 {
				return val
			} 
			return 0
			
		}).(int)
          //insert msg to msg list
          //get insert id
		  //insert to last read db store
          //notify touser last received msg id
		  //broadcast fromuser last received msg id
	}

```

