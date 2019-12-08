package handler

import (
	"hash/fnv"
	"net/http"

	"github.com/champbronc2/amplifile/model"
	"github.com/labstack/echo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type (
	Handler struct {
		DB *mgo.Session
	}
)

const (
	// Key (Should come from somewhere else).
	Key = "secret"
)

func (h *Handler) Index(c echo.Context) (err error) {
	var (
		authenticated = false
	)

	_, err = c.Request().Cookie("Authorization")
	if err == nil {
		// accessToken, _ = url.QueryUnescape(cookie.Value)
		authenticated = true
	}

	// Retrieve 10 recent files
	files := []*model.File{}
	db := h.DB.Clone()
	if err = db.DB("").C("files").
		Find(bson.D{{}}).
		Skip(0).
		Sort("-$natural").
		Limit(10).
		All(&files); err != nil {
		return
	}
	defer db.Close()

	for _, f := range files {
		f.IDtext = f.ID.Hex()
	}

	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"files":         files,
		"authenticated": authenticated,
	})
}

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
