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
	modalHelp
)

type modal struct {
	kind      modalKind
	message   string
	studentId int
	activeBtn int
}

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
	loading       bool
	isDownloading bool
	downloadPct   float64
	status        string
	statusIsErr   bool
	statusTimer   time.Time
	previewPath   string
	iconPaths     map[int]string
	iconPending   map[int]bool
	modal         modal
	hasMagick     bool
	termType      string
	logoFormat    string
	settingsIdx   int
}

func newModel() model {
	return model{
		tab:         TabBrowse,
		gridIdx:     0,
		gridCols:    5,
		gridOffset:  0,
		cache:       NewCache(CacheDir(), DefaultCacheSize),
		downloader:  NewDownloader(5),
		loading:     true,
		iconPaths:   make(map[int]string),
		iconPending: make(map[int]bool),
		termType:    detectTerminalType(),
		logoFormat:  "kitty",
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

type downloadProgressMsg float64

func tickDownload() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return downloadProgressMsg(0.1)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case downloadProgressMsg:
		if m.isDownloading {
			m.downloadPct += float64(msg)
			if m.downloadPct > 0.95 {
				m.downloadPct = 0.95 // Giữ ở 95% cho đến khi xong thực tế
			}
			return m, tickDownload()
		}
		return m, nil
	case magickCheckMsg:
		m.hasMagick = msg.available
		return m, nil

	case manifestLoadedMsg:
		m.manifest = msg.students
		m.filtered = msg.students
		m.loading = false
		if len(m.filtered) > 0 {
			m.previewPath = getPortraitPath(m.filtered[0].Id)
			// Save meta for offline use only if we got fresh data
			if len(m.manifest) > 0 {
				SaveMeta(m.manifest)
			}
		}
		go m.preCachePortraits()
		return m, tea.Batch(m.renderKittyImage(), m.getVisibleIconBatch(), m.renderVisibleIconsCmd())

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		gridColsW := m.width - PreviewW - 9
		if gridColsW < thumbW {
			gridColsW = thumbW
		}
		m.gridCols = gridColsW / thumbW
		if m.gridCols < 1 {
			m.gridCols = 1
		}
		m.clampOffset()
		return m, tea.Batch(m.renderKittyImage(), m.getVisibleIconBatch(), m.renderVisibleIconsCmd())

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

	case showModalMsg:
		m.modal = modal{kind: msg.kind, message: msg.message, studentId: msg.studentId}
		return m, nil

	case cacheSuccessMsg:
		m.isDownloading = false
		m.downloadPct = 1.0
		m.status = fmt.Sprintf("Selected: %s", msg.name)
		m.statusIsErr = false
		m.statusTimer = time.Now().Add(3 * time.Second)
		return m, statusClearCmd()

	case cacheErrorMsg:
		m.isDownloading = false
		m.status = msg.err.Error()
		m.statusIsErr = true
		m.statusTimer = time.Now().Add(3 * time.Second)
		return m, statusClearCmd()

	case statusClearMsg:
		if time.Now().After(m.statusTimer) {
			m.status = ""
		}
		return m, nil

	case iconDownloadedMsg:
		delete(m.iconPending, msg.studentId)
		if msg.err == nil && msg.iconPath != "" {
			m.iconPaths[msg.studentId] = msg.iconPath
		}
		return m, tea.Batch(m.renderKittyImage(), m.renderVisibleIconsCmd())
	}
	return m, nil
}

type statusClearMsg struct{}

func statusClearCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return statusClearMsg{}
	})
}

// --- Grid & Scroll ---

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

	selRow := m.gridIdx / m.gridCols
	firstVisibleRow := m.gridOffset
	lastVisibleRow := m.gridOffset + visibleRows - 1

	if selRow < firstVisibleRow {
		m.gridOffset = selRow
	}
	if selRow > lastVisibleRow {
		m.gridOffset = selRow - visibleRows + 1
	}
	if m.gridOffset < 0 {
		m.gridOffset = 0
	}

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

