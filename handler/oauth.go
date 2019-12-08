package handler

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/champbronc2/amplifile/bottlepay"
	"github.com/champbronc2/amplifile/model"
	"github.com/labstack/echo"
	"gopkg.in/mgo.v2/bson"
)

func (h *Handler) OAuthRedirect(c echo.Context) (err error) {
	var (
		query     = c.Request().URL.Query()
		oauthCode = query.Get("code")
	)

	// Step 1: Fetch AuthResponse from BottlePay
	if oauthCode == "" {
		return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "invalid oauth code"}
	}

	authResponse, err := bottlepay.FetchAccessToken(oauthCode)

	// Step 2: Fetch user information
	userResponse, err := bottlepay.FetchUser(authResponse.AccessToken)

	// Step 3: Hydrate and save user object
	u := &model.User{
		ID:       bson.NewObjectId(),
		Name:     userResponse.Name,
		Email:    userResponse.Email,
		Password: "not_implemented",
		Avatar:   userResponse.Avatar,

		RefreshToken: authResponse.RefreshToken,
		AccessToken:  authResponse.AccessToken,
		BPayID:       userResponse.ID,
		Created:      time.Now(),
	}

	// Save user
	db := h.DB.Clone()
	defer db.Close()
	if err = db.DB("").C("users").Insert(u); err != nil {
		log.Println(err)
		// If already signed up, no worries
		if !strings.Contains(err.Error(), "E11000") {
			return
		}
	}

	http.SetCookie(c.Response().Writer, &http.Cookie{
		Name:    "Authorization",
		Value:   "Bearer " + authResponse.AccessToken,
		Expires: time.Now().Add(time.Second * time.Duration(authResponse.ExpiresIn)),
	})

	return c.Redirect(http.StatusMovedPermanently, "/dashboard?token="+authResponse.AccessToken)
}
