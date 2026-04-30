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

	tea "charm.land/bubbletea/v2"
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
	width         int
	height        int
	screen        screen
	tab           Tab
	manifest      []Student
	manifestByID  map[int]Student
	filtered      []Student
	gridIdx       int
	gridCols      int
	gridOffset    int
	searchQuery   string
	searchMode    bool
	cache         *Cache
	downloader    *Downloader
	loading       bool
	isDownloading bool
	status        string
	statusIsErr   bool
	statusTimer   time.Time
	previewPath   string
	modal         modal
	hasMagick     bool
	termType      string
	logoFormat    string
	settingsIdx   int
	cacheSize     int
	autoBackup    bool
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
		termType:    detectTerminalType(),
		logoFormat:  "kitty",
		cacheSize:   DefaultCacheSize,
		autoBackup:  true,
	}
}

func checkMagick() bool {
	_, err := exec.LookPath("magick")
	return err == nil
}

func (m model) Init() tea.Cmd {
	return tea.Batch(fetchManifestCmd, checkMagickCmd, initTermimgCmd)
}

func initTermimgCmd() tea.Msg {
	if err := InitTerminal(); err != nil {
		return termimgInitErrorMsg{err: err}
	}
	return termimgInitSuccessMsg{}
}

type termimgInitSuccessMsg struct{}
type termimgInitErrorMsg struct{ err error }

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

	case termimgInitErrorMsg:
		m.status = fmt.Sprintf("termimg init: %v", msg.err)
		m.statusIsErr = true
		return m, nil

	case manifestLoadedMsg:
		m.manifest = msg.students
		m.filtered = msg.students
		m.loading = false
		// Build ID index for O(1) lookups
		m.manifestByID = make(map[int]Student, len(msg.students))
		for _, s := range msg.students {
			m.manifestByID[s.Id] = s
		}
		if len(m.filtered) > 0 {
			m.previewPath = getPortraitPath(m.filtered[0].Id)
			if len(m.manifest) > 0 {
				SaveMeta(m.manifest)
			}
		}
		go m.preCachePortraits()
		return m, m.scheduleRenderCmd()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		listW := m.width - PreviewW - 9
		if listW < 16 {
			listW = 16
		}
		colWidth := listW / m.gridCols
		if colWidth < 10 {
			m.gridCols = listW / 10
			if m.gridCols < 1 {
				m.gridCols = 1
			}
		}
		m.clampOffset()
		return m, m.scheduleRenderCmd()

	case tea.KeyPressMsg:
		// Block repeat for action keys; allow repeat for navigation
		if msg.Key().IsRepeat && !isNavKey(msg.String()) {
			return m, nil
		}

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

	case renderImagesMsg:
		return m, func() tea.Msg {
			m.renderImages()
			return nil
		}
	}
	return m, nil
}

// isNavKey returns true for keys that should be repeatable (held down).
func isNavKey(key string) bool {
	switch key {
	case "up", "down", "left", "right", "k", "j", "h", "l",
		"pgup", "pgdown", "home", "end":
		return true
	}
	return false
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
	// marginTop(1) + masterBox borders(2) + gridBox borders(2) + helpLine(1) + statusLine(1) + legendLine(1)
	overhead := 8
	if m.height == 0 {
		return 4
	}
	available := m.height - overhead
	rows := available / gridItemH
	if rows < 1 {
		rows = 1
	}
	return rows
}