func (m model) getInstalledItems() []Student {
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

func (m model) getCurrentItems() []Student {
	if m.tab == TabBrowse {
		return m.filtered
	}
	return m.getInstalledItems()
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

// --- Cache & Select ---

func (m model) cacheAndSelect(studentId int) tea.Cmd {
	return func() tea.Msg {
		bakPath := fastfetchConfigPath() + backupSuffix
		if !fileExists(bakPath) && fileExists(fastfetchConfigPath()) {
			return showModalMsg{
				kind:      modalBackupConfirm,
				message:   "Backup not found. Create backup before saving?",
				studentId: studentId,
			}
		}
		return m.doCacheAndSelect(studentId)()
	}
}

func (m model) cacheAndSelectWithBackup(studentId int) tea.Cmd {
	return func() tea.Msg {
		if fileExists(fastfetchConfigPath()) {
			bakPath := fastfetchConfigPath() + backupSuffix
			if err := copyFile(fastfetchConfigPath(), bakPath); err != nil {
				return cacheErrorMsg{err: fmt.Errorf("backup failed: %w", err)}
			}
		}
		return m.doCacheAndSelect(studentId)()
	}
}

func (m *model) doCacheAndSelect(studentId int) tea.Cmd {
	m.isDownloading = true
	m.downloadPct = 0
	return tea.Batch(
		func() tea.Msg {
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

			if err := updateFastfetchImage(cached.PortraitPath, m.height, m.termType, m.logoFormat); err != nil {
				return cacheErrorMsg{err: err}
			}
			return cacheSuccessMsg{studentId: studentId}
		},
		tickDownload(),
	)
}

func (m model) selectInstalled(studentId int) tea.Cmd {
	return func() tea.Msg {
		cached, ok := m.cache.Get(studentId)
		if !ok {
			return cacheErrorMsg{err: fmt.Errorf("not in cache")}
		}
		if err := updateFastfetchImage(cached.PortraitPath, m.height, m.termType, m.logoFormat); err != nil {
			return cacheErrorMsg{err: err}
		}
		student, _ := m.getStudent(studentId)
		return cacheSuccessMsg{studentId: studentId, name: student.PersonalName}
	}
}

func (m model) getStudent(id int) (Student, bool) {
	for _, s := range m.manifest {
		if s.Id == id {
			return s, true
		}
	}
	return Student{}, false
}

type cacheErrorMsg struct {
	err error
}

type cacheSuccessMsg struct {
	studentId int
	name      string
}

type showModalMsg struct {
	kind      modalKind
	message   string
	studentId int
}

// --- Icon Rendering ---

func (m model) getVisibleIconBatch() tea.Cmd {
	items := m.getCurrentItems()
	if len(items) == 0 {
		return nil
	}

	visibleRows := m.visibleRows()
	startRow := m.gridOffset
	endRow := startRow + visibleRows

	var cmds []tea.Cmd

	for row := startRow; row < endRow; row++ {
		for col := 0; col < m.gridCols; col++ {
			idx := row*m.gridCols + col
			if idx >= len(items) {
				continue
			}

			student := items[idx]
			iconPath := getIconPath(student.Id)

			if _, err := os.Stat(iconPath); err == nil {
				m.iconPaths[student.Id] = iconPath
				continue
			}

			if m.iconPending[student.Id] {
				continue
			}

			m.iconPending[student.Id] = true
			cmds = append(cmds, downloadIconCmd(student.Id, m.downloader))
		}
	}

	if len(cmds) == 0 {
		return nil
	}

	return tea.Batch(cmds...)
}

func (m model) renderVisibleIconsCmd() tea.Cmd {
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
				iconPath, ok := m.iconPaths[student.Id]
				if !ok {
					continue
				}

				visibleRow := row - m.gridOffset
				gridX := 3 + col*thumbW
				gridY := 4 + visibleRow*thumbH

				iconW := thumbW - 2
				iconH := thumbH - 2

				_ = renderImageTermimg(iconPath, gridX, gridY, iconW, iconH)
			}
		}
		return nil
	}
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

		// Unified gridW calculation: Master Box(4) + Grid Box(2) + Space(1) + PreviewW(40) = 11
		gridW := m.width - PreviewW - 11
		if gridW < thumbW+4 {
			gridW = thumbW + 4
		}

		// x offset: MasterBorder(1) + Padding(1) + GridBox(gridW+2) + Space(1) + PreviewBorder(1) = gridW + 6
		// y offset: MarginTop(1) + MasterBorder(1) + PreviewBorder(1) = 3
		x := gridW + 6
		y := 3
		w, h := PreviewW-2, PreviewH-6 // PreviewH-2 for border, further reduced for name tag

		if err := renderImageTermimg(path, x, y, w, h); err != nil {
			return nil
		}
		return nil
	}
}

func downloadIconCmd(studentId int, downloader *Downloader) tea.Cmd {
	return func() tea.Msg {
		iconPath := getIconPath(studentId)
		if _, err := os.Stat(iconPath); err == nil {
			return iconDownloadedMsg{studentId: studentId, iconPath: iconPath}
		}

		if downloader == nil {
			return iconDownloadedMsg{studentId: studentId, err: fmt.Errorf("no downloader")}
		}

		data, err := downloader.DownloadIcon(studentId)
		if err != nil {
			return iconDownloadedMsg{studentId: studentId, err: err}
		}

		if err := os.WriteFile(iconPath, data, 0644); err != nil {
			return iconDownloadedMsg{studentId: studentId, err: err}
		}

		return iconDownloadedMsg{studentId: studentId, iconPath: iconPath}
	}
}

// --- Manifest ---

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

type iconDownloadedMsg struct {
	studentId int
	iconPath  string
	err       error
}

func fetchManifestCmd() tea.Msg {
	data, err := FetchManifest()
	if err != nil {
		// Fallback to offline meta
		students, err := LoadMeta()
		if err != nil {
			return manifestLoadedMsg{students: []Student{}}
		}
		return manifestLoadedMsg{students: students}
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

// --- Fastfetch Integration ---

const backupSuffix = ".bak"

func fastfetchConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	if home == "" {
		home = "/tmp"
	}
	return filepath.Join(home, ".config", "fastfetch", "config.jsonc")
}

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

func updateFastfetchImage(imagePath string, termLines int, termType string, logoFormat string) error {
	configPath := fastfetchConfigPath()
	bakPath := configPath + backupSuffix
	if !fileExists(bakPath) && fileExists(configPath) {
		if err := copyFile(configPath, bakPath); err != nil {
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

	imgType := logoFormat
	if imgType == "" {
		imgType = getFastfetchImageType(termType)
	}

	cmd := exec.Command("jq",
		fmt.Sprintf(`.logo.source = "%s" | .logo.type = "%s" | .logo.width = %d`, imagePath, imgType, fastW),
		configPath,
	)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("jq failed: %w", err)
	}

	return os.WriteFile(configPath, output, 0644)
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


