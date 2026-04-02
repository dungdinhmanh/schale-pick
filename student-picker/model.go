package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

type modalKind int

const (
	modalNone modalKind = iota
	modalBackupConfirm
)

type modal struct {
	kind      modalKind
	message   string
	studentId int
	activeBtn int
}

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

	styleModalBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#F38BA8")).
				Padding(1, 2)

	styleHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086")).
			Italic(true)

	styleSuccess = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1")).
			Bold(true)

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F38BA8")).
			Bold(true)

	styleTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89B4FA")).
			Bold(true)

	styleBtnActive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CDD6F4")).
			Background(lipgloss.Color("#89B4FA")).
			Padding(0, 2).
			Bold(true)

	styleBtnInactive = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6C7086")).
				Background(lipgloss.Color("#45475A")).
				Padding(0, 2)

	styleMasterBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#CBA6F7")).
			Padding(0, 1)

	styleSearchBar = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#89B4FA")).
			Padding(0, 1)
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
	statusTimer time.Time
	previewPath string
	iconPaths   map[int]string
	modal       modal
	hasMagick   bool
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
		hasMagick:  checkMagick(),
	}
}

func checkMagick() bool {
	_, err := exec.LookPath("magick")
	return err == nil
}

func (m model) Init() tea.Cmd {
	return tea.Batch(fetchManifestCmd, checkMagickCmd)
}

func checkMagickCmd() tea.Msg {
	return magickCheckMsg{available: checkMagick()}
}

type magickCheckMsg struct {
	available bool
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case magickCheckMsg:
		m.hasMagick = msg.available
		return m, nil

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
		// Account for: left border(1) + right border(1) + gap(1) + preview border(1) = 4
		m.gridCols = (m.width - PreviewW - 4) / thumbW
		if m.gridCols < 2 {
			m.gridCols = 2
		}
		m.clampOffset()
		return m, tea.Batch(m.renderKittyImage(), m.getVisibleIconBatch())

	case tea.KeyPressMsg:
		if m.modal.kind != modalNone {
			return m.handleModalKey(msg)
		}
		if m.screen == screenSettings {
			return m.handleSettingsKey(msg)
		}
		if m.searchMode {
			return m.handleSearchKey(msg)
		}
		return m.handleNormalKey(msg)

	case statusClearMsg:
		if time.Now().After(m.statusTimer) {
			m.status = ""
		}
		return m, nil
	}
	return m, nil
}

type statusClearMsg struct{}

func statusClearCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return statusClearMsg{}
	})
}

func (m model) handleModalKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		if m.modal.activeBtn > 0 {
			m.modal.activeBtn--
		}
		return m, nil

	case "right", "l":
		if m.modal.activeBtn < 1 {
			m.modal.activeBtn++
		}
		return m, nil

	case "enter":
		if m.modal.activeBtn == 0 {
			// Yes - proceed with backup and save
			studentId := m.modal.studentId
			m.modal = modal{kind: modalNone}
			return m, m.cacheAndSelectWithBackup(studentId)
		}
		// No - cancel
		m.modal = modal{kind: modalNone}
		return m, nil

	case "esc":
		m.modal = modal{kind: modalNone}
		return m, nil
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
		m.gridOffset = 0
		return m, nil

	case "left", "h":
		if m.gridIdx > 0 {
			m.gridIdx--
			m.clampOffset()
			return m, tea.Batch(m.renderKittyImage(), m.getVisibleIconBatch())
		}

	case "right", "l":
		if m.gridIdx < len(m.getCurrentItems())-1 {
			m.gridIdx++
			m.clampOffset()
			return m, tea.Batch(m.renderKittyImage(), m.getVisibleIconBatch())
		}

	case "up", "k":
		m.gridIdx -= m.gridCols
		if m.gridIdx < 0 {
			m.gridIdx = 0
		}
		m.clampOffset()
		return m, tea.Batch(m.renderKittyImage(), m.getVisibleIconBatch())

	case "down", "j":
		m.gridIdx += m.gridCols
		maxIdx := len(m.getCurrentItems()) - 1
		if m.gridIdx > maxIdx {
			m.gridIdx = maxIdx
		}
		m.clampOffset()
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

