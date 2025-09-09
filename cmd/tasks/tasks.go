package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/grik-ai/ricochet-task/pkg/providers"
	providerCmd "github.com/grik-ai/ricochet-task/cmd/providers"
)

var (
	registry *providers.ProviderRegistry
	logger   *logrus.Logger
)

// TasksCmd represents the tasks command
var TasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Manage tasks across providers",
	Long: `Create, update, list, and manage tasks across multiple task management providers.
	
Tasks can be created in specific providers or automatically routed to the optimal provider
based on configured routing rules.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initializeTasks()
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	Long: `Create a new task in the specified provider or using automatic routing.
	
Examples:
  ricochet tasks create --title "Implement OAuth" --provider youtrack-prod
  ricochet tasks create --title "Fix bug" --description "Login issue" --priority high
  ricochet tasks create --title "Research API" --type research --auto-route`,
	RunE: runCreateTask,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long: `List tasks from one or more providers with optional filters.
	
Examples:
  ricochet tasks list --provider youtrack-prod
  ricochet tasks list --providers all --status open
  ricochet tasks list --assignee me --priority high
  ricochet tasks list --project BACKEND --type bug`,
	RunE: runListTasks,
}

var getCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get a specific task",
	Long: `Retrieve detailed information about a specific task.
	
Examples:
  ricochet tasks get PROJ-123 --provider youtrack-prod
  ricochet tasks get 12345 --provider jira-company
  ricochet tasks get --search "OAuth implementation"`,
	RunE: runGetTask,
}

var updateCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "Update a task",
	Long: `Update an existing task's properties.
	
Examples:
  ricochet tasks update PROJ-123 --status "in_progress" --provider youtrack-prod
  ricochet tasks update 12345 --assignee john.doe --priority high
  ricochet tasks update PROJ-123 --title "New title" --description "Updated description"`,
	Args: cobra.ExactArgs(1),
	RunE: runUpdateTask,
}

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a task",
	Long: `Delete a task from the specified provider.
	
Examples:
  ricochet tasks delete PROJ-123 --provider youtrack-prod
  ricochet tasks delete 12345 --provider jira-company --force`,
	Args: cobra.ExactArgs(1),
	RunE: runDeleteTask,
}

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search tasks across providers",
	Long: `Search for tasks across one or more providers using a query string.
	
Examples:
  ricochet tasks search "authentication" --providers all
  ricochet tasks search "bug" --provider youtrack-prod --status open
  ricochet tasks search --query "assignee:me and priority:high"`,
	RunE: runSearchTasks,
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync tasks between providers",
	Long: `Synchronize tasks between different providers.
	
Examples:
  ricochet tasks sync --from youtrack-prod --to jira-company
  ricochet tasks sync --project BACKEND --bidirectional
  ricochet tasks sync --dry-run --from youtrack-prod --to notion-docs`,
	RunE: runSyncTasks,
}

var bulkCreateCmd = &cobra.Command{
	Use:   "bulk-create",
	Short: "Create multiple tasks from a file",
	Long: `Create multiple tasks from a JSON or YAML file.
	
Examples:
  ricochet tasks bulk-create --file tasks.json --provider youtrack-prod
  ricochet tasks bulk-create --file tasks.yaml --auto-route
  ricochet tasks bulk-create --file import.json --dry-run`,
	RunE: runBulkCreateTasks,
}

var bulkUpdateCmd = &cobra.Command{
	Use:   "bulk-update",
	Short: "Update multiple tasks from a file",
	Long: `Update multiple tasks from a JSON or YAML file with task ID to updates mapping.
	
Examples:
  ricochet tasks bulk-update --file updates.json --provider youtrack-prod
  ricochet tasks bulk-update --file batch-updates.yaml --dry-run`,
	RunE: runBulkUpdateTasks,
}

var bulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete",
	Short: "Delete multiple tasks",
	Long: `Delete multiple tasks by IDs from a file or command line.
	
