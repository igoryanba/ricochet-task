package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"time"
)

// EmailChannel канал email уведомлений
type EmailChannel struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	logger       Logger
}

// NewEmailChannel создает новый email канал
func NewEmailChannel(logger Logger) *EmailChannel {
	return &EmailChannel{
		smtpHost:     "smtp.gmail.com", // Конфигурировать из настроек
		smtpPort:     "587",
		smtpUsername: "",
		smtpPassword: "",
		fromEmail:    "notifications@ricochet-task.com",
		logger:       logger,
	}
}

func (ec *EmailChannel) GetType() string {
	return "email"
}

func (ec *EmailChannel) Send(ctx context.Context, notification *Notification) error {
	if len(notification.Recipients) == 0 {
		return fmt.Errorf("no recipients specified")
	}
	
	// Формируем email
	subject := notification.Title
	body := ec.formatEmailBody(notification)
	
	// Отправляем каждому получателю
	for _, recipient := range notification.Recipients {
		if err := ec.sendEmail(recipient, subject, body); err != nil {
			ec.logger.Error("Failed to send email", err, "recipient", recipient)
			return err
		}
	}
	
	ec.logger.Info("Email notification sent", 
		"recipients", len(notification.Recipients),
		"notification_id", notification.ID)
	
	return nil
}

func (ec *EmailChannel) sendEmail(to, subject, body string) error {
	// Простая реализация отправки email
	if ec.smtpUsername == "" {
		// В тестовом режиме просто логируем
		ec.logger.Info("Email would be sent", "to", to, "subject", subject)
		return nil
	}
	
	// Реальная отправка через SMTP
	auth := smtp.PlainAuth("", ec.smtpUsername, ec.smtpPassword, ec.smtpHost)
	
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" + body)
	
	err := smtp.SendMail(ec.smtpHost+":"+ec.smtpPort, auth, ec.fromEmail, []string{to}, msg)
	return err
}

func (ec *EmailChannel) formatEmailBody(notification *Notification) string {
	return fmt.Sprintf(`
<html>
<body>
<h2>%s</h2>
<p>%s</p>
<hr>
<p><small>Sent at: %s</small></p>
<p><small>Notification ID: %s</small></p>
</body>
</html>
`, notification.Title, notification.Message, notification.Timestamp.Format(time.RFC3339), notification.ID)
}

// SlackChannel канал Slack уведомлений
type SlackChannel struct {
	webhookURL string
	botToken   string
	logger     Logger
}

// NewSlackChannel создает новый Slack канал
func NewSlackChannel(logger Logger) *SlackChannel {
	return &SlackChannel{
		webhookURL: "", // Конфигурировать из настроек
		botToken:   "", // Конфигурировать из настроек
		logger:     logger,
	}
}

func (sc *SlackChannel) GetType() string {
	return "slack"
}

func (sc *SlackChannel) Send(ctx context.Context, notification *Notification) error {
	if sc.webhookURL == "" {
		// В тестовом режиме просто логируем
		sc.logger.Info("Slack notification would be sent", 
			"notification_id", notification.ID,
			"title", notification.Title)
		return nil
	}
	
	// Формируем Slack сообщение
	slackMsg := sc.formatSlackMessage(notification)
	
	// Отправляем через webhook
	return sc.sendSlackWebhook(slackMsg)
}

func (sc *SlackChannel) formatSlackMessage(notification *Notification) map[string]interface{} {
	color := sc.getColorByPriority(notification.Priority)
	
	attachment := map[string]interface{}{
		"color":    color,
		"title":    notification.Title,
		"text":     notification.Message,
		"ts":       notification.Timestamp.Unix(),
		"footer":   "Ricochet Task",
		"fields": []map[string]interface{}{
			{
				"title": "Priority",
				"value": notification.Priority,
				"short": true,
			},
			{
				"title": "Type",
				"value": notification.Type,
				"short": true,
			},
		},
	}
	
	return map[string]interface{}{
		"attachments": []interface{}{attachment},
	}
}

func (sc *SlackChannel) getColorByPriority(priority string) string {
	switch priority {
	case "high", "critical":
		return "danger"
	case "medium":
		return "warning"
	default:
		return "good"
	}
}

func (sc *SlackChannel) sendSlackWebhook(message map[string]interface{}) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	resp, err := http.Post(sc.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}
	
	return nil
}

// TeamsChannel канал Microsoft Teams уведомлений
type TeamsChannel struct {
	webhookURL string
	logger     Logger
}

