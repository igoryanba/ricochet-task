package workflow

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GitWebhookHandler обработчик Git webhooks
type GitWebhookHandler struct {
	tracker    *GitProgressTracker
	logger     Logger
	secretKey  string
	server     *http.Server
	processors map[string]WebhookProcessor
}

// WebhookProcessor интерфейс для обработки webhook от разных Git провайдеров
type WebhookProcessor interface {
	ProcessWebhook(body []byte, headers map[string]string) (*GitProgressEvent, error)
	ValidateSignature(body []byte, signature string) bool
	GetProviderName() string
}

// NewGitWebhookHandler создает новый обработчик webhooks
func NewGitWebhookHandler(tracker *GitProgressTracker, logger Logger) *GitWebhookHandler {
	handler := &GitWebhookHandler{
		tracker:    tracker,
		logger:     logger,
		secretKey:  "", // TODO: получать из конфигурации
		processors: make(map[string]WebhookProcessor),
	}
	
	// Регистрируем процессоры для разных провайдеров
	handler.RegisterProcessor(&GitHubWebhookProcessor{secretKey: handler.secretKey})
	handler.RegisterProcessor(&GitLabWebhookProcessor{secretKey: handler.secretKey})
	handler.RegisterProcessor(&BitbucketWebhookProcessor{secretKey: handler.secretKey})
	
	return handler
}

// RegisterProcessor регистрирует процессор webhook
func (gwh *GitWebhookHandler) RegisterProcessor(processor WebhookProcessor) {
	gwh.processors[processor.GetProviderName()] = processor
	gwh.logger.Info("Registered webhook processor", "provider", processor.GetProviderName())
}

// StartServer запускает HTTP сервер для приема webhooks
func (gwh *GitWebhookHandler) StartServer(port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", gwh.handleWebhook)
	mux.HandleFunc("/webhook/github", gwh.handleGitHubWebhook)
	mux.HandleFunc("/webhook/gitlab", gwh.handleGitLabWebhook)
	mux.HandleFunc("/webhook/bitbucket", gwh.handleBitbucketWebhook)
	mux.HandleFunc("/health", gwh.handleHealth)
	
	gwh.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	
	gwh.logger.Info("Starting webhook server", "port", port)
	return gwh.server.ListenAndServe()
}

// StopServer останавливает HTTP сервер
func (gwh *GitWebhookHandler) StopServer(ctx context.Context) error {
	if gwh.server != nil {
		return gwh.server.Shutdown(ctx)
	}
	return nil
}