func (m *model) clampOffset() {
	items := m.getCurrentItems()
	if len(items) == 0 {
		m.gridOffset = 0
		return
	}

	visibleRows := m.visibleRows()
	if visibleRows <= 0 {
		return
	}

	// Calculate which row the current selection is in
	selRow := m.gridIdx / m.gridCols

	// Calculate first and last visible rows
	firstVisibleRow := m.gridOffset
	lastVisibleRow := m.gridOffset + visibleRows - 1

	// Scroll up if selection is above visible area
	if selRow < firstVisibleRow {
		m.gridOffset = selRow
	}

	// Scroll down if selection is below visible area
	if selRow > lastVisibleRow {
		m.gridOffset = selRow - visibleRows + 1
	}

	// Ensure offset is not negative
	if m.gridOffset < 0 {
		m.gridOffset = 0
	}

	// Ensure we don't scroll past the last row
	maxOffset := (len(items)-1)/m.gridCols - visibleRows + 1
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.gridOffset > maxOffset {
		m.gridOffset = maxOffset
	}
}

func (m model) visibleRows() int {
	masterBorder := 2
	titleLine := 1
	headerLine := 1
	statusLine := 1
	helpLine := 1

	availableHeight := m.height - masterBorder - titleLine - headerLine - statusLine - helpLine

	rows := availableHeight / thumbH
	if rows < 1 {
		rows = 1
	}
	return rows
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
	m.gridOffset = 0
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
		// Check if backup exists
		bakPath := fastfetchConfig + backupSuffix
		if !fileExists(bakPath) && fileExists(fastfetchConfig) {
			// Return modal request instead of auto-backup
			return showModalMsg{
				kind:      modalBackupConfirm,
				message:   "Backup not found. Create backup before saving?",
				studentId: studentId,
			}
		}

		// Backup exists, proceed with normal save
		return m.doCacheAndSelect(studentId)()
	}
}

func (m model) cacheAndSelectWithBackup(studentId int) tea.Cmd {
	return func() tea.Msg {
		// Create backup first
		if fileExists(fastfetchConfig) {
			bakPath := fastfetchConfig + backupSuffix
			if err := copyFile(fastfetchConfig, bakPath); err != nil {
				return cacheErrorMsg{err: fmt.Errorf("backup failed: %w", err)}
			}
		}
		return m.doCacheAndSelect(studentId)()
	}
}

func (m model) doCacheAndSelect(studentId int) tea.Cmd {
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
		return cacheSuccessMsg{studentId: studentId}
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
		return cacheSuccessMsg{studentId: studentId}
	}
}

type cacheErrorMsg struct {
	err error
}

type cacheSuccessMsg struct {
	studentId int
}

type showModalMsg struct {
	kind      modalKind
	message   string
	studentId int
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

	browseCount := len(m.filtered)
	installedCount := len(m.cache.List())

	gridW := m.width - PreviewW - 4
	if gridW < thumbW {
		gridW = thumbW
	}

	var leftContent string
	if m.searchMode {
		searchQuery := m.searchQuery + "█"
		searchBar := styleSearchBar.Width(gridW - 2).Render(searchQuery)
		leftContent = searchBar + "\n" + gridContent
	} else {
		tabsContent := m.renderTabs()
		leftContent = tabsContent + "\n" + gridContent
	}

	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#45475A")).
		Render("│")

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftContent, " ", separator, " ", previewContent)

	titleText := fmt.Sprintf(" Browse (%d) / Installed (%d) ", browseCount, installedCount)
	titleStyled := styleTitle.Render(titleText)
	titleLine := lipgloss.NewStyle().
		Width(m.width - 4).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#CBA6F7")).
		Render(titleStyled)

	masterContent := lipgloss.JoinVertical(lipgloss.Left, titleLine, body)
	masterBox := styleMasterBox.Width(m.width - 2).Render(masterContent)

	statusLine := ""
	if m.status != "" {
		if m.statusIsErr {
			statusLine = styleError.Render("! " + m.status)
		} else {
			statusLine = styleSuccess.Render("✓ " + m.status)
		}
	}

	var help string
	if m.searchMode {
		help = styleHelp.Render("Type to search... | Enter: confirm | Esc: cancel")
	} else {
		help = styleHelp.Render("h/j/k/l: move | /: search | Tab: switch | i: settings | Enter: select | q: quit")
	}

	mainContent := lipgloss.JoinVertical(lipgloss.Left, masterBox, statusLine, help)

	if m.modal.kind != modalNone {
		return m.renderModalOverlay(mainContent)
	}

	return mainContent
}

