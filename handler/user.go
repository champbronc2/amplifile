package handler

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/champbronc2/amplifile/amazon"
	"github.com/champbronc2/amplifile/bottlepay"
	"github.com/champbronc2/amplifile/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func (h *Handler) Dashboard(c echo.Context) (err error) {
	var (
		query       = c.Request().URL.Query()
		accessToken = query.Get("token")
	)

	if accessToken != "" {
		// Set cookie; possibly remove logic later
		http.SetCookie(c.Response().Writer, &http.Cookie{
			Name:    "Authorization",
			Value:   accessToken,
			Expires: time.Now().Add(time.Second * 31622400),
		})

		// Make sure user exists in database
	} else {
		// Fetch token from cookie
		cookie, _ := c.Request().Cookie("Authorization")
		accessToken, _ = url.QueryUnescape(cookie.Value)
	}

	userResponse, err := bottlepay.FetchUser(accessToken)
	// Hydrate and save user object
	u := &model.User{
		Name:   userResponse.Name,
		Email:  userResponse.Email,
		Avatar: userResponse.Avatar,
	}

	// // TO-DO: fetch user information from DB
	db := h.DB.Clone()
	defer db.Close()
	if err = db.DB("").C("users").
		Find(bson.M{"email": userResponse.Email}).One(u); err != nil {
		if err == mgo.ErrNotFound {
			return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "something is wrong with your registration"}
		}
		return
	}

	// Fetch AWS pre-signed URL to prepare for upload
	url, key := amazon.GeneratePostSign()

	// Retrieve files from database
	files := []*model.File{}
	if err = db.DB("").C("files").
		Find(bson.D{{"bpay_id", u.BPayID}}).
		Skip((1 - 1) * 1000).
		Sort("-$natural").
		Limit(1000).
		All(&files); err != nil {
		return
	}

	for _, f := range files {
		f.IDtext = f.ID.Hex()
	}

	return c.Render(http.StatusOK, "dashboard.html", map[string]interface{}{
		"user":          u,
		"url":           url,
		"key":           strconv.Itoa(u.BPayID) + "/" + key,
		"authenticated": true,
		"files":         files,
	})
}

// TO-DO: implement
func (h *Handler) UpdateUser(c echo.Context) (err error) {
	// Fetch user from Bottlepay

	// Find user in database
	u := &model.User{}
	db := h.DB.Clone()
	defer db.Close()
	if err = db.DB("").C("users").
		Find(bson.M{"bpay_id": u.BPayID}).One(u); err != nil {
		if err == mgo.ErrNotFound {
			return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "invalid token"}
		}
		return
	}

	// Check for differences

	// Save user if different
	db = h.DB.Clone()
	defer db.Close()
	if err = db.DB("").C("users").Update(bson.M{"bpay_id": u.BPayID}, u); err != nil {
		return
	}

	return c.Redirect(http.StatusMovedPermanently, "/dashboard")
}

func usernameFromToken(c echo.Context) string {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return claims["username"].(string)
}
