package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) View() tea.View {
	if m.loading {
		return tea.NewView("Loading student manifest...")
	}

	var mainContent string
	if m.screen == screenSettings {
		mainContent = m.viewSettings()
	} else {
		mainContent = m.viewMain()
	}

	// Overlay modals if active
	if m.modal.kind != modalNone {
		var modalView string
		if m.modal.kind == modalHelp {
			modalView = m.renderHelpModal()
		} else {
			modalView = m.renderModal()
		}
		mainContent = m.overlayModal(mainContent, modalView)
	}

	v := tea.NewView(mainContent)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (m model) overlayModal(base, modal string) string {
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal, lipgloss.WithWhitespaceChars(base))
}

func (m model) renderModal() string {
	btnOk := styleBtnInactive.Render(" Yes (Enter) ")
	btnCancel := styleBtnInactive.Render(" No (Esc) ")

	if m.modal.activeBtn == 0 {
		btnOk = styleBtnActive.Render(" Yes (Enter) ")
	} else {
		btnCancel = styleBtnActive.Render(" No (Esc) ")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, btnOk, "  ", btnCancel)
	content := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Width(40).Align(lipgloss.Center).Render(m.modal.message),
		"",
		buttons,
	)

	return styleModalBorder.Render(content)
}

func (m model) renderHelpModal() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#89B4FA")).
		Padding(1, 2).
		Background(lipgloss.Color("#1E1E2E"))

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F38BA8")).Render("HELP & KEYBINDS")
	content := []string{
		title,
		"",
		"↑/k       : Move up",
		"↓/j       : Move down",
		"←/h       : Move left",
		"→/l       : Move right",
		"Enter     : Select & set as Fastfetch logo",
		"Tab       : Switch between Browse/Installed",
		"/         : Search students",
		"i         : Open Settings",
		"?         : Show this Help",
		"q/Ctrl+C  : Quit",
		"Esc       : Back / Cancel",
		"",
		lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#6C7086")).Render("Press any key to close"),
	}

	return style.Render(strings.Join(content, "\n"))
}

func (m model) viewMain() string {
	rows := m.visibleRows()
	// gridBoxH = tab-line border(1) + rows*gridItemH + bottom border(1)
	gridBoxH := 2 + rows*gridItemH

	gridW := m.width - PreviewW - 9
	if gridW < 16 {
		gridW = 16
	}

	gridContent := m.renderList()
	previewContent := m.renderPreview(gridBoxH)

	leftContent := m.renderGridBoxWithTabs(gridContent, gridW)

	// Body = Grid Box + Space + Preview Box
	body := lipgloss.JoinHorizontal(lipgloss.Top, leftContent, " ", previewContent)

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
		help = styleHelp.Render("Type to search  |  Enter: confirm  |  Esc: cancel")
	} else {
		help = styleHelp.Render("Tab: switch  |  /: search  |  Enter: select  |  i: settings  |  ?: help  |  q: quit")
	}

	// Legend bar: show abbreviations used in current page
	legendLine := m.renderLegendBar()

	lines := []string{body, statusLine, help}
	if legendLine != "" {
		lines = append(lines, legendLine)
	}
	mainCol := lipgloss.JoinVertical(lipgloss.Left, lines...)

	return styleMasterBox.Render(mainCol)
}