// handleWebhook универсальный обработчик webhook
func (gwh *GitWebhookHandler) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	body, err := io.ReadAll(r.Body)
	if err != nil {
		gwh.logger.Error("Failed to read webhook body", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	// Извлекаем заголовки
	headers := make(map[string]string)
	for name, values := range r.Header {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}
	
	// Определяем провайдера по заголовкам
	provider := gwh.detectProvider(headers)
	if provider == "" {
		gwh.logger.Warn("Unknown webhook provider", "headers", headers)
		http.Error(w, "Unknown provider", http.StatusBadRequest)
		return
	}
	
	// Обрабатываем webhook
	processor, exists := gwh.processors[provider]
	if !exists {
		gwh.logger.Error("No processor for provider", nil, "provider", provider)
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}
	
	event, err := processor.ProcessWebhook(body, headers)
	if err != nil {
		gwh.logger.Error("Failed to process webhook", err, "provider", provider)
		http.Error(w, "Processing failed", http.StatusInternalServerError)
		return
	}
	
	// Передаем событие в tracker
	if err := gwh.tracker.ProcessGitEvent(r.Context(), event); err != nil {
		gwh.logger.Error("Failed to process git event", err)
		http.Error(w, "Event processing failed", http.StatusInternalServerError)
		return
	}
	
	gwh.logger.Info("Webhook processed successfully", "provider", provider, "event_type", event.Type)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// detectProvider определяет Git провайдера по заголовкам
func (gwh *GitWebhookHandler) detectProvider(headers map[string]string) string {
	if headers["X-GitHub-Event"] != "" {
		return "github"
	}
	if headers["X-Gitlab-Event"] != "" {
		return "gitlab"
	}
	if headers["X-Event-Key"] != "" {
		return "bitbucket"
	}
	return ""
}

// Специализированные обработчики

func (gwh *GitWebhookHandler) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	gwh.handleProviderWebhook(w, r, "github")
}

func (gwh *GitWebhookHandler) handleGitLabWebhook(w http.ResponseWriter, r *http.Request) {
	gwh.handleProviderWebhook(w, r, "gitlab")
}

func (gwh *GitWebhookHandler) handleBitbucketWebhook(w http.ResponseWriter, r *http.Request) {
	gwh.handleProviderWebhook(w, r, "bitbucket")
}

func (gwh *GitWebhookHandler) handleProviderWebhook(w http.ResponseWriter, r *http.Request, provider string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	body, err := io.ReadAll(r.Body)
	if err != nil {
		gwh.logger.Error("Failed to read webhook body", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	// Извлекаем заголовки
	headers := make(map[string]string)
	for name, values := range r.Header {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}
	
	processor, exists := gwh.processors[provider]
	if !exists {
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}
	
	event, err := processor.ProcessWebhook(body, headers)
	if err != nil {
		gwh.logger.Error("Failed to process webhook", err, "provider", provider)
		http.Error(w, "Processing failed", http.StatusInternalServerError)
		return
	}
	
	if err := gwh.tracker.ProcessGitEvent(r.Context(), event); err != nil {
		gwh.logger.Error("Failed to process git event", err)
		http.Error(w, "Event processing failed", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (gwh *GitWebhookHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"service": "git-webhook-handler",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// GitHubWebhookProcessor процессор для GitHub webhooks
type GitHubWebhookProcessor struct {
	secretKey string
}

func (p *GitHubWebhookProcessor) GetProviderName() string {
	return "github"
}

func (p *GitHubWebhookProcessor) ValidateSignature(body []byte, signature string) bool {
	if p.secretKey == "" {
		return true // Пропускаем валидацию если секрет не установлен
	}
	
	mac := hmac.New(sha256.New, []byte(p.secretKey))
	mac.Write(body)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func (p *GitHubWebhookProcessor) ProcessWebhook(body []byte, headers map[string]string) (*GitProgressEvent, error) {
	eventType := headers["X-GitHub-Event"]
	
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub webhook payload: %w", err)
	}
	
	switch eventType {
	case "push":
		return p.processPushEvent(payload)
	case "pull_request":
		return p.processPullRequestEvent(payload)
	default:
		return nil, fmt.Errorf("unsupported GitHub event type: %s", eventType)
	}
}

func (p *GitHubWebhookProcessor) processPushEvent(payload map[string]interface{}) (*GitProgressEvent, error) {
	repository := p.getString(payload, "repository.full_name")
	ref := p.getString(payload, "ref")
	branch := strings.TrimPrefix(ref, "refs/heads/")
	
	// Обрабатываем коммиты
	commits, ok := payload["commits"].([]interface{})
	if !ok || len(commits) == 0 {
		return nil, fmt.Errorf("no commits in push event")
	}
	
	// Берем последний коммит
	lastCommit := commits[len(commits)-1].(map[string]interface{})
	
	commit := &CommitInfo{
		SHA:       p.getString(lastCommit, "id"),
		Message:   p.getString(lastCommit, "message"),
		Author:    p.getString(lastCommit, "author.username"),
		Timestamp: p.parseTimestamp(p.getString(lastCommit, "timestamp")),
	}
	
	// Извлекаем измененные файлы
	if added, ok := lastCommit["added"].([]interface{}); ok {
		for _, file := range added {
			commit.FilesChanged = append(commit.FilesChanged, file.(string))
		}
	}
	if modified, ok := lastCommit["modified"].([]interface{}); ok {
		for _, file := range modified {
			commit.FilesChanged = append(commit.FilesChanged, file.(string))
		}
	}
	
	return &GitProgressEvent{
		Type:       "push",
		Repository: repository,
		Branch:     branch,
		Commit:     commit,
		Data:       payload,
		Timestamp:  time.Now(),
	}, nil
}

func (p *GitHubWebhookProcessor) processPullRequestEvent(payload map[string]interface{}) (*GitProgressEvent, error) {
	repository := p.getString(payload, "repository.full_name")
	action := p.getString(payload, "action")
	
	prData, ok := payload["pull_request"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no pull_request data in payload")
	}
	
	pr := &PullRequestInfo{
		ID:           int(p.getFloat64(prData, "number")),
		Title:        p.getString(prData, "title"),
		State:        action,
		Author:       p.getString(prData, "user.login"),
		SourceBranch: p.getString(prData, "head.ref"),
		TargetBranch: p.getString(prData, "base.ref"),
		CreatedAt:    p.parseTimestamp(p.getString(prData, "created_at")),
		UpdatedAt:    p.parseTimestamp(p.getString(prData, "updated_at")),
	}
	
	return &GitProgressEvent{
		Type:        "pull_request",
		Repository:  repository,
		Branch:      pr.SourceBranch,
		PullRequest: pr,
		Data:        payload,
		Timestamp:   time.Now(),
	}, nil
}

// GitLabWebhookProcessor процессор для GitLab webhooks
type GitLabWebhookProcessor struct {
	secretKey string
}

func (p *GitLabWebhookProcessor) GetProviderName() string {
	return "gitlab"
}

func (p *GitLabWebhookProcessor) ValidateSignature(body []byte, signature string) bool {
	// GitLab использует X-Gitlab-Token заголовок
	return p.secretKey == "" || signature == p.secretKey
}

func (p *GitLabWebhookProcessor) ProcessWebhook(body []byte, headers map[string]string) (*GitProgressEvent, error) {
	eventType := headers["X-Gitlab-Event"]
	
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse GitLab webhook payload: %w", err)
	}
	
	switch eventType {
	case "Push Hook":
		return p.processPushEvent(payload)
	case "Merge Request Hook":
		return p.processMergeRequestEvent(payload)
	default:
		return nil, fmt.Errorf("unsupported GitLab event type: %s", eventType)
	}
}

func (p *GitLabWebhookProcessor) processPushEvent(payload map[string]interface{}) (*GitProgressEvent, error) {
	repository := p.getString(payload, "project.path_with_namespace")
	ref := p.getString(payload, "ref")
	branch := strings.TrimPrefix(ref, "refs/heads/")
	
	commits, ok := payload["commits"].([]interface{})
	if !ok || len(commits) == 0 {
		return nil, fmt.Errorf("no commits in GitLab push event")
	}
	
	lastCommit := commits[len(commits)-1].(map[string]interface{})
	
	commit := &CommitInfo{
		SHA:       p.getString(lastCommit, "id"),
		Message:   p.getString(lastCommit, "message"),
		Author:    p.getString(lastCommit, "author.username"),
		Timestamp: p.parseTimestamp(p.getString(lastCommit, "timestamp")),
	}
	
	return &GitProgressEvent{
		Type:       "push",
		Repository: repository,
		Branch:     branch,
		Commit:     commit,
		Data:       payload,
		Timestamp:  time.Now(),
	}, nil
}

func (p *GitLabWebhookProcessor) processMergeRequestEvent(payload map[string]interface{}) (*GitProgressEvent, error) {
	repository := p.getString(payload, "project.path_with_namespace")
	
	mrData, ok := payload["object_attributes"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no object_attributes in GitLab MR payload")
	}
	
	pr := &PullRequestInfo{
		ID:           int(p.getFloat64(mrData, "iid")),
		Title:        p.getString(mrData, "title"),
		State:        p.getString(mrData, "action"),
		Author:       p.getString(payload, "user.username"),
		SourceBranch: p.getString(mrData, "source_branch"),
		TargetBranch: p.getString(mrData, "target_branch"),
		CreatedAt:    p.parseTimestamp(p.getString(mrData, "created_at")),
		UpdatedAt:    p.parseTimestamp(p.getString(mrData, "updated_at")),
	}
	
	return &GitProgressEvent{
		Type:        "pull_request",
		Repository:  repository,
		Branch:      pr.SourceBranch,
		PullRequest: pr,
		Data:        payload,
		Timestamp:   time.Now(),
	}, nil
}

// BitbucketWebhookProcessor процессор для Bitbucket webhooks
type BitbucketWebhookProcessor struct {
	secretKey string
}

func (p *BitbucketWebhookProcessor) GetProviderName() string {
	return "bitbucket"
}

func (p *BitbucketWebhookProcessor) ValidateSignature(body []byte, signature string) bool {
	// Bitbucket может использовать разные методы подписи
	return true // Упрощенная валидация
}

func (p *BitbucketWebhookProcessor) ProcessWebhook(body []byte, headers map[string]string) (*GitProgressEvent, error) {
	eventType := headers["X-Event-Key"]
	
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse Bitbucket webhook payload: %w", err)
	}
	
	switch eventType {
	case "repo:push":
		return p.processPushEvent(payload)
	case "pullrequest:created", "pullrequest:updated", "pullrequest:fulfilled":
		return p.processPullRequestEvent(payload)
	default:
		return nil, fmt.Errorf("unsupported Bitbucket event type: %s", eventType)
	}
}

func (p *BitbucketWebhookProcessor) processPushEvent(payload map[string]interface{}) (*GitProgressEvent, error) {
	repository := p.getString(payload, "repository.full_name")
	
	changes, ok := payload["push"].(map[string]interface{})["changes"].([]interface{})
	if !ok || len(changes) == 0 {
		return nil, fmt.Errorf("no changes in Bitbucket push event")
	}
	
	change := changes[0].(map[string]interface{})
	branch := p.getString(change, "new.name")
	
	// Создаем базовую информацию о коммите
	commit := &CommitInfo{
		SHA:       p.getString(change, "new.target.hash"),
		Message:   p.getString(change, "new.target.message"),
		Author:    p.getString(change, "new.target.author.raw"),
		Timestamp: p.parseTimestamp(p.getString(change, "new.target.date")),
	}
	
	return &GitProgressEvent{
		Type:       "push",
		Repository: repository,
		Branch:     branch,
		Commit:     commit,
		Data:       payload,
		Timestamp:  time.Now(),
	}, nil
}

func (p *BitbucketWebhookProcessor) processPullRequestEvent(payload map[string]interface{}) (*GitProgressEvent, error) {
	repository := p.getString(payload, "repository.full_name")
	
	prData, ok := payload["pullrequest"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no pullrequest data in Bitbucket payload")
	}
	
	state := p.getString(prData, "state")
	
	pr := &PullRequestInfo{
		ID:           int(p.getFloat64(prData, "id")),
		Title:        p.getString(prData, "title"),
		State:        state,
		Author:       p.getString(prData, "author.username"),
		SourceBranch: p.getString(prData, "source.branch.name"),
		TargetBranch: p.getString(prData, "destination.branch.name"),
		CreatedAt:    p.parseTimestamp(p.getString(prData, "created_on")),
		UpdatedAt:    p.parseTimestamp(p.getString(prData, "updated_on")),
	}
	
	return &GitProgressEvent{
		Type:        "pull_request",
		Repository:  repository,
		Branch:      pr.SourceBranch,
		PullRequest: pr,
		Data:        payload,
		Timestamp:   time.Now(),
	}, nil
}

// Утилиты для извлечения данных из payload

func (p *GitHubWebhookProcessor) getString(data map[string]interface{}, path string) string {
	return getString(data, path)
}

func (p *GitHubWebhookProcessor) getFloat64(data map[string]interface{}, path string) float64 {
	return getFloat64(data, path)
}

func (p *GitHubWebhookProcessor) parseTimestamp(timeStr string) time.Time {
	return parseTimestamp(timeStr)
}

func (p *GitLabWebhookProcessor) getString(data map[string]interface{}, path string) string {
	return getString(data, path)
}

func (p *GitLabWebhookProcessor) getFloat64(data map[string]interface{}, path string) float64 {
	return getFloat64(data, path)
}

func (p *GitLabWebhookProcessor) parseTimestamp(timeStr string) time.Time {
	return parseTimestamp(timeStr)
}

func (p *BitbucketWebhookProcessor) getString(data map[string]interface{}, path string) string {
	return getString(data, path)
}

func (p *BitbucketWebhookProcessor) getFloat64(data map[string]interface{}, path string) float64 {
	return getFloat64(data, path)
}

func (p *BitbucketWebhookProcessor) parseTimestamp(timeStr string) time.Time {
	return parseTimestamp(timeStr)
}

// Общие утилиты

func getString(data map[string]interface{}, path string) string {
	parts := strings.Split(path, ".")
	current := data
	
	for i, part := range parts {
		if i == len(parts)-1 {
			if val, ok := current[part].(string); ok {
				return val
			}
			return ""
		}
		
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return ""
		}
	}
	
	return ""
}

func getFloat64(data map[string]interface{}, path string) float64 {
	parts := strings.Split(path, ".")
	current := data
	
	for i, part := range parts {
		if i == len(parts)-1 {
			if val, ok := current[part].(float64); ok {
				return val
			}
			return 0
		}
		
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return 0
		}
	}
	
	return 0
}

func parseTimestamp(timeStr string) time.Time {
	if timeStr == "" {
		return time.Now()
	}
	
	// Пробуем разные форматы
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t
		}
	}
	
	return time.Now()
}