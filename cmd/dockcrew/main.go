package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/TacoronteRiveroCristian/dockcrew/internal/config"
	"github.com/TacoronteRiveroCristian/dockcrew/internal/docker"
	"github.com/TacoronteRiveroCristian/dockcrew/internal/ui"
)

// version se sobreescribe en build con -ldflags "-X main.version=..."
var version = "0.0.0-dev"

func printUsage() {
	fmt.Println("dockcrew", version)
	fmt.Println("Uso:")
	fmt.Println("  dockcrew                    # abre la TUI interactiva")
	fmt.Println("  dockcrew version | --version | -v")
	fmt.Println("  dockcrew containers [--all]   # lista contenedores")
	fmt.Println("  dockcrew images       # lista imágenes")
	fmt.Println("  dockcrew image rm [-f] <ref>")
	fmt.Println("  dockcrew image prune")
	fmt.Println("  dockcrew --export-default-config   # imprime YAML de configuración por defecto")
	fmt.Println("  dockcrew start <id|name>")
	fmt.Println("  dockcrew stop <id|name>")
	fmt.Println("  dockcrew restart <id|name>")
	fmt.Println("  dockcrew rm [-f] <id|name>")
	fmt.Println("  dockcrew tui            # abre la TUI (alias explícito)")
}

func main() {
	if len(os.Args) <= 1 {
		runTUI()
		return
	}

	if len(os.Args) > 1 {
		arg := os.Args[1]
		switch arg {
		case "version", "--version", "-v":
			fmt.Println("dockcrew", version)
			return
		case "--export-default-config":
			runExportDefaultConfig()
			return
		case "containers":
			runContainers(os.Args[2:])
			return
		case "images":
			runImages()
			return
		case "image":
			runImage(os.Args[2:])
			return
		case "start":
			runStart(os.Args[2:])
			return
		case "stop":
			runStop(os.Args[2:])
			return
		case "restart":
			runRestart(os.Args[2:])
			return
		case "rm":
			runRm(os.Args[2:])
			return
		case "tui":
			runTUI()
			return
		case "help", "-h", "--help":
			printUsage()
			return
		}
	}
	printUsage()
}

