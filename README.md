# IM

## IM_MESSSAGE 
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
createdstamp timestamp
) ngram_len='1' ngram_chars='cjk'
```
### Comment
1. work with both cjk and non-cjk languages
2. message from user1[UUID] To user2[UUID]
3. msg contents
4. user type: 1 normal user | 2 enterprise user 
5. unixtime for order by
6. timestamp for readable

### SQL
```
select * from im_message where match('@touser t @fromuser f') and id >= {last_read_id};
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
1. How to generate an unique ID ?  