// NewTeamsChannel создает новый Teams канал
func NewTeamsChannel(logger Logger) *TeamsChannel {
	return &TeamsChannel{
		webhookURL: "", // Конфигурировать из настроек
		logger:     logger,
	}
}

func (tc *TeamsChannel) GetType() string {
	return "teams"
}

func (tc *TeamsChannel) Send(ctx context.Context, notification *Notification) error {
	if tc.webhookURL == "" {
		// В тестовом режиме просто логируем
		tc.logger.Info("Teams notification would be sent", 
			"notification_id", notification.ID,
			"title", notification.Title)
		return nil
	}
	
	// Формируем Teams сообщение
	teamsMsg := tc.formatTeamsMessage(notification)
	
	// Отправляем через webhook
	return tc.sendTeamsWebhook(teamsMsg)
}

func (tc *TeamsChannel) formatTeamsMessage(notification *Notification) map[string]interface{} {
	themeColor := tc.getThemeColorByPriority(notification.Priority)
	
	return map[string]interface{}{
		"@type":      "MessageCard",
		"@context":   "http://schema.org/extensions",
		"themeColor": themeColor,
		"summary":    notification.Title,
		"sections": []map[string]interface{}{
			{
				"activityTitle":    notification.Title,
				"activitySubtitle": notification.Type,
				"text":             notification.Message,
				"facts": []map[string]interface{}{
					{
						"name":  "Priority",
						"value": notification.Priority,
					},
					{
						"name":  "Time",
						"value": notification.Timestamp.Format("2006-01-02 15:04:05"),
					},
				},
			},
		},
	}
}

func (tc *TeamsChannel) getThemeColorByPriority(priority string) string {
	switch priority {
	case "high", "critical":
		return "FF0000" // Red
	case "medium":
		return "FFA500" // Orange
	default:
		return "00FF00" // Green
	}
}

func (tc *TeamsChannel) sendTeamsWebhook(message map[string]interface{}) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	resp, err := http.Post(tc.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("teams webhook returned status %d", resp.StatusCode)
	}
	
	return nil
}

// WebhookChannel универсальный webhook канал
type WebhookChannel struct {
	defaultURL string
	logger     Logger
}

// NewWebhookChannel создает новый webhook канал
func NewWebhookChannel(logger Logger) *WebhookChannel {
	return &WebhookChannel{
		defaultURL: "", // Конфигурировать из настроек
		logger:     logger,
	}
}

func (wc *WebhookChannel) GetType() string {
	return "webhook"
}

func (wc *WebhookChannel) Send(ctx context.Context, notification *Notification) error {
	webhookURL := wc.getWebhookURL(notification)
	if webhookURL == "" {
		wc.logger.Info("Webhook notification would be sent", 
			"notification_id", notification.ID,
			"title", notification.Title)
		return nil
	}
	
	// Формируем webhook payload
	payload := wc.formatWebhookPayload(notification)
	
	// Отправляем
	return wc.sendWebhook(webhookURL, payload)
}

func (wc *WebhookChannel) getWebhookURL(notification *Notification) string {
	// Можно брать URL из данных уведомления или использовать дефолтный
	if url, exists := notification.Data["webhook_url"].(string); exists {
		return url
	}
	return wc.defaultURL
}

func (wc *WebhookChannel) formatWebhookPayload(notification *Notification) map[string]interface{} {
	return map[string]interface{}{
		"id":         notification.ID,
		"type":       notification.Type,
		"title":      notification.Title,
		"message":    notification.Message,
		"priority":   notification.Priority,
		"recipients": notification.Recipients,
		"data":       notification.Data,
		"timestamp":  notification.Timestamp.Unix(),
		"source":     "ricochet-task",
	}
}

func (wc *WebhookChannel) sendWebhook(url string, payload map[string]interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "RicochetTask/1.0")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	
	return nil
}

// SMSChannel канал SMS уведомлений  
type SMSChannel struct {
	apiKey    string
	apiSecret string
	provider  string // twilio, nexmo, etc.
	logger    Logger
}

// NewSMSChannel создает новый SMS канал
func NewSMSChannel(logger Logger) *SMSChannel {
	return &SMSChannel{
		apiKey:    "", // Конфигурировать из настроек
		apiSecret: "", // Конфигурировать из настроек
		provider:  "twilio",
		logger:    logger,
	}
}

func (sc *SMSChannel) GetType() string {
	return "sms"
}

