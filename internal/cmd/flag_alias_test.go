package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestFlagAlias(t *testing.T) {
	t.Run("alias shares value with original", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		var val string
		fs.StringVar(&val, "status", "default", "")
		flagAlias(fs, "status", "st")
		if err := fs.Parse([]string{"--st", "PAID"}); err != nil {
			t.Fatal(err)
		}
		if val != "PAID" {
			t.Errorf("expected val=PAID, got %q", val)
		}
	})

	t.Run("alias is hidden", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		var val string
		fs.StringVar(&val, "status", "", "")
		flagAlias(fs, "status", "st")
		f := fs.Lookup("st")
		if f == nil {
			t.Fatal("alias not found")
		}
		if !f.Hidden {
			t.Error("alias should be hidden")
		}
	})

	t.Run("flagOrAliasChanged detects alias", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		var val string
		cmd.Flags().StringVar(&val, "status", "", "")
		flagAlias(cmd.Flags(), "status", "st")
		_ = cmd.Flags().Parse([]string{"--st", "PAID"})
		if !flagOrAliasChanged(cmd, "status") {
			t.Error("flagOrAliasChanged should detect alias")
		}
	})

	t.Run("flagOrAliasChanged detects original", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		var val string
		cmd.Flags().StringVar(&val, "status", "", "")
		flagAlias(cmd.Flags(), "status", "st")
		_ = cmd.Flags().Parse([]string{"--status", "PAID"})
		if !flagOrAliasChanged(cmd, "status") {
			t.Error("flagOrAliasChanged should detect original")
		}
	})

	t.Run("flagOrAliasChanged false when neither set", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		var val string
		cmd.Flags().StringVar(&val, "status", "", "")
		flagAlias(cmd.Flags(), "status", "st")
		_ = cmd.Flags().Parse([]string{})
		if flagOrAliasChanged(cmd, "status") {
			t.Error("flagOrAliasChanged should be false")
		}
	})

	t.Run("panics on missing flag", func(t *testing.T) {
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for missing flag")
			}
		}()
		flagAlias(fs, "nonexistent", "ne")
	})
}
