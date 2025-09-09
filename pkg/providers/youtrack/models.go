package youtrack

import (
	"fmt"
	"time"
)

// YouTrackIssue represents an issue in YouTrack
type YouTrackIssue struct {
	ID          string             `json:"id,omitempty"`
	IDReadable  string             `json:"idReadable,omitempty"`
	Summary     string             `json:"summary"`
	Description string             `json:"description,omitempty"`
	Project     *YouTrackProject   `json:"project,omitempty"`
	State       *YouTrackState     `json:"state,omitempty"`
	Assignee    *YouTrackUser      `json:"assignee,omitempty"`
	Reporter    *YouTrackUser      `json:"reporter,omitempty"`
	Priority    *YouTrackPriority  `json:"priority,omitempty"`
	Type        *YouTrackIssueType `json:"type,omitempty"`
	Created     int64              `json:"created,omitempty"`
	Updated     int64              `json:"updated,omitempty"`
	Resolved    *int64             `json:"resolved,omitempty"`
	
	// Hierarchical structure
	Parent      *YouTrackIssue     `json:"parent,omitempty"`
	Subtasks    []*YouTrackIssue   `json:"subtasks,omitempty"`
	
	// Time tracking
	Estimation  *YouTrackDuration  `json:"estimation,omitempty"`
	TimeSpent   *YouTrackDuration  `json:"spent,omitempty"`
	
	// Custom fields
	CustomFields []*YouTrackCustomField `json:"customFields,omitempty"`
	
	// Related items
	Comments    []*YouTrackComment    `json:"comments,omitempty"`
	Attachments []*YouTrackAttachment `json:"attachments,omitempty"`
	Links       []*YouTrackIssueLink  `json:"links,omitempty"`
	Tags        []*YouTrackTag        `json:"tags,omitempty"`
	
	// Workflow
	WorkItems   []*YouTrackWorkItem   `json:"workItems,omitempty"`
	
	// Board-related
	Board       *YouTrackBoard        `json:"board,omitempty"`
	Column      *YouTrackBoardColumn  `json:"column,omitempty"`
	Sprint      *YouTrackSprint       `json:"sprint,omitempty"`
}

// YouTrackIssueUpdate represents updates to an issue
type YouTrackIssueUpdate struct {
	Summary     *string            `json:"summary,omitempty"`
	Description *string            `json:"description,omitempty"`
	State       *YouTrackState     `json:"state,omitempty"`
	Assignee    *YouTrackUser      `json:"assignee,omitempty"`
	Priority    *YouTrackPriority  `json:"priority,omitempty"`
	Type        *YouTrackIssueType `json:"type,omitempty"`
	Estimation  *YouTrackDuration  `json:"estimation,omitempty"`
	
	// Custom fields updates
	CustomFields []*YouTrackCustomFieldUpdate `json:"customFields,omitempty"`
	
	// Tags
	Tags        []*YouTrackTag     `json:"tags,omitempty"`
}

// YouTrackProject represents a project in YouTrack
type YouTrackProject struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ShortName   string `json:"shortName,omitempty"`
	Description string `json:"description,omitempty"`
}

// YouTrackState represents an issue state/status
type YouTrackState struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	IsResolved  bool   `json:"isResolved,omitempty"`
	LocalizedName string `json:"localizedName,omitempty"`
}

