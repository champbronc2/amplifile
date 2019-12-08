package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type (
	Download struct {
		ID          bson.ObjectId `json:"id" form:"id" bson:"_id,omitempty"`
		FileID      bson.ObjectId `json:"file_id" form:"file_id" bson:"file_id"`
		Amount      int           `json:"amount" form:"amount" bson:"amount"`
		Invoice     string        `json:"invoice" form:"invoice" bson:"invoice"`
		RHash       string        `json:"r_hash" form:"r_hash" bson:"r_hash"`
		Paid        bool          `json:"paid" form:"paid" bson:"paid"`
		DownloadURL string        `json:"download_url" form:"download_url" bson:"download_url"`
		Created     time.Time     `json:"created" form:"created" bson:"created"`
		Completed   time.Time     `json:"completed" form:"completed" bson:"completed"`
	}
)
