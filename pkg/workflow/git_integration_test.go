package workflow

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// MockLogger для тестов
type MockLogger struct{}

func (ml *MockLogger) Debug(msg string, args ...interface{}) {}
func (ml *MockLogger) Info(msg string, args ...interface{})  {}
func (ml *MockLogger) Warn(msg string, args ...interface{})  {}
func (ml *MockLogger) Error(msg string, err error, args ...interface{}) {}

// TestGitProgressTrackerIntegration тестирует полную интеграцию Git tracking
func TestGitProgressTrackerIntegration(t *testing.T) {
	logger := &MockLogger{}
	tracker := NewGitProgressTracker(nil, logger)
	
	// Тестируем обработку push события
	t.Run("ProcessPushEvent", func(t *testing.T) {
		gitEvent := &GitProgressEvent{
			Type:       "push",
			Repository: "test/repo",
			Branch:     "feature/123",
			Commit: &CommitInfo{
				SHA:         "abc123",
				Message:     "feat: implement task #123 - new feature",
				Author:      "developer1",
				Timestamp:   time.Now(),
				FilesChanged: []string{"src/main.go", "test/main_test.go"},
				LinesAdded:   50,
				LinesDeleted: 10,
			},
			Data:      map[string]interface{}{},
			Timestamp: time.Now(),
		}
		
		err := tracker.ProcessGitEvent(context.Background(), gitEvent)
		if err != nil {
			t.Fatalf("Failed to process git event: %v", err)
		}
		
		// Проверяем, что прогресс был создан
		progress := tracker.GetTaskProgress("123")
		if progress == nil {
			t.Fatal("Expected task progress to be created")
		}
		
		if progress.TaskID != "123" {
			t.Errorf("Expected task ID 123, got %s", progress.TaskID)
		}
		
		if len(progress.GitActivity) != 1 {
			t.Errorf("Expected 1 git activity, got %d", len(progress.GitActivity))
		}
		
		if progress.ProgressPercent <= 0 {
			t.Errorf("Expected progress > 0, got %f", progress.ProgressPercent)
		}
	})
	
	// Тестируем обработку PR события
	t.Run("ProcessPullRequestEvent", func(t *testing.T) {
		gitEvent := &GitProgressEvent{
			Type:       "pull_request",
			Repository: "test/repo",
			Branch:     "feature/456",
			PullRequest: &PullRequestInfo{
				ID:           1,
				Title:        "Implement feature #456",
				State:        "opened",
				Author:       "developer2",
				SourceBranch: "feature/456",
				TargetBranch: "main",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			Data:      map[string]interface{}{},
			Timestamp: time.Now(),
		}
		
		err := tracker.ProcessGitEvent(context.Background(), gitEvent)
		if err != nil {
			t.Fatalf("Failed to process git event: %v", err)
		}
		
		progress := tracker.GetTaskProgress("456")
		if progress == nil {
			t.Fatal("Expected task progress to be created")
		}
		
		// PR должен установить прогресс минимум в 70%
		if progress.ProgressPercent < 70 {
			t.Errorf("Expected progress >= 70%% for PR, got %f", progress.ProgressPercent)
		}
		
		if progress.CurrentStage != "code_review" {
			t.Errorf("Expected stage 'code_review', got %s", progress.CurrentStage)
		}
	})
}

// TestCommitPatterns тестирует анализ паттернов в коммитах
func TestCommitPatterns(t *testing.T) {
	patterns := NewCommitPatterns()
	
	testCases := []struct {
		name        string
		message     string
		expectedIDs []string
		expectedType string
	}{
		{
			name:        "Feature commit with task ID",
			message:     "feat: implement new feature for task #123",
			expectedIDs: []string{"123"},
			expectedType: "feature",
		},
		{
			name:        "Bug fix with multiple formats",
			message:     "fix: resolve issue #456 and bug-789",
			expectedIDs: []string{"456", "789", "BUG-789"},
			expectedType: "fix",
		},
		{
			name:        "JIRA style ticket",
			message:     "PROJ-123: implement user authentication",
			expectedIDs: []string{"PROJ-123"},
			expectedType: "other",
		},
		{
			name:        "Multiple task references",
			message:     "refactor: update code for tasks #111, #222, and issue-333",
			expectedIDs: []string{"111", "222", "333", "ISSUE-333"},
			expectedType: "other",
		},
		{
			name:        "Test commit",
			message:     "test: add unit tests for user service",
			expectedIDs: []string{},
			expectedType: "test",
		},
		{
			name:        "Documentation",
			message:     "docs: update README with installation instructions",
			expectedIDs: []string{},
			expectedType: "docs",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Тестируем извлечение ID задач
			extractedIDs := patterns.ExtractTaskIDs(tc.message)
			if len(extractedIDs) != len(tc.expectedIDs) {
				t.Errorf("Expected %d task IDs, got %d: %v", 
					len(tc.expectedIDs), len(extractedIDs), extractedIDs)
			}
			
			for _, expectedID := range tc.expectedIDs {
				found := false
				for _, extractedID := range extractedIDs {
					if extractedID == expectedID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find task ID %s in %v", expectedID, extractedIDs)
				}
			}
			
			// Тестируем определение типа коммита
			commitType := patterns.GetCommitType(tc.message)
			if commitType != tc.expectedType {
				t.Errorf("Expected commit type %s, got %s", tc.expectedType, commitType)
			}
		})
	}
}

