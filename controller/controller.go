package controller

import "imserver/model"

type Controller struct {
	M *model.Model
}

func (c *Controller) GetDescribe(table string) (*[]model.Describe, error) {
	return c.M.GetDescribe(table)
}

func (c *Controller) InsertMessages(data model.ImMessage) (int64, error) {
	return c.M.InsertMessages(data)
}

func (c *Controller) GetMessagesByStartId(data model.MessagesByStartId) (*[]model.ImMessage, error) {
	return c.M.GetMessagesByStartId(data)
}

func (c *Controller) GetAllMessages() (*[]model.ImMessage, error) {
	return c.M.GetAllMessages()
}

func (c *Controller) ClearAllMessages() error {
	return c.M.ClearAllMessages()
}

func (c *Controller) DeleteMessages() {

}

func (c *Controller) SetMsgIdWithTTL(msgId string) (bool, error) {
	return c.M.SetMsgIdWithTTL(msgId)
}

func (c *Controller) SetLastReadId(lastId model.SetLastReadId) error {
	return c.M.SetLastReadId(lastId)
}
func (c *Controller) GetLastReadId() {

}
func (c *Controller) DeleteLastReadId() {

}

func (c *Controller) CheckOrSetFriends(friend model.ImFreindList) error {
	ok, err := c.M.CheckOrSetFriends(friend)
	if err != nil {
		return err
	}
	//已经设置完成了
	if ok {
		//存入内存结构
	}
	return nil
}
func (c *Controller) GetAllFreinds() (*[]model.ImFreindList, error) {
	return c.M.GetAllFreinds()
}
