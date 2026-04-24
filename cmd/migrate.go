package cmd

import (
	"context"

	cfg "github.com/nicovak/miqio-go/config"
	"github.com/nicovak/miqio-proxy/config"
	mig "github.com/nicovak/miqio-proxy/migrate"
	"github.com/spf13/cobra"
)

func NewMigrateCmd(_ context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Run: func(_ *cobra.Command, _ []string) {
			mig.Run()
		},
	}
}

func init() {
	c := config.Config{}
	cfg.Setup("miqio-proxy", &c)

	ctx := context.Background()

	migrateCmd := NewMigrateCmd(ctx)
	rootCmd.AddCommand(migrateCmd)
}
