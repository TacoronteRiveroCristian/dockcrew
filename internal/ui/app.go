package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/TacoronteRiveroCristian/dockcrew/internal/docker"
	"github.com/TacoronteRiveroCristian/dockcrew/internal/domain"
)

const (
	defaultRefreshInterval = 5 * time.Second
	fetchTimeout           = 4 * time.Second
	actionTimeout          = 20 * time.Second
	notificationLifetime   = 4 * time.Second
)

type actionType int

const (
	actionStart actionType = iota
	actionStop
	actionRestart
	actionRemove
)

type actionPrompt struct {
	action         actionType
	container      domain.Container
	requireConfirm bool
}

type containersMsg struct {
	items []domain.Container
	err   error
	all   bool
}

type refreshTickMsg time.Time

type actionResultMsg struct {
	action    actionType
	container domain.Container
	err       error
}

type notificationKind int

const (
	notifNone notificationKind = iota
	notifInfo
	notifError
)

func Run(client docker.Client) error {
	m := newModel(client)
	p := tea.NewProgram(m, tea.WithAltScreen())
	return p.Start()
}

type model struct {
	client docker.Client

	table         table.Model
	containers    []domain.Container
	fetching      bool
	showAll       bool
	lastErr       error
	lastUpdated   time.Time
	width         int
	height        int
	statusLine    string
	refreshEvery  time.Duration
	pendingAction *actionPrompt
	notification  string
	notifUntil    time.Time
	notifKind     notificationKind
}

func newModel(client docker.Client) model {
	t := table.New(table.WithColumns([]table.Column{
		{Title: "STATUS", Width: 12},
		{Title: "NAME", Width: 24},
		{Title: "IMAGE", Width: 24},
		{Title: "STACK", Width: 16},
		{Title: "PORTS", Width: 22},
		{Title: "CREATED", Width: 14},
	}),
	)
	t.SetStyles(tableStyles())
	t.SetHeight(12)
	t.Focus()
	return model{
		client:       client,
		table:        t,
		refreshEvery: defaultRefreshInterval,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.loadContainers(), scheduleRefresh(m.refreshEvery))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.table.SetWidth(max(60, msg.Width-4))
		m.table.SetHeight(max(10, msg.Height-10))
		return m, nil
	case containersMsg:
		m.fetching = false
		m.lastErr = msg.err
		if msg.err == nil {
			m.containers = msg.items
			m.statusLine = fmt.Sprintf("%d containers (%s)", len(msg.items), ternary(msg.all, "all", "running"))
			m.lastUpdated = time.Now()
			m.table.SetRows(buildRows(msg.items))
		}
		return m, scheduleRefresh(m.refreshEvery)
	case refreshTickMsg:
		if !m.fetching {
			return m, tea.Batch(m.loadContainers(), scheduleRefresh(m.refreshEvery))
		}
		return m, scheduleRefresh(m.refreshEvery)
	case actionResultMsg:
		m.fetching = false
		m.pendingAction = nil
		if msg.err != nil {
			m.lastErr = msg.err
			m.flash(fmt.Sprintf("Error al %s %s: %v", actionVerb(msg.action), msg.container.Name, msg.err), notifError)
		} else {
			m.lastErr = nil
			m.flash(fmt.Sprintf("Listo: %s %s", actionVerbPast(msg.action), msg.container.Name), notifInfo)
		}
		cmds := []tea.Cmd{scheduleRefresh(m.refreshEvery)}
		if msg.err == nil {
			cmds = append(cmds, m.loadContainers())
		}
		return m, tea.Batch(cmds...)
	case tea.KeyMsg:
		if m.pendingAction != nil {
			switch msg.String() {
			case "y", "Y":
				cmd := m.executePendingAction()
				return m, cmd
			case "n", "N", "esc":
				m.flash("Accion cancelada", notifInfo)
				m.pendingAction = nil
				return m, nil
			default:
				return m, nil
			}
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			if !m.fetching {
				return m, tea.Batch(m.loadContainers(), scheduleRefresh(m.refreshEvery))
			}
		case "a":
			if !m.fetching {
				m.showAll = !m.showAll
				return m, tea.Batch(m.loadContainers(), scheduleRefresh(m.refreshEvery))
			}
		case "s":
			cmd := m.queueAction(actionStart)
			return m, cmd
		case "x":
			cmd := m.queueAction(actionStop)
			return m, cmd
		case "R":
			cmd := m.queueAction(actionRestart)
			return m, cmd
		case "d":
			cmd := m.queueAction(actionRemove)
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	header := lipgloss.JoinHorizontal(lipgloss.Top, titleStyle.Render("DockCrew"), subtitleStyle.Render(statusText(m)))
	var body string
	if m.fetching {
		body = infoStyle.Render("Actualizando contenedores...")
	}
	if m.lastErr != nil {
		body = errorStyle.Render("Error: " + m.lastErr.Error())
	}
	tableView := m.table.View()
	detail := m.renderDetail()
	notice := m.notificationView()
	prompt := m.renderPrompt()
	footer := footerStyle.Render("↑/↓ mover • a mostrar todos • r refrescar • s start • x stop • R restart • d remove • q salir")

	sections := []string{header}
	if body != "" {
		sections = append(sections, body)
	}
	sections = append(sections, tableView)
	if notice != "" {
		sections = append(sections, notice)
	}
	sections = append(sections, detail)
	if prompt != "" {
		sections = append(sections, prompt)
	}
	sections = append(sections, footer)
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m model) renderDetail() string {
	if len(m.containers) == 0 {
		return infoStyle.Render("Sin contenedor seleccionado")
	}
	idx := m.table.Cursor()
	if idx < 0 || idx >= len(m.containers) {
		return ""
	}
	ct := m.containers[idx]
	builder := strings.Builder{}
	fmt.Fprintf(&builder, "ID: %s\n", ct.ID)
	fmt.Fprintf(&builder, "Estado: %s (%s)\n", ct.State, ct.Status)
	if ct.Stack != "" {
		fmt.Fprintf(&builder, "Stack: %s\n", ct.Stack)
	}
	if len(ct.Labels) > 0 {
		fmt.Fprintf(&builder, "Labels: %s\n", formatMap(ct.Labels))
	}
	if len(ct.Ports) > 0 {
		fmt.Fprintf(&builder, "Puertos: %s\n", formatPorts(ct.Ports))
	}
	if !ct.CreatedAt.IsZero() {
		fmt.Fprintf(&builder, "Creado: %s (%s)\n", ct.CreatedAt.Format(time.RFC822), humanizeDuration(time.Since(ct.CreatedAt)))
	}
	return detailStyle.Render(builder.String())
}

func (m *model) loadContainers() tea.Cmd {
	m.fetching = true
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
		defer cancel()
		items, err := m.client.ListContainersAll(ctx, m.showAll)
		return containersMsg{items: items, err: err, all: m.showAll}
	}
}

