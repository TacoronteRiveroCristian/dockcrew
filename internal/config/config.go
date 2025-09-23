package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

type App struct {
	RefreshInterval float64 `yaml:"refresh_interval"`
	CheckForUpdate  bool    `yaml:"check_for_update"`
	Telemetry       bool    `yaml:"telemetry"`
}

type UI struct {
	Theme           string `yaml:"theme"`
	Mouse           bool   `yaml:"mouse"`
	WrapMainPanel   bool   `yaml:"wrap_main_panel"`
	PaletteShortcut string `yaml:"palette_shortcut"`
}

type Logs struct {
	MaxLines       int    `yaml:"max_lines"`
	Tail           int    `yaml:"tail"`
	Since          string `yaml:"since"`
	HighlightLevel bool   `yaml:"highlight_levels"`
	PrettyJSON     bool   `yaml:"pretty_json"`
	RegexTimeoutMS int    `yaml:"regex_timeout_ms"`
}

type Docker struct {
	Host string `yaml:"host"`
}

type Compose struct {
	Enabled        bool   `yaml:"enabled"`
	Binary         string `yaml:"binary"`
	DefaultProject string `yaml:"default_project"`
}

type Colors struct {
	Title      string `yaml:"title"`
	Help       string `yaml:"help"`
	Background string `yaml:"background"`
	Footer     string `yaml:"footer"`
	Success    string `yaml:"success"`
	Error      string `yaml:"error"`
}

type Config struct {
	App     App     `yaml:"app"`
	UI      UI      `yaml:"ui"`
	Logs    Logs    `yaml:"logs"`
	Docker  Docker  `yaml:"docker"`
	Compose Compose `yaml:"compose"`
	Colors  Colors  `yaml:"colors"`
}

// Default devuelve una configuración por defecto razonable
func Default() Config {
	return Config{
		App: App{RefreshInterval: 5.0, CheckForUpdate: true, Telemetry: false},
		UI: UI{Theme: "default", Mouse: true, WrapMainPanel: false, PaletteShortcut: "ctrl+p"},
		Logs: Logs{MaxLines: 4000, Tail: 400, Since: "30m", HighlightLevel: true, PrettyJSON: true, RegexTimeoutMS: 50},
		Docker: Docker{Host: defaultDockerHost()},
		Compose: Compose{Enabled: true, Binary: "docker"},
	}
}

func defaultDockerHost() string {
	if runtime.GOOS == "windows" {
		return "npipe:////./pipe/docker_engine"
	}
	return "unix:///var/run/docker.sock"
}

// ResolvePath determina la ruta de config por defecto del usuario
func ResolvePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil { return "", err }
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "DockCrew", "config.yaml"), nil
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "DockCrew", "config.yaml"), nil
	default:
		return filepath.Join(home, ".config", "dockcrew", "config.yaml"), nil
	}
}

// Load implementará la carga desde archivo/env más adelante. Por ahora devuelve defaults.
func Load() (Config, error) {
	cfg := Default()
	// NOTA: para evitar dependencias externas, de momento no parseamos YAML.
	// La ruta se resuelve para futura compatibilidad, pero se ignora su contenido.
	// Próximo paso: habilitar YAML con gopkg.in/yaml.v3 cuando se pueda ejecutar `go get`.
	_ = os.Getenv("DOCKCREW_CONFIG")
	// Overrides por variables de entorno (subset útil)
	if v := os.Getenv("DOCKCREW_APP_REFRESH_INTERVAL"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil { cfg.App.RefreshInterval = f }
	}
	if v := os.Getenv("DOCKCREW_APP_CHECK_FOR_UPDATE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil { cfg.App.CheckForUpdate = b }
	}
	if v := os.Getenv("DOCKCREW_APP_TELEMETRY"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil { cfg.App.Telemetry = b }
	}
	if v := os.Getenv("DOCKCREW_UI_THEME"); v != "" { cfg.UI.Theme = v }
	if v := os.Getenv("DOCKCREW_LOGS_MAX_LINES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil { cfg.Logs.MaxLines = n }
	}
	if v := os.Getenv("DOCKCREW_DOCKER_HOST"); v != "" { cfg.Docker.Host = v }
	if v := os.Getenv("DOCKCREW_COMPOSE_ENABLED"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil { cfg.Compose.Enabled = b }
	}
	return cfg, nil
}

// Validate valida rangos básicos
func (c *Config) Validate() error {
	if c.App.RefreshInterval <= 0 { return errors.New("app.refresh_interval debe ser > 0") }
	if c.Logs.MaxLines < 100 { return errors.New("logs.max_lines demasiado bajo (>=100)") }
	if c.Logs.Tail < 0 { return errors.New("logs.tail no puede ser negativo") }
	if c.Logs.RegexTimeoutMS < 0 { return errors.New("logs.regex_timeout_ms no puede ser negativo") }
	return nil
}
