package youtrack

import (
	"strings"

	"github.com/grik-ai/ricochet-task/pkg/providers"
)

// YouTrackTranslator handles conversion between YouTrack and Universal formats
type YouTrackTranslator struct {
	statusMapping   map[string]providers.TaskStatus
	priorityMapping map[string]providers.TaskPriority
	typeMapping     map[string]providers.TaskType
}

// NewYouTrackTranslator creates a new translator
func NewYouTrackTranslator() *YouTrackTranslator {
	return &YouTrackTranslator{
		statusMapping: map[string]providers.TaskStatus{
			"Open":        {ID: "open", Name: "Open", Category: providers.StatusCategoryTodo},
			"In Progress": {ID: "in_progress", Name: "In Progress", Category: providers.StatusCategoryInProgress},
			"Fixed":       {ID: "fixed", Name: "Fixed", Category: providers.StatusCategoryDone, IsFinal: true},
			"Done":        {ID: "done", Name: "Done", Category: providers.StatusCategoryDone, IsFinal: true},
			"Verified":    {ID: "verified", Name: "Verified", Category: providers.StatusCategoryDone, IsFinal: true},
			"Incomplete":  {ID: "incomplete", Name: "Incomplete", Category: providers.StatusCategoryTodo},
			"To be discussed": {ID: "to_be_discussed", Name: "To be discussed", Category: providers.StatusCategoryTodo},
			"Reopened":    {ID: "reopened", Name: "Reopened", Category: providers.StatusCategoryTodo},
			"Duplicate":   {ID: "duplicate", Name: "Duplicate", Category: providers.StatusCategoryDone, IsFinal: true},
			"Won't fix":   {ID: "wont_fix", Name: "Won't fix", Category: providers.StatusCategoryCancelled, IsFinal: true},
			"Can't Reproduce": {ID: "cant_reproduce", Name: "Can't Reproduce", Category: providers.StatusCategoryCancelled, IsFinal: true},
			"Obsolete":    {ID: "obsolete", Name: "Obsolete", Category: providers.StatusCategoryCancelled, IsFinal: true},
			"Blocked":     {ID: "blocked", Name: "Blocked", Category: providers.StatusCategoryBlocked},
		},
		priorityMapping: map[string]providers.TaskPriority{
			"Show-stopper": providers.TaskPriorityCritical,
			"Critical":     providers.TaskPriorityHighest,
			"Major":        providers.TaskPriorityHigh,
			"Normal":       providers.TaskPriorityMedium,
			"Minor":        providers.TaskPriorityLow,
			"Cosmetic":     providers.TaskPriorityLowest,
		},
		typeMapping: map[string]providers.TaskType{
			"Feature":      providers.TaskTypeFeature,
			"Bug":          providers.TaskTypeBug,
			"Task":         providers.TaskTypeTask,
			"Epic":         providers.TaskTypeEpic,
			"Story":        providers.TaskTypeStory,
			"Sub-task":     providers.TaskTypeSubtask,
			"Improvement":  providers.TaskTypeImprovement,
			"Research":     providers.TaskTypeResearch,
		},
	}
}

