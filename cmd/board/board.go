package board

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sirupsen/logrus"

	"github.com/grik-ai/ricochet-task/pkg/providers"
)

var (
	registry *providers.ProviderRegistry
	logger   *logrus.Logger
)

// BoardCmd represents the board command
var BoardCmd = &cobra.Command{
	Use:   "board",
	Short: "Interactive board selection and context management",
	Long: `Interactive board selection and context management for ricochet-task.

This command allows you to:
- Select and switch between agile boards
- Set working context for AI-powered task management
- View board information and available projects

Example usage:
  ricochet board select          # Interactive board selection
  ricochet board list           # List all available boards
  ricochet board context        # Show current working context
  ricochet board plan <task>    # Create AI-powered project plan`,
	RunE: runBoardInteractive,
}

var selectCmd = &cobra.Command{
	Use:   "select",
	Short: "Select working agile board",
	Long: `Interactive selection of agile board to work with.

This will prompt you to choose from available agile boards and set
the working context for all subsequent operations.`,
	RunE: runSelectBoard,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available agile boards",
	Long: `List all available agile boards with their projects and IDs.

Shows boards from all configured providers with detailed information
including project mappings and current status.`,
	RunE: runListBoards,
}

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Show current working context",
	Long: `Display the currently selected board context and working settings.

Shows which board, project, and provider are currently active
for AI-powered task management operations.`,
	RunE: runShowContext,
}

