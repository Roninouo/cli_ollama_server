package app

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"cli_ollama_server/internal/config"
	"cli_ollama_server/internal/i18n"
	"cli_ollama_server/internal/ui"
)

func runUI(tr *i18n.Bundle, loaded config.Config, meta config.LoadMeta, opts globalOpts, args []string) int {
	eff, _ := config.ResolveEffective(config.EffectiveOptions{
		GlobalHostFlag:      opts.Host,
		GlobalLangFlag:      opts.Lang,
		GlobalOllamaExeFlag: opts.OllamaExe,
		GlobalModeFlag:      opts.Mode,
		GlobalUnsafeFlag:    opts.Unsafe,
		LoadedConfig:        loaded,
	})

	listen := "127.0.0.1:0"
	if len(args) >= 2 && args[0] == "--listen" {
		listen = args[1]
	}
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		fmt.Fprintln(os.Stderr, tr.Sprintf("error.ui_listen", "error", err.Error()))
		return 1
	}
	defer ln.Close()

	srv := ui.Server{
		Listener:   ln,
		Translator: tr,
		Effective:  eff,
		ConfigPath: meta.PrimaryPath,
		BaseEnv:    os.Environ(),
	}
	url, err := srv.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, tr.Sprintf("error.ui_start", "error", err.Error()))
		return 1
	}

	fmt.Println(tr.Sprintf("ui.started", "url", url))
	fmt.Println(tr.Sprintf("ui.stop_hint"))

	// Set up signal handling for graceful shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Wait for either server error or shutdown signal.
	done := make(chan int, 1)
	go func() {
		done <- srv.Wait()
	}()

	select {
	case <-sigCh:
		fmt.Println() // Newline after ^C
		if err := srv.Shutdown(); err != nil {
			fmt.Fprintln(os.Stderr, tr.Sprintf("error.ui_shutdown", "error", err.Error()))
		}
		return 0
	case code := <-done:
		return code
	}
}
