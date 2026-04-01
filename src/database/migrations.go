package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func RunMigrations(cfg *DatabaseConfig) {
	url := PostgresURLFromConfig(cfg)
	dir := "migrations"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	if err := RunAtlasMigrations(context.Background(), url, dir); err != nil {
		log.Fatal(err)
		panic(err)
	}
	log.Println("atlas migrate apply: ok")
}

func RunAtlasMigrations(ctx context.Context, databaseURL string, migrationsDir string) error {
	abs, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("migrations dir: %w", err)
	}
	cmd := exec.CommandContext(ctx, "atlas", "migrate", "apply",
		"--url", databaseURL,
		"--dir", "file://"+filepath.ToSlash(abs),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("atlas migrate apply: %w\n%s", err, string(out))
	}
	return nil
}
