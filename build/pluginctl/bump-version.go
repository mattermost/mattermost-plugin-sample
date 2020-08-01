package main

import (
	"github.com/blang/semver/v4"
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

	err = applyManifest(manifest)
	if err != nil {
		return errors.Wrap(err, "failed to write manifest after bumping version")
	}

	return nil
}
