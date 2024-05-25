package util

import (
  "encoding/json"
  "github.com/deatil/go-cryptobin/cryptobin/crypto"
  gonanoid "github.com/matoous/go-nanoid/v2"
  "imserver/config"
  "imserver/models"
  "time"
)

// 验证头部
func CheckAuthHeader(auth string) (token models.UserToken, err error) {
  cyptde := crypto.FromBase64String(auth).SetKey(config.TOKEN_KEY).Aes().ECB().PKCS7Padding().Decrypt()
  errs := cyptde.Errors
  if len(errs) > 0 {
    return token, errs[0]
  }
  tokeByte := cyptde.ToBytes()
  err = json.Unmarshal(tokeByte, &token)
  if err != nil {
    return
  }
  return
}

// 验证usertoken是否超时
func CheckAuthHeaderIsExpired(tokenTime int64) bool {
  //return time.Now().Unix() < tokenTime
  return time.Now().Unix() > tokenTime
  //return false
}

// 获取唯一连接的TOKEN标识
func GetAuthToken() (token string) {
  token, _ = gonanoid.New(128)
  return
}