func (m model) getInstalledItems() []Student {
	cached := m.cache.List()
	result := make([]Student, 0, len(cached))
	for _, c := range cached {
		if s, ok := m.manifestByID[c.StudentId]; ok {
			result = append(result, s)
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
	items := m.getCurrentItems()
	count := len(items)
	if count == 0 {
		return
	}

	if m.gridIdx < count {
		studentId := items[m.gridIdx].Id
		data, err := m.downloader.DownloadPortrait(studentId)
		if err == nil {
			m.cache.Put(studentId, data)
		}
	}
}

// --- Cache & Select ---

func (m model) cacheAndSelect(studentId int) tea.Cmd {
	return func() tea.Msg {
		if m.autoBackup {
			if fileExists(fastfetchConfigPath()) {
				bakPath := fastfetchConfigPath() + backupSuffix
				if err := copyFile(fastfetchConfigPath(), bakPath); err != nil {
					return cacheErrorMsg{err: fmt.Errorf("backup failed: %w", err)}
				}
			}
			return m.doCacheAndSelect(studentId)()
		}
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

func (m model) doCacheAndSelect(studentId int) tea.Cmd {
	student, _ := m.getStudent(studentId)
	studentName := student.Name
	if studentName == "" {
		studentName = student.PersonalName
	}
	return func() tea.Msg {
		portraitData, err := m.downloader.DownloadPortrait(studentId)
		if err != nil {
			return cacheErrorMsg{err: err}
		}
		if err := m.cache.Put(studentId, portraitData); err != nil {
			return cacheErrorMsg{err: fmt.Errorf("cache write failed: %w", err)}
		}

		cached, ok := m.cache.Get(studentId)
		if !ok {
			return cacheErrorMsg{err: fmt.Errorf("cache miss after put")}
		}

		if err := updateFastfetchImage(cached.PortraitPath, m.height, m.termType, m.logoFormat); err != nil {
			return cacheErrorMsg{err: err}
		}
		return cacheSuccessMsg{studentId: studentId, name: studentName}
	}
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
	if s, ok := m.manifestByID[id]; ok {
		return s, true
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

func (m model) renderImages() {
	items := m.getCurrentItems()
	if len(items) == 0 || terminal == nil {
		return
	}

	if m.gridIdx < 0 || m.gridIdx >= len(items) {
		return
	}

	student := items[m.gridIdx]
	path := ensurePortraitCached(student.Id, m.downloader)
	if path == "" {
		return
	}

	listW := m.width - PreviewW - 9
	if listW < 16 {
		listW = 16
	}

	// Compute image height to match renderPreview's imageH (innerH - 1)
	rows := m.visibleRows()
	gridBoxH := 2 + rows*gridItemH
	innerH := gridBoxH - 2
	imageH := innerH - 1
	if imageH < 2 {
		imageH = 2
	}

	x := listW + 8
	y := 3
	w, h := PreviewW-2, imageH

	_ = drawPreviewImage(path, x, y, w, h)
}

func (m model) scheduleRenderCmd() tea.Cmd {
	return tea.Tick(16*time.Millisecond, func(time.Time) tea.Msg {
		return renderImagesMsg{}
	})
}

type renderImagesMsg struct{}

// --- Manifest ---

func filterStudents(students []Student, query string) []Student {
	if query == "" {
		return students
	}
	result := []Student{}
	lowerQuery := strings.ToLower(query)
	for _, s := range students {
		if strings.Contains(strings.ToLower(s.Name), lowerQuery) ||
			strings.Contains(strings.ToLower(s.FamilyName), lowerQuery) ||
			strings.Contains(strings.ToLower(s.PersonalName), lowerQuery) ||
			strings.Contains(strings.ToLower(s.Base), lowerQuery) ||
			strings.Contains(strings.ToLower(s.Variant), lowerQuery) ||
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
		// Fallback to offline meta
		students, err := LoadMeta()
		if err != nil {
			return manifestLoadedMsg{students: []Student{}}
		}
		return manifestLoadedMsg{students: students}
	}
	students, err := parseManifest(data)
	if err != nil {
		debugLog.Printf("parseManifest error: %v", err)
		return manifestLoadedMsg{students: []Student{}}
	}
	return manifestLoadedMsg{students: students}
}

func parseManifest(data []byte) ([]Student, error) {
	var students []Student
	if err := json.Unmarshal(data, &students); err != nil {
		return nil, err
	}
	for i := range students {
		computeVariant(&students[i])
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
		"--arg", "source", imagePath,
		"--arg", "itype", imgType,
		"--argjson", "width", strconv.Itoa(fastW),
		`.logo.source = $source | .logo.type = $itype | .logo.width = $width | .logo.padding.left = 2`,
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