// TestBranchPatterns тестирует извлечение задач из названий веток
func TestBranchPatterns(t *testing.T) {
	patterns := NewCommitPatterns()
	
	testCases := []struct {
		branch      string
		expectedIDs []string
	}{
		{"feature/123", []string{"123"}},
		{"fix/456-bug-fix", []string{"456"}},
		{"PROJ-789", []string{"PROJ-789"}},
		{"bugfix/999-critical", []string{"999"}},
		{"task/111/implementation", []string{"111"}},
		{"main", []string{}},
		{"develop", []string{}},
	}
	
	for _, tc := range testCases {
		t.Run(tc.branch, func(t *testing.T) {
			extractedIDs := patterns.ExtractTaskIDsFromBranch(tc.branch)
			if len(extractedIDs) != len(tc.expectedIDs) {
				t.Errorf("Expected %d task IDs, got %d: %v", 
					len(tc.expectedIDs), len(extractedIDs), extractedIDs)
			}
			
			for _, expectedID := range tc.expectedIDs {
				found := false
				for _, extractedID := range extractedIDs {
					if extractedID == expectedID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find task ID %s in %v", expectedID, extractedIDs)
				}
			}
		})
	}
}

// TestWebhookHandling тестирует обработку webhook
func TestWebhookHandling(t *testing.T) {
	logger := &MockLogger{}
	tracker := NewGitProgressTracker(nil, logger)
	handler := tracker.GetWebhookHandler()
	
	// Тестируем GitHub webhook
	t.Run("GitHubPushWebhook", func(t *testing.T) {
		githubPayload := `{
			"ref": "refs/heads/feature/123",
			"repository": {
				"full_name": "test/repo"
			},
			"commits": [{
				"id": "abc123",
				"message": "feat: implement task #123",
				"author": {
					"username": "developer1"
				},
				"timestamp": "2024-01-01T12:00:00Z",
				"added": ["src/main.go"],
				"modified": ["README.md"]
			}]
		}`
		
		req := httptest.NewRequest("POST", "/webhook", strings.NewReader(githubPayload))
		req.Header.Set("X-GitHub-Event", "push")
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		handler.handleWebhook(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		// Проверяем, что прогресс был создан
		progress := tracker.GetTaskProgress("123")
		if progress == nil {
			t.Fatal("Expected task progress to be created from webhook")
		}
	})
	
	// Тестируем GitLab webhook
	t.Run("GitLabPushWebhook", func(t *testing.T) {
		gitlabPayload := `{
			"ref": "refs/heads/feature/456",
			"project": {
				"path_with_namespace": "test/repo"
			},
			"commits": [{
				"id": "def456",
				"message": "fix: resolve issue #456",
				"author": {
					"username": "developer2"
				},
				"timestamp": "2024-01-01T12:00:00Z"
			}]
		}`
		
		req := httptest.NewRequest("POST", "/webhook", strings.NewReader(gitlabPayload))
		req.Header.Set("X-Gitlab-Event", "Push Hook")
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		handler.handleWebhook(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		progress := tracker.GetTaskProgress("456")
		if progress == nil {
			t.Fatal("Expected task progress to be created from GitLab webhook")
		}
	})
}

// TestProgressEngine тестирует движок прогресса
func TestProgressEngine(t *testing.T) {
	logger := &MockLogger{}
	tracker := NewGitProgressTracker(nil, logger)
	engine := tracker.progressEngine
	
	t.Run("SaveAndLoadProgress", func(t *testing.T) {
		progress := &TaskProgress{
			TaskID:          "TEST-001",
			CurrentStage:    "development",
			CompletedStages: []string{"planning"},
			ProgressPercent: 25.0,
			LastActivity:    time.Now(),
			GitActivity:     []GitActivity{},
			Metrics:         &ProgressMetrics{TotalCommits: 3},
		}
		
		err := engine.SaveTaskProgress(progress)
		if err != nil {
			t.Fatalf("Failed to save progress: %v", err)
		}
		
		loaded := engine.GetTaskProgress("TEST-001")
		if loaded == nil {
			t.Fatal("Failed to load saved progress")
		}
		
		if loaded.TaskID != progress.TaskID {
			t.Errorf("Expected TaskID %s, got %s", progress.TaskID, loaded.TaskID)
		}
		
		if loaded.ProgressPercent != progress.ProgressPercent {
			t.Errorf("Expected progress %f, got %f", progress.ProgressPercent, loaded.ProgressPercent)
		}
	})
	
	t.Run("StageStatistics", func(t *testing.T) {
		// Добавляем несколько задач в разных этапах
		tasks := []*TaskProgress{
			{TaskID: "T1", CurrentStage: "development", ProgressPercent: 30},
			{TaskID: "T2", CurrentStage: "development", ProgressPercent: 50},
			{TaskID: "T3", CurrentStage: "testing", ProgressPercent: 80},
			{TaskID: "T4", CurrentStage: "completed", ProgressPercent: 100},
		}
		
		for _, task := range tasks {
			engine.SaveTaskProgress(task)
		}
		
		stats := engine.GetStageStatistics()
		
		if stats["development"].Count != 2 {
			t.Errorf("Expected 2 tasks in development, got %d", stats["development"].Count)
		}
		
		if stats["testing"].Count != 1 {
			t.Errorf("Expected 1 task in testing, got %d", stats["testing"].Count)
		}
		
		if stats["completed"].Count != 1 {
			t.Errorf("Expected 1 task completed, got %d", stats["completed"].Count)
		}
	})
	
	t.Run("ProgressReport", func(t *testing.T) {
		report := engine.GenerateProgressReport()
		
		if report.TotalTasks == 0 {
			t.Error("Expected some tasks in progress report")
		}
		
		if report.GeneratedAt.IsZero() {
			t.Error("Expected valid generated timestamp")
		}
		
		if report.Summary == nil {
			t.Error("Expected summary to be present")
		}
	})
}

// TestCommitComplexity тестирует анализ сложности коммитов
func TestCommitComplexity(t *testing.T) {
	patterns := NewCommitPatterns()
	
	testCases := []struct {
		name         string
		message      string
		files        []string
		linesChanged int
		expectedComplexity string
	}{
		{
			name:         "Simple feature",
			message:      "feat: add simple function",
			files:        []string{"src/utils.go"},
			linesChanged: 20,
			expectedComplexity: "Feature implementation",
		},
		{
			name:         "Complex refactoring",
			message:      "refactor: restructure core engine",
			files:        []string{"src/main.go", "src/engine.go", "src/core.go"},
			linesChanged: 500,
			expectedComplexity: "Code refactoring",
		},
		{
			name:         "Bug fix with tests",
			message:      "fix: resolve memory leak",
			files:        []string{"src/memory.go", "test/memory_test.go"},
			linesChanged: 50,
			expectedComplexity: "Bug fix",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			complexity := patterns.GetCommitComplexity(tc.message, tc.files, tc.linesChanged)
			
			if !strings.Contains(complexity.Description, tc.expectedComplexity) {
				t.Errorf("Expected complexity description to contain '%s', got '%s'", 
					tc.expectedComplexity, complexity.Description)
			}
			
			if complexity.Score <= 0 {
				t.Errorf("Expected positive complexity score, got %f", complexity.Score)
			}
			
			if complexity.Type == "" {
				t.Error("Expected non-empty complexity type")
			}
		})
	}
}

// BenchmarkPatternMatching бенчмарк для анализа паттернов
func BenchmarkPatternMatching(b *testing.B) {
	patterns := NewCommitPatterns()
	testMessage := "feat(auth): implement user authentication for task #123 and issue-456"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patterns.ExtractTaskIDs(testMessage)
		patterns.GetCommitType(testMessage)
	}
}

// BenchmarkProgressUpdate бенчмарк для обновления прогресса
func BenchmarkProgressUpdate(b *testing.B) {
	logger := &MockLogger{}
	tracker := NewGitProgressTracker(nil, logger)
	
	gitEvent := &GitProgressEvent{
		Type:       "push",
		Repository: "test/repo",
		Branch:     "feature/bench",
		Commit: &CommitInfo{
			SHA:         "bench123",
			Message:     "feat: benchmark test #999",
			Author:      "benchmarker",
			Timestamp:   time.Now(),
			FilesChanged: []string{"bench.go"},
			LinesAdded:   10,
			LinesDeleted: 5,
		},
		Data:      map[string]interface{}{},
		Timestamp: time.Now(),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.ProcessGitEvent(context.Background(), gitEvent)
	}
}