func (m model) viewSettings() string {
	cacheDir := CacheDir()
	inner := m.width - 6 // masterBox border(2) + padding(2) + settingsBox border(2)
	if inner < 20 {
		inner = 20
	}
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))

	cfgLine := lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA")).Render("Cache      ") + dim.Render(cacheDir)
	cachedItems := m.cache.List()
	statsLine := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Render("Installed  ") + dim.Render(fmt.Sprintf("%d students", len(cachedItems)))
	magickStatus := "not available"
	if m.hasMagick {
		magickStatus = "available"
	}
	magickLine := lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7")).Render("ImageMagick") + dim.Render(" "+magickStatus)
	if !m.hasMagick {
		magickLine += "  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8")).Italic(true).Render("→ install: sudo pacman -S imagemagick  or  sudo apt install imagemagick")
	}
	termLine := lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA")).Render("Terminal   ") + dim.Render(" "+m.termType)

	infoTitle := styleTitle.Render("  System Info")
	infoBox := styleSettingsBorder.Width(inner).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			infoTitle,
			"  "+cfgLine,
			"  "+statsLine,
			"  "+magickLine,
			"  "+termLine,
		),
	)

	// Interactive Settings Section
	highlight := lipgloss.NewStyle().Background(lipgloss.Color("#313244"))
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("#CDD6F4"))
	active := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Bold(true)

	// Setting 0: Logo Format
	logoLabel := label.Render("Logo Format  ")
	var kittyOption, rawOption string
	if m.logoFormat == "kitty" {
		kittyOption = active.Render("[kitty]")
		rawOption = dim.Render("  raw")
	} else {
		kittyOption = dim.Render("kitty  ")
		rawOption = active.Render("[raw]")
	}
	logoFormatRow := lipgloss.JoinHorizontal(lipgloss.Left, logoLabel, kittyOption, rawOption)
	if m.settingsIdx == 0 {
		logoFormatRow = highlight.Render(" ▶ " + logoFormatRow + " ")
	} else {
		logoFormatRow = "   " + logoFormatRow
	}

	// Setting 1: Cache Size
	cacheSizeRow := label.Render("Cache Size   ") +
		dim.Render("◀  ") + active.Render(fmt.Sprintf("%2d", m.cacheSize)) + dim.Render("  ▶") +
		dim.Render(fmt.Sprintf("  (portraits kept on disk, 1–20)"))
	if m.settingsIdx == 1 {
		cacheSizeRow = highlight.Render(" ▶ " + cacheSizeRow + " ")
	} else {
		cacheSizeRow = "   " + cacheSizeRow
	}

	// Setting 2: Auto Backup
	var backupOn, backupOff string
	if m.autoBackup {
		backupOn = active.Render("[on] ")
		backupOff = dim.Render(" off")
	} else {
		backupOn = dim.Render(" on  ")
		backupOff = active.Render("[off]")
	}
	autoBackupRow := label.Render("Auto Backup  ") + backupOn + backupOff +
		dim.Render("  (backup fastfetch config before overwriting)")
	if m.settingsIdx == 2 {
		autoBackupRow = highlight.Render(" ▶ " + autoBackupRow + " ")
	} else {
		autoBackupRow = "   " + autoBackupRow
	}

	settingsTitle := styleTitle.Render("  Settings")
	settingsBox := styleSettingsBorder.Width(inner).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			settingsTitle,
			" "+logoFormatRow,
			" "+cacheSizeRow,
			" "+autoBackupRow,
			"",
			dim.Render("  ←/→: change  |  ↑/↓: navigate"),
		),
	)

	statusLine := ""
	if m.status != "" {
		if m.statusIsErr {
			statusLine = styleError.Render("! " + m.status)
		} else {
			statusLine = styleSuccess.Render("✓ " + m.status)
		}
	}

	help := styleHelp.Render("esc: back  |  q: quit")
	mainCol := lipgloss.JoinVertical(lipgloss.Left, infoBox, "", settingsBox, "", statusLine, help)
	return styleMasterBox.Width(m.width - 4).Render(mainCol)
}

func renderBorderTitle(title string, panelWidth int) string {
	titleStyled := styleTitle.Render(title)
	titleLen := lipgloss.Width(titleStyled)

	remaining := panelWidth - titleLen
	if remaining < 0 {
		remaining = 0
	}
	leftPad := remaining / 2

	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA"))
	leftBorder := borderStyle.Render(strings.Repeat("─", leftPad))
	rightBorder := borderStyle.Render(strings.Repeat("─", remaining-leftPad))

	return borderStyle.Render("┌") + leftBorder + titleStyled + rightBorder + borderStyle.Render("┐")
}

