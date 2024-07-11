package http

type RestResult struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ApiCode struct {
	Msg  string `json:"Msg"`
	Code int    `json:"Code"`
}
type ApiRestResult struct {
	Code    ApiCode     `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewApiRestResult(restResult RestResult) ApiRestResult {
	return ApiRestResult{
		Code: ApiCode{
			Msg:  restResult.Message,
			Code: restResult.Code,
		},
		Message: restResult.Message,
		Data:    restResult.Data,
	}
}
