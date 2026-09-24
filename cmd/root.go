package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/benjuh/stew/cmd/add"
	"github.com/benjuh/stew/cmd/browse"
	"github.com/benjuh/stew/cmd/completion"
	"github.com/benjuh/stew/cmd/diff"
	"github.com/benjuh/stew/cmd/doctor"
	"github.com/benjuh/stew/cmd/edit"
	"github.com/benjuh/stew/cmd/export"
	"github.com/benjuh/stew/cmd/get"
	importcmd "github.com/benjuh/stew/cmd/importcmd"
	"github.com/benjuh/stew/cmd/info"
	"github.com/benjuh/stew/cmd/install"
	"github.com/benjuh/stew/cmd/list"
	"github.com/benjuh/stew/cmd/new"
	"github.com/benjuh/stew/cmd/outdated"
	"github.com/benjuh/stew/cmd/remove"
	"github.com/benjuh/stew/cmd/replace"
	"github.com/benjuh/stew/cmd/run"
	"github.com/benjuh/stew/cmd/search"
	cmdtasks "github.com/benjuh/stew/cmd/tasks"
	"github.com/benjuh/stew/cmd/update"
	"github.com/benjuh/stew/cmd/upgrade"
	"github.com/benjuh/stew/cmd/validate"
	"github.com/benjuh/stew/cmd/verify"
	"github.com/benjuh/stew/util"
)

var (
	cfgFile string

	// Version is set at build time with -ldflags and defaults to dev.
	Version = "dev"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "stew",
	Short: "A project template manager",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		flag, _ := cmd.Flags().GetBool("version")
		if flag {
			fmt.Println(Version)
			os.Exit(0)
		}

		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func setDefaults() {
	viper.SetDefault("stewsPath", util.GetHomeDir()+"/.stews.json")
	viper.SetDefault("timeFormat", "2006-01-02 15:04:05")
	viper.SetDefault("templatesPath", util.GetHomeDir()+"/.config/stew/templates")
	viper.SetDefault("catalogURL", "https://benjuh.com/stew/catalog/index.json")

}

func addSubCommands() {
	rootCmd.AddCommand(add.SaveCmd)
	rootCmd.AddCommand(browse.BrowseCmd)
	rootCmd.AddCommand(completion.CompletionCmd)
	rootCmd.AddCommand(info.InfoCmd)
	rootCmd.AddCommand(edit.EditCmd)
	rootCmd.AddCommand(list.ListCmd)
	rootCmd.AddCommand(remove.RemoveCmd)
	rootCmd.AddCommand(new.CreateCmd)
	rootCmd.AddCommand(get.ViewCmd)
	rootCmd.AddCommand(replace.ReplaceCmd)
	rootCmd.AddCommand(validate.ValidateCmd)
	rootCmd.AddCommand(export.ExportCmd)
	rootCmd.AddCommand(importcmd.ImportCmd)
	rootCmd.AddCommand(install.InstallCmd)
	rootCmd.AddCommand(update.UpdateCmd)
	rootCmd.AddCommand(upgrade.UpgradeCmd)
	rootCmd.AddCommand(outdated.OutdatedCmd)
	rootCmd.AddCommand(diff.DiffCmd)
	rootCmd.AddCommand(doctor.DoctorCmd)
	rootCmd.AddCommand(verify.VerifyCmd)
	rootCmd.AddCommand(run.RunCmd)
	rootCmd.AddCommand(cmdtasks.TasksCmd)
	rootCmd.AddCommand(search.SearchCmd)

}

func flags() {
	// -v, --version flag to print the version
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Print the version")
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/stew/config.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// add sub commands
	addSubCommands()

	flags()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	setDefaults()
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		configPath := home + "/.config/stew"

		// Search config in home directory with name ".stew" (without extension).
		viper.AddConfigPath(configPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")

	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			cobra.CheckErr(err)
		}
	}
}