func (m model) renderModalOverlay(bg string) string {
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

	overlay := lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modalBox,
	)

	return placeOverlay(0, 0, overlay, bg, false)
}

func placeOverlay(x, y int, fg, bg string, _ bool) string {
	fgLines := strings.Split(fg, "\n")
	bgLines := strings.Split(bg, "\n")

	fgHeight := len(fgLines)
	fgWidth := 0
	for _, l := range fgLines {
		if w := lipgloss.Width(l); w > fgWidth {
			fgWidth = w
		}
	}

	if fgHeight >= len(bgLines) && fgWidth >= lipgloss.Width(bg) {
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

func (m model) renderGridBoxWithTabs(content string, gridW int) string {
	var tabsText string
	if m.searchMode {
		searchQuery := m.searchQuery + "█"
		tabsText = lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA")).Render(" Search: "),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#CDD6F4")).Render(searchQuery+" "),
		)
	} else {
		tabsText = m.renderTabs()
	}
	tabsLen := lipgloss.Width(tabsText)

	targetWidth := gridW + 4

	remaining := targetWidth - tabsLen - 7
	if remaining < 0 {
		remaining = 0
	}

	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#45475A"))
	leftBorder := borderStyle.Render("╭── ")
	rightBorder := borderStyle.Render(" ─" + strings.Repeat("─", remaining) + "╮")
	topLine := leftBorder + tabsText + rightBorder

	bottomLine := borderStyle.Render("╰" + strings.Repeat("─", targetWidth-2) + "╯")

	var bodyLines []string
	if content != "" {
		lines := strings.Split(content, "\n")
		leftEdge := borderStyle.Render("│ ")
		rightEdge := borderStyle.Render(" │")
		for _, line := range lines {
			contentPadded := lipgloss.PlaceHorizontal(targetWidth-4, lipgloss.Left, line)
			bodyLines = append(bodyLines, leftEdge+contentPadded+rightEdge)
		}
	}

	res := []string{topLine}
	res = append(res, bodyLines...)
	res = append(res, bottomLine)

	return strings.Join(res, "\n")
}

func (m model) renderTabs() string {
	installedCount := len(m.getInstalledItems())
	var browse, installed string
	if m.tab == TabBrowse {
		browse = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A6E3A1")).Render(" Browse ")
		installed = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(fmt.Sprintf(" Installed (%d) ", installedCount))
	} else {
		browse = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(" Browse ")
		installed = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A6E3A1")).Render(fmt.Sprintf(" Installed (%d) ", installedCount))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, browse, installed)
}

func (m model) renderList() string {
	items := m.getCurrentItems()

	listW := m.width - PreviewW - 9
	if listW < 16 {
		listW = 16
	}

	if len(items) == 0 {
		return lipgloss.NewStyle().
			Width(listW).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("#6C7086")).
			Render("No students")
	}

	visibleRows := m.visibleRows()
	startIdx := m.gridOffset * m.gridCols
	if startIdx >= len(items) {
		startIdx = 0
		m.gridOffset = 0
	}

	var rows []string
	for row := 0; row < visibleRows; row++ {
		var cols []string
		for col := 0; col < m.gridCols; col++ {
			idx := startIdx + row*m.gridCols + col
			if idx >= len(items) {
				// Empty cell to fill the row
				colWidth := listW / m.gridCols
				if colWidth < 10 {
					colWidth = 10
				}
				innerW := colWidth - 2
				if innerW < 1 {
					innerW = 1
				}
				cols = append(cols, lipgloss.NewStyle().Width(innerW).Render(""))
				continue
			}
			cols = append(cols, m.renderListItem(items[idx], idx == m.gridIdx))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cols...))
	}

	if len(rows) == 0 {
		return ""
	}

	return strings.Join(rows, "\n")
}