// UniversalToYouTrack converts a Universal task to YouTrack issue
func (t *YouTrackTranslator) UniversalToYouTrack(task *providers.UniversalTask) *YouTrackIssue {
	issue := &YouTrackIssue{
		Summary:     task.Title,
		Description: task.Description,
	}

	// Set ID if updating existing issue
	if task.ExternalID != "" {
		issue.ID = task.ExternalID
	}

	// Set project
	if task.ProjectID != "" {
		issue.Project = &YouTrackProject{
			ID: task.ProjectID,
		}
	}

	// Convert status
	if ytStatus := t.findYouTrackStatus(task.Status); ytStatus != "" {
		issue.State = &YouTrackState{
			Name: ytStatus,
		}
	}

	// Convert priority
	if ytPriority := t.findYouTrackPriority(task.Priority); ytPriority != "" {
		issue.Priority = &YouTrackPriority{
			Name: ytPriority,
		}
	}

	// Convert type
	if ytType := t.findYouTrackType(task.Type); ytType != "" {
		issue.Type = &YouTrackIssueType{
			Name: ytType,
		}
	}

	// Set assignee
	if task.AssigneeID != "" {
		issue.Assignee = &YouTrackUser{
			ID: task.AssigneeID,
		}
	}

	// Set reporter
	if task.ReporterID != "" {
		issue.Reporter = &YouTrackUser{
			ID: task.ReporterID,
		}
	}

	// Convert time tracking
	if task.EstimatedTime != nil {
		issue.Estimation = DurationToYouTrackDuration(*task.EstimatedTime)
	}

	// Convert custom fields
	if task.CustomFields != nil {
		issue.CustomFields = t.convertCustomFieldsToYouTrack(task.CustomFields)
	}

	// Convert tags/labels
	// NOTE: For now, we skip tags during creation as YouTrack requires tag IDs
	// TODO: Implement proper tag creation/lookup by adding tags after issue creation
	if len(task.Labels) > 0 && task.ExternalID != "" {
		// Only set tags when updating existing issues
		issue.Tags = make([]*YouTrackTag, len(task.Labels))
		for i, label := range task.Labels {
			issue.Tags[i] = &YouTrackTag{
				Name: label,
			}
		}
	}

	// Set timestamps
	if !task.CreatedAt.IsZero() {
		issue.Created = task.CreatedAt.Unix() * 1000
	}
	if !task.UpdatedAt.IsZero() {
		issue.Updated = task.UpdatedAt.Unix() * 1000
	}
	if task.ResolvedAt != nil {
		resolved := task.ResolvedAt.Unix() * 1000
		issue.Resolved = &resolved
	}

	return issue
}

// YouTrackToUniversal converts a YouTrack issue to Universal task
func (t *YouTrackTranslator) YouTrackToUniversal(issue *YouTrackIssue) *providers.UniversalTask {
	task := &providers.UniversalTask{
		ID:          issue.ID,
		ExternalID:  issue.ID,
		Key:         issue.IDReadable,
		Title:       issue.Summary,
		Description: issue.Description,
		CreatedAt:   issue.GetCreatedTime(),
		UpdatedAt:   issue.GetUpdatedTime(),
	}

	// Set project info
	if issue.Project != nil {
		task.ProjectID = issue.Project.ID
		task.ProjectKey = issue.Project.ShortName
	}

	// Convert status
	if issue.State != nil {
		if status, exists := t.statusMapping[issue.State.Name]; exists {
			task.Status = status
		} else {
			// Create a dynamic status mapping
			task.Status = providers.TaskStatus{
				ID:   strings.ToLower(strings.ReplaceAll(issue.State.Name, " ", "_")),
				Name: issue.State.Name,
				Category: t.inferStatusCategory(issue.State.Name, issue.State.IsResolved),
				IsFinal: issue.State.IsResolved,
			}
		}
	}

	// Convert priority
	if issue.Priority != nil {
		if priority, exists := t.priorityMapping[issue.Priority.Name]; exists {
			task.Priority = priority
		} else {
			// Default to medium if unknown
			task.Priority = providers.TaskPriorityMedium
		}
	}

	// Convert type
	if issue.Type != nil {
		if taskType, exists := t.typeMapping[issue.Type.Name]; exists {
			task.Type = taskType
		} else {
			// Default to task if unknown
			task.Type = providers.TaskTypeTask
		}
	}

	// Set assignee
	if issue.Assignee != nil {
		task.AssigneeID = issue.Assignee.ID
	}

	// Set reporter
	if issue.Reporter != nil {
		task.ReporterID = issue.Reporter.ID
		task.CreatorID = issue.Reporter.ID
	}

	// Convert time tracking
	if issue.Estimation != nil {
		duration := issue.Estimation.ToDuration()
		task.EstimatedTime = &duration
	}

	// Set resolved time
	if resolvedTime := issue.GetResolvedTime(); resolvedTime != nil {
		task.ResolvedAt = resolvedTime
	}

	// Convert custom fields
	if issue.CustomFields != nil {
		task.CustomFields = t.convertCustomFieldsFromYouTrack(issue.CustomFields)
	}

	// Convert tags to labels
	if len(issue.Tags) > 0 {
		task.Labels = make([]string, len(issue.Tags))
		for i, tag := range issue.Tags {
			task.Labels[i] = tag.Name
		}
	}

	// Convert comments
	if len(issue.Comments) > 0 {
		task.Comments = make([]*providers.Comment, len(issue.Comments))
		for i, comment := range issue.Comments {
			task.Comments[i] = t.YouTrackCommentToUniversal(comment)
		}
	}

	// Convert attachments
	if len(issue.Attachments) > 0 {
		task.Attachments = make([]*providers.Attachment, len(issue.Attachments))
		for i, attachment := range issue.Attachments {
			task.Attachments[i] = t.youTrackAttachmentToUniversal(attachment)
		}
	}

	// Handle hierarchical structure
	if issue.Parent != nil {
		task.ParentID = issue.Parent.ID
	}

	if len(issue.Subtasks) > 0 {
		task.SubtaskIDs = make([]string, len(issue.Subtasks))
		for i, subtask := range issue.Subtasks {
			task.SubtaskIDs[i] = subtask.ID
		}
	}

	// Store original YouTrack data
	task.ProviderData = map[string]interface{}{
		"youtrack_original": issue,
	}

	return task
}