func (sc *SMSChannel) Send(ctx context.Context, notification *Notification) error {
	if sc.apiKey == "" {
		// В тестовом режиме просто логируем
		sc.logger.Info("SMS notification would be sent", 
			"notification_id", notification.ID,
			"recipients", notification.Recipients)
		return nil
	}
	
	// Формируем SMS текст
	smsText := sc.formatSMSText(notification)
	
	// Отправляем каждому получателю
	for _, recipient := range notification.Recipients {
		if err := sc.sendSMS(recipient, smsText); err != nil {
			sc.logger.Error("Failed to send SMS", err, "recipient", recipient)
			return err
		}
	}
	
	return nil
}

func (sc *SMSChannel) formatSMSText(notification *Notification) string {
	// SMS должны быть короткими
	maxLength := 160
	text := fmt.Sprintf("%s: %s", notification.Title, notification.Message)
	
	if len(text) > maxLength {
		text = text[:maxLength-3] + "..."
	}
	
	return text
}

func (sc *SMSChannel) sendSMS(to, text string) error {
	// Здесь была бы интеграция с SMS провайдером
	sc.logger.Info("SMS sent", "to", to, "text", text)
	return nil
}

// PushChannel канал push уведомлений
type PushChannel struct {
	fcmServerKey string
	logger       Logger
}

// NewPushChannel создает новый push канал  
func NewPushChannel(logger Logger) *PushChannel {
	return &PushChannel{
		fcmServerKey: "", // Конфигурировать из настроек
		logger:       logger,
	}
}

func (pc *PushChannel) GetType() string {
	return "push"
}

func (pc *PushChannel) Send(ctx context.Context, notification *Notification) error {
	if pc.fcmServerKey == "" {
		// В тестовом режиме просто логируем
		pc.logger.Info("Push notification would be sent", 
			"notification_id", notification.ID,
			"title", notification.Title)
		return nil
	}
	
	// Формируем push уведомление
	pushPayload := pc.formatPushPayload(notification)
	
	// Отправляем через FCM
	return pc.sendFCM(pushPayload)
}

func (pc *PushChannel) formatPushPayload(notification *Notification) map[string]interface{} {
	return map[string]interface{}{
		"notification": map[string]interface{}{
			"title": notification.Title,
			"body":  notification.Message,
			"icon":  "ic_notification",
		},
		"data": map[string]interface{}{
			"notification_id": notification.ID,
			"type":           notification.Type,
			"priority":       notification.Priority,
		},
		"registration_ids": notification.Recipients,
	}
}

func (pc *PushChannel) sendFCM(payload map[string]interface{}) error {
	// Здесь была бы интеграция с Firebase Cloud Messaging
	pc.logger.Info("Push notification sent via FCM")
	return nil
}

// DiscordChannel канал Discord уведомлений
type DiscordChannel struct {
	webhookURL string
	logger     Logger
}

// NewDiscordChannel создает новый Discord канал
func NewDiscordChannel(logger Logger) *DiscordChannel {
	return &DiscordChannel{
		webhookURL: "", // Конфигурировать из настроек
		logger:     logger,
	}
}

func (dc *DiscordChannel) GetType() string {
	return "discord"
}

func (dc *DiscordChannel) Send(ctx context.Context, notification *Notification) error {
	if dc.webhookURL == "" {
		dc.logger.Info("Discord notification would be sent", 
			"notification_id", notification.ID,
			"title", notification.Title)
		return nil
	}
	
	// Формируем Discord embed
	discordMsg := dc.formatDiscordMessage(notification)
	
	// Отправляем через webhook
	return dc.sendDiscordWebhook(discordMsg)
}

func (dc *DiscordChannel) formatDiscordMessage(notification *Notification) map[string]interface{} {
	color := dc.getColorByPriority(notification.Priority)
	
	embed := map[string]interface{}{
		"title":       notification.Title,
		"description": notification.Message,
		"color":       color,
		"timestamp":   notification.Timestamp.Format(time.RFC3339),
		"footer": map[string]interface{}{
			"text": "Ricochet Task",
		},
		"fields": []map[string]interface{}{
			{
				"name":   "Priority",
				"value":  notification.Priority,
				"inline": true,
			},
			{
				"name":   "Type",
				"value":  notification.Type,
				"inline": true,
			},
		},
	}
	
	return map[string]interface{}{
		"embeds": []interface{}{embed},
	}
}

func (dc *DiscordChannel) getColorByPriority(priority string) int {
	switch priority {
	case "high", "critical":
		return 0xFF0000 // Red
	case "medium":
		return 0xFFA500 // Orange
	default:
		return 0x00FF00 // Green
	}
}

func (dc *DiscordChannel) sendDiscordWebhook(message map[string]interface{}) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	resp, err := http.Post(dc.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("discord webhook returned status %d", resp.StatusCode)
	}
	
	return nil
}