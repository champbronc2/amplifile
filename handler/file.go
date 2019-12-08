package handler

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/champbronc2/amplifile/amazon"
	"github.com/champbronc2/amplifile/bottlepay"
	"github.com/champbronc2/amplifile/model"
	"github.com/labstack/echo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func (h *Handler) CreateFile(c echo.Context) (err error) {
	// u := &model.User{}
	f := &model.File{
		ID:      bson.NewObjectId(),
		Created: time.Now(),
	}
	// Bind remaing payload
	if err = c.Bind(f); err != nil {
		return
	}

	// Validation
	if f.FileLocation == "" {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "invalid to or message fields"}
	}

	// Find user from database
	db := h.DB.Clone()
	defer db.Close()
	/*
		if err = db.DB("").C("users").Find(bson.D{{"username", p.EmailHash}}).One(u); err != nil {
			if err == mgo.ErrNotFound {
				return echo.ErrNotFound
			}
			return
		}*/

	// Save file in database
	if err = db.DB("").C("files").Insert(f); err != nil {
		return
	}

	return c.Redirect(http.StatusMovedPermanently, "/file/"+f.ID.Hex())
}

func (h *Handler) FetchFile(c echo.Context) (err error) {
	var (
		authenticated  = false
		invoiceCreated = false
		paymentRequest string
		d              = &model.Download{}
		id             = bson.ObjectIdHex(c.Param("id"))
	)

	_, err = c.Request().Cookie("Authorization")
	if err == nil {
		// accessToken, _ = url.QueryUnescape(cookie.Value)
		authenticated = true
	}

	// Retrieve file from database
	files := []*model.File{}
	db := h.DB.Clone()
	if err = db.DB("").C("files").
		Find(bson.M{"_id": id}).
		Skip(0).
		Limit(100).
		All(&files); err != nil {
		return
	}
	defer db.Close()

	f := files[0]

	cookie, err := c.Request().Cookie("File_" + f.ID.Hex())
	if err == nil {
		paymentRequest, _ = url.QueryUnescape(cookie.Value)
		invoiceCreated = true
	}

	/*
		If cookie for file ID + invoice present {
			if file is paid {
				show download URL that is tied to invoice
			} else {
				show unpaid invoice QR code
			}
		} else {
			fetch, store and display invoice
		}
	*/

	if invoiceCreated {
		// Fetch invoice from Cookie
		if err = db.DB("").C("downloads").
			Find(bson.M{"invoice": paymentRequest}).One(d); err != nil {
			if err == mgo.ErrNotFound {
				return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "could not find uploader by file"}
			}
			return
		}
	} else {

		// Fetch token from user who uploaded the file originally
		u := &model.User{}
		if err = db.DB("").C("users").
			Find(bson.M{"bpay_id": f.BPayID}).One(u); err != nil {
			if err == mgo.ErrNotFound {
				return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "could not find uploader by file"}
			}
			return
		}

		// Fetch invoice from Bottlepay
		invoice, err := bottlepay.FetchUserInvoice(u.AccessToken, f.Cost, "Download "+f.FileName)
		if err != nil {
			log.Println(err)
			return err
		}

		// TO-DO: Store invoice in DB, and create cookie for user
		// Step 3: Hydrate and save user object
		d = &model.Download{
			ID:          bson.NewObjectId(),
			FileID:      f.ID,
			Amount:      f.Cost,
			Invoice:     invoice.PaymentRequest,
			RHash:       invoice.RHash,
			Paid:        false,
			DownloadURL: "http://not_paid_yet",
			Created:     time.Now(),
		}

		// Save download + create cookie
		if err = db.DB("").C("downloads").Insert(d); err != nil {
			log.Println(err)
			return err
		}
		http.SetCookie(c.Response().Writer, &http.Cookie{
			Name:    "File_" + f.ID.Hex(),
			Value:   invoice.PaymentRequest,
			Expires: time.Now().Add(time.Second * 36000),
		})
	}

	return c.Render(http.StatusCreated, "file.html", map[string]interface{}{
		"id":             f.ID.Hex(),
		"fileName":       f.FileName,
		"fileType":       f.FileType,
		"fileSize":       f.FileSize,
		"category":       f.Category,
		"tags":           f.Tags,
		"cost":           f.Cost,
		"created":        f.Created,
		"paymentRequest": d.Invoice,
		"authenticated":  authenticated,
		"paid":           d.Paid,
		"downloadURL":    d.DownloadURL,
	})
}

func (h *Handler) FileDownloadWebhook(c echo.Context) (err error) {
	var (
		request = &bottlepay.WebhookRequest{}
		d       = &model.Download{}
	)
	// Bind remaing payload
	if err = c.Bind(request); err != nil {
		return
	}

	// Locate download by r_hash
	db := h.DB.Clone()
	if err = db.DB("").C("downloads").
		Find(bson.M{"r_hash": request.RHash}).One(d); err != nil {
		if err == mgo.ErrNotFound {
			return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "could not find file by r_hash"}
		}
		return
	}
	defer db.Close()

	// Select file
	f := &model.File{}
	if err = db.DB("").C("files").
		Find(bson.M{"_id": d.FileID}).One(f); err != nil {
		if err == mgo.ErrNotFound {
			return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "could not find file by download"}
		}
		return
	}

	// Generate pre-signed URL from AWS
	downloadURL, err := amazon.GenerateGetSign(f.FileLocation)
	if err != nil {
		log.Println(err)
		return err
	}

	// Update Download URL and mark paid
	d.DownloadURL = downloadURL
	d.Paid = true
	if err = db.DB("").C("downloads").Update(bson.M{"_id": d.ID}, d); err != nil {
		return
	}

	return c.JSON(http.StatusOK, d)
}