// UniversalUpdatesToYouTrack converts universal updates to YouTrack format
func (t *YouTrackTranslator) UniversalUpdatesToYouTrack(updates *providers.TaskUpdate) *YouTrackIssueUpdate {
	ytUpdates := &YouTrackIssueUpdate{}

	if updates.Title != nil {
		ytUpdates.Summary = updates.Title
	}

	if updates.Description != nil {
		ytUpdates.Description = updates.Description
	}

	if updates.Status != nil {
		if ytStatus := t.findYouTrackStatus(*updates.Status); ytStatus != "" {
			ytUpdates.State = &YouTrackState{
				Name: ytStatus,
			}
		}
	}

	if updates.Priority != nil {
		if ytPriority := t.findYouTrackPriority(*updates.Priority); ytPriority != "" {
			ytUpdates.Priority = &YouTrackPriority{
				Name: ytPriority,
			}
		}
	}

	if updates.AssigneeID != nil {
		ytUpdates.Assignee = &YouTrackUser{
			ID: *updates.AssigneeID,
		}
	}

	if updates.EstimatedTime != nil {
		ytUpdates.Estimation = DurationToYouTrackDuration(*updates.EstimatedTime)
	}

	if updates.CustomFields != nil {
		ytUpdates.CustomFields = t.convertCustomFieldUpdatesToYouTrack(updates.CustomFields)
	}

	if len(updates.Labels) > 0 {
		ytUpdates.Tags = make([]*YouTrackTag, len(updates.Labels))
		for i, label := range updates.Labels {
			ytUpdates.Tags[i] = &YouTrackTag{
				Name: label,
			}
		}
	}

	return ytUpdates
}

