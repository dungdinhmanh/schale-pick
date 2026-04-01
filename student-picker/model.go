package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Tab int

const (
	TabBrowse Tab = iota
	TabInstalled
)

type screen int

const (
	screenMain screen = iota
	screenSettings
)

var (
	styleGridBorder = lipgloss.NewStyle().
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
)

type model struct {
	width       int
	height      int
	screen      screen
	tab         Tab
	manifest    []Student
	filtered    []Student
	gridIdx     int
	gridCols    int
	gridOffset  int
	searchQuery string
	searchMode  bool
	cache       *Cache
	downloader  *Downloader
	loading     bool
	status      string
	statusIsErr bool
	previewPath string
	iconPaths   map[int]string
}

func newModel() model {
	return model{
		tab:        TabBrowse,
		gridIdx:    0,
		gridCols:   5,
		gridOffset: 0,
		cache:      NewCache(CacheDir(), DefaultCacheSize),
		downloader: NewDownloader(5),
		loading:    true,
		iconPaths:  make(map[int]string),
	}
}

func (m model) Init() tea.Cmd {
	return fetchManifestCmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case manifestLoadedMsg:
		m.manifest = msg.students
		m.filtered = msg.students
		m.loading = false
		if len(m.filtered) > 0 {
			m.previewPath = ensurePortraitCached(m.filtered[0].Id, m.downloader)
		}
		m.preCachePortraits()
		return m, tea.Batch(m.renderKittyImage(), m.getVisibleIconBatch())

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.gridCols = (m.width - PreviewW - 6) / thumbW
		if m.gridCols < 2 {
			m.gridCols = 2
		}
		return m, tea.Batch(m.renderKittyImage(), m.getVisibleIconBatch())

	case tea.KeyPressMsg:
		if m.screen == screenSettings {
			return m.handleSettingsKey(msg)
		}
		if m.searchMode {
			return m.handleSearchKey(msg)
		}
		return m.handleNormalKey(msg)
	}
	return m, nil
}

func (m model) handleNormalKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "i":
		m.screen = screenSettings
		m.status = ""
		return m, nil

	case "tab":
		if m.tab == TabBrowse {
			m.tab = TabInstalled
		} else {
			m.tab = TabBrowse
		}
		m.gridIdx = 0
		return m, nil

	case "left", "h":
		if m.gridIdx > 0 {
			m.gridIdx--
			return m, tea.Batch(m.renderKittyImage(), m.getVisibleIconBatch())
		}

	case "right", "l":
		if m.gridIdx < len(m.getCurrentItems())-1 {
			m.gridIdx++
			return m, tea.Batch(m.renderKittyImage(), m.getVisibleIconBatch())
		}

	case "up", "k":
		m.gridIdx -= m.gridCols
		if m.gridIdx < 0 {
			m.gridIdx = 0
		}
		return m, tea.Batch(m.renderKittyImage(), m.getVisibleIconBatch())

	case "down", "j":
		m.gridIdx += m.gridCols
		maxIdx := len(m.getCurrentItems()) - 1
		if m.gridIdx > maxIdx {
			m.gridIdx = maxIdx
		}
		return m, tea.Batch(m.renderKittyImage(), m.getVisibleIconBatch())

	case "/":
		m.searchMode = true
		m.searchQuery = ""

	case "enter":
		items := m.getCurrentItems()
		if m.gridIdx < 0 || m.gridIdx >= len(items) {
			return m, nil
		}
		student := items[m.gridIdx]
		if m.tab == TabBrowse {
			return m, m.cacheAndSelect(student.Id)
		}
		return m, m.selectInstalled(student.Id)

	case "esc":
		m.searchMode = false
		m.searchQuery = ""
		m.applyFilter()
	}
	return m, nil
}

func (m model) handleSearchKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchMode = false
		m.searchQuery = ""
		m.applyFilter()
		return m, nil

	case "enter":
		m.searchMode = false
		m.applyFilter()
		return m, nil

	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.applyFilter()
		}
		return m, nil

	default:
		if len(msg.Text) > 0 {
			m.searchQuery += msg.Text
			m.applyFilter()
		}
		return m, nil
	}
}

func (m model) handleSettingsKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q", "b", "esc":
		m.screen = screenMain
		m.status = ""
		return m, nil
	}
	return m, nil
}

func (m *model) getCurrentItems() []Student {
	if m.tab == TabBrowse {
		return m.filtered
	}
	cached := m.cache.List()
	result := make([]Student, 0, len(cached))
	for _, c := range cached {
		for _, s := range m.manifest {
			if s.Id == c.StudentId {
				result = append(result, s)
				break
			}
		}
	}
	return result
}

func (m *model) applyFilter() {
	if m.searchQuery == "" {
		m.filtered = m.manifest
	} else {
		m.filtered = filterStudents(m.manifest, m.searchQuery)
	}
	m.gridIdx = 0
}