Examples:
  ricochet tasks bulk-delete --file task-ids.txt --provider youtrack-prod
  ricochet tasks bulk-delete --ids PROJ-123,PROJ-124,PROJ-125 --provider youtrack-prod
  ricochet tasks bulk-delete --query "status:obsolete" --provider youtrack-prod --dry-run`,
	RunE: runBulkDeleteTasks,
}

func init() {
	// Add subcommands
	TasksCmd.AddCommand(createCmd)
	TasksCmd.AddCommand(listCmd)
	TasksCmd.AddCommand(getCmd)
	TasksCmd.AddCommand(updateCmd)
	TasksCmd.AddCommand(deleteCmd)
	TasksCmd.AddCommand(searchCmd)
	TasksCmd.AddCommand(syncCmd)
	TasksCmd.AddCommand(bulkCreateCmd)
	TasksCmd.AddCommand(bulkUpdateCmd)
	TasksCmd.AddCommand(bulkDeleteCmd)

	// Global task flags
	TasksCmd.PersistentFlags().StringP("provider", "p", "", "Target provider name")
	TasksCmd.PersistentFlags().StringSlice("providers", []string{}, "Multiple providers (use 'all' for all enabled)")
	TasksCmd.PersistentFlags().StringP("output", "o", "table", "Output format: table, json, yaml")

	// Create command flags
	createCmd.Flags().StringP("title", "t", "", "Task title")
	createCmd.Flags().StringP("description", "d", "", "Task description")
	createCmd.Flags().String("project", "", "Project ID")
	createCmd.Flags().String("type", "task", "Task type (task, bug, feature, etc.)")
	createCmd.Flags().String("priority", "medium", "Task priority (low, medium, high, critical)")
	createCmd.Flags().String("status", "", "Initial status")
	createCmd.Flags().String("assignee", "", "Assignee ID or username")
	createCmd.Flags().StringSlice("labels", []string{}, "Task labels")
	createCmd.Flags().Bool("auto-route", false, "Automatically route to optimal provider")
	createCmd.MarkFlagRequired("title")

	// List command flags
	listCmd.Flags().String("project", "", "Filter by project")
	listCmd.Flags().String("status", "", "Filter by status")
	listCmd.Flags().String("assignee", "", "Filter by assignee")
	listCmd.Flags().String("type", "", "Filter by type")
	listCmd.Flags().String("priority", "", "Filter by priority")
	listCmd.Flags().StringSlice("labels", []string{}, "Filter by labels")
	listCmd.Flags().Int("limit", 50, "Maximum number of tasks to return")
	listCmd.Flags().Int("offset", 0, "Number of tasks to skip")

	// Get command flags
	getCmd.Flags().String("search", "", "Search for task by title/description")

	// Update command flags
	updateCmd.Flags().StringP("title", "t", "", "New title")
	updateCmd.Flags().StringP("description", "d", "", "New description")
	updateCmd.Flags().String("status", "", "New status")
	updateCmd.Flags().String("priority", "", "New priority")
	updateCmd.Flags().String("assignee", "", "New assignee")
	updateCmd.Flags().StringSlice("labels", []string{}, "New labels (replaces existing)")
	updateCmd.Flags().StringSlice("add-labels", []string{}, "Add labels")
	updateCmd.Flags().StringSlice("remove-labels", []string{}, "Remove labels")

	// Delete command flags
	deleteCmd.Flags().Bool("force", false, "Force deletion without confirmation")

	// Search command flags
	searchCmd.Flags().String("query", "", "Search query")
	searchCmd.Flags().String("status", "", "Filter by status")
	searchCmd.Flags().String("assignee", "", "Filter by assignee")
	searchCmd.Flags().String("type", "", "Filter by type")
	searchCmd.Flags().String("priority", "", "Filter by priority")
	searchCmd.Flags().Int("limit", 100, "Maximum number of results")

	// Sync command flags
	syncCmd.Flags().String("from", "", "Source provider")
	syncCmd.Flags().String("to", "", "Target provider")
	syncCmd.Flags().String("project", "", "Sync specific project")
	syncCmd.Flags().Bool("bidirectional", false, "Bidirectional sync")
	syncCmd.Flags().Bool("dry-run", false, "Show what would be synced without making changes")
	syncCmd.MarkFlagRequired("from")
	syncCmd.MarkFlagRequired("to")

	// Bulk create command flags
	bulkCreateCmd.Flags().StringP("file", "f", "", "Input file (JSON or YAML)")
	bulkCreateCmd.Flags().Bool("auto-route", false, "Automatically route to optimal provider")
	bulkCreateCmd.Flags().Bool("dry-run", false, "Show what would be created without making changes")
	bulkCreateCmd.MarkFlagRequired("file")

	// Bulk update command flags
	bulkUpdateCmd.Flags().StringP("file", "f", "", "Input file (JSON or YAML)")
	bulkUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")
	bulkUpdateCmd.MarkFlagRequired("file")

	// Bulk delete command flags
	bulkDeleteCmd.Flags().StringP("file", "f", "", "File containing task IDs (one per line)")
	bulkDeleteCmd.Flags().String("ids", "", "Comma-separated list of task IDs")
	bulkDeleteCmd.Flags().String("query", "", "Query to select tasks for deletion")
	bulkDeleteCmd.Flags().Bool("dry-run", false, "Show what would be deleted without making changes")
	bulkDeleteCmd.Flags().Bool("force", false, "Force deletion without confirmation")
}

func initializeTasks() {
	// Reuse the provider registry initialization
	providerCmd.ProvidersCmd.PersistentPreRun(nil, nil)
	registry = providerCmd.GetRegistry() // We'd need to expose this
	logger = logrus.New()
}

func runCreateTask(cmd *cobra.Command, args []string) error {
	title, _ := cmd.Flags().GetString("title")
	description, _ := cmd.Flags().GetString("description")
	project, _ := cmd.Flags().GetString("project")
	taskType, _ := cmd.Flags().GetString("type")
	priority, _ := cmd.Flags().GetString("priority")
	status, _ := cmd.Flags().GetString("status")
	assignee, _ := cmd.Flags().GetString("assignee")
	labels, _ := cmd.Flags().GetStringSlice("labels")
	autoRoute, _ := cmd.Flags().GetBool("auto-route")
	providerName, _ := cmd.Flags().GetString("provider")

	// Create universal task
	task := &providers.UniversalTask{
		Title:       title,
		Description: description,
		ProjectID:   project,
		Type:        providers.TaskType(taskType),
		Priority:    mapPriority(priority),
		AssigneeID:  assignee,
		Labels:      labels,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if status != "" {
		task.Status = providers.TaskStatus{
			ID:   strings.ToLower(strings.ReplaceAll(status, " ", "_")),
			Name: status,
		}
	}

	// Determine target provider
	var provider providers.TaskProvider
	var err error

	if autoRoute {
		// TODO: Implement smart routing based on rules
		provider, err = registry.GetDefaultProvider()
	} else if providerName != "" {
		provider, err = registry.GetProvider(providerName)
	} else {
		provider, err = registry.GetDefaultProvider()
	}

	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Create task
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	createdTask, err := provider.CreateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Printf("✅ Task created successfully\n")
	fmt.Printf("ID: %s\n", createdTask.GetDisplayID())
	fmt.Printf("Title: %s\n", createdTask.Title)
	fmt.Printf("Provider: %s\n", createdTask.ProviderName)

	return nil
}

func runListTasks(cmd *cobra.Command, args []string) error {
	providerName, _ := cmd.Flags().GetString("provider")
	providerNames, _ := cmd.Flags().GetStringSlice("providers")
	output, _ := cmd.Flags().GetString("output")

	// Build filters
	filters := &providers.TaskFilters{
		ProjectID:  getStringFlag(cmd, "project"),
		AssigneeID: getStringFlag(cmd, "assignee"),
		Query:      getStringFlag(cmd, "query"),
		Limit:      getIntFlag(cmd, "limit"),
		Offset:     getIntFlag(cmd, "offset"),
	}

	if status := getStringFlag(cmd, "status"); status != "" {
		filters.Status = []string{status}
	}
	if taskType := getStringFlag(cmd, "type"); taskType != "" {
		filters.Type = []string{taskType}
	}
	if priority := getStringFlag(cmd, "priority"); priority != "" {
		filters.Priority = []string{priority}
	}
	if labels, _ := cmd.Flags().GetStringSlice("labels"); len(labels) > 0 {
		filters.Labels = labels
	}

	// Determine target providers
	var targetProviders []string
	if len(providerNames) > 0 && providerNames[0] == "all" {
		enabledProviders := registry.ListEnabledProviders()
		for name := range enabledProviders {
			targetProviders = append(targetProviders, name)
		}
	} else if len(providerNames) > 0 {
		targetProviders = providerNames
	} else if providerName != "" {
		targetProviders = []string{providerName}
	} else {
		// Use default provider
		if defaultProvider, err := registry.GetDefaultProvider(); err == nil {
			info := defaultProvider.GetProviderInfo()
			targetProviders = []string{info.Name}
		}
	}

	// Collect tasks from all target providers
	var allTasks []*providers.UniversalTask
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for _, providerName := range targetProviders {
		provider, err := registry.GetProvider(providerName)
		if err != nil {
			logger.Warnf("Failed to get provider %s: %v", providerName, err)
			continue
		}

		tasks, err := provider.ListTasks(ctx, filters)
		if err != nil {
			logger.Warnf("Failed to list tasks from %s: %v", providerName, err)
			continue
		}

		// Set provider name for display
		for _, task := range tasks {
			task.ProviderName = providerName
		}

		allTasks = append(allTasks, tasks...)
	}

	// Output results
	switch output {
	case "json":
		return outputJSON(allTasks)
	case "yaml":
		return outputYAML(allTasks)
	default:
		return outputTaskTable(allTasks)
	}
}

func runGetTask(cmd *cobra.Command, args []string) error {
	search, _ := cmd.Flags().GetString("search")
	providerName, _ := cmd.Flags().GetString("provider")
	output, _ := cmd.Flags().GetString("output")

	if search != "" {
		return runSearchTasks(cmd, []string{search})
	}

	if len(args) == 0 {
		return fmt.Errorf("task ID is required")
	}

	taskID := args[0]

	// Get provider
	var provider providers.TaskProvider
	var err error

	if providerName != "" {
		provider, err = registry.GetProvider(providerName)
	} else {
		provider, err = registry.GetDefaultProvider()
	}

	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Get task
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	task, err := provider.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Output result
	switch output {
	case "json":
		return outputJSON(task)
	case "yaml":
		return outputYAML(task)
	default:
		return outputTaskDetails(task)
	}
}

func runUpdateTask(cmd *cobra.Command, args []string) error {
	taskID := args[0]
	providerName, _ := cmd.Flags().GetString("provider")

	// Get provider
	var provider providers.TaskProvider
	var err error

	if providerName != "" {
		provider, err = registry.GetProvider(providerName)
	} else {
		provider, err = registry.GetDefaultProvider()
	}

	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Build updates
	updates := &providers.TaskUpdate{}

	if title := getStringFlag(cmd, "title"); title != "" {
		updates.Title = &title
	}
	if description := getStringFlag(cmd, "description"); description != "" {
		updates.Description = &description
	}
	if status := getStringFlag(cmd, "status"); status != "" {
		taskStatus := providers.TaskStatus{
			ID:   strings.ToLower(strings.ReplaceAll(status, " ", "_")),
			Name: status,
		}
		updates.Status = &taskStatus
	}
	if priority := getStringFlag(cmd, "priority"); priority != "" {
		taskPriority := mapPriority(priority)
		updates.Priority = &taskPriority
	}
	if assignee := getStringFlag(cmd, "assignee"); assignee != "" {
		updates.AssigneeID = &assignee
	}

	// Handle labels
	if labels, _ := cmd.Flags().GetStringSlice("labels"); len(labels) > 0 {
		updates.Labels = labels
	}

	// TODO: Handle add-labels and remove-labels

	// Update task
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := provider.UpdateTask(ctx, taskID, updates); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	fmt.Printf("✅ Task %s updated successfully\n", taskID)
	return nil
}

func runDeleteTask(cmd *cobra.Command, args []string) error {
	taskID := args[0]
	providerName, _ := cmd.Flags().GetString("provider")
	force, _ := cmd.Flags().GetBool("force")

	// Get provider
	var provider providers.TaskProvider
	var err error

	if providerName != "" {
		provider, err = registry.GetProvider(providerName)
	} else {
		provider, err = registry.GetDefaultProvider()
	}

	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Confirmation
	if !force {
		fmt.Printf("Are you sure you want to delete task '%s'? (y/N): ", taskID)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	// Delete task
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := provider.DeleteTask(ctx, taskID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	fmt.Printf("✅ Task %s deleted successfully\n", taskID)
	return nil
}

func runSearchTasks(cmd *cobra.Command, args []string) error {
	var query string
	if len(args) > 0 {
		query = args[0]
	}
	if q := getStringFlag(cmd, "query"); q != "" {
		query = q
	}

	if query == "" {
		return fmt.Errorf("search query is required")
	}

	providerNames, _ := cmd.Flags().GetStringSlice("providers")
	output, _ := cmd.Flags().GetString("output")
	limit, _ := cmd.Flags().GetInt("limit")

	// Build search filters
	filters := &providers.TaskFilters{
		Query:  query,
		Limit:  limit,
		Status: getStringSliceFlag(cmd, "status"),
		Type:   getStringSliceFlag(cmd, "type"),
		Priority: getStringSliceFlag(cmd, "priority"),
	}

	if assignee := getStringFlag(cmd, "assignee"); assignee != "" {
		filters.AssigneeID = assignee
	}

	// Determine target providers
	var targetProviders []string
	if len(providerNames) > 0 && providerNames[0] == "all" {
		enabledProviders := registry.ListEnabledProviders()
		for name := range enabledProviders {
			targetProviders = append(targetProviders, name)
		}
	} else if len(providerNames) > 0 {
		targetProviders = providerNames
	} else {
		// Use default provider
		if defaultProvider, err := registry.GetDefaultProvider(); err == nil {
			info := defaultProvider.GetProviderInfo()
			targetProviders = []string{info.Name}
		}
	}

	// Search across providers
	var allTasks []*providers.UniversalTask
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for _, providerName := range targetProviders {
		provider, err := registry.GetProvider(providerName)
		if err != nil {
			logger.Warnf("Failed to get provider %s: %v", providerName, err)
			continue
		}

		tasks, err := provider.ListTasks(ctx, filters)
		if err != nil {
			logger.Warnf("Failed to search tasks in %s: %v", providerName, err)
			continue
		}

		for _, task := range tasks {
			task.ProviderName = providerName
		}

		allTasks = append(allTasks, tasks...)
	}

	fmt.Printf("Found %d tasks matching '%s'\n\n", len(allTasks), query)

	// Output results
	switch output {
	case "json":
		return outputJSON(allTasks)
	case "yaml":
		return outputYAML(allTasks)
	default:
		return outputTaskTable(allTasks)
	}
}

func runSyncTasks(cmd *cobra.Command, args []string) error {
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	project, _ := cmd.Flags().GetString("project")
	bidirectional, _ := cmd.Flags().GetBool("bidirectional")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// TODO: Implement task synchronization between providers
	// This would involve:
	// 1. Getting tasks from source provider
	// 2. Comparing with target provider
	// 3. Creating/updating tasks in target
	// 4. Handling conflicts
	// 5. Bidirectional sync if enabled

	fmt.Printf("Sync from %s to %s", from, to)
	if project != "" {
		fmt.Printf(" (project: %s)", project)
	}
	if bidirectional {
		fmt.Printf(" (bidirectional)")
	}
	if dryRun {
		fmt.Printf(" (dry run)")
	}
	fmt.Println()

	fmt.Println("🚧 Task synchronization not yet implemented")
	return nil
}

// Helper functions
func getStringFlag(cmd *cobra.Command, name string) string {
	value, _ := cmd.Flags().GetString(name)
	return value
}

func getIntFlag(cmd *cobra.Command, name string) int {
	value, _ := cmd.Flags().GetInt(name)
	return value
}

func getStringSliceFlag(cmd *cobra.Command, name string) []string {
	value, _ := cmd.Flags().GetStringSlice(name)
	if len(value) == 1 && value[0] == "" {
		return []string{}
	}
	return value
}

func mapPriority(priority string) providers.TaskPriority {
	switch strings.ToLower(priority) {
	case "lowest":
		return providers.TaskPriorityLowest
	case "low":
		return providers.TaskPriorityLow
	case "medium":
		return providers.TaskPriorityMedium
	case "high":
		return providers.TaskPriorityHigh
	case "highest":
		return providers.TaskPriorityHighest
	case "critical":
		return providers.TaskPriorityCritical
	default:
		return providers.TaskPriorityMedium
	}
}

func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func outputYAML(data interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(data)
}

func outputTaskTable(tasks []*providers.UniversalTask) error {
	fmt.Printf("%-15s %-12s %-40s %-12s %-10s %-15s\n", "ID", "PROVIDER", "TITLE", "STATUS", "PRIORITY", "ASSIGNEE")
	fmt.Printf("%-15s %-12s %-40s %-12s %-10s %-15s\n", "--", "--------", "-----", "------", "--------", "--------")

	for _, task := range tasks {
		title := task.Title
		if len(title) > 37 {
			title = title[:37] + "..."
		}

		assignee := task.AssigneeID
		if len(assignee) > 12 {
			assignee = assignee[:12] + "..."
		}

		fmt.Printf("%-15s %-12s %-40s %-12s %-10s %-15s\n",
			task.GetDisplayID(),
			task.ProviderName,
			title,
			task.Status.Name,
			string(task.Priority),
			assignee,
		)
	}

	return nil
}

func outputTaskDetails(task *providers.UniversalTask) error {
	fmt.Printf("Task Details\n")
	fmt.Printf("============\n\n")
	fmt.Printf("ID:           %s\n", task.GetDisplayID())
	fmt.Printf("Title:        %s\n", task.Title)
	fmt.Printf("Provider:     %s\n", task.ProviderName)
	fmt.Printf("Status:       %s\n", task.Status.Name)
	fmt.Printf("Priority:     %s\n", string(task.Priority))
	fmt.Printf("Type:         %s\n", string(task.Type))
	
	if task.AssigneeID != "" {
		fmt.Printf("Assignee:     %s\n", task.AssigneeID)
	}
	
	if task.ProjectID != "" {
		fmt.Printf("Project:      %s\n", task.ProjectID)
	}
	
	if len(task.Labels) > 0 {
		fmt.Printf("Labels:       %s\n", strings.Join(task.Labels, ", "))
	}
	
	fmt.Printf("Created:      %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated:      %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05"))
	
	if task.Description != "" {
		fmt.Printf("\nDescription:\n%s\n", task.Description)
	}

	return nil
}

// Bulk operation implementations

func runBulkCreateTasks(cmd *cobra.Command, args []string) error {
	fileName, _ := cmd.Flags().GetString("file")
	autoRoute, _ := cmd.Flags().GetBool("auto-route")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	providerName, _ := cmd.Flags().GetString("provider")
	
	// Read and parse file
	data, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", fileName, err)
	}
	
	var tasks []*providers.UniversalTask
	if strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml") {
		err = yaml.Unmarshal(data, &tasks)
	} else {
		err = json.Unmarshal(data, &tasks)
	}
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", fileName, err)
	}
	
	fmt.Printf("Found %d tasks to create\n", len(tasks))
	
	if dryRun {
		fmt.Println("\nDry run - would create the following tasks:")
		for i, task := range tasks {
			fmt.Printf("%d. %s (Project: %s, Type: %s)\n", i+1, task.Title, task.ProjectID, task.Type)
		}
		return nil
	}
	
	// Determine provider
	if !autoRoute && providerName == "" {
		return fmt.Errorf("either --provider or --auto-route must be specified")
	}
	
	var provider providers.TaskProvider
	if autoRoute {
		// TODO: Implement smart routing
		return fmt.Errorf("auto-routing not yet implemented")
	} else {
		p, err := registry.GetProvider(providerName)
		if err != nil {
			return fmt.Errorf("failed to get provider %s: %w", providerName, err)
		}
		provider = p
	}
	
	// Create tasks in batches
	ctx := context.Background()
	createdTasks, err := provider.BulkCreateTasks(ctx, tasks)
	if err != nil {
		return fmt.Errorf("failed to create tasks: %w", err)
	}
	
	fmt.Printf("Successfully created %d tasks\n", len(createdTasks))
	for _, task := range createdTasks {
		fmt.Printf("- %s: %s\n", task.GetDisplayID(), task.Title)
	}
	
	return nil
}

func runBulkUpdateTasks(cmd *cobra.Command, args []string) error {
	fileName, _ := cmd.Flags().GetString("file")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	providerName, _ := cmd.Flags().GetString("provider")
	
	if providerName == "" {
		return fmt.Errorf("--provider must be specified")
	}
	
	// Read and parse file
	data, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", fileName, err)
	}
	
	var updates map[string]*providers.TaskUpdate
	if strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml") {
		err = yaml.Unmarshal(data, &updates)
	} else {
		err = json.Unmarshal(data, &updates)
	}
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", fileName, err)
	}
	
	fmt.Printf("Found updates for %d tasks\n", len(updates))
	
	if dryRun {
		fmt.Println("\nDry run - would update the following tasks:")
		for taskID, update := range updates {
			fmt.Printf("- %s: ", taskID)
			parts := []string{}
			if update.Title != nil {
				parts = append(parts, fmt.Sprintf("title='%s'", *update.Title))
			}
			if update.Status != nil {
				parts = append(parts, fmt.Sprintf("status='%s'", update.Status.Name))
			}
			if update.Priority != nil {
				parts = append(parts, fmt.Sprintf("priority='%s'", *update.Priority))
			}
			fmt.Println(strings.Join(parts, ", "))
		}
		return nil
	}
	
	// Get provider
	provider, err := registry.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("failed to get provider %s: %w", providerName, err)
	}
	
	// Update tasks in batch
	ctx := context.Background()
	err = provider.BulkUpdateTasks(ctx, updates)
	if err != nil {
		return fmt.Errorf("failed to update tasks: %w", err)
	}
	
	fmt.Printf("Successfully updated %d tasks\n", len(updates))
	
	return nil
}

func runBulkDeleteTasks(cmd *cobra.Command, args []string) error {
	fileName, _ := cmd.Flags().GetString("file")
	idsStr, _ := cmd.Flags().GetString("ids")
	query, _ := cmd.Flags().GetString("query")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")
	providerName, _ := cmd.Flags().GetString("provider")
	
	if providerName == "" {
		return fmt.Errorf("--provider must be specified")
	}
	
	var taskIDs []string
	
	// Collect task IDs from different sources
	if fileName != "" {
		data, err := os.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", fileName, err)
		}
		taskIDs = strings.Split(strings.TrimSpace(string(data)), "\n")
	} else if idsStr != "" {
		taskIDs = strings.Split(idsStr, ",")
		for i, id := range taskIDs {
			taskIDs[i] = strings.TrimSpace(id)
		}
	} else if query != "" {
		// Get provider and search for tasks
		provider, err := registry.GetProvider(providerName)
		if err != nil {
			return fmt.Errorf("failed to get provider %s: %w", providerName, err)
		}
		
		ctx := context.Background()
		filters := &providers.TaskFilters{
			Query: query,
		}
		tasks, err := provider.ListTasks(ctx, filters)
		if err != nil {
			return fmt.Errorf("failed to search tasks: %w", err)
		}
		
		for _, task := range tasks {
			taskIDs = append(taskIDs, task.GetDisplayID())
		}
	} else {
		return fmt.Errorf("one of --file, --ids, or --query must be specified")
	}
	
	if len(taskIDs) == 0 {
		fmt.Println("No tasks found to delete")
		return nil
	}
	
	fmt.Printf("Found %d tasks to delete\n", len(taskIDs))
	
	if dryRun {
		fmt.Println("\nDry run - would delete the following tasks:")
		for _, taskID := range taskIDs {
			fmt.Printf("- %s\n", taskID)
		}
		return nil
	}
	
	// Confirmation unless force is used
	if !force {
		fmt.Printf("Are you sure you want to delete %d tasks? (y/N): ", len(taskIDs))
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}
	
	// Get provider
	provider, err := registry.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("failed to get provider %s: %w", providerName, err)
	}
	
	// Delete tasks
	ctx := context.Background()
	successCount := 0
	for _, taskID := range taskIDs {
		err := provider.DeleteTask(ctx, taskID)
		if err != nil {
			fmt.Printf("Failed to delete task %s: %v\n", taskID, err)
		} else {
			fmt.Printf("Deleted task %s\n", taskID)
			successCount++
		}
	}
	
	fmt.Printf("Successfully deleted %d out of %d tasks\n", successCount, len(taskIDs))
	
	return nil
}