func renderBorderTitle(title string, panelWidth int) string {
	// Create a line like: ┌─ Title ───────────────────
	// The title sits in the middle of the top border
	titleStyled := styleTitle.Render(title)
	titleLen := lipgloss.Width(titleStyled)

	// Calculate padding
	remaining := panelWidth - titleLen - 2 // -2 for corner chars
	if remaining < 0 {
		remaining = 0
	}
	leftPad := remaining / 2

	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA"))
	leftBorder := borderStyle.Render("─" + strings.Repeat("─", leftPad))
	rightBorder := borderStyle.Render(strings.Repeat("─", remaining-leftPad) + "┐")

	return "┌" + leftBorder + titleStyled + rightBorder
}

func (m model) renderModalOverlay(bg string) string {
	// Render modal content
	btnYes := "Yes"
	btnNo := "No"

	if m.modal.activeBtn == 0 {
		btnYes = styleBtnActive.Render("[Yes]")
		btnNo = styleBtnInactive.Render(" No ")
	} else {
		btnYes = styleBtnInactive.Render(" Yes ")
		btnNo = styleBtnActive.Render("[No]")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, btnYes, "  ", btnNo)
	modalContent := lipgloss.JoinVertical(lipgloss.Center,
		"",
		m.modal.message,
		"",
		buttons,
		"",
	)

	modalBox := styleModalBorder.Width(50).Render(modalContent)

	// Center modal on screen
	overlay := lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modalBox,
	)

	// Overlay modal on background
	return placeOverlay(0, 0, overlay, bg, false)
}

func placeOverlay(x, y int, fg, bg string, shadow bool) string {
	fgLines := strings.Split(fg, "\n")
	bgLines := strings.Split(bg, "\n")

	fgHeight := len(fgLines)
	bgHeight := len(bgLines)
	fgWidth := 0
	for _, l := range fgLines {
		if w := lipgloss.Width(l); w > fgWidth {
			fgWidth = w
		}
	}

	if fgHeight >= bgHeight && fgWidth >= lipgloss.Width(bg) {
		return fg
	}

	var result strings.Builder
	for i, bgLine := range bgLines {
		if i > 0 {
			result.WriteByte('\n')
		}
		if i < y || i >= y+fgHeight {
			result.WriteString(bgLine)
			continue
		}

		// Overlay fg line onto bg line at position x
		fgLine := fgLines[i-y]
		bgRunes := []rune(bgLine)
		fgRunes := []rune(fgLine)

		for j := 0; j < len(bgRunes); j++ {
			if j >= x && j < x+len(fgRunes) && j-x < len(fgRunes) {
				result.WriteRune(fgRunes[j-x])
			} else {
				result.WriteRune(bgRunes[j])
			}
		}
	}

	return result.String()
}