func (m *model) preCachePortraits() {
	visibleCount := m.gridCols * 3

	items := m.getCurrentItems()
	count := len(items)
	if count == 0 {
		return
	}

	for i := 0; i < count && i < visibleCount; i++ {
		studentId := items[i].Id
		data, err := m.downloader.DownloadPortrait(studentId)
		if err == nil {
			portraitPath := filepath.Join(CacheDir(), "portrait", strconv.Itoa(studentId)+".webp")
			os.MkdirAll(filepath.Dir(portraitPath), 0755)
			os.WriteFile(portraitPath, data, 0644)
		}
	}
}

func (m model) cacheAndSelect(studentId int) tea.Cmd {
	return func() tea.Msg {
		iconData, err := m.downloader.DownloadIcon(studentId)
		if err != nil {
			return cacheErrorMsg{err: err}
		}
		portraitData, err := m.downloader.DownloadPortrait(studentId)
		if err != nil {
			return cacheErrorMsg{err: err}
		}
		m.cache.Put(studentId, iconData, portraitData)

		cached, ok := m.cache.Get(studentId)
		if !ok {
			return cacheErrorMsg{err: fmt.Errorf("cache miss after put")}
		}

		if err := updateFastfetchImage(cached.PortraitPath, m.height); err != nil {
			return cacheErrorMsg{err: err}
		}
		return nil
	}
}

func (m model) selectInstalled(studentId int) tea.Cmd {
	return func() tea.Msg {
		cached, ok := m.cache.Get(studentId)
		if !ok {
			return cacheErrorMsg{err: fmt.Errorf("not in cache")}
		}
		if err := updateFastfetchImage(cached.PortraitPath, m.height); err != nil {
			return cacheErrorMsg{err: err}
		}
		return nil
	}
}

type cacheErrorMsg struct {
	err error
}

func (m model) View() tea.View {
	if m.loading {
		return tea.NewView("Loading student manifest...")
	}
	var content string
	if m.screen == screenSettings {
		content = m.viewSettings()
	} else {
		content = m.viewMain()
	}
	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (m model) viewMain() string {
	gridContent := m.renderGrid()
	previewContent := m.renderPreview()
	tabsContent := m.renderTabs()

	gridW := m.width - PreviewW - 4
	if gridW < thumbW {
		gridW = thumbW
	}

	leftPanel := styleGridBorder.Width(gridW).Render(tabsContent + "\n" + gridContent)
	rightPanel := stylePreviewBorder.Width(PreviewW).Height(PreviewH).Render(previewContent)
	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, " ", rightPanel)

	statusLine := ""
	if m.status != "" {
		if m.statusIsErr {
			statusLine = styleError.Render("! " + m.status)
		} else {
			statusLine = styleSuccess.Render("OK " + m.status)
		}
	}

	help := styleHelp.Render("h/j/k/l: move | /: search | Tab: switch | i: settings | Enter: select | q: quit")

	return lipgloss.JoinVertical(lipgloss.Left, body, statusLine, help)
}

func (m model) viewSettings() string {
	cacheDir := CacheDir()

	cfgLine := lipgloss.Style{}.Foreground(lipgloss.Color("#89B4FA")).Render("Cache: ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(cacheDir)

	cachedItems := m.cache.List()
	statsLine := lipgloss.Style{}.Foreground(lipgloss.Color("#A6E3A1")).Render("Installed: ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(fmt.Sprintf("%d students", len(cachedItems)))

	infoBlock := lipgloss.JoinVertical(lipgloss.Left,
		" "+cfgLine,
		" "+statsLine,
	)

	contentBlock := styleSettingsBorder.Width(m.width - 4).Height(8).Render(
		lipgloss.NewStyle().Padding(1, 2).Render("Settings\n\nPress 'q', 'b', or 'esc' to go back"),
	)

	statusLine := ""
	if m.status != "" {
		if m.statusIsErr {
			statusLine = styleError.Render("! " + m.status)
		} else {
			statusLine = styleSuccess.Render("OK " + m.status)
		}
	}

	help := styleHelp.Render("q/b/esc: back | q: quit")

	return lipgloss.JoinVertical(lipgloss.Left, "", infoBlock, "", contentBlock, statusLine, "", help)
}

func (m model) renderTabs() string {
	var browse, installed string
	if m.tab == TabBrowse {
		browse = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A6E3A1")).Render("[Browse]")
		installed = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(" Installed ")
	} else {
		browse = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(" Browse ")
		installed = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A6E3A1")).Render("[Installed]")
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, browse, installed)
}

func (m model) renderGrid() string {
	items := m.getCurrentItems()
	if len(items) == 0 {
		return lipgloss.NewStyle().
			Width(thumbW*m.gridCols).
			Height(m.height-4).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("#6C7086")).
			Render("No students")
	}

	var lines []string
	rowCount := (len(items) + m.gridCols - 1) / m.gridCols

	for row := 0; row < rowCount; row++ {
		var rowItems []string
		for col := 0; col < m.gridCols; col++ {
			idx := row*m.gridCols + col
			if idx >= len(items) {
				rowItems = append(rowItems, strings.Repeat(" ", thumbW))
			} else {
				student := items[idx]
				item := m.renderGridItem(student, idx == m.gridIdx)
				rowItems = append(rowItems, item)
			}
		}
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, rowItems...))
	}

	return lipgloss.NewStyle().
		Width(m.gridCols * thumbW).
		Render(strings.Join(lines, ""))
}

