package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/holedaemon/hole.casa/internal/web"
)

func die(msg string, args ...interface{}) {
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}

	fmt.Fprintf(os.Stderr, msg, args...)
}

func main() {
	addr := os.Getenv("HOLE_ADDR")
	token := os.Getenv("HOLE_TOKEN")
	guildID := os.Getenv("HOLE_GUILD_ID")
	ignoreBots := os.Getenv("HOLE_IGNORE_BOTS")

	igb, err := strconv.ParseBool(ignoreBots)
	if err != nil {
		igb = false
	}

	srv, err := web.New(&web.Options{
		Addr:       addr,
		Token:      token,
		GuildID:    guildID,
		IgnoreBots: igb,
	})
	if err != nil {
		die("error creating server: %s", err)
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Kill)
	defer cancel()

	if err := srv.Start(ctx); err != nil {
		die("error occurred during server runtime: %s", err)
	}
}