func (m model) renderListItem(s Student, selected bool) string {
	var borderColor, nameColor string
	if selected {
		borderColor = "#A6E3A1"
		nameColor = "#A6E3A1"
	} else {
		borderColor = "#45475A"
		nameColor = "#CDD6F4"
	}

	listW := m.width - PreviewW - 9
	if listW < 16 {
		listW = 16
	}
	colWidth := listW / m.gridCols
	if colWidth < 4 {
		colWidth = 4
	}
	innerW := colWidth - 2
	if innerW < 1 {
		innerW = 1
	}

	name, _ := displayName(s, innerW)

	nameRendered := lipgloss.NewStyle().
		Width(innerW).
		MaxWidth(innerW).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color(nameColor)).
		Render(name)

	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor))

	return border.Render(nameRendered)
}

func (m model) renderPreview(containerH int) string {
	// containerH is the total height including the border (2 lines)
	innerH := containerH - 2
	if innerH < 4 {
		innerH = 4
	}
	// reserve 2 rows (spacer + name tag) so the name never overlaps the image
	imageH := innerH - 2
	if imageH < 2 {
		imageH = 2
	}

	items := m.getCurrentItems()

	var content string
	if len(items) == 0 || m.gridIdx >= len(items) {
		content = lipgloss.NewStyle().
			Width(PreviewW-2).Height(innerH).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("#6C7086")).
			Render("No image")
	} else {
		student := items[m.gridIdx]
		fullName := student.FamilyName
		if student.PersonalName != "" {
			if fullName != "" {
				fullName += " "
			}
			fullName += student.PersonalName
		}
		if student.Variant != "" {
			fullName += " (" + student.Variant + ")"
		}
		if fullName == "" {
			fullName = student.Name
		}

		imagePlaceholder := lipgloss.NewStyle().
			Width(PreviewW - 2).
			Height(imageH).
			Align(lipgloss.Center, lipgloss.Center).
			Render("")

		nameTag := lipgloss.NewStyle().
			Width(PreviewW - 2).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("#A6E3A1")).
			Bold(true).
			Render(fullName)

		spacer := lipgloss.NewStyle().Width(PreviewW - 2).Height(1).Render("")
		content = lipgloss.JoinVertical(lipgloss.Center, imagePlaceholder, spacer, nameTag)
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#A6E3A1")).
		Render(content)
}
// renderLegendBar builds the abbr legend for variants visible on the current page.
func (m model) renderLegendBar() string {
	items := m.getCurrentItems()
	if len(items) == 0 {
		return ""
	}

	listW := m.width - PreviewW - 9
	if listW < 16 {
		listW = 16
	}
	colWidth := listW / m.gridCols
	if colWidth < 4 {
		colWidth = 4
	}
	innerW := colWidth - 2
	if innerW < 1 {
		innerW = 1
	}

	visibleRows := m.visibleRows()
	startIdx := m.gridOffset * m.gridCols
	endIdx := startIdx + visibleRows*m.gridCols
	if endIdx > len(items) {
		endIdx = len(items)
	}

	// Collect unique abbr→variant pairs where abbr was used
	seen := map[string]string{}
	for i := startIdx; i < endIdx; i++ {
		s := items[i]
		_, used := displayName(s, innerW)
		if used && s.Abbr != "" && s.Variant != "" {
			seen[s.Abbr] = s.Variant
		}
	}
	if len(seen) == 0 {
		return ""
	}

	parts := make([]string, 0, len(seen))
	for abbr, variant := range seen {
		parts = append(parts, abbr+"="+variant)
	}
	// stable sort for consistent output
	for i := 0; i < len(parts)-1; i++ {
		for j := i + 1; j < len(parts); j++ {
			if parts[i] > parts[j] {
				parts[i], parts[j] = parts[j], parts[i]
			}
		}
	}

	legend := strings.Join(parts, "  ")
	return styleHelp.Render("abbr: " + legend)
}