// YouTrackUser represents a user in YouTrack
type YouTrackUser struct {
	ID       string `json:"id,omitempty"`
	Login    string `json:"login,omitempty"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	FullName string `json:"fullName,omitempty"`
}

// YouTrackPriority represents an issue priority
type YouTrackPriority struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       *YouTrackColor `json:"color,omitempty"`
}

// YouTrackIssueType represents an issue type
type YouTrackIssueType struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	AutoAttached bool  `json:"autoAttached,omitempty"`
}

// YouTrackDuration represents time duration in YouTrack
type YouTrackDuration struct {
	Minutes     int    `json:"minutes,omitempty"`
	Presentation string `json:"presentation,omitempty"`
}

// YouTrackCustomField represents a custom field value
type YouTrackCustomField struct {
	ID    string      `json:"id,omitempty"`
	Name  string      `json:"name"`
	Value interface{} `json:"value,omitempty"`
	ProjectCustomField *YouTrackProjectCustomField `json:"projectCustomField,omitempty"`
}

// YouTrackCustomFieldUpdate represents custom field update
type YouTrackCustomFieldUpdate struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
}

// YouTrackProjectCustomField represents project-level custom field definition
type YouTrackProjectCustomField struct {
	ID    string `json:"id,omitempty"`
	Field *YouTrackCustomFieldDefinition `json:"field,omitempty"`
}

// YouTrackCustomFieldDefinition represents custom field definition
type YouTrackCustomFieldDefinition struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	LocalizedName string `json:"localizedName,omitempty"`
	FieldType   *YouTrackFieldType `json:"fieldType,omitempty"`
}

// YouTrackFieldType represents custom field type
type YouTrackFieldType struct {
	ID           string `json:"id,omitempty"`
	Presentation string `json:"presentation,omitempty"`
}

// YouTrackComment represents a comment on an issue
type YouTrackComment struct {
	ID       string        `json:"id,omitempty"`
	Text     string        `json:"text"`
	Author   *YouTrackUser `json:"author,omitempty"`
	Created  int64         `json:"created,omitempty"`
	Updated  int64         `json:"updated,omitempty"`
	Deleted  bool          `json:"deleted,omitempty"`
	
	// Visibility
	PermittedGroup *YouTrackUserGroup `json:"permittedGroup,omitempty"`
}

// YouTrackAttachment represents a file attachment
type YouTrackAttachment struct {
	ID          string        `json:"id,omitempty"`
	Name        string        `json:"name"`
	Author      *YouTrackUser `json:"author,omitempty"`
	Created     int64         `json:"created,omitempty"`
	Size        int64         `json:"size,omitempty"`
	Extension   string        `json:"extension,omitempty"`
	MimeType    string        `json:"mimeType,omitempty"`
	URL         string        `json:"url,omitempty"`
	ThumbnailURL string       `json:"thumbnailURL,omitempty"`
}

// YouTrackIssueLink represents a link between issues
type YouTrackIssueLink struct {
	ID          string               `json:"id,omitempty"`
	Direction   string               `json:"direction,omitempty"`
	LinkType    *YouTrackIssueLinkType `json:"linkType,omitempty"`
	Issues      []*YouTrackIssue     `json:"issues,omitempty"`
}

// YouTrackIssueLinkType represents a type of link between issues
type YouTrackIssueLinkType struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	LocalizedName string `json:"localizedName,omitempty"`
	SourceToTarget string `json:"sourceToTarget,omitempty"`
	TargetToSource string `json:"targetToSource,omitempty"`
	Directed    bool   `json:"directed,omitempty"`
}

// YouTrackTag represents a tag
type YouTrackTag struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name"`
	Query string `json:"query,omitempty"`
	Color *YouTrackColor `json:"color,omitempty"`
}

// YouTrackColor represents a color
type YouTrackColor struct {
	ID         string `json:"id,omitempty"`
	Background string `json:"background,omitempty"`
	Foreground string `json:"foreground,omitempty"`
}

// YouTrackWorkItem represents a work item (time tracking)
type YouTrackWorkItem struct {
	ID          string            `json:"id,omitempty"`
	Author      *YouTrackUser     `json:"author,omitempty"`
	Creator     *YouTrackUser     `json:"creator,omitempty"`
	Text        string            `json:"text,omitempty"`
	Type        *YouTrackWorkItemType `json:"type,omitempty"`
	Duration    *YouTrackDuration `json:"duration,omitempty"`
	Date        int64             `json:"date,omitempty"`
	Created     int64             `json:"created,omitempty"`
	Updated     int64             `json:"updated,omitempty"`
}

// YouTrackWorkItemType represents a work item type
type YouTrackWorkItemType struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	AutoAttached bool  `json:"autoAttached,omitempty"`
}

// YouTrackUserGroup represents a user group
type YouTrackUserGroup struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}

// Board-related structures
// YouTrackBoard represents an agile board
type YouTrackBoard struct {
	ID          string                `json:"id,omitempty"`
	Name        string                `json:"name"`
	Columns     []*YouTrackBoardColumn `json:"columns,omitempty"`
	Projects    []*YouTrackProject    `json:"projects,omitempty"`
	Sprints     []*YouTrackSprint     `json:"sprints,omitempty"`
	EstimationField *YouTrackCustomFieldDefinition `json:"estimationField,omitempty"`
}

// YouTrackBoardColumn represents a board column
type YouTrackBoardColumn struct {
	ID          string         `json:"id,omitempty"`
	Presentation string        `json:"presentation,omitempty"`
	IsResolved  bool           `json:"isResolved,omitempty"`
	Color       *YouTrackColor `json:"color,omitempty"`
	WIPLimit    *YouTrackWIPLimit `json:"wipLimit,omitempty"`
}

// YouTrackWIPLimit represents work-in-progress limit
type YouTrackWIPLimit struct {
	Max int `json:"max,omitempty"`
	Min int `json:"min,omitempty"`
}

// YouTrackSprint represents a sprint
type YouTrackSprint struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Goal        string `json:"goal,omitempty"`
	Start       int64  `json:"start,omitempty"`
	Finish      int64  `json:"finish,omitempty"`
	Archived    bool   `json:"archived,omitempty"`
	IsDefault   bool   `json:"isDefault,omitempty"`
}

