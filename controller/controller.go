package controller

import "imserver/model"

type Controller struct {
	M *model.Model
}

func (c *Controller) GetDescribe(table string) (*[]model.Describe, error) {
	res, err := c.M.GetDescribe(table)
	return res, err
}

func (c *Controller) InsertMessages() {

}
func (c *Controller) GetMessages() {

}
func (c *Controller) DeleteMessages() {

}

func (c *Controller) InsertLastReadId() {

}
func (c *Controller) GetLastReadId() {

}
func (c *Controller) DeleteLastReadId() {

}
