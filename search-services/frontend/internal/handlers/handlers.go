package handlers

import (
	"fmt"
	"net/http"

	"yadro.com/course/frontend/internal/api"
	"yadro.com/course/frontend/internal/models"
	"yadro.com/course/frontend/internal/templates"
)

type Handler struct {
	apiClient *api.Client
}

func NewHandler() *Handler {
	return &Handler{
		apiClient: api.NewClient(),
	}
}

func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := templates.Get("index")
	if tmpl == nil {
		http.Error(w, "Шаблон не найден", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		Title: "Поиск комиксов XKCD",
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка при отображении шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	query := r.FormValue("query")
	fmt.Printf("query %s\n", query)
	if query == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	limit := r.FormValue("limit")
	if limit == "" {
		limit = "10"
	}

	comics, total, err := h.apiClient.Search(query, limit)
	if err != nil {
		h.renderError(w, "Ошибка при выполнении поиска: "+err.Error(), query, limit)
		return
	}

	tmpl := templates.Get("index")
	if tmpl == nil {
		http.Error(w, "Шаблон не найден", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		Title:  fmt.Sprintf("Результаты поиска: %s (найдено: %d)", query, total),
		Comics: comics,
		Query:  query,
		Limit:  limit,
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка при отображении шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		token, err := h.apiClient.Login(username, password)
		if err != nil {
			tmpl := templates.Get("login")
			if tmpl == nil {
				http.Error(w, "Шаблон не найден", http.StatusInternalServerError)
				return
			}

			data := models.PageData{
				Title: "Вход в систему",
				Error: "Неверные учетные данные или ошибка сервера",
			}
			if err := tmpl.Execute(w, data); err != nil {
				http.Error(w, "Ошибка при отображении шаблона: "+err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
		})

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	tmpl := templates.Get("login")
	if tmpl == nil {
		http.Error(w, "Шаблон не найден", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		Title: "Вход в систему",
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка при отображении шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) AdminHandler(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	stats, err := h.apiClient.GetStats(token.Value)
	if err != nil {
		h.renderError(w, "Ошибка при получении статистики: "+err.Error(), "", "")
		return
	}

	status, err := h.apiClient.GetStatus(token.Value)
	if err != nil {
		h.renderError(w, "Ошибка при получении статуса: "+err.Error(), "", "")
		return
	}

	tmpl := templates.Get("admin")
	if tmpl == nil {
		http.Error(w, "Шаблон не найден", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		Title:    "Панель администратора",
		IsAdmin:  true,
		Username: "admin",
		Stats:    stats,
		Status:   status,
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка при отображении шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	token, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	status, err := h.apiClient.GetStatus(token.Value)
	if err != nil {
		h.renderError(w, "Ошибка при получении статуса базы данных: "+err.Error(), "", "")
		return
	}

	if status == "updating" || status == "running" {
		stats, err := h.apiClient.GetStats(token.Value)
		if err != nil {
			h.renderError(w, "Ошибка при получении статистики: "+err.Error(), "", "")
			return
		}

		tmpl := templates.Get("admin")
		if tmpl == nil {
			http.Error(w, "Шаблон не найден", http.StatusInternalServerError)
			return
		}

		data := models.PageData{
			Title:    "Панель администратора",
			IsAdmin:  true,
			Username: "admin",
			Stats:    stats,
			Status:   status,
		}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Ошибка при отображении шаблона: "+err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	err = h.apiClient.UpdateDB(token.Value)
	if err != nil {
		h.renderError(w, "Ошибка при обновлении базы данных: "+err.Error(), "", "")
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) DropHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	token, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = h.apiClient.DropDB(token.Value)
	if err != nil {
		stats, statsErr := h.apiClient.GetStats(token.Value)
		status, statusErr := h.apiClient.GetStatus(token.Value)

		if statsErr != nil || statusErr != nil {
			h.renderError(w, "Ошибка при очистке базы данных: "+err.Error(), "", "")
			return
		}

		tmpl := templates.Get("admin")
		if tmpl == nil {
			http.Error(w, "Шаблон не найден", http.StatusInternalServerError)
			return
		}

		data := models.PageData{
			Title:    "Панель администратора",
			IsAdmin:  true,
			Username: "admin",
			Stats:    stats,
			Status:   status,
			Error:    "Ошибка при очистке базы данных: " + err.Error(),
		}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Ошибка при отображении шаблона: "+err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) renderError(w http.ResponseWriter, message string, query string, limit string) {
	tmpl := templates.Get("index")
	if tmpl == nil {
		http.Error(w, "Шаблон не найден", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		Title: "Ошибка",
		Error: message,
		Query: query,
		Limit: limit,
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка при отображении шаблона ошибки: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
