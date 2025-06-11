package ricochet_task

import (
	"github.com/spf13/cobra"
)

// Команда для управления задачами
var TaskCmd = &cobra.Command{
	Use:   "task",
	Short: "Управление задачами Ricochet Task",
	Long:  `Команды для создания и управления задачами в Ricochet Task.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
