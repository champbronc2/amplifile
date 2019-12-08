package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type (
	User struct {
		ID       bson.ObjectId `json:"id" bson:"_id,omitempty"`
		Created  time.Time     `json:"created" bson:"created"`
		Name     string        `json:"name" form:"name" bson:"name"`
		Email    string        `json:"email" form:"email" bson:"email"`
		Password string        `json:"password,omitempty" form:"password" bson:"password"`
		Avatar   string        `json:"avatar, omitempty" form:"avatar" bson:"avatar"`

		Token        string `json:"token,omitempty" bson:"-"`
		AccessToken  string `json:"access_token,omitempty" bson:"access_token"`
		RefreshToken string `json:"refresh_token,omitempty" bson:"refresh_token"`
		BPayID       int    `json:"bpay_id,omitempty" bson:"bpay_id"`
	}
)
