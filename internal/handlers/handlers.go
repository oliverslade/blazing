package handlers

import (
	"embed"
	"html/template"

	"blazing/internal/app"
)

//go:embed templates/*.html
var templateFS embed.FS

type Handlers struct {
	app               *app.App
	loginTemplate     *template.Template
	dashboardTemplate *template.Template
}

func New(app *app.App) (*Handlers, error) {
	loginTmpl, err := template.New("login").ParseFS(templateFS, "templates/base.html", "templates/login.html")
	if err != nil {
		return nil, err
	}

	dashboardTmpl, err := template.New("dashboard").ParseFS(templateFS, "templates/base.html", "templates/dashboard.html")
	if err != nil {
		return nil, err
	}

	return &Handlers{
		app:               app,
		loginTemplate:     loginTmpl,
		dashboardTemplate: dashboardTmpl,
	}, nil
}
