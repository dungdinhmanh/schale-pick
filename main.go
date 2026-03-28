package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	imageDir        string
	studentListFile string
	fastfetchConfig string
	backupSuffix    = ".bak"
	previewW        = 40
	previewH        = 22
)

func init() {
	cwd, _ := os.Getwd()
	testDir := filepath.Join(cwd, "assets")
	if _, err := os.Stat(testDir); err == nil {
		imageDir = testDir
		studentListFile = ""
	} else {
		home, _ := user.Current()
		h := home.HomeDir
		imageDir = filepath.Join(h, "students")
		studentListFile = filepath.Join(h, "students", "students.txt")
	}
	fastfetchConfig = filepath.Join(os.Getenv("HOME"), ".config", "fastfetch", "config.jsonc")
}

type screen int

const (
	screenMain screen = iota
	screenSettings
)

type Student struct {
	ID       string
	Name     string
	ImageExt string
}

func (s Student) ImagePath() string {
	return filepath.Join(imageDir, s.ID+s.ImageExt)
}
func (s Student) Title() string       { return s.Name }
func (s Student) Description() string { return s.ID }
func (s Student) FilterValue() string { return s.Name + " " + s.ID }

type menuItem struct{ title, desc string }

func (m menuItem) Title() string       { return m.title }
func (m menuItem) Description() string { return m.desc }
func (m menuItem) FilterValue() string { return m.title }

type settingsActionMsg struct{ err error }
type renderImageMsg struct{ path string }

// ─── Kitty Graphics Protocol ──────────────────────────────────────────────────

const (
	kittyChunkSize = 4096
)

