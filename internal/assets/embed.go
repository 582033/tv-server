package assets

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed css/* js/*
var StaticFS embed.FS

//go:embed template/*
var TemplateFS embed.FS

func LoadHTMLFromEmbedFS(engine *gin.Engine, embedFS embed.FS, pattern string) {
	root := template.New("")
	tmpl := template.Must(root, loadAndAddToRoot(engine.FuncMap, root, embedFS, pattern))
	engine.SetHTMLTemplate(tmpl)
}

func loadAndAddToRoot(funcMap template.FuncMap, rootTemplate *template.Template, embedFS embed.FS, pattern string) error {
	pattern = strings.ReplaceAll(pattern, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "*", ".*")

	err := fs.WalkDir(embedFS, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			fmt.Println("Walk error:", walkErr)
			return walkErr
		}

		if matched, _ := regexp.MatchString(pattern+"$", path); !d.IsDir() && matched {
			fmt.Println("Matched path:", path)
			data, readErr := embedFS.ReadFile(path)
			if readErr != nil {
				fmt.Println("Read error:", readErr)
				return readErr
			}
			templatePath := filepath.ToSlash(path)
			templatePath = strings.TrimPrefix(templatePath, "static/")
			t := rootTemplate.New(templatePath).Funcs(funcMap)
			if _, parseErr := t.Parse(string(data)); parseErr != nil {
				fmt.Println("Parse error:", parseErr)
				return parseErr
			}
		}
		return nil
	})
	return err
}