func (m model) getVisibleIconBatch() tea.Cmd {
	return func() tea.Msg {
		clearImagesTermimg()

		items := m.getCurrentItems()
		if len(items) == 0 {
			return nil
		}

		for i, student := range items {
			iconPath := ensureIconCached(student.Id, m.downloader)
			if iconPath == "" {
				continue
			}

			col := i % m.gridCols
			row := i / m.gridCols

			gridX := 2 + col*thumbW
			gridY := 3 + row*thumbH

			iconW := thumbW - 4
			iconH := thumbH - 3

			_ = renderImageTermimg(iconPath, gridX, gridY, iconW, iconH)
		}
		return nil
	}
}

func (m model) renderGridItem(s Student, selected bool) string {
	var borderColor, nameColor string
	if selected {
		borderColor = "#A6E3A1"
		nameColor = "#A6E3A1"
	} else {
		borderColor = "#45475A"
		nameColor = "#CDD6F4"
	}

	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(thumbW - 1).
		Height(thumbH - 1)

	name := s.PersonalName
	if len(name) > thumbW-4 {
		name = name[:thumbW-4] + ".."
	}

	iconH := thumbH - 3
	iconLines := strings.Repeat("\n", iconH-1)
	iconArea := lipgloss.NewStyle().
		Width(thumbW - 3).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color(nameColor)).
		Render(iconLines)

	nameStr := lipgloss.NewStyle().
		Width(thumbW - 3).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color(nameColor)).
		Render(name)

	content := lipgloss.JoinVertical(lipgloss.Center, iconArea, nameStr)

	return border.Render(content)
}

func (m model) renderPreview() string {
	items := m.getCurrentItems()
	if len(items) == 0 || m.gridIdx >= len(items) {
		return lipgloss.NewStyle().
			Width(PreviewW-2).Height(PreviewH-2).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("#6C7086")).
			Render("No image")
	}

	student := items[m.gridIdx]
	name := student.PersonalName
	if student.FamilyName != "" {
		name += " " + student.FamilyName
	}
	return lipgloss.NewStyle().
		Width(PreviewW-2).
		Height(PreviewH-2).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("#A6E3A1")).
		Render(name)
}

func (m model) renderKittyImage() tea.Cmd {
	return func() tea.Msg {
		items := m.getCurrentItems()
		if m.gridIdx < 0 || m.gridIdx >= len(items) {
			return nil
		}

		student := items[m.gridIdx]
		path := ensurePortraitCached(student.Id, m.downloader)
		if path == "" {
			return nil
		}

		clearImagesTermimg()

		x := m.width - PreviewW - 1
		y := 2
		w, h := PreviewW-2, PreviewH-2

		if err := renderImageTermimg(path, x, y, w, h); err != nil {
			return nil
		}
		return nil
	}
}

func filterStudents(students []Student, query string) []Student {
	if query == "" {
		return students
	}
	result := []Student{}
	lowerQuery := strings.ToLower(query)
	for _, s := range students {
		if strings.Contains(strings.ToLower(s.FamilyName), lowerQuery) ||
			strings.Contains(strings.ToLower(s.PersonalName), lowerQuery) ||
			strings.Contains(strings.ToLower(strconv.Itoa(s.Id)), lowerQuery) {
			result = append(result, s)
		}
	}
	return result
}

type manifestLoadedMsg struct {
	students []Student
}

func fetchManifestCmd() tea.Msg {
	data, err := FetchManifest()
	if err != nil {
		return manifestLoadedMsg{students: []Student{}}
	}
	students, err := parseManifest(data)
	if err != nil {
		return manifestLoadedMsg{students: []Student{}}
	}
	return manifestLoadedMsg{students: students}
}

func parseManifest(data []byte) ([]Student, error) {
	var students []Student
	if err := json.Unmarshal(data, &students); err != nil {
		return nil, err
	}
	return students, nil
}

const (
	fastfetchConfig = "/home/kazukisatou/.config/fastfetch/config.jsonc"
	backupSuffix    = ".bak"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
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

const (
	thumbW = 18
	thumbH = 10
)