func (m model) viewSettings() string {
	cacheDir := CacheDir()

	cfgLine := lipgloss.Style{}.Foreground(lipgloss.Color("#89B4FA")).Render("Cache: ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(cacheDir)

	cachedItems := m.cache.List()
	statsLine := lipgloss.Style{}.Foreground(lipgloss.Color("#A6E3A1")).Render("Installed: ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(fmt.Sprintf("%d students", len(cachedItems)))

	magickStatus := "not available"
	if m.hasMagick {
		magickStatus = "available"
	}
	magickLine := lipgloss.Style{}.Foreground(lipgloss.Color("#CBA6F7")).Render("ImageMagick: ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(magickStatus)

	infoBlock := lipgloss.JoinVertical(lipgloss.Left,
		" "+cfgLine,
		" "+statsLine,
		" "+magickLine,
	)

	settingsTitleLine := renderBorderTitle(" Settings ", m.width-4)
	contentBlock := styleSettingsBorder.Width(m.width - 4).Height(8).Render(
		lipgloss.NewStyle().Padding(1, 2).Render("Settings\n\nPress 'q', 'b', or 'esc' to go back"),
	)
	// Prepend title line
	contentBlock = settingsTitleLine + "\n" + contentBlock

	contentBlock = settingsTitleLine + "\n" + styleSettingsBorder.Width(m.width-4).Render("Press 'q', 'b', or 'esc' to go back")

	statusLine := ""
	if m.status != "" {
		if m.statusIsErr {
			statusLine = styleError.Render("! " + m.status)
		} else {
			statusLine = styleSuccess.Render("✓ " + m.status)
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
	gridW := m.width - PreviewW - 4
	if gridW < thumbW {
		gridW = thumbW
	}

	if len(items) == 0 {
		return lipgloss.NewStyle().
			Width(thumbW*m.gridCols).
			MaxWidth(gridW).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("#6C7086")).
			Render("No students")
	}

	visibleRows := m.visibleRows()
	startRow := m.gridOffset
	endRow := startRow + visibleRows

	totalRows := (len(items) + m.gridCols - 1) / m.gridCols
	if endRow > totalRows {
		endRow = totalRows
	}

	var lines []string

	for row := startRow; row < endRow; row++ {
		var rowItems []string
		for col := 0; col < m.gridCols; col++ {
			idx := row*m.gridCols + col
			if idx >= len(items) {
				rowItems = append(rowItems, strings.Repeat(" ", thumbW-2))
			} else {
				student := items[idx]
				item := m.renderGridItem(student, idx == m.gridIdx)
				rowItems = append(rowItems, item)
			}
		}
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, rowItems...))
	}

	if len(lines) == 0 {
		return ""
	}

	return strings.Join(lines, "\n")
}

func (m model) getVisibleIconBatch() tea.Cmd {
	return func() tea.Msg {
		clearImagesTermimg()

		items := m.getCurrentItems()
		if len(items) == 0 {
			return nil
		}

		visibleRows := m.visibleRows()
		startRow := m.gridOffset
		endRow := startRow + visibleRows

		for row := startRow; row < endRow; row++ {
			for col := 0; col < m.gridCols; col++ {
				idx := row*m.gridCols + col
				if idx >= len(items) {
					continue
				}

				student := items[idx]
				iconPath := ensureIconCached(student.Id, m.downloader)
				if iconPath == "" {
					continue
				}

				visibleRow := row - m.gridOffset
				gridX := 1 + col*thumbW
				gridY := 3 + visibleRow*thumbH

				iconW := thumbW - 4
				iconH := thumbH - 3

				_ = renderImageTermimg(iconPath, gridX, gridY, iconW, iconH)
			}
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

	innerW := thumbW - 2
	innerH := thumbH - 2

	name := s.PersonalName
	maxNameLen := innerW - 2
	if len(name) > maxNameLen {
		name = name[:maxNameLen-1] + "…"
	}

	var lines []string
	for i := 0; i < innerH; i++ {
		lines = append(lines, strings.Repeat(" ", innerW))
	}

	lines[innerH-1] = lipgloss.NewStyle().
		Width(innerW).
		MaxWidth(innerW).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color(nameColor)).
		Render(name)

	content := strings.Join(lines, "\n")

	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor))

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

		x := m.width - PreviewW - 2
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
	thumbW = 14
	thumbH = 7
)
