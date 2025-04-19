package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"yadro.com/course/frontend/internal/models"
)

type Client struct {
	httpClient *http.Client
	apiAddress string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiAddress: os.Getenv("API_ADDRESS"),
	}
}

func (c *Client) Search(query, limit string) ([]models.Comic, int, error) {
	encodedQuery := template.URLQueryEscaper(query)

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/search?phrase=%s&limit=%s", c.apiAddress, encodedQuery, limit))
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("ошибка при выполнении поиска: код %d", resp.StatusCode)
	}

	var result struct {
		Comics []models.Comic `json:"comics"`
		Total  int            `json:"total"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}

	for i := range result.Comics {
		result.Comics[i].Title = fmt.Sprintf("XKCD #%d", result.Comics[i].ID)
		result.Comics[i].Image = result.Comics[i].URL
		result.Comics[i].PageURL = fmt.Sprintf("https://xkcd.com/%d/", result.Comics[i].ID)
	}

	return result.Comics, result.Total, nil
}

func (c *Client) Login(username, password string) (string, error) {
	loginReq := models.LoginRequest{
		Name:     username,
		Password: password,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Post(fmt.Sprintf("%s/api/login", c.apiAddress), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка аутентификации: код %d", resp.StatusCode)
	}

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(token), nil
}

func (c *Client) GetStats(token string) (*models.Stats, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/db/stats", c.apiAddress), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Token "+token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ошибка при получении статистики: код %d", resp.StatusCode)
	}

	var stats models.Stats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

func (c *Client) GetStatus(token string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/db/status", c.apiAddress), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Token "+token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка при получении статуса: код %d", resp.StatusCode)
	}

	var result struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Status, nil
}

func (c *Client) UpdateDB(token string) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/db/update", c.apiAddress), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Token "+token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("ошибка при обновлении базы данных: код %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) DropDB(token string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/db", c.apiAddress), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Token "+token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ошибка при очистке базы данных: код %d", resp.StatusCode)
	}

	return nil
}
