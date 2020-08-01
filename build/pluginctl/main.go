// main handles deployment of the plugin to a development server using the Client4 API.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use: "pluginctl",
}

func init() {
	rootCmd.AddCommand(
		deployCmd,
		disableCmd,
		enableCmd,
		resetCmd,
	)
}

var deployCmd = &cobra.Command{
	Use:     "deploy",
	Short:   "Deploy the plugin",
	Example: "deploy dist/com.mattermost.plugin-starter-template-0.1.0.tar.gz",
	Args:    cobra.ExactArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		m, err := findManifest()
		if err != nil {
			return errors.Wrap(err, "failed to find manifest")
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		return deploy(client, m.Id, args[0])
	},
}

var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable the plugin",
	Args:  cobra.ExactArgs(0),
	RunE: func(command *cobra.Command, args []string) error {
		m, err := findManifest()
		if err != nil {
			return errors.Wrap(err, "failed to find manifest")
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		return disablePlugin(client, m.Id)
	},
}

var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable the plugin",
	Args:  cobra.ExactArgs(0),
	RunE: func(command *cobra.Command, args []string) error {
		m, err := findManifest()
		if err != nil {
			return errors.Wrap(err, "failed to find manifest")
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		return enablePlugin(client, m.Id)
	},
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Disable and enable the plugin",
	Args:  cobra.ExactArgs(0),
	RunE: func(command *cobra.Command, args []string) error {
		m, err := findManifest()
		if err != nil {
			return errors.Wrap(err, "failed to find manifest")
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		return resetPlugin(client, m.Id)
	},
}

func getClient() (*model.Client4, error) {
	siteURL := os.Getenv("MM_SERVICESETTINGS_SITEURL")
	adminToken := os.Getenv("MM_ADMIN_TOKEN")
	adminUsername := os.Getenv("MM_ADMIN_USERNAME")
	adminPassword := os.Getenv("MM_ADMIN_PASSWORD")

	if siteURL == "" {
		return nil, errors.New("MM_SERVICESETTINGS_SITEURL is not set")
	}

	client := model.NewAPIv4Client(siteURL)

	if adminToken != "" {
		log.Printf("Authenticating using token against %s.", siteURL)
		client.SetToken(adminToken)
		return client, nil
	}

	if adminUsername != "" && adminPassword != "" {
		client := model.NewAPIv4Client(siteURL)
		log.Printf("Authenticating as %s against %s.", adminUsername, siteURL)
		_, resp := client.Login(adminUsername, adminPassword)
		if resp.Error != nil {
			return nil, fmt.Errorf("failed to login as %s: %w", adminUsername, resp.Error)
		}
		return client, nil
	}

	return nil, errors.New("one of MM_ADMIN_TOKEN or MM_ADMIN_USERNAME/MM_ADMIN_PASSWORD must be defined")
}

// deploy attempts to upload and enable a plugin via the Client4 API.
// It will fail if plugin uploads are disabled.
func deploy(client *model.Client4, pluginID, bundlePath string) error {
	pluginBundle, err := os.Open(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", bundlePath, err)
	}
	defer pluginBundle.Close()

	log.Print("Uploading plugin via API.")
	_, resp := client.UploadPluginForced(pluginBundle)
	if resp.Error != nil {
		return fmt.Errorf("failed to upload plugin bundle: %s", resp.Error.Error())
	}

	log.Print("Enabling plugin.")
	_, resp = client.EnablePlugin(pluginID)
	if resp.Error != nil {
		return fmt.Errorf("failed to enable plugin: %s", resp.Error.Error())
	}

	return nil
}

// disablePlugin attempts to disable the plugin via the Client4 API.
func disablePlugin(client *model.Client4, pluginID string) error {
	log.Print("Disabling plugin.")
	_, resp := client.DisablePlugin(pluginID)
	if resp.Error != nil {
		return fmt.Errorf("failed to disable plugin: %w", resp.Error)
	}

	return nil
}

// enablePlugin attempts to enable the plugin via the Client4 API.
func enablePlugin(client *model.Client4, pluginID string) error {
	log.Print("Enabling plugin.")
	_, resp := client.EnablePlugin(pluginID)
	if resp.Error != nil {
		return fmt.Errorf("failed to enable plugin: %w", resp.Error)
	}

	return nil
}

// resetPlugin attempts to reset the plugin via the Client4 API.
func resetPlugin(client *model.Client4, pluginID string) error {
	err := disablePlugin(client, pluginID)
	if err != nil {
		return err
	}

	err = enablePlugin(client, pluginID)
	if err != nil {
		return err
	}

	return nil
}