func (m *model) queueAction(action actionType) tea.Cmd {
	ct, ok := m.selectedContainer()
	if !ok {
		m.flash("Selecciona un contenedor", notifInfo)
		return nil
	}
	prompt := &actionPrompt{action: action, container: ct, requireConfirm: action != actionStart}
	if prompt.requireConfirm {
		m.pendingAction = prompt
		return nil
	}
	m.pendingAction = nil
	return m.executeAction(*prompt)
}

func (m *model) executePendingAction() tea.Cmd {
	if m.pendingAction == nil {
		return nil
	}
	p := *m.pendingAction
	m.pendingAction = nil
	return m.executeAction(p)
}

func (m *model) executeAction(p actionPrompt) tea.Cmd {
	m.fetching = true
	m.statusLine = fmt.Sprintf("%s %s...", actionVerb(p.action), p.container.Name)
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), actionTimeout)
		defer cancel()
		var err error
		switch p.action {
		case actionStart:
			err = m.client.StartContainer(ctx, p.container.ID)
		case actionStop:
			err = m.client.StopContainer(ctx, p.container.ID)
		case actionRestart:
			err = m.client.RestartContainer(ctx, p.container.ID)
		case actionRemove:
			err = m.client.RemoveContainer(ctx, p.container.ID, false)
		}
		return actionResultMsg{action: p.action, container: p.container, err: err}
	}
}

func (m model) selectedContainer() (domain.Container, bool) {
	if len(m.containers) == 0 {
		return domain.Container{}, false
	}
	idx := m.table.Cursor()
	if idx < 0 || idx >= len(m.containers) {
		return domain.Container{}, false
	}
	return m.containers[idx], true
}

func (m *model) flash(msg string, kind notificationKind) {
	if msg == "" {
		return
	}
	m.notification = msg
	m.notifUntil = time.Now().Add(notificationLifetime)
	m.notifKind = kind
}

func (m *model) notificationView() string {
	if m.notification == "" {
		return ""
	}
	if time.Now().After(m.notifUntil) {
		m.notification = ""
		m.notifKind = notifNone
		return ""
	}
	switch m.notifKind {
	case notifError:
		return notificationErrorStyle.Render(m.notification)
	case notifInfo:
		return notificationInfoStyle.Render(m.notification)
	default:
		return notificationInfoStyle.Render(m.notification)
	}
}

