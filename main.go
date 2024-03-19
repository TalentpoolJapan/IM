package main

import (
	"fmt"
	"imserver/manticore"
	"imserver/rosedb"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	manticore := manticore.NewManticore(manticore.ManticoreMysqlConfig{
		Host: "127.0.0.1",
		Port: 9306,
	})
	// res, _ := manticore.DB.Query("describe im_message")
	// fmt.Println(res)
	res, err := manticore.Describe("im_message")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res)

	kv, err := rosedb.NewRoseDB(rosedb.RoseDBConfig{
		DBPath: "IMKVDB",
	})
	if err != nil {
		fmt.Println(err.Error())
	}
	startT := time.Now()
	err = kv.Set("aaa", "bbb")
	if err != nil {
		fmt.Println(err.Error())
	}

	v, err := kv.Get("aaa")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(v)
	fmt.Println("--------------")

	ok, err := kv.SetIfNotExist("eee", "ddd")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(ok)
	fmt.Println("--------------")
	v, err = kv.Get("eee")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(v)
	tc := time.Since(startT) //计算耗时
	fmt.Printf("time cost = %v\n", tc)

	// db, err := xorm.NewEngine("mysql", "``:``@tcp(127.0.0.1:9306)/Manticore")
	// fmt.Println(err)
	// res, _ := db.Query(`describe im_message`)
	// fmt.Println(res)
	// body := `SHOW TABLES`
	// // body := `CREATE TABLE im_message (
	// // 	id bigint,
	// // 	touser text,
	// // 	fromuser text,
	// // 	msg text,
	// // 	msgtype integer,
	// // 	totype integer,
	// // 	fromtype integer,
	// // 	created bigint,
	// // 	msgid string attribute
	// // 	) ngram_len='1' ngram_chars='cjk'`
	// // string | A query parameter string.
	// rawResponse := true // bool | Optional parameter, defines a format of response. Can be set to `False` for Select only queries and set to `True` or omitted for any type of queries:  (optional) (default to true)

	// configuration := openapiclient.NewConfiguration()
	// apiClient := openapiclient.NewAPIClient(configuration)
	// resp, r, err := apiClient.UtilsAPI.Sql(context.Background()).Body(body).RawResponse(rawResponse).Execute()
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error when calling `UtilsAPI.Sql``: %v\n", err)
	// 	fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	// }
	// // response from `Sql`: []map[string]interface{}
	// fmt.Fprintf(os.Stdout, "Response from `UtilsAPI.Sql`: %v\n", resp)
}