// Filters for querying
// YouTrackIssueFilters represents filters for listing issues
type YouTrackIssueFilters struct {
	ProjectID     string     `json:"projectId,omitempty"`
	State         string     `json:"state,omitempty"`
	Assignee      string     `json:"assignee,omitempty"`
	Reporter      string     `json:"reporter,omitempty"`
	Type          string     `json:"type,omitempty"`
	Priority      string     `json:"priority,omitempty"`
	CreatedAfter  *time.Time `json:"createdAfter,omitempty"`
	CreatedBefore *time.Time `json:"createdBefore,omitempty"`
	UpdatedAfter  *time.Time `json:"updatedAfter,omitempty"`
	UpdatedBefore *time.Time `json:"updatedBefore,omitempty"`
	Query         string     `json:"query,omitempty"`
	Top           int        `json:"top,omitempty"`
	Skip          int        `json:"skip,omitempty"`
}

// Helper methods for time conversion
func (i *YouTrackIssue) GetCreatedTime() time.Time {
	if i.Created == 0 {
		return time.Time{}
	}
	return time.Unix(i.Created/1000, 0)
}

func (i *YouTrackIssue) GetUpdatedTime() time.Time {
	if i.Updated == 0 {
		return time.Time{}
	}
	return time.Unix(i.Updated/1000, 0)
}

func (i *YouTrackIssue) GetResolvedTime() *time.Time {
	if i.Resolved == nil || *i.Resolved == 0 {
		return nil
	}
	t := time.Unix(*i.Resolved/1000, 0)
	return &t
}

func (c *YouTrackComment) GetCreatedTime() time.Time {
	if c.Created == 0 {
		return time.Time{}
	}
	return time.Unix(c.Created/1000, 0)
}

func (c *YouTrackComment) GetUpdatedTime() time.Time {
	if c.Updated == 0 {
		return time.Time{}
	}
	return time.Unix(c.Updated/1000, 0)
}

func (a *YouTrackAttachment) GetCreatedTime() time.Time {
	if a.Created == 0 {
		return time.Time{}
	}
	return time.Unix(a.Created/1000, 0)
}

func (w *YouTrackWorkItem) GetDateTime() time.Time {
	if w.Date == 0 {
		return time.Time{}
	}
	return time.Unix(w.Date/1000, 0)
}

func (w *YouTrackWorkItem) GetCreatedTime() time.Time {
	if w.Created == 0 {
		return time.Time{}
	}
	return time.Unix(w.Created/1000, 0)
}

func (w *YouTrackWorkItem) GetUpdatedTime() time.Time {
	if w.Updated == 0 {
		return time.Time{}
	}
	return time.Unix(w.Updated/1000, 0)
}

func (s *YouTrackSprint) GetStartTime() time.Time {
	if s.Start == 0 {
		return time.Time{}
	}
	return time.Unix(s.Start/1000, 0)
}

func (s *YouTrackSprint) GetFinishTime() time.Time {
	if s.Finish == 0 {
		return time.Time{}
	}
	return time.Unix(s.Finish/1000, 0)
}

// Helper methods for duration conversion
func (d *YouTrackDuration) ToDuration() time.Duration {
	if d == nil || d.Minutes == 0 {
		return 0
	}
	return time.Duration(d.Minutes) * time.Minute
}

func DurationToYouTrackDuration(duration time.Duration) *YouTrackDuration {
	if duration == 0 {
		return nil
	}
	
	minutes := int(duration.Minutes())
	return &YouTrackDuration{
		Minutes: minutes,
		Presentation: formatMinutes(minutes),
	}
}

func formatMinutes(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	
	hours := minutes / 60
	remainingMinutes := minutes % 60
	
	if remainingMinutes == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	
	return fmt.Sprintf("%dh %dm", hours, remainingMinutes)
}

// Helper methods for finding custom fields
func (i *YouTrackIssue) GetCustomFieldValue(fieldName string) interface{} {
	if i.CustomFields == nil {
		return nil
	}
	
	for _, field := range i.CustomFields {
		if field.Name == fieldName {
			return field.Value
		}
	}
	
	return nil
}

func (i *YouTrackIssue) GetCustomFieldStringValue(fieldName string) string {
	value := i.GetCustomFieldValue(fieldName)
	if value == nil {
		return ""
	}
	
	if str, ok := value.(string); ok {
		return str
	}
	
	return ""
}

// Helper methods for issue hierarchy
func (i *YouTrackIssue) IsSubtask() bool {
	return i.Parent != nil
}

func (i *YouTrackIssue) HasSubtasks() bool {
	return len(i.Subtasks) > 0
}

// Helper methods for issue state
func (i *YouTrackIssue) IsResolved() bool {
	return i.State != nil && i.State.IsResolved
}

func (i *YouTrackIssue) GetDisplayID() string {
	if i.IDReadable != "" {
		return i.IDReadable
	}
	return i.ID
}