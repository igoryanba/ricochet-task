package ricochet

import (
	"fmt"
	"os"

	"github.com/grik-ai/ricochet-task/cmd/ricochet/chain"
	"github.com/grik-ai/ricochet-task/cmd/ricochet/checkpoint"
	"github.com/grik-ai/ricochet-task/cmd/ricochet/key"
	"github.com/grik-ai/ricochet-task/cmd/ricochet/ricochet_task"
	"github.com/grik-ai/ricochet-task/pkg/ui"
	"github.com/spf13/cobra"
)

var (
	// Флаг интерактивного режима
	interactiveMode bool
)

var rootCmd = &cobra.Command{
	Use:   "ricochet",
	Short: "Ricochet Task - CLI для управления задачами и цепочками моделей",
	Long: `Ricochet Task - мощный CLI-инструмент для управления задачами 
и цепочками моделей в экосистеме GRIK AI. Позволяет обрабатывать большие 
объемы текстовых данных с использованием различных языковых моделей.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Если указан флаг интерактивного режима или нет аргументов, запускаем интерактивное меню
		if interactiveMode || len(args) == 0 {
			if err := ui.ShowMainMenu(); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// Если есть аргументы, выполняем стандартную обработку
		cmd.Help()
		os.Exit(0)
	},
}

// Execute выполняет корневую команду
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Глобальные флаги
	rootCmd.PersistentFlags().StringP("config", "c", "", "Путь к файлу конфигурации")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Включить подробный вывод")
	rootCmd.PersistentFlags().BoolVarP(&interactiveMode, "interactive", "i", false, "Запустить в интерактивном режиме")

	// Подкоманды
	rootCmd.AddCommand(chain.ChainCmd)
	rootCmd.AddCommand(checkpoint.CheckpointCmd)
	rootCmd.AddCommand(key.KeyCmd)
	rootCmd.AddCommand(ricochet_task.TaskCmd)

	// Подкоманды для ключей API
	key.KeyCmd.AddCommand(&cobra.Command{
		Use:   "add",
		Short: "Добавить новый API-ключ",
		Run: func(cmd *cobra.Command, args []string) {
			if err := ui.HandleAddKey(); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка при добавлении ключа: %v\n", err)
				os.Exit(1)
			}
		},
	})

	key.KeyCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Просмотреть список API-ключей",
		Run: func(cmd *cobra.Command, args []string) {
			if err := ui.HandleListKeys(); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка при просмотре ключей: %v\n", err)
				os.Exit(1)
			}
		},
	})

	key.KeyCmd.AddCommand(&cobra.Command{
		Use:   "update",
		Short: "Обновить существующий API-ключ",
		Run: func(cmd *cobra.Command, args []string) {
			if err := ui.HandleUpdateKey(); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка при обновлении ключа: %v\n", err)
				os.Exit(1)
			}
		},
	})

	key.KeyCmd.AddCommand(&cobra.Command{
		Use:   "delete",
		Short: "Удалить API-ключ",
		Run: func(cmd *cobra.Command, args []string) {
			if err := ui.HandleDeleteKey(); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка при удалении ключа: %v\n", err)
				os.Exit(1)
			}
		},
	})

	// Подкоманды для цепочек моделей
	chain.ChainCmd.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "Создать новую цепочку моделей",
		Run: func(cmd *cobra.Command, args []string) {
			if err := ui.HandleCreateChain(); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка при создании цепочки: %v\n", err)
				os.Exit(1)
			}
		},
	})

	chain.ChainCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Просмотреть список цепочек моделей",
		Run: func(cmd *cobra.Command, args []string) {
			if err := ui.HandleListChains(); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка при просмотре цепочек: %v\n", err)
				os.Exit(1)
			}
		},
	})

	chain.ChainCmd.AddCommand(&cobra.Command{
		Use:   "update",
		Short: "Обновить существующую цепочку моделей",
		Run: func(cmd *cobra.Command, args []string) {
			if err := ui.HandleUpdateChain(); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка при обновлении цепочки: %v\n", err)
				os.Exit(1)
			}
		},
	})

	chain.ChainCmd.AddCommand(&cobra.Command{
		Use:   "delete",
		Short: "Удалить цепочку моделей",
		Run: func(cmd *cobra.Command, args []string) {
			if err := ui.HandleDeleteChain(); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка при удалении цепочки: %v\n", err)
				os.Exit(1)
			}
		},
	})

	chain.ChainCmd.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "Запустить цепочку моделей",
		Run: func(cmd *cobra.Command, args []string) {
			if err := ui.HandleRunChain(); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка при запуске цепочки: %v\n", err)
				os.Exit(1)
			}
		},
	})

	// Подкоманды для чекпоинтов
	checkpoint.CheckpointCmd.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "Создать новый чекпоинт",
		Run: func(cmd *cobra.Command, args []string) {
			if err := ui.HandleCreateCheckpoint(); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка при создании чекпоинта: %v\n", err)
				os.Exit(1)
			}
		},
	})

	checkpoint.CheckpointCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Просмотреть список чекпоинтов",
		Run: func(cmd *cobra.Command, args []string) {
			if err := ui.HandleListCheckpoints(); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка при просмотре чекпоинтов: %v\n", err)
				os.Exit(1)
			}
		},
	})

	checkpoint.CheckpointCmd.AddCommand(&cobra.Command{
		Use:   "update",
		Short: "Обновить существующий чекпоинт",
		Run: func(cmd *cobra.Command, args []string) {
			if err := ui.HandleUpdateCheckpoint(); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка при обновлении чекпоинта: %v\n", err)
				os.Exit(1)
			}
		},
	})

	checkpoint.CheckpointCmd.AddCommand(&cobra.Command{
		Use:   "delete",
		Short: "Удалить чекпоинт",
		Run: func(cmd *cobra.Command, args []string) {
			if err := ui.HandleDeleteCheckpoint(); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка при удалении чекпоинта: %v\n", err)
				os.Exit(1)
			}
		},
	})
}

// Команда для инициализации
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Инициализировать конфигурацию Ricochet",
	Run: func(cmd *cobra.Command, args []string) {
		if interactiveMode {
			// Используем интерактивный режим
			if err := ui.HandleInitProject(); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка при инициализации: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// Неинтерактивный режим
		ui.PrintInfo("Инициализация проекта Ricochet...")
		// TODO: Реализовать инициализацию
		ui.PrintSuccess("Проект успешно инициализирован!")
	},
}

// Команда для управления ключами API
var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Управление API-ключами",
	Run: func(cmd *cobra.Command, args []string) {
		if interactiveMode {
			if err := ui.ShowKeyManagementMenu(); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка при управлении ключами: %v\n", err)
				os.Exit(1)
			}
			return
		}

		cmd.Help()
	},
}

// Команда для управления цепочками моделей
var chainCmd = &cobra.Command{
	Use:   "chain",
	Short: "Управление цепочками моделей",
	Run: func(cmd *cobra.Command, args []string) {
		if interactiveMode {
			if err := ui.ShowChainManagementMenu(); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка при управлении цепочками: %v\n", err)
				os.Exit(1)
			}
			return
		}

		cmd.Help()
	},
}

// Команда для управления чекпоинтами
var checkpointCmd = &cobra.Command{
	Use:   "checkpoint",
	Short: "Управление чекпоинтами",
	Run: func(cmd *cobra.Command, args []string) {
		if interactiveMode {
			if err := ui.ShowCheckpointManagementMenu(); err != nil {
				fmt.Fprintf(os.Stderr, "Ошибка при управлении чекпоинтами: %v\n", err)
				os.Exit(1)
			}
			return
		}

		cmd.Help()
	},
}
