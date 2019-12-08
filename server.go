package main

import (
	"errors"
	"html/template"
	"io"
	"os"

	"github.com/champbronc2/amplifile/handler"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"gopkg.in/mgo.v2"
)

// Implement e.Renderer interface
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		err := errors.New("Template not found -> " + name)
		return err
	}
	return tmpl.ExecuteTemplate(w, "base.html", data)
}

// Define the template registry struct
type TemplateRegistry struct {
	templates map[string]*template.Template
}

func main() {
	e := echo.New()

	templates := make(map[string]*template.Template)
	templates["index.html"] = template.Must(template.ParseFiles("templates/index.html", "templates/base.html"))
	templates["user.html"] = template.Must(template.ParseFiles("templates/user.html", "templates/base.html"))
	templates["dashboard.html"] = template.Must(template.ParseFiles("templates/dashboard.html", "templates/base.html"))
	templates["file.html"] = template.Must(template.ParseFiles("templates/file.html", "templates/base.html"))
	e.Renderer = &TemplateRegistry{
		templates: templates,
	}

	e.Logger.SetLevel(log.ERROR)
	e.Use(middleware.Logger())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	})

	// Database connection
	mongoURL := "localhost"
	if value, ok := os.LookupEnv("MONGODB_URI"); ok {
		mongoURL = value
	}
	db, err := mgo.Dial(mongoURL)
	if err != nil {
		e.Logger.Fatal(err)
	}

	// Create indices
	if err = db.Copy().DB("").C("users").EnsureIndex(mgo.Index{
		Key:    []string{"bpay_id"},
		Unique: true,
	}); err != nil {
		log.Fatal(err)
	}

	// Initialize handler
	h := &handler.Handler{DB: db}

	// Routes
	e.GET("/", h.Index)
	e.GET("/file/:id", h.FetchFile)
	e.GET("/dashboard", h.Dashboard)
	// e.GET("/files", h.ListFiles) -- eventually show all public files
	e.GET("/oauth/redirect", h.OAuthRedirect)

	// e.PUT("/dashboard", h.UpdateUser) -- eventually fetch any updates from BottlePay

	e.POST("/file", h.CreateFile)
	e.POST("/webhooks", h.FileDownloadWebhook)

	e.Static("/static", "static")
	e.File("/favicon.ico", "static/images/favicon.ico")

	// Start server
	port := "1323"
	if value, ok := os.LookupEnv("PORT"); ok {
		port = value
	}
	e.Logger.Fatal(e.Start(":" + port))
}
