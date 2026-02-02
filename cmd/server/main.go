package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"poke/internal/server"
	"syscall"
)

// main wires CLI flags into server startup.
func main() {
	configPath, err := resolveConfigPath()
	if err != nil {
		log.Fatalf("config path: %v", err)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("config load: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runtime, err := server.Start(ctx, cfg)
	if err != nil {
		log.Fatalf("server start: %v", err)
	}

	log.Printf("server started with config %s", configPath)

	<-ctx.Done()
	close(runtime.RequestChannel)
	log.Printf("server shutting down")
}

// resolveConfigPath parses flags and selects the configuration file path.
func resolveConfigPath() (string, error) {
	shortFlag := flag.String("c", "", "path to poke server config file")
	longFlag := flag.String("config", "", "path to poke server config file")
	flag.Parse()

	if *shortFlag != "" && *longFlag != "" && *shortFlag != *longFlag {
		return "", fmt.Errorf("conflicting config flags: -c=%s --config=%s", *shortFlag, *longFlag)
	}

	if *shortFlag != "" {
		return *shortFlag, nil
	}
	if *longFlag != "" {
		return *longFlag, nil
	}

	return findDefaultConfigPath()
}

// findDefaultConfigPath returns the first existing default configuration path.
func findDefaultConfigPath() (string, error) {
	candidates := []string{"/etc/poke/poke.yml"}

	if xdg, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok && xdg != "" {
		candidates = append(candidates, filepath.Join(xdg, "poke", "poke.yml"))
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	candidates = append(candidates,
		filepath.Join(home, "config", "poke", "poke.yml"),
		filepath.Join(home, ".poke", "poke.yml"),
	)

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}

	return "", fmt.Errorf("config file not found; searched: %v", candidates)
}

// loadConfig reads and parses the server configuration file.
func loadConfig(path string) (server.Config, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- by design, comes as CLI arg
	if err != nil {
		return server.Config{}, err
	}
	return server.Parse(data)
}
