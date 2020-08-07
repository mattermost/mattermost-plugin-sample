package main

import (
	"fmt"
	"os/exec"

	"github.com/blang/semver/v4"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	bumpVersionCmd.AddCommand(
		bumpMajorCmd,
		bumpMinorCmd,
		bumpPatchCmd,
	)
	rootCmd.AddCommand(bumpVersionCmd)
}

var bumpVersionCmd = &cobra.Command{
	Use:   "bump-version",
	Short: "Bump the plugin version",
	Args:  cobra.ExactArgs(0),
}

var bumpMajorCmd = &cobra.Command{
	Use:   "major",
	Short: "Bump to next major version",
	Args:  cobra.ExactArgs(0),
	RunE: func(command *cobra.Command, args []string) error {
		return bumpVersion("major")
	},
}

var bumpMinorCmd = &cobra.Command{
	Use:   "minor",
	Short: "Bump to next minor version",
	Args:  cobra.ExactArgs(0),
	RunE: func(command *cobra.Command, args []string) error {
		return bumpVersion("minor")
	},
}

var bumpPatchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Bump to next patch version",
	Args:  cobra.ExactArgs(0),
	RunE: func(command *cobra.Command, args []string) error {
		return bumpVersion("patch")
	},
}

// bumpVersion
func bumpVersion(mode string) error {
	manifest, err := findManifest()
	if err != nil {
		return errors.Wrap(err, "failed to find manifest")
	}

	oldVersion, err := semver.Parse(manifest.Version)
	if err != nil {
		return errors.Wrap(err, "failed to parse version in manifest")
	}

	newVersion := oldVersion
	switch mode {
	case "major":
		err = newVersion.IncrementMajor()
	case "minor":
		err = newVersion.IncrementMinor()
	case "patch":
		err = newVersion.IncrementPatch()
	default:
		return errors.Errorf("unknown mode %s", mode)
	}
	if err != nil {
		return errors.Wrap(err, "failed up bump manifest version")
	}

	manifest.Version = newVersion.String()
	err = writeManifest(manifest)
	if err != nil {
		return errors.Wrap(err, "failed to writing manifest after bumping version")
	}

	err = applyManifest(manifest)
	if err != nil {
		return errors.Wrap(err, "failed to applying manifest after bumping version")
	}

	files := []string{"plugin.json"}
	if manifest.HasServer() {
		files = append(files, "server/manifest.go")
	}
	if manifest.HasWebapp() {
		files = append(files, "webapp/src/manifest.js")
	}

	cmd := exec.Command("git", append([]string{"diff"}, files...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Print(string(out))
		return err
	}
	fmt.Print(string(out))

	prompt := promptui.Prompt{
		Label:     "Does the diff look good",
		IsConfirm: true,
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Println("Diff wasn't confirmed. Exiting.")
		return nil
	}

	branch := fmt.Sprintf("release_v%s", newVersion)

	gitCommands := [][]string{
		//{"checkout", "master"},
		//{"pull"},
		{"checkout", "-b", branch},
		append([]string{"add"}, files...),
		{"commit", "-m", fmt.Sprintf("Bump version to %s", newVersion)},
		{"push", "--set-upstream", "origin", branch},
		{"checkout", "master"},
	}

	for _, command := range gitCommands {
		cmd := exec.Command("git", command...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Print(string(out))
			return err
		}
		fmt.Print(string(out))
	}

	return nil
}