func runTUI() {
	client := docker.NewShellClient()
	client.Timeout = 8 * time.Second
	if err := ui.Run(client); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func runContainers(args []string) {
	c := docker.NewShellClient()
	c.Timeout = 5 * time.Second
	ctx := context.Background()
	all := false
	if len(args) > 0 {
		if args[0] == "--all" || args[0] == "-a" {
			all = true
		}
	}
	list, err := c.ListContainersAll(ctx, all)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	if len(list) == 0 {
		if all {
			fmt.Println("(sin contenedores)")
		} else {
			fmt.Println("(sin contenedores en ejecución; prueba --all para todos)")
		}
		return
	}
	fmt.Printf("%-12s  %-25s  %-25s  %-20s  %-15s\n", "ID", "NAME", "IMAGE", "STATUS", "STACK")
	for _, ct := range list {
		fmt.Printf("%-12s  %-25s  %-25s  %-20s  %-15s\n", shortID(ct.ID), trunc(ct.Name, 25), trunc(ct.Image, 25), trunc(ct.Status, 20), ct.Stack)
	}
}

func runImages() {
	c := docker.NewShellClient()
	c.Timeout = 5 * time.Second
	ctx := context.Background()
	list, err := c.ListImages(ctx)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	if len(list) == 0 {
		fmt.Println("(sin imágenes)")
		return
	}
	for _, im := range list {
		tag := ""
		if len(im.RepoTags) > 0 {
			tag = im.RepoTags[0]
		}
		fmt.Printf("%s\t%s\n", im.ID, tag)
	}
}

func runImage(args []string) {
	if len(args) == 0 {
		fmt.Println("uso: dockcrew image [rm|prune] ...")
		os.Exit(2)
	}
	sub := args[0]
	c := docker.NewShellClient()
	c.Timeout = 30 * time.Second
	switch sub {
	case "rm":
		if len(args) < 2 {
			fmt.Println("uso: dockcrew image rm [-f] <ref>")
			os.Exit(2)
		}
		force := false
		idx := 1
		if args[1] == "-f" {
			force = true
			idx = 2
		}
		if idx >= len(args) {
			fmt.Println("uso: dockcrew image rm [-f] <ref>")
			os.Exit(2)
		}
		ref := args[idx]
		if err := c.RemoveImage(context.Background(), ref, force); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
	case "prune":
		if err := c.PruneImages(context.Background()); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
	default:
		fmt.Println("uso: dockcrew image [rm|prune] ...")
		os.Exit(2)
	}
}

func runStart(args []string) {
	if len(args) < 1 {
		fmt.Println("uso: dockcrew start <id|name>")
		os.Exit(2)
	}
	id := args[0]
	c := docker.NewShellClient()
	c.Timeout = 10 * time.Second
	if err := c.StartContainer(context.Background(), id); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func runStop(args []string) {
	if len(args) < 1 {
		fmt.Println("uso: dockcrew stop <id|name>")
		os.Exit(2)
	}
	id := args[0]
	c := docker.NewShellClient()
	c.Timeout = 30 * time.Second
	if err := c.StopContainer(context.Background(), id); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func runRestart(args []string) {
	if len(args) < 1 {
		fmt.Println("uso: dockcrew restart <id|name>")
		os.Exit(2)
	}
	id := args[0]
	c := docker.NewShellClient()
	c.Timeout = 30 * time.Second
	if err := c.RestartContainer(context.Background(), id); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func runRm(args []string) {
	if len(args) < 1 {
		fmt.Println("uso: dockcrew rm [-f] <id|name>")
		os.Exit(2)
	}
	force := false
	targetIdx := 0
	if args[0] == "-f" {
		force = true
		targetIdx = 1
	}
	if targetIdx >= len(args) {
		fmt.Println("uso: dockcrew rm [-f] <id|name>")
		os.Exit(2)
	}
	id := args[targetIdx]
	c := docker.NewShellClient()
	c.Timeout = 10 * time.Second
	if err := c.RemoveContainer(context.Background(), id, force); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func runExportDefaultConfig() {
	cfg := config.Default()
	// YAML manual sencillo para evitar traer librerías; suficiente para bootstrap
	fmt.Println("app:")
	fmt.Printf("  refresh_interval: %.1f\n", cfg.App.RefreshInterval)
	fmt.Printf("  check_for_update: %t\n", cfg.App.CheckForUpdate)
	fmt.Printf("  telemetry: %t\n", cfg.App.Telemetry)
	fmt.Println("ui:")
	fmt.Printf("  theme: \"%s\"\n", cfg.UI.Theme)
	fmt.Printf("  mouse: %t\n", cfg.UI.Mouse)
	fmt.Printf("  wrap_main_panel: %t\n", cfg.UI.WrapMainPanel)
	fmt.Printf("  palette_shortcut: \"%s\"\n", cfg.UI.PaletteShortcut)
	fmt.Println("logs:")
	fmt.Printf("  max_lines: %d\n", cfg.Logs.MaxLines)
	fmt.Printf("  tail: %d\n", cfg.Logs.Tail)
	fmt.Printf("  since: \"%s\"\n", cfg.Logs.Since)
	fmt.Printf("  highlight_levels: %t\n", cfg.Logs.HighlightLevel)
	fmt.Printf("  pretty_json: %t\n", cfg.Logs.PrettyJSON)
	fmt.Printf("  regex_timeout_ms: %d\n", cfg.Logs.RegexTimeoutMS)
	fmt.Println("docker:")
	fmt.Printf("  host: \"%s\"\n", cfg.Docker.Host)
	fmt.Println("compose:")
	fmt.Printf("  enabled: %t\n", cfg.Compose.Enabled)
	fmt.Printf("  binary: \"%s\"\n", cfg.Compose.Binary)
	fmt.Printf("  default_project: \"%s\"\n", cfg.Compose.DefaultProject)
	if (cfg.Colors != config.Colors{}) {
		fmt.Println("colors:")
		fmt.Printf("  title: \"%s\"\n", cfg.Colors.Title)
		fmt.Printf("  help: \"%s\"\n", cfg.Colors.Help)
		fmt.Printf("  background: \"%s\"\n", cfg.Colors.Background)
		fmt.Printf("  footer: \"%s\"\n", cfg.Colors.Footer)
		fmt.Printf("  success: \"%s\"\n", cfg.Colors.Success)
		fmt.Printf("  error: \"%s\"\n", cfg.Colors.Error)
	}
}

func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func trunc(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 1 {
		return string(runes[:max])
	}
	if max <= 3 {
		return string(runes[:max])
	}
	return string(runes[:max-3]) + "..."
}
