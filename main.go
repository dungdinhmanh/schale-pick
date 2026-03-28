package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
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

var debugLog *os.File

func init() {
	var err error
	debugLog, err = os.OpenFile("/tmp/student-picker.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		debugLog = nil
	}
}

func debug(format string, args ...interface{}) {
	if debugLog != nil {
		msg := fmt.Sprintf(format, args...)
		debugLog.WriteString(msg + "\n")
		debugLog.Sync()
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
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

	if strings.Contains(os.Getenv("TERM"), "kitty") || os.Getenv("KITTY_WINDOW_ID") != "" {
		debug("TERM=%s, using termimg with Kitty protocol", os.Getenv("TERM"))
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
		return m, nil

	case tea.KeyMsg:
		switch m.screen {
		case screenMain:
			return m.updateMain(msg)
		case screenSettings:
			return m.updateSettings(msg)
		}

	default:
		var cmd tea.Cmd
		if m.screen == screenMain {
			prevIdx := m.list.Index()
			m.list, cmd = m.list.Update(msg)
			if m.list.Index() != prevIdx {
				if s, ok := m.list.SelectedItem().(Student); ok {
					m.currentImage = s.ImagePath()
					cmd = m.renderKittyImage()
				}
			}
		} else {
			m.settingsList, cmd = m.settingsList.Update(msg)
		}
		return m, cmd
	}

	return m, nil
}

func (m model) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		if err := updateFastfetchImage(s.ImagePath(), m.height); err != nil {
			m.status = fmt.Sprintf("ERROR: %v", err)
			m.statusIsErr = true
		} else {
			m.status = fmt.Sprintf("SUCCESS: %s", s.Name)
			m.statusIsErr = false
		}
		return m, tea.Quit
	}

	prevIdx := m.list.Index()
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	if m.list.Index() != prevIdx {
		if s, ok := m.list.SelectedItem().(Student); ok {
			m.currentImage = s.ImagePath()
			cmd = m.renderKittyImage()
		}
	}
	return m, cmd
}

func (m model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "q", "b", "esc":
		m.screen = screenMain
		m.status = ""
		m.viewingConfig = ""
		return m, nil

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
	default:
		m.settingsList, cmd = m.settingsList.Update(msg)
		return m, cmd
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
			Render("No image\n\n\n\n\n")
	}

	imgW, imgH := getImageDimensions(m.currentImage)
	fastW, fastH := calculateFastfetchSize(imgH, imgW)
	info, _ := os.Stat(m.currentImage)
	size := formatFileSize(info.Size())

	isKitty := strings.Contains(os.Getenv("TERM"), "kitty") || os.Getenv("KITTY_WINDOW_ID") != ""
	if !isKitty {
		return lipgloss.NewStyle().
			Width(previewW-2).Height(previewH-2).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("#6C7086")).
			Render(fmt.Sprintf(`No Kitty terminal

%s
%d×%d (%s)
fastfetch: %d×%d`, filepath.Base(m.currentImage), imgW, imgH, size, fastW, fastH))
	}

	return lipgloss.NewStyle().
		Width(previewW - 2).
		Height(previewH - 2).
		Render("")
}

func (m model) renderKittyImage() tea.Cmd {
	return func() tea.Msg {
		if m.currentImage == "" || !fileExists(m.currentImage) {
			return nil
		}

		isKitty := strings.Contains(os.Getenv("TERM"), "kitty") || os.Getenv("KITTY_WINDOW_ID") != ""
		if !isKitty {
			return nil
		}

		imagePath := m.currentImage
		needsConversion := strings.ToLower(filepath.Ext(m.currentImage)) == ".webp"
		if needsConversion {
			tmpFile := filepath.Join(os.TempDir(), "student-picker-"+filepath.Base(m.currentImage)+".png")
			cmd := exec.Command("magick", "convert", m.currentImage, tmpFile)
			if err := cmd.Run(); err != nil {
				debug("FAIL: magick convert: %v", err)
				return nil
			}
			imagePath = tmpFile
			defer os.Remove(tmpFile)
		}

		ext := strings.ToLower(filepath.Ext(imagePath))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			return nil
		}

		tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
		if err != nil {
			debug("FAIL: open /dev/tty: %v", err)
			return nil
		}
		defer tty.Close()

		tty.WriteString("\x1b_Ga=d,d=A\x1b\\")

		x := m.width - previewW - 1
		y := 2
		w, h := previewW-3, previewH-2

		buf := make([]byte, 0, w*h)
		for i := 0; i < h; i++ {
			buf = append(buf, fmt.Sprintf("\x1b[%d;%dH", y+i, x+1)...)
			buf = append(buf, strings.Repeat(" ", w)...)
		}
		tty.Write(buf)

		cmd := exec.Command("kitty", "+icat", "--silent",
			"--stdin=no",
			"--transfer-mode=stream",
			"--place", fmt.Sprintf("%dx%d@%dx%d", w, h, x, y),
			imagePath,
		)
		cmd.Stdin = nil
		cmd.Stdout = tty
		cmd.Stderr = tty

		if err := cmd.Run(); err != nil {
			debug("FAIL: kitty icat: %v", err)
			return nil
		}

		debug("OK: rendered image via kitty icat at x=%d,y=%d", x, y)
		return nil
	}
}

func calculateFastfetchSize(imgH, imgW int) (int, int) {
	maxW, maxH := 40, 20
	minW, minH := 20, 10

	if imgW == 0 {
		imgW = 1
	}
	aspect := float64(imgH) / float64(imgW)

	if aspect >= 1.0 {
		h := maxH
		w := int(float64(h) / aspect)
		if w < minW {
			w = minW
		}
		return w, h
	}

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

func updateFastfetchImage(imagePath string, termLines int) error {
	bakPath := fastfetchConfig + backupSuffix
	if !fileExists(bakPath) && fileExists(fastfetchConfig) {
		if err := copyFile(fastfetchConfig, bakPath); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
	}

	imgW, imgH := getImageDimensions(imagePath)
	if imgW == 0 || imgH == 0 {
		imgW, imgH = 30, 20
	}

	targetHeight := float64(termLines) * 0.88
	cellRatio := 0.544
	fastW := int(targetHeight * float64(imgW) / float64(imgH) / cellRatio)

	minW := 15
	maxW := 50
	if fastW < minW {
		fastW = minW
	}
	if fastW > maxW {
		fastW = maxW
	}

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
