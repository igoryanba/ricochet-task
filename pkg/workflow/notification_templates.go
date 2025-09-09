package workflow

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/ai"
)

// NotificationTemplates система шаблонов уведомлений
type NotificationTemplates struct {
	templates map[string]*template.Template
	functions template.FuncMap
}

// NewNotificationTemplates создает новую систему шаблонов
func NewNotificationTemplates() *NotificationTemplates {
	nt := &NotificationTemplates{
		templates: make(map[string]*template.Template),
		functions: template.FuncMap{
			"formatTime":     formatTime,
			"timeAgo":        timeAgo,
			"titleCase":      strings.Title,
			"upper":          strings.ToUpper,
			"lower":          strings.ToLower,
			"join":           strings.Join,
			"pluralize":      pluralize,
			"truncate":       truncate,
			"highlightCode":  highlightCode,
			"markdownToHTML": markdownToHTML,
		},
	}
	
	// Загружаем предустановленные шаблоны
	nt.loadDefaultTemplates()
	
	return nt
}

// loadDefaultTemplates загружает шаблоны по умолчанию
func (nt *NotificationTemplates) loadDefaultTemplates() {
	// Шаблоны для задач
	nt.RegisterTemplate("task_created_title", "New Task: {{.title}}")
	nt.RegisterTemplate("task_created_body", `A new task "{{.title}}" has been created.
	
Priority: {{.priority | titleCase}}
Assignee: {{.assignee}}
Due Date: {{.due_date | formatTime}}

{{if .description}}Description:
{{.description}}{{end}}`)

	nt.RegisterTemplate("task_assigned_title", "Task Assigned: {{.title}}")
	nt.RegisterTemplate("task_assigned_body", `You have been assigned to task "{{.title}}".

Priority: {{.priority | titleCase}}
Due Date: {{.due_date | formatTime}}
Project: {{.project}}

{{if .description}}Description:
{{.description}}{{end}}

Please review and start working on this task.`)

	nt.RegisterTemplate("task_completed_title", "Task Completed: {{.title}}")
	nt.RegisterTemplate("task_completed_body", `Task "{{.title}}" has been completed by {{.assignee}}.

Completion Time: {{.completed_at | formatTime}}
Duration: {{.duration}}
Quality Score: {{.quality_score}}

{{if .notes}}Completion Notes:
{{.notes}}{{end}}`)

	// Шаблоны для Git событий
	nt.RegisterTemplate("git_push_title", "{{.commits | len}} new {{.commits | len | pluralize \"commit\" \"commits\"}} in {{.repository}}")
	nt.RegisterTemplate("git_push_body", `{{.author}} pushed {{.commits | len}} {{.commits | len | pluralize "commit" "commits"}} to {{.branch}} in {{.repository}}.

Recent commits:
{{range .commits}}• {{.message | truncate 100}}
{{end}}

Files changed: {{.files_changed | join ", " | truncate 200}}`)

	nt.RegisterTemplate("pull_request_title", "Pull Request {{.action | titleCase}}: {{.title}}")
	nt.RegisterTemplate("pull_request_body", `Pull request "{{.title}}" has been {{.action}}.

Author: {{.author}}
Branch: {{.source_branch}} → {{.target_branch}}
{{if .reviewers}}Reviewers: {{.reviewers | join ", "}}{{end}}

{{if .description}}{{.description | truncate 300}}{{end}}`)

	// Шаблоны для workflow
	nt.RegisterTemplate("workflow_stage_change_title", "Stage Changed: {{.task_title}} → {{.new_stage | titleCase}}")
	nt.RegisterTemplate("workflow_stage_change_body", `Task "{{.task_title}}" has moved to {{.new_stage | titleCase}} stage.

Previous Stage: {{.old_stage | titleCase}}
Progress: {{.progress}}%
Assignee: {{.assignee}}

{{if .next_actions}}Next Actions:
{{range .next_actions}}• {{.}}
{{end}}{{end}}`)

	// Шаблоны для уведомлений о прогрессе
	nt.RegisterTemplate("progress_milestone_title", "Milestone Reached: {{.milestone}} for {{.task_title}}")
	nt.RegisterTemplate("progress_milestone_body", `Great progress! Task "{{.task_title}}" has reached {{.milestone}}.

Progress: {{.progress}}%
Time Spent: {{.time_spent}}
Velocity: {{.velocity}}

{{if .blockers}}Current Blockers:
{{range .blockers}}• {{.}}
{{end}}{{end}}`)

	// Шаблоны для команды
	nt.RegisterTemplate("team_daily_summary_title", "Daily Team Summary - {{.date | formatTime}}")
	nt.RegisterTemplate("team_daily_summary_body", `Here's your team's progress for {{.date | formatTime}}:

📊 Overview:
• Completed Tasks: {{.completed_tasks}}
• Active Tasks: {{.active_tasks}}
• Team Velocity: {{.velocity}}

🏆 Top Performers:
{{range .top_performers}}• {{.name}} - {{.score}} points
{{end}}

⚠️ Attention Needed:
{{range .attention_needed}}• {{.task}} - {{.reason}}
{{end}}`)

	// Шаблоны для AI анализа
	nt.RegisterTemplate("ai_insight_title", "AI Insight: {{.type | titleCase}}")
	nt.RegisterTemplate("ai_insight_body", `AI Analysis has identified an important insight:

{{.insight}}

Confidence: {{.confidence}}%
Impact: {{.impact | titleCase}}

{{if .recommendations}}Recommendations:
{{range .recommendations}}• {{.}}
{{end}}{{end}}`)

	// Шаблоны для экстренных уведомлений
	nt.RegisterTemplate("critical_alert_title", "🚨 CRITICAL: {{.alert_type | titleCase}}")
	nt.RegisterTemplate("critical_alert_body", `CRITICAL ALERT: {{.alert_type | titleCase}}

{{.description}}

Affected: {{.affected}}
Started: {{.started_at | formatTime}}
Severity: {{.severity | upper}}

{{if .action_required}}IMMEDIATE ACTION REQUIRED:
{{.action_required}}{{end}}

{{if .contact}}Emergency Contact: {{.contact}}{{end}}`)

	// Персонализированные шаблоны
	nt.RegisterTemplate("personalized_daily_title", "Your Daily Update, {{.user_name}}")
	nt.RegisterTemplate("personalized_daily_body", `Good {{.time_of_day}}, {{.user_name}}!

Here's your personalized update:

🎯 Your Tasks ({{.user_tasks | len}}):
{{range .user_tasks}}• {{.title}} - {{.status}} ({{.progress}}%)
{{end}}

📈 Your Progress:
• Velocity: {{.user_velocity}}
• Quality Score: {{.quality_score}}
• Completed This Week: {{.completed_this_week}}

{{if .user_insights}}💡 AI Insights for You:
{{range .user_insights}}• {{.}}
{{end}}{{end}}

{{if .suggested_actions}}📋 Suggested Actions:
{{range .suggested_actions}}• {{.}}
{{end}}{{end}}`)
}

