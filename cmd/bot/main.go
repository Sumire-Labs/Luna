package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sumire-Labs/Luna/config"
	"github.com/Sumire-Labs/Luna/di"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if cfg.Bot.Debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Debug mode enabled")
	}

	container, err := di.NewContainer(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}
	defer container.Cleanup()

	if err := container.Bot.Start(); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}

	if err := container.CommandRegistry.RegisterSlashCommands(); err != nil {
		log.Fatalf("Failed to register slash commands: %v", err)
	}

	log.Println("Luna Bot is now running. Press CTRL+C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Shutting down Luna Bot...")

	if err := container.CommandRegistry.UnregisterSlashCommands(); err != nil {
		log.Printf("Failed to unregister slash commands: %v", err)
	}

	if err := container.Bot.Stop(); err != nil {
		log.Printf("Failed to stop bot gracefully: %v", err)
	}

	log.Println("Luna Bot has been shut down successfully.")
}