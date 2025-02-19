package templates

import (
	"html/template"
	"time"
)

var FuncMap = template.FuncMap{
	"formatTime": func(timestamp int64) string {
		t := time.Unix(timestamp, 0)
		return t.Format("2006-01-02 15:04:05")
	},
}