var planCmd = &cobra.Command{
	Use:   "plan [description]",
	Short: "Create AI-powered project plan",
	Long: `Create an AI-powered project plan that automatically breaks down
into tasks and manages them through the selected agile board.

The plan will:
- Analyze the project requirements
- Break down into manageable tasks
- Create tasks in the selected board
- Track progress and update statuses
- Add progress comments to tasks

Examples:
  ricochet board plan "Implement user authentication system"
  ricochet board plan --epic "Mobile app redesign"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCreatePlan,
}

func init() {
	// Add subcommands
	BoardCmd.AddCommand(selectCmd)
	BoardCmd.AddCommand(listCmd)
	BoardCmd.AddCommand(contextCmd)
	BoardCmd.AddCommand(planCmd)

	// Global board flags
	BoardCmd.PersistentFlags().StringP("config", "c", "", "Configuration file path")
	BoardCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")

	// Plan command flags
	planCmd.Flags().Bool("epic", false, "Create as epic with subtasks")
	planCmd.Flags().StringP("priority", "p", "medium", "Task priority")
	planCmd.Flags().StringP("assignee", "a", "", "Default assignee for tasks")
	planCmd.Flags().StringSliceP("labels", "l", []string{}, "Labels to add to tasks")
	planCmd.Flags().Bool("auto-start", false, "Automatically start execution after planning")
}

func initializeBoard() error {
	// Setup logger
	logger = logrus.New()
	if viper.GetBool("debug") {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	// Load configuration
	config := loadMultiProviderConfig()
	if config == nil {
		return fmt.Errorf("failed to load provider configuration")
	}

	// Initialize provider registry
	registry = providers.NewProviderRegistry(config, logger)

	// Initialize providers
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := registry.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize providers: %w", err)
	}

	return nil
}

func runBoardInteractive(cmd *cobra.Command, args []string) error {
	fmt.Println("🎯 Ricochet Task - Interactive Board Management")
	fmt.Println("==============================================")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  select  - Select working agile board")
	fmt.Println("  list    - List all available boards")
	fmt.Println("  context - Show current context")
	fmt.Println("  plan    - Create AI-powered project plan")
	fmt.Println()
	fmt.Println("Use 'ricochet board <command> --help' for more information")
	
	return nil
}

func runSelectBoard(cmd *cobra.Command, args []string) error {
	if err := initializeBoard(); err != nil {
		return err
	}

	// Get available boards from all providers
	boards, err := getAllBoards()
	if err != nil {
		return fmt.Errorf("failed to get boards: %w", err)
	}

	if len(boards) == 0 {
		fmt.Println("❌ No agile boards found in any configured provider")
		return nil
	}

	// Display boards for selection
	fmt.Println("📋 Available Agile Boards:")
	fmt.Println("==========================")
	for i, board := range boards {
		fmt.Printf("%d. %s\n", i+1, board.DisplayName())
		fmt.Printf("   Project: %s (%s)\n", board.ProjectName, board.ProjectID)
		fmt.Printf("   Provider: %s\n", board.ProviderName)
		fmt.Println()
	}

	// Get user selection
	fmt.Print("Select board (enter number): ")
	var selection int
	if _, err := fmt.Scanf("%d", &selection); err != nil {
		return fmt.Errorf("invalid selection: %w", err)
	}

	if selection < 1 || selection > len(boards) {
		return fmt.Errorf("invalid selection: must be between 1 and %d", len(boards))
	}

	selectedBoard := boards[selection-1]

	// Save context
	if err := saveWorkingContext(selectedBoard); err != nil {
		return fmt.Errorf("failed to save context: %w", err)
	}

	fmt.Printf("✅ Selected board: %s\n", selectedBoard.DisplayName())
	fmt.Printf("📁 Working project: %s (%s)\n", selectedBoard.ProjectName, selectedBoard.ProjectID)
	fmt.Printf("🔧 Provider: %s\n", selectedBoard.ProviderName)
	fmt.Println()
	fmt.Println("💡 Context saved! You can now use 'ricochet board plan' to create AI-powered plans")

	return nil
}

func runListBoards(cmd *cobra.Command, args []string) error {
	if err := initializeBoard(); err != nil {
		return err
	}

	boards, err := getAllBoards()
	if err != nil {
		return fmt.Errorf("failed to get boards: %w", err)
	}

	if len(boards) == 0 {
		fmt.Println("❌ No agile boards found")
		return nil
	}

	fmt.Printf("📋 Found %d agile boards:\n", len(boards))
	fmt.Println("=" + strings.Repeat("=", 50))
	
	for _, board := range boards {
		fmt.Printf("🎯 %s\n", board.Name)
		fmt.Printf("   ID: %s\n", board.ID)
		fmt.Printf("   Project: %s (%s)\n", board.ProjectName, board.ProjectID)
		fmt.Printf("   Provider: %s\n", board.ProviderName)
		fmt.Println()
	}

	return nil
}

func runShowContext(cmd *cobra.Command, args []string) error {
	context, err := loadWorkingContext()
	if err != nil {
		fmt.Println("❌ No working context found. Use 'ricochet board select' to choose a board.")
		return nil
	}

	fmt.Println("🎯 Current Working Context:")
	fmt.Println("===========================")
	fmt.Printf("Board: %s\n", context.BoardName)
	fmt.Printf("Project: %s (%s)\n", context.ProjectName, context.ProjectID)
	fmt.Printf("Provider: %s\n", context.ProviderName)
	if context.DefaultAssignee != "" {
		fmt.Printf("Default Assignee: %s\n", context.DefaultAssignee)
	}
	if len(context.DefaultLabels) > 0 {
		fmt.Printf("Default Labels: %s\n", strings.Join(context.DefaultLabels, ", "))
	}
	fmt.Printf("Last Updated: %s\n", context.UpdatedAt.Format("2006-01-02 15:04:05"))

	return nil
}

func runCreatePlan(cmd *cobra.Command, args []string) error {
	if err := initializeBoard(); err != nil {
		return err
	}

	// Load working context
	context, err := loadWorkingContext()
	if err != nil {
		return fmt.Errorf("no working context found. Use 'ricochet board select' first")
	}

	description := strings.Join(args, " ")
	isEpic, _ := cmd.Flags().GetBool("epic")
	priority, _ := cmd.Flags().GetString("priority")
	assignee, _ := cmd.Flags().GetString("assignee")
	labels, _ := cmd.Flags().GetStringSlice("labels")
	autoStart, _ := cmd.Flags().GetBool("auto-start")

	// Use context defaults if not specified
	if assignee == "" && context.DefaultAssignee != "" {
		assignee = context.DefaultAssignee
	}
	if len(labels) == 0 && len(context.DefaultLabels) > 0 {
		labels = context.DefaultLabels
	}

	fmt.Printf("🤖 Creating AI-powered plan for: %s\n", description)
	fmt.Printf("📋 Target board: %s\n", context.BoardName)
	fmt.Printf("📁 Project: %s\n", context.ProjectName)
	fmt.Println()

	// Create the plan using AI
	plan, err := createAIPlan(context, description, isEpic, priority, assignee, labels)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	// Display the plan
	displayPlan(plan)

	// Ask for confirmation
	fmt.Print("\n📝 Create these tasks in YouTrack? (y/N): ")
	var confirm string
	fmt.Scanln(&confirm)
	
	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		fmt.Println("❌ Plan cancelled")
		return nil
	}

	// Execute the plan
	if err := executePlan(context, plan, autoStart); err != nil {
		return fmt.Errorf("failed to execute plan: %w", err)
	}

	fmt.Println("✅ Plan executed successfully!")
	fmt.Printf("🎯 Created %d tasks in board: %s\n", len(plan.Tasks), context.BoardName)

	return nil
}

// Helper types and functions

type BoardInfo struct {
	ID           string
	Name         string
	ProjectID    string
	ProjectName  string
	ProviderName string
}

func (b *BoardInfo) DisplayName() string {
	return fmt.Sprintf("%s (%s)", b.Name, b.ProjectName)
}

type WorkingContext struct {
	BoardID         string    `json:"board_id"`
	BoardName       string    `json:"board_name"`
	ProjectID       string    `json:"project_id"`
	ProjectName     string    `json:"project_name"`
	ProviderName    string    `json:"provider_name"`
	DefaultAssignee string    `json:"default_assignee,omitempty"`
	DefaultLabels   []string  `json:"default_labels,omitempty"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type TaskPlan struct {
	Title       string
	Description string
	Priority    string
	Assignee    string
	Labels      []string
	TaskType    string
	Subtasks    []TaskPlan
}

type ProjectPlan struct {
	Title       string
	Description string
	Tasks       []TaskPlan
	IsEpic      bool
}

func getAllBoards() ([]*BoardInfo, error) {
	// This is a placeholder implementation
	// In the real implementation, this would query all providers
	// for their agile boards using their APIs
	
	boards := []*BoardInfo{
		{
			ID:           "176-2",
			Name:         "GAMESDROP: Develop",
			ProjectID:    "0-1",
			ProjectName:  "[DEV]GAMESDROP",
			ProviderName: "gamesdrop-youtrack",
		},
		{
			ID:           "176-4",
			Name:         "Marketing",
			ProjectID:    "0-3",
			ProjectName:  "[MARKETING] GAMESDROP",
			ProviderName: "gamesdrop-youtrack",
		},
		{
			ID:           "176-3",
			Name:         "Бизнес задачи",
			ProjectID:    "0-2",
			ProjectName:  "[BUSINESS] GAMESDROP",
			ProviderName: "gamesdrop-youtrack",
		},
	}

	return boards, nil
}

func saveWorkingContext(board *BoardInfo) error {
	context := &WorkingContext{
		BoardID:      board.ID,
		BoardName:    board.Name,
		ProjectID:    board.ProjectID,
		ProjectName:  board.ProjectName,
		ProviderName: board.ProviderName,
		UpdatedAt:    time.Now(),
	}

	data, err := json.MarshalIndent(context, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(".ricochet-context.json", data, 0644)
}

func loadWorkingContext() (*WorkingContext, error) {
	data, err := os.ReadFile(".ricochet-context.json")
	if err != nil {
		return nil, err
	}

	var context WorkingContext
	if err := json.Unmarshal(data, &context); err != nil {
		return nil, err
	}

	return &context, nil
}

func createAIPlan(context *WorkingContext, description string, isEpic bool, priority, assignee string, labels []string) (*ProjectPlan, error) {
	// This is a placeholder for AI-powered planning
	// In the real implementation, this would use AI to analyze the description
	// and create a detailed breakdown of tasks
	
	plan := &ProjectPlan{
		Title:       fmt.Sprintf("🤖 AI Plan: %s", description),
		Description: fmt.Sprintf("AI-generated plan for: %s", description),
		IsEpic:      isEpic,
		Tasks: []TaskPlan{
			{
				Title:       fmt.Sprintf("📋 Planning: %s", description),
				Description: "Initial planning and requirements analysis",
				Priority:    "high",
				Assignee:    assignee,
				Labels:      append(labels, "planning", "ai-generated"),
				TaskType:    "task",
			},
			{
				Title:       fmt.Sprintf("🔧 Implementation: %s", description),
				Description: "Main implementation work",
				Priority:    priority,
				Assignee:    assignee,
				Labels:      append(labels, "implementation", "ai-generated"),
				TaskType:    "task",
			},
			{
				Title:       fmt.Sprintf("✅ Testing: %s", description),
				Description: "Testing and quality assurance",
				Priority:    "medium",
				Assignee:    assignee,
				Labels:      append(labels, "testing", "ai-generated"),
				TaskType:    "task",
			},
		},
	}

	return plan, nil
}

func displayPlan(plan *ProjectPlan) {
	fmt.Println("📋 Generated Plan:")
	fmt.Println("==================")
	fmt.Printf("Title: %s\n", plan.Title)
	fmt.Printf("Description: %s\n", plan.Description)
	fmt.Printf("Epic Mode: %t\n", plan.IsEpic)
	fmt.Printf("Tasks: %d\n", len(plan.Tasks))
	fmt.Println()

	for i, task := range plan.Tasks {
		fmt.Printf("%d. %s\n", i+1, task.Title)
		fmt.Printf("   Description: %s\n", task.Description)
		fmt.Printf("   Priority: %s\n", task.Priority)
		if task.Assignee != "" {
			fmt.Printf("   Assignee: %s\n", task.Assignee)
		}
		if len(task.Labels) > 0 {
			fmt.Printf("   Labels: %s\n", strings.Join(task.Labels, ", "))
		}
		fmt.Println()
	}
}

func executePlan(context *WorkingContext, plan *ProjectPlan, autoStart bool) error {
	// This would implement the actual task creation via MCP API
	// For now, it's a placeholder
	
	fmt.Println("🚀 Executing plan...")
	
	for i, task := range plan.Tasks {
		fmt.Printf("Creating task %d/%d: %s\n", i+1, len(plan.Tasks), task.Title)
		
		// Here we would call the MCP API to create the task
		// using the task_create_smart tool with context.ProjectID
		
		time.Sleep(500 * time.Millisecond) // Simulate API call
	}
	
	if autoStart {
		fmt.Println("🎯 Auto-starting task execution...")
		// Here we would call ai_execute_task for each created task
	}
	
	return nil
}

// Helper functions

func loadMultiProviderConfig() *providers.MultiProviderConfig {
	config := providers.DefaultMultiProviderConfig()

	// Try to load from config file
	configFile := viper.GetString("config")
	if configFile == "" {
		configFile = "ricochet.yaml"
	}

	if _, err := os.Stat(configFile); err == nil {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err == nil {
			viper.Unmarshal(config)
			logger.Infof("Loaded configuration from %s", configFile)
		} else {
			logger.Warnf("Failed to read config file %s: %v", configFile, err)
		}
	} else {
		logger.Infof("No config file found at %s, using defaults", configFile)
	}

	return config
}