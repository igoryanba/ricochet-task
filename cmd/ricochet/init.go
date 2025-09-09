package ricochet

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/grik-ai/ricochet-task/pkg/ai"
	"github.com/grik-ai/ricochet-task/pkg/key"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive setup for Ricochet-Task (like Claude CLI)",
	Long: `Initialize Ricochet-Task with interactive setup.
This command will guide you through:
- Setting up AI providers (BYOK or GRIK subscription)
- Configuring your first workflow
- Testing the connection
- Creating example chains

Just like Claude CLI, this makes getting started super simple.`,
	Run: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	printWelcome()
	
	// Check if already initialized
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("❌ Error getting home directory: %v\n", err)
		return
	}
	
	configDir := filepath.Join(homeDir, ".ricochet")
	if _, err := os.Stat(configDir); err == nil {
		fmt.Print("🔄 Ricochet-Task is already configured. Reconfigure? (y/N): ")
		if !askYesNo(false) {
			fmt.Println("✨ Run 'ricochet --help' to see available commands")
			return
		}
	}

	// Create config directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("❌ Error creating config directory: %v\n", err)
		return
	}

	// Step 1: Choose AI strategy
	aiStrategy := chooseAIStrategy()
	
	// Step 2: Setup API keys (if BYOK)
	var userKeys *ai.UserAPIKeys
	if aiStrategy == "byok" || aiStrategy == "hybrid" {
		userKeys = setupAPIKeys(configDir)
	}

	// Step 3: Test connection
	if !testConnection(aiStrategy, userKeys) {
		fmt.Println("❌ Setup failed. Please check your configuration.")
		return
	}

	// Step 4: Create example workflow
	createExampleWorkflow(configDir)

	// Step 5: Success!
	printSuccess()
}

func printWelcome() {
	fmt.Println("")
	fmt.Println(" ____  _                _          _     _____         _    ")
	fmt.Println("|  _ \\(_) ___ ___   ___| |__   ___| |_  |_   _|_ _ ___| | __")
	fmt.Println("| |_) | |/ __/ _ \\ / __| '_ \\ / _ \\ __|   | |/ _` / __| |/ /")
	fmt.Println("|  _ <| | (_| (_) | (__| | | |  __/ |_    | | (_| \\__ \\   < ")
	fmt.Println("|_| \\_\\_|\\___\\___/ \\___|_| |_|\\___|\\__|   |_|\\__,_|___/_|\\_\\")
	fmt.Println("")
	fmt.Println("🚀 Enterprise AI Workflow Orchestrator")
	fmt.Println("")
	fmt.Println("Welcome! Let's get you set up in under 2 minutes.")
	fmt.Println("")
}

func chooseAIStrategy() string {
	fmt.Println("🤖 How do you want to use AI models?")
	fmt.Println()
	fmt.Println("1. 🔐 BYOK (Bring Your Own Keys) - Use your OpenAI/Anthropic/etc. keys")
	fmt.Println("2. ☁️  GRIK Subscription - Use our managed AI service")
	fmt.Println("3. 🔀 Hybrid - Use your keys with GRIK fallback")
	fmt.Println()
	
	for {
		fmt.Print("Choose option (1/2/3): ")
		choice := askString("")
		
		switch choice {
		case "1":
			fmt.Println("✅ BYOK selected - You'll provide your own API keys")
			return "byok"
		case "2":
			fmt.Println("✅ GRIK Subscription selected - We'll handle AI for you")
			return "subscription"
		case "3":
			fmt.Println("✅ Hybrid selected - Best of both worlds")
			return "hybrid"
		default:
			fmt.Println("❌ Please choose 1, 2, or 3")
		}
	}
}

func setupAPIKeys(configDir string) *ai.UserAPIKeys {
	fmt.Println("\n🔑 Setting up your API keys...")
	fmt.Println("You can add these now or skip and add them later with 'ricochet keys add'")
	fmt.Println()

	keys := &ai.UserAPIKeys{}
	
	// OpenAI
	fmt.Print("🟢 OpenAI API Key (optional, press Enter to skip): ")
	if openaiKey := askString(""); openaiKey != "" {
		keys.OpenAI = &ai.APIKeyConfig{APIKey: openaiKey, Enabled: true}
		fmt.Println("✅ OpenAI key saved")
	}

	// Anthropic Claude
	fmt.Print("🟣 Anthropic API Key (optional, press Enter to skip): ")
	if anthropicKey := askString(""); anthropicKey != "" {
		keys.Anthropic = &ai.APIKeyConfig{APIKey: anthropicKey, Enabled: true}
		fmt.Println("✅ Anthropic key saved")
	}

	// DeepSeek
	fmt.Print("🔵 DeepSeek API Key (optional, press Enter to skip): ")
	if deepseekKey := askString(""); deepseekKey != "" {
		keys.DeepSeek = &ai.APIKeyConfig{APIKey: deepseekKey, Enabled: true}
		fmt.Println("✅ DeepSeek key saved")
	}

	// Save keys to file
	if keys.OpenAI != nil || keys.Anthropic != nil || keys.DeepSeek != nil {
		keyStore, err := key.NewFileKeyStore(configDir)
		if err != nil {
			fmt.Printf("❌ Error creating key store: %v\n", err)
			return keys
		}

		if keys.OpenAI != nil {
			keyStore.Add(key.Key{
				ID: "default-openai", 
				Provider: "openai", 
				Value: keys.OpenAI.APIKey,
				Name: "Default OpenAI",
			})
		}
		if keys.Anthropic != nil {
			keyStore.Add(key.Key{
				ID: "default-anthropic", 
				Provider: "anthropic", 
				Value: keys.Anthropic.APIKey,
				Name: "Default Anthropic",
			})
		}
		if keys.DeepSeek != nil {
			keyStore.Add(key.Key{
				ID: "default-deepseek", 
				Provider: "deepseek", 
				Value: keys.DeepSeek.APIKey,
				Name: "Default DeepSeek",
			})
		}
	}

	return keys
}

