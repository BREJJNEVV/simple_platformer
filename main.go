package main

import (
	"flag"
	"log"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"platformer/internal/config"
	"platformer/internal/game"
)

// main - точка входа в программу
func main() {
	modeFlag := flag.String("mode", string(game.ModeLocal), "Game mode: local, host, client")
	addrFlag := flag.String("addr", "", "Address for host or client connection (e.g. :4000 or 192.168.0.5:4000)")
	flag.Parse()

	modeValue := strings.ToLower(strings.TrimSpace(*modeFlag))
	if modeValue == "" {
		modeValue = string(game.ModeLocal)
	}
	mode := game.Mode(modeValue)

	switch mode {
	case game.ModeLocal, game.ModeHost, game.ModeClient:
	default:
		log.Fatalf("unknown mode %q, expected local, host or client", modeValue)
	}

	gameInstance, err := game.NewGameWithOptions(game.Options{
		Mode:    mode,
		Address: strings.TrimSpace(*addrFlag),
	})
	if err != nil {
		log.Fatalf("failed to start game: %v", err)
	}

	// Настраиваем параметры окна
	ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
	ebiten.SetWindowTitle("Платформер на Go")

	// Запускаем игровой цикл
	// RunGame будет вызывать Update и Draw в цикле до тех пор, пока игра не завершится
	if err := ebiten.RunGame(gameInstance); err != nil {
		log.Fatalf("game error: %v", err)
	}
}
