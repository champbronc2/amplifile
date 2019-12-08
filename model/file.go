package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type (
	File struct {
		ID           bson.ObjectId `json:"id" form:"id" bson:"_id,omitempty"`
		IDtext       string
		BPayID       int       `json:"bpay_id" form:"bpay_id" bson:"bpay_id"`
		FileLocation string    `json:"file_location" form:"file_location" bson:"file_location"`
		FileName     string    `json:"file_name" form:"file_name" bson:"file_name"`
		FileType     string    `json:"file_type" form:"file_type" bson:"file_type"`
		FileSize     int       `json:"file_size" form:"file_size" bson:"file_size"`
		Category     string    `json:"category" form:"category" bson:"category"`
		Tags         string    `json:"tags" form:"tags" bson:"tags"`
		Cost         int       `json:"cost" form:"cost" bson:"cost"`
		Created      time.Time `json:"created" bson:"created"`
	}
)
