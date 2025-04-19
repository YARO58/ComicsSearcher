package templates

import (
	"html/template"
	"sync"
)

var (
	templates map[string]*template.Template
	mu        sync.RWMutex
)

func Init() {
	templates = make(map[string]*template.Template)

	load("index", "templates/index.html")
	load("login", "templates/login.html")
	load("admin", "templates/admin.html")
}

func load(name, path string) {
	mu.Lock()
	defer mu.Unlock()

	templates[name] = template.Must(template.ParseFiles(path))
}

func Get(name string) *template.Template {
	mu.RLock()
	defer mu.RUnlock()

	return templates[name]
}
