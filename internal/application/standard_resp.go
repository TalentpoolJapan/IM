package application

type SingleResp[T any] struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Data    T      `json:"data"`
}

type MultiResp[T any] struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Data    []*T   `json:"data"`
}

func SingleRespOf[T any](data T, msg string) SingleResp[T] {
	return SingleResp[T]{
		Success: true,
		Msg:     msg,
		Data:    data,
	}
}

func SingleRespFail[T any](msg string) SingleResp[T] {
	var emptyData T
	return SingleResp[T]{
		Success: false,
		Msg:     msg,
		Data:    emptyData,
	}
}

func MultiRespOf[T any](data []*T, msg string) MultiResp[T] {
	return MultiResp[T]{
		Success: true,
		Msg:     msg,
		Data:    data,
	}
}

func MultiRespFail[T any](msg string) MultiResp[T] {
	return MultiResp[T]{
		Success: false,
		Msg:     msg,
		Data:    nil,
	}
}