func (m model) renderPrompt() string {
	if m.pendingAction == nil {
		return ""
	}
	ct := m.pendingAction.container
	msg := fmt.Sprintf("Confirmar %s %s? [y/n]", actionVerb(m.pendingAction.action), ct.Name)
	return promptStyle.Render(msg)
}

func scheduleRefresh(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return refreshTickMsg(t)
	})
}

func buildRows(containers []domain.Container) []table.Row {
	rows := make([]table.Row, 0, len(containers))
	for _, ct := range containers {
		rows = append(rows, table.Row{
			stilizeStatus(ct.State, ct.Status),
			truncate(ct.Name, 24),
			truncate(ct.Image, 24),
			truncate(ct.Stack, 16),
			truncate(formatPorts(ct.Ports), 22),
			relativeTime(ct.CreatedAt),
		})
	}
	return rows
}

func relativeTime(ts time.Time) string {
	if ts.IsZero() {
		return "" // Docker a veces lo omite
	}
	diff := time.Since(ts)
	if diff < 0 {
		diff = 0
	}
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		return fmt.Sprintf("%dm ago", int(diff/time.Minute))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(diff/time.Hour))
	case diff < 48*time.Hour:
		return "yesterday"
	default:
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	}
}

func stilizeStatus(state, status string) string {
	switch strings.ToLower(state) {
	case "running":
		return statusRunning.Render(status)
	case "exited", "dead":
		return statusStopped.Render(status)
	case "paused":
		return statusPaused.Render(status)
	default:
		return statusOther.Render(status)
	}
}

func statusText(m model) string {
	parts := []string{}
	if m.statusLine != "" {
		parts = append(parts, m.statusLine)
	}
	if m.showAll {
		parts = append(parts, "mostrando todos")
	} else {
		parts = append(parts, "en ejecución")
	}
	if !m.lastUpdated.IsZero() {
		parts = append(parts, "actualizado "+humanizeDuration(time.Since(m.lastUpdated))+" atrás")
	}
	if m.fetching {
		parts = append(parts, "sincronizando...")
	}
	return strings.Join(parts, " • ")
}

func humanizeDuration(d time.Duration) string {
	if d < time.Second {
		return "justo ahora"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d/time.Second))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d/time.Minute))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d/time.Hour))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

func formatMap(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	pairs := make([]string, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(pairs, ", ")
}

func formatPorts(ports []domain.PortBinding) string {
	if len(ports) == 0 {
		return ""
	}
	parts := make([]string, 0, len(ports))
	for _, p := range ports {
		if p.Public != 0 {
			parts = append(parts, fmt.Sprintf("%s:%d->%d/%s", displayIP(p.IP), p.Public, p.Private, p.Type))
		} else {
			parts = append(parts, fmt.Sprintf("%d/%s", p.Private, p.Type))
		}
	}
	return strings.Join(parts, ", ")
}

func displayIP(ip string) string {
	if ip == "" {
		return "*"
	}
	return ip
}

func truncate(s string, limit int) string {
	if len([]rune(s)) <= limit {
		return s
	}
	if limit <= 3 {
		return string([]rune(s)[:limit])
	}
	return string([]rune(s)[:limit-3]) + "..."
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func ternary[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

var (
	titleStyle             = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Background(lipgloss.Color("57")).Padding(0, 1).MarginBottom(1)
	subtitleStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("105")).MarginLeft(2)
	infoStyle              = lipgloss.NewStyle().Foreground(lipgloss.Color("111")).Padding(0, 1)
	errorStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Background(lipgloss.Color("52")).Padding(0, 1)
	footerStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1)
	detailStyle            = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("62")).Padding(0, 1).MarginTop(1)
	notificationInfoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Padding(0, 1)
	notificationErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("124")).Padding(0, 1)
	promptStyle            = lipgloss.NewStyle().BorderStyle(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1).MarginTop(1)

	statusRunning = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	statusStopped = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	statusPaused  = lipgloss.NewStyle().Foreground(lipgloss.Color("178")).Bold(true)
	statusOther   = lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true)
)

func tableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("63")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	return s
}

func actionVerb(a actionType) string {
	switch a {
	case actionStart:
		return "iniciar"
	case actionStop:
		return "detener"
	case actionRestart:
		return "reiniciar"
	case actionRemove:
		return "eliminar"
	default:
		return "accion"
	}
}

func actionVerbPast(a actionType) string {
	switch a {
	case actionStart:
		return "iniciado"
	case actionStop:
		return "detenido"
	case actionRestart:
		return "reiniciado"
	case actionRemove:
		return "eliminado"
	default:
		return "accion"
	}
}