// UniversalFiltersToYouTrack converts universal filters to YouTrack format
func (t *YouTrackTranslator) UniversalFiltersToYouTrack(filters *providers.TaskFilters) *YouTrackIssueFilters {
	ytFilters := &YouTrackIssueFilters{
		ProjectID: filters.ProjectID,
		Assignee:  filters.AssigneeID,
		Reporter:  filters.ReporterID,
		Query:     filters.Query,
		Top:       filters.Limit,
		Skip:      filters.Offset,
	}

	// Convert status filters
	if len(filters.Status) > 0 {
		// For YouTrack, we need to convert status names
		ytStatuses := make([]string, 0, len(filters.Status))
		for _, status := range filters.Status {
			if ytStatus := t.findYouTrackStatusByID(status); ytStatus != "" {
				ytStatuses = append(ytStatuses, ytStatus)
			}
		}
		if len(ytStatuses) == 1 {
			ytFilters.State = ytStatuses[0]
		} else if len(ytStatuses) > 1 {
			// Multiple statuses need to be handled in query
			ytFilters.Query = t.buildMultiStatusQuery(ytStatuses, ytFilters.Query)
		}
	}

	// Convert priority filters
	if len(filters.Priority) > 0 {
		ytPriorities := make([]string, 0, len(filters.Priority))
		for _, priority := range filters.Priority {
			if ytPriority := t.findYouTrackPriorityByTaskPriority(providers.TaskPriority(priority)); ytPriority != "" {
				ytPriorities = append(ytPriorities, ytPriority)
			}
		}
		if len(ytPriorities) == 1 {
			ytFilters.Priority = ytPriorities[0]
		} else if len(ytPriorities) > 1 {
			ytFilters.Query = t.buildMultiPriorityQuery(ytPriorities, ytFilters.Query)
		}
	}

	// Convert type filters
	if len(filters.Type) > 0 {
		ytTypes := make([]string, 0, len(filters.Type))
		for _, taskType := range filters.Type {
			if ytType := t.findYouTrackTypeByTaskType(providers.TaskType(taskType)); ytType != "" {
				ytTypes = append(ytTypes, ytType)
			}
		}
		if len(ytTypes) == 1 {
			ytFilters.Type = ytTypes[0]
		} else if len(ytTypes) > 1 {
			ytFilters.Query = t.buildMultiTypeQuery(ytTypes, ytFilters.Query)
		}
	}

	// Set time filters
	ytFilters.CreatedAfter = filters.CreatedAfter
	ytFilters.CreatedBefore = filters.CreatedBefore
	ytFilters.UpdatedAfter = filters.UpdatedAfter
	ytFilters.UpdatedBefore = filters.UpdatedBefore

	return ytFilters
}

// Status conversion helpers
func (t *YouTrackTranslator) UniversalStatusToYouTrack(status providers.TaskStatus) string {
	return t.findYouTrackStatus(status)
}

func (t *YouTrackTranslator) YouTrackStatusToUniversal(status *YouTrackState) providers.TaskStatus {
	if status == nil {
		return providers.TaskStatus{}
	}

	if universalStatus, exists := t.statusMapping[status.Name]; exists {
		return universalStatus
	}

	// Create dynamic mapping
	return providers.TaskStatus{
		ID:       strings.ToLower(strings.ReplaceAll(status.Name, " ", "_")),
		Name:     status.Name,
		Category: t.inferStatusCategory(status.Name, status.IsResolved),
		IsFinal:  status.IsResolved,
	}
}

// Comment conversion
func (t *YouTrackTranslator) YouTrackCommentToUniversal(comment *YouTrackComment) *providers.Comment {
	universalComment := &providers.Comment{
		ID:        comment.ID,
		Content:   comment.Text,
		CreatedAt: comment.GetCreatedTime(),
		UpdatedAt: comment.GetUpdatedTime(),
		IsEdited:  comment.GetUpdatedTime().After(comment.GetCreatedTime()),
	}

	if comment.Author != nil {
		universalComment.AuthorID = comment.Author.ID
	}

	return universalComment
}

// Attachment conversion
func (t *YouTrackTranslator) youTrackAttachmentToUniversal(attachment *YouTrackAttachment) *providers.Attachment {
	universalAttachment := &providers.Attachment{
		ID:          attachment.ID,
		Filename:    attachment.Name,
		ContentType: attachment.MimeType,
		Size:        attachment.Size,
		URL:         attachment.URL,
		UploadedAt:  attachment.GetCreatedTime(),
	}

	if attachment.Author != nil {
		universalAttachment.UploadedBy = attachment.Author.ID
	}

	return universalAttachment
}

// Custom field conversions
func (t *YouTrackTranslator) convertCustomFieldsToYouTrack(fields map[string]interface{}) []*YouTrackCustomField {
	ytFields := make([]*YouTrackCustomField, 0, len(fields))

	for name, value := range fields {
		ytField := &YouTrackCustomField{
			Name:  name,
			Value: value,
		}
		ytFields = append(ytFields, ytField)
	}

	return ytFields
}

