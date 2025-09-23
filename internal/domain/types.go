package domain

import "time"

// PortBinding representa un puerto mapeado
type PortBinding struct {
	IP      string
	Private uint16
	Public  uint16
	Type    string // tcp|udp
}

type Container struct {
	ID        string
	Name      string
	Image     string
	State     string // running|exited|restarting|paused|created|dead
	Status    string // human readable
	Ports     []PortBinding
	CreatedAt time.Time
	Labels    map[string]string
	Stack     string // compose project
}

type Image struct {
	ID        string
	RepoTags  []string
	SizeBytes int64
	CreatedAt time.Time
	Dangling  bool
}

type Volume struct {
	Name       string
	Driver     string
	Mountpoint string
	CreatedAt  time.Time
	InUseBy    []string
}

type Network struct {
	ID        string
	Name      string
	Driver    string
	Scope     string
	CreatedAt time.Time
}

type Stack struct {
	Name       string
	Services   []string
	Containers []string
}

type Stats struct {
	CPUPercent       float64
	MemBytes         uint64
	MemPercent       float64
	NetRxBytes       uint64
	NetTxBytes       uint64
	BlockReadBytes   uint64
	BlockWriteBytes  uint64
}

type LogLevel int

const (
	TRACE LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
)

type LogLine struct {
	TS   *time.Time
	Lvl  LogLevel
	Text string
	Meta map[string]string
}
