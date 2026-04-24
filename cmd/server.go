package cmd

import (
	"context"

	cfg "github.com/nicovak/miqio-go/config"
	"github.com/nicovak/miqio-go/driver"
	"github.com/nicovak/miqio-go/logger"
	"github.com/nicovak/miqio-go/repository"
	"github.com/nicovak/miqio-proxy/config"
	"github.com/nicovak/miqio-proxy/internal/r2"
	"github.com/nicovak/miqio-proxy/server"
	"github.com/spf13/cobra"
)

func NewServerCmd(_ context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Start the HTTP proxy server",
		Run: func(_ *cobra.Command, _ []string) {
			conf := config.Get()

			pg := driver.NewPostgres(conf.Postgres.ConnectionString)
			repo := &repository.PgxRepository{DB: pg.ConnPool}

			r2Client := r2.NewClient(
				conf.R2.AccountID,
				conf.R2.AccessKeyID,
				conf.R2.SecretAccessKey,
				conf.R2.BucketName,
			)

			srv := &server.Server{
				Logger:   logger.Get(),
				Pool:     pg.ConnPool,
				Repo:     repo,
				R2Client: r2Client,
			}

			server.Run(srv)
		},
	}
}

func init() {
	c := config.Config{}
	cfg.Setup("miqio-proxy", &c)

	ctx := context.Background()

	serverCmd := NewServerCmd(ctx)
	rootCmd.AddCommand(serverCmd)
}