func (t *YouTrackTranslator) convertCustomFieldsFromYouTrack(fields []*YouTrackCustomField) map[string]interface{} {
	customFields := make(map[string]interface{})

	for _, field := range fields {
		customFields[field.Name] = field.Value
	}

	return customFields
}

func (t *YouTrackTranslator) convertCustomFieldUpdatesToYouTrack(fields map[string]interface{}) []*YouTrackCustomFieldUpdate {
	ytUpdates := make([]*YouTrackCustomFieldUpdate, 0, len(fields))

	for name, value := range fields {
		ytUpdate := &YouTrackCustomFieldUpdate{
			ID:    name, // In practice, you'd need to map field names to IDs
			Value: value,
		}
		ytUpdates = append(ytUpdates, ytUpdate)
	}

	return ytUpdates
}

// Helper methods for finding mappings
func (t *YouTrackTranslator) findYouTrackStatus(status providers.TaskStatus) string {
	for ytStatus, universalStatus := range t.statusMapping {
		if universalStatus.ID == status.ID || universalStatus.Name == status.Name {
			return ytStatus
		}
	}
	return status.Name // Fallback to the original name
}

func (t *YouTrackTranslator) findYouTrackStatusByID(statusID string) string {
	for ytStatus, universalStatus := range t.statusMapping {
		if universalStatus.ID == statusID {
			return ytStatus
		}
	}
	return ""
}

func (t *YouTrackTranslator) findYouTrackPriority(priority providers.TaskPriority) string {
	for ytPriority, universalPriority := range t.priorityMapping {
		if universalPriority == priority {
			return ytPriority
		}
	}
	return "Normal" // Default fallback
}

func (t *YouTrackTranslator) findYouTrackPriorityByTaskPriority(priority providers.TaskPriority) string {
	return t.findYouTrackPriority(priority)
}

func (t *YouTrackTranslator) findYouTrackType(taskType providers.TaskType) string {
	for ytType, universalType := range t.typeMapping {
		if universalType == taskType {
			return ytType
		}
	}
	return "Task" // Default fallback
}

func (t *YouTrackTranslator) findYouTrackTypeByTaskType(taskType providers.TaskType) string {
	return t.findYouTrackType(taskType)
}

// Helper method to infer status category
func (t *YouTrackTranslator) inferStatusCategory(statusName string, isResolved bool) providers.StatusCategory {
	statusLower := strings.ToLower(statusName)
	
	if isResolved {
		if strings.Contains(statusLower, "duplicate") || 
		   strings.Contains(statusLower, "won't") || 
		   strings.Contains(statusLower, "obsolete") ||
		   strings.Contains(statusLower, "can't reproduce") {
			return providers.StatusCategoryCancelled
		}
		return providers.StatusCategoryDone
	}
	
	if strings.Contains(statusLower, "progress") || strings.Contains(statusLower, "dev") {
		return providers.StatusCategoryInProgress
	}
	
	if strings.Contains(statusLower, "block") {
		return providers.StatusCategoryBlocked
	}
	
	if strings.Contains(statusLower, "review") || strings.Contains(statusLower, "test") {
		return providers.StatusCategoryReview
	}
	
	return providers.StatusCategoryTodo
}

// Query building helpers
func (t *YouTrackTranslator) buildMultiStatusQuery(statuses []string, existingQuery string) string {
	statusQuery := "State: {" + strings.Join(statuses, "} or State: {") + "}"
	return t.combineQueries(existingQuery, statusQuery)
}

func (t *YouTrackTranslator) buildMultiPriorityQuery(priorities []string, existingQuery string) string {
	priorityQuery := "Priority: {" + strings.Join(priorities, "} or Priority: {") + "}"
	return t.combineQueries(existingQuery, priorityQuery)
}

func (t *YouTrackTranslator) buildMultiTypeQuery(types []string, existingQuery string) string {
	typeQuery := "Type: {" + strings.Join(types, "} or Type: {") + "}"
	return t.combineQueries(existingQuery, typeQuery)
}

func (t *YouTrackTranslator) combineQueries(existing, new string) string {
	if existing == "" {
		return new
	}
	return "(" + existing + ") and (" + new + ")"
}