package cmd_store

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sasano8/kvtool/pkg/config"
	"github.com/sasano8/kvtool/pkg/server"
	"github.com/spf13/cobra"
)

func init() {
	cmd := CmdServe
	cmd.Flags().StringP("config", "c", ".kvtool.yml", "path to config file")
	cmd.Flags().BoolP("global", "g", false, "use global config (~/.config/kvtool/.kvtool.yml)")
	cmd.Flags().IntP("port", "p", 8080, "port to listen on")
}

var CmdServe = &cobra.Command{
	Use:   "serve",
	Short: "Start HTTP server to serve store contents",
	Long: `Start an HTTP server that exposes store contents via REST API.

Endpoints:
  GET /v1/namespaces/{namespace}/kv/{store}/{path}      JSON response
  GET /v1/namespaces/{namespace}/storage/{store}/{path} Raw response
  GET /health                                           Health check

Examples:
  kvtool store serve
  kvtool store serve --port 3000

  curl http://localhost:8080/v1/namespaces/production/kv/vault/app/settings
  curl http://localhost:8080/v1/namespaces/default/storage/local/config.json`,
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, _ := cmd.Flags().GetString("config")
		useGlobal, _ := cmd.Flags().GetBool("global")
		port, _ := cmd.Flags().GetInt("port")

		// If --global flag is set, use global config
		if useGlobal {
			globalPath, err := config.GetGlobalConfigPath()
			if err != nil {
				return err
			}
			configPath = globalPath
		}

		// Load config
		cfg, err := config.LoadConfigAuto(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Create server
		ctx := context.Background()
		srv := server.NewServer(ctx, cfg)

		// HTTP server
		addr := fmt.Sprintf(":%d", port)
		httpServer := &http.Server{
			Addr:    addr,
			Handler: srv.Handler(),
		}

		// Graceful shutdown
		done := make(chan os.Signal, 1)
		signal.Notify(done, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-done
			fmt.Println("\nShutting down...")
			httpServer.Close()
		}()

		fmt.Printf("Starting server on http://localhost:%d\n", port)
		fmt.Println("\nEndpoints:")
		fmt.Println("  GET /v1/namespaces/{ns}/kv/{store}/{path}      - JSON")
		fmt.Println("  GET /v1/namespaces/{ns}/storage/{store}/{path} - Raw")
		fmt.Println("  GET /health                                    - Health check")
		fmt.Println("\nPress Ctrl+C to stop")

		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}

		return nil
	},
}