// renderKittyImage sends image using Kitty Graphics Protocol
// Based on fastfetch implementation: uses base64-encoded PNG data
func renderKittyImage(path string, col, row, w, h int) {
	if path == "" {
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	// Convert to PNG using ImageMagick for consistent format
	cmd := exec.Command("magick", "convert", path, "png:-")
	var pngBuf strings.Builder
	cmd.Stdout = &pngBuf
	if err := cmd.Run(); err != nil {
		// Fallback: use original file data
		pngBuf.Reset()
		pngBuf.Write(data)
	}

	imgData := pngBuf.String()
	if len(imgData) == 0 {
		return
	}

	// Base64 encode the image data
	encoded := base64.StdEncoding.EncodeToString([]byte(imgData))

	// Send in chunks (Kitty protocol requires chunking for large data)
	for i := 0; i < len(encoded); i += kittyChunkSize {
		end := i + kittyChunkSize
		if end > len(encoded) {
			end = len(encoded)
		}
		chunk := encoded[i:end]

		isLast := (end >= len(encoded))

		var seq string
		if i == 0 {
			// First chunk: full header with dimensions
			// a=T: transmit-and-display action
			// f=32: RGBA format
			// s=,v=: pixel dimensions
			// o=z: zlib compression
			// m=1/m=0: more chunks indicator
			seq = fmt.Sprintf("\033_Ga=T,f=32,s=%d,v=%d,o=z,m=%d;%s\033\\",
				w*8, h*16, boolToInt(!isLast), chunk)
		} else {
			// Subsequent chunks
			seq = fmt.Sprintf("\033_Gm=%d;%s\033\\",
				boolToInt(!isLast), chunk)
		}
		os.Stderr.WriteString(seq)
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func renderImagePreview(path string, xOffset, w, h int) {
	if path == "" || !fileExists(path) {
		return
	}

	// Check TERM for reliable Kitty detection
	if strings.Contains(os.Getenv("TERM"), "kitty") {
		// xOffset is calculated by caller based on terminal width
		// The key is C=1 composition mode so image persists
		renderKittyImage(path, xOffset, 1, w, h)
		return
	}

	if cmd := exec.Command("which", "chafa"); cmd.Run() == nil {
		cmd := exec.Command("chafa", "--size", fmt.Sprintf("%dx%d", w, h), path)
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

var (
	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#CDD6F4")).
			Background(lipgloss.Color("#313244")).
			Padding(0, 1)

	styleListBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#89B4FA"))

	stylePreviewBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#A6E3A1"))

	styleSettingsBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#CBA6F7"))

	styleHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086")).
			Italic(true)

	styleSuccess = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1")).
			Bold(true)

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F38BA8")).
			Bold(true)

	styleTagOK = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1E1E2E")).
			Background(lipgloss.Color("#A6E3A1")).
			Padding(0, 1).Bold(true)

	styleTagWarn = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1E1E2E")).
			Background(lipgloss.Color("#FAB387")).
			Padding(0, 1).Bold(true)

	styleSettingsTitle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#CDD6F4")).
				Background(lipgloss.Color("#313244")).
				Padding(0, 1)
)

type model struct {
	screen        screen
	width         int
	height        int
	status        string
	statusIsErr   bool
	list          list.Model
	students      []Student
	currentImage  string
	settingsList  list.Model
	viewingConfig string
}

func newModel(students []Student) model {
	items := make([]list.Item, len(students))
	for i := range students {
		items[i] = students[i]
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#89B4FA")).
		Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#74C7EC"))
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.
		Foreground(lipgloss.Color("#6C7086"))

	l := list.New(items, delegate, 40, 20)
	l.Title = "Select Student Image"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.Styles.Title = styleTitle
	l.KeyMap.CursorUp.SetKeys("up", "ctrl+p", "k")
	l.KeyMap.CursorDown.SetKeys("down", "ctrl+n", "j")

	m := model{
		list:         l,
		students:     students,
		settingsList: newSettingsList(),
	}
	if len(students) > 0 {
		m.currentImage = students[0].ImagePath()
	}
	return m
}

func newSettingsList() list.Model {
	backupPath := fastfetchConfig + backupSuffix
	backupExists := fileExists(backupPath)
	cfgExists := fileExists(fastfetchConfig)

	items := []list.Item{
		menuItem{title: "Backup Config", desc: "Save current config"},
		menuItem{title: "Restore Config", desc: "Restore from backup"},
		menuItem{title: "Open Config", desc: "Open config folder"},
		menuItem{title: "Open Images", desc: "Open images folder"},
		menuItem{title: "View Config", desc: "Display config"},
	}

	if !cfgExists {
		items[0] = menuItem{title: "Backup Config", desc: "Config not found"}
		items[1] = menuItem{title: "Restore Config", desc: "Config not found"}
	}
	if !backupExists {
		items[1] = menuItem{title: "Restore Config", desc: "No backup found"}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#CBA6F7")).
		Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#F5C2E7"))

	sl := list.New(items, delegate, 50, 6)
	sl.Title = "Settings"
	sl.SetShowStatusBar(false)
	sl.SetFilteringEnabled(false)
	sl.SetShowHelp(false)
	sl.Styles.Title = styleSettingsTitle
	return sl
}

func (m model) Init() tea.Cmd {
	if m.currentImage != "" {
		return func() tea.Msg { return renderImageMsg{path: m.currentImage} }
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		listW := m.width - previewW - 6
		if listW < 30 {
			listW = 30
		}
		m.list.SetSize(listW, m.height-4)
		m.settingsList.SetSize(m.width-8, 8)
		if m.screen == screenMain {
			return m, func() tea.Msg { return renderImageMsg{path: m.currentImage} }
		}
		return m, nil

	case renderImageMsg:
		if m.screen == screenMain && msg.path != "" {
			// Calculate x offset for right panel: left panel width + gap + border
			listW := m.width - previewW - 6
			if listW < 30 {
				listW = 30
			}
			xOffset := listW + 3 // gap(1) + border(1) + padding(1) = 3
			renderImagePreview(msg.path, xOffset, previewW-2, previewH-2)
		}
		return m, nil

	case settingsActionMsg:
		if msg.err != nil {
			m.status = "ERROR: " + msg.err.Error()
			m.statusIsErr = true
		} else {
			m.status = "SUCCESS"
			m.statusIsErr = false
		}
		m.settingsList = newSettingsList()
		return m, nil

	case tea.KeyMsg:
		switch m.screen {
		case screenMain:
			return m.updateMain(msg)
		case screenSettings:
			return m.updateSettings(msg)
		}
	}

	var cmd tea.Cmd
	if m.screen == screenMain {
		prevIdx := m.list.Index()
		m.list, cmd = m.list.Update(msg)
		if m.list.Index() != prevIdx {
			if s, ok := m.list.SelectedItem().(Student); ok {
				m.currentImage = s.ImagePath()
			}
		}
	} else {
		m.settingsList, cmd = m.settingsList.Update(msg)
	}
	return m, cmd
}

func (m model) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Nếu đang filter thì pass hết vào list, không intercept
	if m.list.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "i":
		m.screen = screenSettings
		m.status = ""
		m.settingsList = newSettingsList()
		return m, nil

	case "enter":
		s, ok := m.list.SelectedItem().(Student)
		if !ok {
			return m, nil
		}
		if err := updateFastfetchImage(s.ImagePath()); err != nil {
			m.status = fmt.Sprintf("ERROR: %v", err)
			m.statusIsErr = true
		} else {
			m.status = fmt.Sprintf("SUCCESS: %s", s.Name)
			m.statusIsErr = false
		}
		return m, tea.Quit
	}

	// Pass tất cả key khác (j/k/↑/↓//) xuống list
	prevIdx := m.list.Index()
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	if m.list.Index() != prevIdx {
		if s, ok := m.list.SelectedItem().(Student); ok {
			m.currentImage = s.ImagePath()
			// Render ảnh mới khi selection thay đổi
			return m, tea.Batch(cmd, func() tea.Msg {
				return renderImageMsg{path: m.currentImage}
			})
		}
	}
	return m, cmd
}

func (m model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "q", "b", "esc":
		m.screen = screenMain
		m.status = ""
		m.viewingConfig = ""
		// Re-render ảnh khi quay về main
		return m, func() tea.Msg { return renderImageMsg{path: m.currentImage} }

	case "enter":
		item, ok := m.settingsList.SelectedItem().(menuItem)
		if !ok {
			return m, nil
		}
		switch item.title {
		case "Backup Config":
			return m, func() tea.Msg { return settingsActionMsg{err: installConfig()} }
		case "Restore Config":
			return m, func() tea.Msg { return settingsActionMsg{err: uninstallConfig()} }
		case "Open Config":
			cfgDir := filepath.Dir(fastfetchConfig)
			exec.Command("xdg-open", cfgDir).Start()
			m.status = "Opened"
			m.statusIsErr = false
			return m, nil
		case "Open Images":
			exec.Command("xdg-open", imageDir).Start()
			m.status = "Opened"
			m.statusIsErr = false
			return m, nil
		case "View Config":
			data, err := os.ReadFile(fastfetchConfig)
			if err != nil {
				m.status = fmt.Sprintf("Error: %v", err)
				m.statusIsErr = true
			} else {
				m.viewingConfig = string(data)
			}
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}
	switch m.screen {
	case screenSettings:
		return m.viewSettings()
	default:
		return m.viewMain()
	}
}

func (m model) viewMain() string {
	listW := m.width - previewW - 6
	if listW < 30 {
		listW = 30
	}

	leftPanel := styleListBorder.Width(listW).Render(m.list.View())
	previewContent := m.renderPreview()
	rightPanel := stylePreviewBorder.Width(previewW).Height(previewH).Render(previewContent)
	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, " ", rightPanel)

	statusLine := ""
	if m.status != "" {
		if m.statusIsErr {
			statusLine = styleError.Render("! " + m.status)
		} else {
			statusLine = styleSuccess.Render("OK " + m.status)
		}
	}

	help := styleHelp.Render("j/k: up/down | /: filter | Enter: select | i: settings | q: quit")

	return lipgloss.JoinVertical(lipgloss.Left, body, statusLine, help)
}

func (m model) renderPreview() string {
	if m.currentImage == "" || !fileExists(m.currentImage) {
		return lipgloss.NewStyle().
			Width(previewW-2).Height(previewH-2).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("#6C7086")).
			Render("No image")
	}

	imgW, imgH := getImageDimensions(m.currentImage)
	fastW, fastH := calculateFastfetchSize(imgH, imgW)

	info, _ := os.Stat(m.currentImage)
	size := formatFileSize(info.Size())

	preview := lipgloss.NewStyle().
		Width(previewW-2).Height(previewH-2).
		Align(lipgloss.Left, lipgloss.Top).
		Foreground(lipgloss.Color("#A6E3A1"))

	return preview.Render(fmt.Sprintf(`
  [Preview]

  %s
  %dx%d (%s)
  fastfetch: %dx%d

  (Run fastfetch to
   see actual image)
`, filepath.Base(m.currentImage), imgW, imgH, size, fastW, fastH))
}

func calculateFastfetchSize(imgH, imgW int) (int, int) {
	maxW, maxH := 40, 20
	minW, minH := 20, 10

	if imgW == 0 {
		imgW = 1
	}
	aspect := float64(imgH) / float64(imgW)

	// Portrait: height > width
	if aspect >= 1.0 {
		// Use max height → gives narrower display for tall images
		h := maxH
		w := int(float64(h) / aspect)
		if w < minW {
			w = minW
		}
		return w, h
	}

	// Landscape or nearly-square: use max width → gives wider display
	w := maxW
	h := int(float64(w) * aspect)
	if h < minH {
		h = minH
	}
	return w, h
}

func (m model) viewSettings() string {
	backupPath := fastfetchConfig + backupSuffix
	backupExists := fileExists(backupPath)

	var badge string
	if backupExists {
		badge = styleTagOK.Render("BACKUP OK")
	} else {
		badge = styleTagWarn.Render("NO BACKUP")
	}

	cfgLine := lipgloss.Style{}.Foreground(lipgloss.Color("#89B4FA")).Render("Config: ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(truncatePath(fastfetchConfig))
	bakLine := lipgloss.Style{}.Foreground(lipgloss.Color("#A6E3A1")).Render("Backup: ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(truncatePath(backupPath))
	imgLine := lipgloss.Style{}.Foreground(lipgloss.Color("#CBA6F7")).Render("Images: ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(truncatePath(imageDir))

	infoBlock := lipgloss.JoinVertical(lipgloss.Left,
		" "+badge,
		" "+cfgLine,
		" "+bakLine,
		" "+imgLine,
	)

	var menuBlock string
	if m.viewingConfig != "" {
		lines := strings.Split(m.viewingConfig, "\n")
		var displayLines []string
		for i := 0; i < len(lines) && i < 20; i++ {
			displayLines = append(displayLines, lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Render("  "+lines[i]))
		}
		menuBlock = styleSettingsBorder.Width(m.width - 4).Render(
			lipgloss.JoinVertical(lipgloss.Left, displayLines...),
		)
	} else {
		menuBlock = styleSettingsBorder.Width(m.width - 4).Render(m.settingsList.View())
	}

	statusLine := ""
	if m.status != "" {
		if m.statusIsErr {
			statusLine = styleError.Render("! " + m.status)
		} else {
			statusLine = styleSuccess.Render("OK " + m.status)
		}
	}

	help := styleHelp.Render("j/k: move | Enter: execute | q: back")

	return lipgloss.JoinVertical(lipgloss.Left, "", infoBlock, "", menuBlock, statusLine, "", help)
}

func truncatePath(path string) string {
	if len(path) > 50 {
		return "..." + path[len(path)-47:]
	}
	return path
}

func installConfig() error {
	bakPath := fastfetchConfig + backupSuffix
	if fileExists(bakPath) {
		return fmt.Errorf("backup already exists")
	}
	if !fileExists(fastfetchConfig) {
		return fmt.Errorf("config not found")
	}
	return copyFile(fastfetchConfig, bakPath)
}

func uninstallConfig() error {
	bakPath := fastfetchConfig + backupSuffix
	if !fileExists(bakPath) {
		return fmt.Errorf("no backup to restore")
	}
	if err := copyFile(bakPath, fastfetchConfig); err != nil {
		return err
	}
	return os.Remove(bakPath)
}

func updateFastfetchImage(imagePath string) error {
	// Auto backup if no backup exists
	bakPath := fastfetchConfig + backupSuffix
	if !fileExists(bakPath) && fileExists(fastfetchConfig) {
		if err := copyFile(fastfetchConfig, bakPath); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
	}

	// Calculate width based on image aspect ratio
	imgW, imgH := getImageDimensions(imagePath)
	fastW := 30
	if imgW > 0 {
		aspect := float64(imgH) / float64(imgW)
		if aspect >= 1.0 {
			fastW = 45
		} else {
			fastW = 35
		}
		if fastW < 15 {
			fastW = 15
		}
	}

	// Use jq to safely update only the fields we need
	cmd := exec.Command("jq",
		fmt.Sprintf(`.logo.source = "%s" | .logo.type = "kitty" | .logo.width = %d`, imagePath, fastW),
		fastfetchConfig,
	)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("jq failed: %w", err)
	}

	return os.WriteFile(fastfetchConfig, output, 0644)
}

// replaceConfigValue removed - using jq instead

func getImageDimensions(imagePath string) (int, int) {
	cmd := exec.Command("magick", "identify", "-format", "%w %h", imagePath)
	out, err := cmd.Output()
	if err != nil {
		return 30, 20
	}
	var w, h int
	fmt.Sscanf(string(out), "%d %d", &w, &h)
	if w == 0 || h == 0 {
		return 30, 20
	}
	return w, h
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, info.Mode())
}

func loadStudents() []Student {
	var students []Student

	if data, err := os.ReadFile(studentListFile); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "|", 2)
			id := strings.TrimSpace(parts[0])
			name := id
			if len(parts) == 2 {
				name = strings.TrimSpace(parts[1])
			}
			ext := findImageExt(id)
			if ext == "" {
				continue
			}
			students = append(students, Student{ID: id, Name: name, ImageExt: ext})
		}
		if len(students) > 0 {
			return students
		}
	}

	entries, err := os.ReadDir(imageDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext != ".webp" && ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ext)
		students = append(students, Student{ID: id, Name: id, ImageExt: ext})
	}
	sort.Slice(students, func(i, j int) bool { return students[i].ID < students[j].ID })
	return students
}

func findImageExt(id string) string {
	for _, ext := range []string{".webp", ".png", ".jpg", ".jpeg"} {
		if _, err := os.Stat(filepath.Join(imageDir, id+ext)); err == nil {
			return ext
		}
	}
	return ""
}

func main() {
	students := loadStudents()
	if len(students) == 0 {
		fmt.Fprintf(os.Stderr, "No students found in %s\n", imageDir)
		os.Exit(1)
	}

	p := tea.NewProgram(newModel(students), tea.WithAltScreen(), tea.WithMouseCellMotion())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fm := finalModel.(model)
	if fm.status != "" && !fm.statusIsErr {
		fmt.Println(fm.status)
	}
}
