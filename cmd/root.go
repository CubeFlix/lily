// cmd/root.go
// Server main command.

package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/cubeflix/lily/version"
	"github.com/spf13/cobra"
)

var serverFile string
var host string
var port int
var quiet bool
var clearance int

var driveFile string
var accessClearance int
var modifyClearance int

// Base Lily command.
var RootCmd = &cobra.Command{
	Use:   "lilys",
	Short: "A secure file server.",
	Long:  `lilys is the Lily file server program.`,
}

// Version command.
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version and exit.",
	Long:  `Print the Lily server version number.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("lily", version.VERSION, runtime.GOOS)
	},
}

// Serve command.
var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Lily server.",
	Long:  `Start the Lily server.`,
	Run:   ServeCommand,
}

// Config command.
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure the Lily server.",
	Long:  `Configure the Lily server.`,
}

// Init server subcommand.
var ConfigInitCmd = &cobra.Command{
	Use:   "init <config>",
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Short: "Initialize a new server.",
	Long:  `Initialize a new Lily server with a config file.`,
	Run:   ConfigInit,
}

// Set server setting subcommand.
var ConfigSetCmd = &cobra.Command{
	Use:   "set <name> <value>",
	Args:  cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs),
	Short: "Set a setting on the server.",
	Long:  `Set a setting on the server.`,
	Run:   ConfigSet,
}

// Get server setting subcommand.
var ConfigGetCmd = &cobra.Command{
	Use:   "get <name>",
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Short: "Get a setting on the server.",
	Long:  `Get a setting on the server.`,
	Run:   ConfigGet,
}

// List server setting subcommand.
var ConfigListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the server settings.",
	Long:  `List the server settings.`,
	Run:   ConfigList,
}

// Add drive file subcommand.
var ConfigAddDriveCmd = &cobra.Command{
	Use:   "add-drive <name> <file>",
	Args:  cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs),
	Short: "Add a drive file to the server.",
	Long:  `Add a drive file to the server (must be absolute path).`,
	Run:   ConfigAddDrive,
}

// Rename drive subcommand.
var ConfigRenameDriveCmd = &cobra.Command{
	Use:   "rename-drive <name> <newName>",
	Args:  cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs),
	Short: "Rename a drive.",
	Long:  `Rename a drive (will rename on drive file as well).`,
	Run:   ConfigRenameDrive,
}

// Remove drive file subcommand.
var ConfigRemoveDriveCmd = &cobra.Command{
	Use:   "remove-drive <name>",
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Short: "Remove a drive file from the server.",
	Long:  `Remove a drive file from the server.`,
	Run:   ConfigRemoveDrive,
}

// List drive files subcommand.
var ConfigListDriveCmd = &cobra.Command{
	Use:   "list-drive",
	Short: "List the drives on the server.",
	Long:  `List the drives on the server.`,
	Run:   ConfigListDrive,
}

// Add user subcommand.
var ConfigAddUserCmd = &cobra.Command{
	Use:   "add-user <username> <password>",
	Args:  cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs),
	Short: "Add a user to the server.",
	Long:  `Add a user to the server.`,
	Run:   ConfigAddUser,
}

// Remove user subcommand.
var ConfigRemoveUserCmd = &cobra.Command{
	Use:   "remove-user <name>",
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Short: "Remove a user from the server.",
	Long:  `Remove a user from the server.`,
	Run:   ConfigRemoveUser,
}

// Set certificates subcommand.
var ConfigSetCertsCmd = &cobra.Command{
	Use:   "set-certs <certFiles> <keyFiles>",
	Args:  cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs),
	Short: "Set the certificates.",
	Long:  `Set the certificates, given a comma-separated list of cert files and key files (must be absolute paths).`,
	Run:   ConfigSetCerts,
}

// Drive command.
var DriveCmd = &cobra.Command{
	Use:   "drive",
	Short: "Configure a Lily drive.",
	Long:  `Configure a Lily drive.`,
}

// Drive init command.
var DriveInitCmd = &cobra.Command{
	Use:   "init <name> <path>",
	Args:  cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs),
	Short: "Initialize a new drive.",
	Long:  `Initialize a new Lily drive with the name and path (must be absolute path).`,
	Run:   DriveInit,
}

// Drive set path command.
var DriveSetPathCmd = &cobra.Command{
	Use:   "set-path <path>",
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Short: "Set the drive path.",
	Long:  `Set the drive path (must be absolute path).`,
	Run:   DriveSetPath,
}

// Drive reimport path command.
var DriveReimportCmd = &cobra.Command{
	Use:   "reimport",
	Short: "Reimport the drive directory.",
	Long:  `Reimport all the directories and files in in drive path.`,
	Run:   DriveReimport,
}

// Drive params command.
var DriveSettingsCmd = &cobra.Command{
	Use:   "params",
	Short: "Get the drive parameters.",
	Long:  `Get the drive name and path.`,
	Run:   DriveSettings,
}

// Drive list FS command.
var DriveListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the drive directory.",
	Long:  `List all the directories and files recursively.`,
	Run:   DriveList,
}

// Execute the root command.
func Execute() {
	// Execute the main command.
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Init cobra.
func init() {
	// Set the arguments.
	ServeCmd.PersistentFlags().StringVarP(&serverFile, "file", "f", ".server.lily", "The server file to use")
	ServeCmd.PersistentFlags().StringVar(&host, "host", "", "The host to listen on (defaults to server file)")
	ServeCmd.PersistentFlags().IntVarP(&port, "port", "p", 0, "The port to listen on (defaults to server file)")
	ServeCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "If we should not log (defaults to false)")
	ConfigCmd.PersistentFlags().StringVarP(&serverFile, "file", "f", ".server.lily", "The server file to use")
	DriveCmd.PersistentFlags().StringVarP(&driveFile, "file", "f", ".%name%.lilyd", "The drive file to use")
	ConfigAddUserCmd.PersistentFlags().IntVarP(&clearance, "clearance", "c", 5, "The clearance level for the new user")
	DriveInitCmd.PersistentFlags().IntVarP(&accessClearance, "access-clearance", "a", 1, "The access clearance level")
	DriveInitCmd.PersistentFlags().IntVarP(&modifyClearance, "modify-clearance", "m", 1, "The modify clearance level")
	DriveReimportCmd.PersistentFlags().IntVarP(&accessClearance, "access-clearance", "a", 1, "The access clearance level")
	DriveReimportCmd.PersistentFlags().IntVarP(&modifyClearance, "modify-clearance", "m", 1, "The modify clearance level")

	// Add the commands.
	RootCmd.AddCommand(VersionCmd)
	RootCmd.AddCommand(ServeCmd)
	RootCmd.AddCommand(ConfigCmd)
	RootCmd.AddCommand(DriveCmd)
	ConfigCmd.AddCommand(ConfigInitCmd)
	ConfigCmd.AddCommand(ConfigSetCmd)
	ConfigCmd.AddCommand(ConfigGetCmd)
	ConfigCmd.AddCommand(ConfigListCmd)
	ConfigCmd.AddCommand(ConfigAddDriveCmd)
	ConfigCmd.AddCommand(ConfigRenameDriveCmd)
	ConfigCmd.AddCommand(ConfigRemoveDriveCmd)
	ConfigCmd.AddCommand(ConfigListDriveCmd)
	ConfigCmd.AddCommand(ConfigAddUserCmd)
	ConfigCmd.AddCommand(ConfigRemoveUserCmd)
	ConfigCmd.AddCommand(ConfigSetCertsCmd)
	DriveCmd.AddCommand(DriveInitCmd)
	DriveCmd.AddCommand(DriveSetPathCmd)
	DriveCmd.AddCommand(DriveReimportCmd)
	DriveCmd.AddCommand(DriveSettingsCmd)
	DriveCmd.AddCommand(DriveListCmd)
}