// RegisterTemplate регистрирует новый шаблон
func (nt *NotificationTemplates) RegisterTemplate(name, templateStr string) error {
	tmpl, err := template.New(name).Funcs(nt.functions).Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}
	
	nt.templates[name] = tmpl
	return nil
}

// RenderTemplate рендерит шаблон с данными
func (nt *NotificationTemplates) RenderTemplate(name string, data map[string]interface{}) string {
	tmpl, exists := nt.templates[name]
	if !exists {
		return fmt.Sprintf("Template %s not found", name)
	}
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("Error rendering template %s: %v", name, err)
	}
	
	return buf.String()
}

// GetAvailableTemplates возвращает список доступных шаблонов
func (nt *NotificationTemplates) GetAvailableTemplates() []string {
	var names []string
	for name := range nt.templates {
		names = append(names, name)
	}
	return names
}

// ValidateTemplate проверяет валидность шаблона
func (nt *NotificationTemplates) ValidateTemplate(templateStr string) error {
	_, err := template.New("test").Funcs(nt.functions).Parse(templateStr)
	return err
}

// Template helper functions

// formatTime форматирует время
func formatTime(t interface{}) string {
	switch v := t.(type) {
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	case string:
		if parsed, err := time.Parse(time.RFC3339, v); err == nil {
			return parsed.Format("2006-01-02 15:04:05")
		}
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// timeAgo возвращает "X ago" формат
func timeAgo(t interface{}) string {
	var targetTime time.Time
	
	switch v := t.(type) {
	case time.Time:
		targetTime = v
	case string:
		if parsed, err := time.Parse(time.RFC3339, v); err == nil {
			targetTime = parsed
		} else {
			return v
		}
	default:
		return fmt.Sprintf("%v", v)
	}
	
	duration := time.Since(targetTime)
	
	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%d %s ago", minutes, pluralize(minutes, "minute", "minutes"))
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%d %s ago", hours, pluralize(hours, "hour", "hours"))
	} else {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%d %s ago", days, pluralize(days, "day", "days"))
	}
}

// pluralize возвращает правильную форму множественного числа
func pluralize(count interface{}, singular, plural string) string {
	var n int
	switch v := count.(type) {
	case int:
		n = v
	case float64:
		n = int(v)
	default:
		return singular
	}
	
	if n == 1 {
		return singular
	}
	return plural
}

// truncate обрезает строку до указанной длины
func truncate(s interface{}, length int) string {
	str := fmt.Sprintf("%v", s)
	if len(str) <= length {
		return str
	}
	return str[:length-3] + "..."
}

// highlightCode добавляет подсветку кода
func highlightCode(code string) string {
	// Простая подсветка для markdown
	return fmt.Sprintf("```\n%s\n```", code)
}

// markdownToHTML конвертирует markdown в HTML (упрощенно)
func markdownToHTML(markdown string) string {
	// Упрощенная конвертация
	html := strings.ReplaceAll(markdown, "\n", "<br>")
	html = strings.ReplaceAll(html, "**", "<strong>")
	html = strings.ReplaceAll(html, "**", "</strong>")
	html = strings.ReplaceAll(html, "*", "<em>")
	html = strings.ReplaceAll(html, "*", "</em>")
	return html
}

// TemplateBuilder помощник для создания динамических шаблонов
type TemplateBuilder struct {
	title string
	body  string
	data  map[string]interface{}
}

// NewTemplateBuilder создает новый builder шаблонов
func NewTemplateBuilder() *TemplateBuilder {
	return &TemplateBuilder{
		data: make(map[string]interface{}),
	}
}

// SetTitle устанавливает заголовок
func (tb *TemplateBuilder) SetTitle(title string) *TemplateBuilder {
	tb.title = title
	return tb
}

// SetBody устанавливает тело
func (tb *TemplateBuilder) SetBody(body string) *TemplateBuilder {
	tb.body = body
	return tb
}

// AddData добавляет данные
func (tb *TemplateBuilder) AddData(key string, value interface{}) *TemplateBuilder {
	tb.data[key] = value
	return tb
}

// Build создает шаблон
func (tb *TemplateBuilder) Build() (string, string, map[string]interface{}) {
	return tb.title, tb.body, tb.data
}

// PersonalizedTemplateEngine движок персонализированных шаблонов
type PersonalizedTemplateEngine struct {
	templates *NotificationTemplates
	aiChains  *ai.AIChains
	logger    Logger
}

// NewPersonalizedTemplateEngine создает новый движок персонализации
func NewPersonalizedTemplateEngine(templates *NotificationTemplates, aiChains *ai.AIChains, logger Logger) *PersonalizedTemplateEngine {
	return &PersonalizedTemplateEngine{
		templates: templates,
		aiChains:  aiChains,
		logger:    logger,
	}
}

// PersonalizeContent персонализирует контент уведомления
func (pte *PersonalizedTemplateEngine) PersonalizeContent(ctx context.Context, notification *Notification, subscriber *NotificationSubscriber, context *NotificationContext) (*PersonalizedContent, error) {
	// Базовый контент
	content := &PersonalizedContent{
		Subject: notification.Title,
		Body:    notification.Message,
		Summary: pte.generateSummary(notification),
		Context: make(map[string]string),
	}
	
	// AI персонализация если включена
	if subscriber.Preferences.AIPersonalization && pte.aiChains != nil {
		aiContent, err := pte.generateAIPersonalizedContent(ctx, notification, subscriber, context)
		if err != nil {
			pte.logger.Error("AI personalization failed", err)
		} else {
			content = aiContent
		}
	}
	
	// Генерируем action items
	content.ActionItems = pte.generateActionItems(notification, subscriber, context)
	
	return content, nil
}

// generateAIPersonalizedContent генерирует персонализированный контент с помощью AI
func (pte *PersonalizedTemplateEngine) generateAIPersonalizedContent(ctx context.Context, notification *Notification, subscriber *NotificationSubscriber, context *NotificationContext) (*PersonalizedContent, error) {
	prompt := fmt.Sprintf(`Personalize this notification for the user:

NOTIFICATION:
Title: %s
Message: %s
Type: %s
Priority: %s

USER CONTEXT:
- User ID: %s
- Preferences: %v
- Role: %s
- Recent Activity: %d events

Please create a personalized version that:
1. Uses appropriate tone for the user
2. Highlights relevant information
3. Provides actionable insights
4. Adapts to user's working style

Respond with personalized subject and body.`,
		notification.Title,
		notification.Message,
		notification.Type,
		notification.Priority,
		subscriber.UserID,
		subscriber.Preferences,
		subscriber.Context["role"],
		len(context.RecentActivity))
	
	response, err := pte.aiChains.ExecuteTask("Content Personalization", prompt, "personalization")
	if err != nil {
		return nil, err
	}
	
	// Парсим ответ AI (упрощенно)
	lines := strings.Split(response, "\n")
	subject := notification.Title
	body := notification.Message
	
	for _, line := range lines {
		if strings.HasPrefix(line, "Subject:") {
			subject = strings.TrimSpace(strings.TrimPrefix(line, "Subject:"))
		} else if strings.HasPrefix(line, "Body:") {
			body = strings.TrimSpace(strings.TrimPrefix(line, "Body:"))
		}
	}
	
	return &PersonalizedContent{
		Subject: subject,
		Body:    body,
		Summary: pte.generateSummary(notification),
		Context: map[string]string{
			"personalization_source": "ai",
			"tone": "adaptive",
		},
	}, nil
}

// generateSummary генерирует краткое описание
func (pte *PersonalizedTemplateEngine) generateSummary(notification *Notification) string {
	maxLength := 100
	if len(notification.Message) <= maxLength {
		return notification.Message
	}
	return notification.Message[:maxLength-3] + "..."
}

// generateActionItems генерирует список действий
func (pte *PersonalizedTemplateEngine) generateActionItems(notification *Notification, subscriber *NotificationSubscriber, context *NotificationContext) []string {
	var actions []string
	
	switch notification.Type {
	case "task_assigned":
		actions = append(actions, "Review task details and requirements")
		actions = append(actions, "Check dependencies and blockers")
		actions = append(actions, "Update task status when starting work")
		
	case "task_completed":
		actions = append(actions, "Review completed work")
		actions = append(actions, "Provide feedback if necessary")
		
	case "git_push":
		actions = append(actions, "Review code changes")
		actions = append(actions, "Update related documentation")
		
	case "pull_request":
		if prAction, ok := notification.Data["action"].(string); ok && prAction == "opened" {
			actions = append(actions, "Conduct code review")
			actions = append(actions, "Test the changes")
			actions = append(actions, "Approve or request changes")
		}
		
	case "workflow_stage_change":
		actions = append(actions, "Check new stage requirements")
		actions = append(actions, "Update progress tracking")
		
	default:
		actions = append(actions, "Review notification details")
	}
	
	return actions
}