func testConnection(strategy string, userKeys *ai.UserAPIKeys) bool {
	fmt.Println("\n🔍 Testing AI connection...")
	
	// Create a simple logger
	logger := &SimpleLogger{}
	
	// Test based on strategy
	switch strategy {
	case "byok":
		if userKeys != nil && (userKeys.OpenAI != nil || userKeys.Anthropic != nil || userKeys.DeepSeek != nil) {
			// Test user keys
			hybridClient := ai.NewHybridAIClient("", "", "test-user", userKeys, logger)
			results := hybridClient.ValidateUserKeys(nil)
			
			hasValidKey := false
			for provider, err := range results {
				if err == nil {
					fmt.Printf("✅ %s connection successful\n", provider)
					hasValidKey = true
				} else {
					fmt.Printf("❌ %s connection failed: %v\n", provider, err)
				}
			}
			return hasValidKey
		} else {
			fmt.Println("⚠️  No API keys provided. You can add them later with 'ricochet keys add'")
			return true
		}
	case "subscription":
		fmt.Println("✅ GRIK subscription will be tested on first use")
		return true
	case "hybrid":
		fmt.Println("✅ Hybrid mode configured - will try user keys first, then GRIK")
		return true
	}
	
	return false
}

func createExampleWorkflow(configDir string) {
	fmt.Println("\n📝 Creating example workflow...")
	
	// Create example chain configuration
	exampleChain := `{
  "id": "welcome-chain",
  "name": "Welcome to Ricochet",
  "description": "A simple example chain to get you started",
  "models": [
    {
      "id": "welcome-model",
      "name": "gpt-4",
      "type": "openai",
      "role": "assistant",
      "prompt": "You are a helpful assistant. Respond with a warm welcome message and briefly explain what Ricochet-Task can do.",
      "max_tokens": 150,
      "temperature": 0.7
    }
  ]
}`

	chainFile := filepath.Join(configDir, "chains", "welcome.json")
	os.MkdirAll(filepath.Dir(chainFile), 0755)
	
	if err := os.WriteFile(chainFile, []byte(exampleChain), 0644); err == nil {
		fmt.Println("✅ Example chain created: welcome-chain")
		fmt.Println("   Try: ricochet chain run welcome-chain \"Hello!\"")
	}
}

func printSuccess() {
	fmt.Println("")
	fmt.Println("🎉 Setup complete! Ricochet-Task is ready to use.")
	fmt.Println("")
	fmt.Println("🚀 Quick start:")
	fmt.Println("   ricochet chain list              # See available chains")
	fmt.Println("   ricochet chain run welcome-chain \"Hello!\"  # Test your setup")
	fmt.Println("   ricochet workflow list           # See workflow templates")
	fmt.Println("")
	fmt.Println("📚 Learn more:")
	fmt.Println("   ricochet --help                  # All commands")
	fmt.Println("   ricochet docs                    # Documentation")
	fmt.Println("   ricochet examples                # Example workflows")
	fmt.Println("")
	fmt.Println("💡 Pro tip: Try 'ricochet chain create \"Build login API\"' to see AI in action!")
	fmt.Println("")
}

// Helper functions
func askString(prompt string) string {
	if prompt != "" {
		fmt.Print(prompt)
	}
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func askYesNo(defaultYes bool) bool {
	input := askString("")
	input = strings.ToLower(input)
	
	if input == "" {
		return defaultYes
	}
	
	return input == "y" || input == "yes"
}

// SimpleLogger for init command
type SimpleLogger struct{}

func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	// Silent for init
}

func (l *SimpleLogger) Error(msg string, err error, args ...interface{}) {
	// Silent for init unless needed
}

func (l *SimpleLogger) Warn(msg string, args ...interface{}) {
	// Silent for init
}

func (l *SimpleLogger) Debug(msg string, args ...interface{}) {
	// Silent for init
}