package main

import (
	"fmt"
	"strings"

	"charm.land/bubbletea/v2"
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

	// Move down 2px as requested
	mainContent = "\n\n" + mainContent

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
		"↑/↓/←/→ : Navigate grid",
		"Enter   : Select & set as Fastfetch logo",
		"Tab     : Switch between Browse/Installed",
		"/       : Search students",
		"i       : Open Settings",
		"h       : Show this Help",
		"q/Ctrl+C: Quit",
		"Esc     : Back / Cancel",
		"",
		lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#6C7086")).Render("Press any key to close"),
	}

	return style.Render(strings.Join(content, "\n"))
}

func (m model) viewMain() string {
	gridContent := m.renderGrid()
	previewContent := m.renderPreview()

	// Reserved: Master Border (2) + Padding (2) + Middle Space (1) = 5
	gridW := m.width - PreviewW - 5
	if gridW < thumbW {
		gridW = thumbW
	}

	leftContent := m.renderGridBoxWithTabs(gridContent, gridW)

	upperPart := lipgloss.JoinHorizontal(lipgloss.Top, leftContent, " ", previewContent)
	
	statusLine := ""
	if m.status != "" {
		if m.statusIsErr {
			statusLine = " " + styleError.Render("! "+m.status)
		} else {
			statusLine = " " + styleSuccess.Render("✓ "+m.status)
		}
	}

	if m.isDownloading {
		prog := renderProgressBar(m.downloadPct, 30)
		statusLine = " " + lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA")).Render("Downloading: ") + prog
	}

	var help string
	if m.searchMode {
		help = " " + styleHelp.Render("Type to search... | Enter: confirm | Esc: cancel")
	} else {
		help = " " + styleHelp.Render("h/j/k/l: move | /: search | Tab: switch | i: settings | Enter: select | q: quit")
	}

	// Calculate inner height to make it fill the terminal
	innerHeight := m.height - 5 // Top offset (2) + Master Border (2) + Some buffer
	if innerHeight < 10 {
		innerHeight = 10
	}

	mainCol := lipgloss.JoinVertical(lipgloss.Left, upperPart, statusLine, help)
	mainCol = lipgloss.NewStyle().Height(innerHeight).Render(mainCol)

	return styleMasterBox.Width(m.width - 2).Render(mainCol)
}

func (m model) viewSettings() string {
	var b strings.Builder
	
	title := styleTitle.Render(" SETTINGS (Interactive) ")
	b.WriteString(title + "\n\n")

	// Setting Item: Fastfetch Logo Format
	var logoLabel string
	if m.settingsIdx == 0 {
		logoLabel = styleSettingSelected.Render("> Fastfetch Logo Format: ")
	} else {
		logoLabel = styleSettingNormal.Render("  Fastfetch Logo Format: ")
	}
	b.WriteString(logoLabel)

	options := []string{"kitty", "raw"}
	for _, opt := range options {
		style := styleSettingNormal
		if m.logoFormat == opt {
			style = styleSettingSelected.Copy().Background(lipgloss.Color("#FAB387"))
		}
		b.WriteString(style.Render(" " + opt + " "))
		b.WriteString(" ")
	}
	b.WriteString("\n\n")
	
	// App Info
	b.WriteString(styleSubTitle.Render(" Application Info "))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Cache Dir:  %s\n", CacheDir()))
	b.WriteString(fmt.Sprintf("  Terminal:   %s\n", m.termType))
	if m.hasMagick {
		b.WriteString("  Magick:     Available\n")
	}
	b.WriteString("\n")

	// Navigation Help
	b.WriteString(styleSubTitle.Render(" Navigation "))
	b.WriteString("\n")
	b.WriteString(styleSettingNormal.Render(" ↑/↓: Select Setting | ←/→: Change Value "))
	b.WriteString("\n")
	b.WriteString(styleSettingNormal.Render(" i: Back to Browse "))

	content := b.String()
	settingsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#CBA6F7")).
		Padding(1, 4).
		Width(m.width - 6).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, settingsBox)
}

func renderBorderTitle(title string, panelWidth int) string {
	titleStyled := styleTitle.Render(title)
	titleLen := lipgloss.Width(titleStyled)

	remaining := panelWidth - titleLen - 2
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

func placeOverlay(x, y int, fg, bg string, shadow bool) string {
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

	targetWidth := gridW + 2

	remaining := targetWidth - tabsLen - 8
	if remaining < 0 {
		remaining = 0
	}

	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#45475A"))
	leftBorder := borderStyle.Render("╭── ")
	rightBorder := borderStyle.Render(" ─" + strings.Repeat("─", remaining) + "──╮")
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

func (m model) renderGrid() string {
	items := m.getCurrentItems()
	gridW := m.width - PreviewW - 7
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
	
	var content string
	if len(items) == 0 || m.gridIdx >= len(items) {
		content = lipgloss.NewStyle().
			Width(PreviewW-2).Height(PreviewH-2).
			Align(lipgloss.Center, lipgloss.Center).
			Foreground(lipgloss.Color("#6C7086")).
			Render("No image")
	} else {
		student := items[m.gridIdx]
		fullName := student.PersonalName
		if student.FamilyName != "" {
			fullName += " " + student.FamilyName
		}
		
		imagePlaceholder := lipgloss.NewStyle().
			Width(PreviewW - 2).
			Height(PreviewH - 6).
			Align(lipgloss.Center, lipgloss.Center).
			Render("") // Termimg will draw over this

		nameTag := lipgloss.NewStyle().
			Width(PreviewW - 2).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("#A6E3A1")).
			Bold(true).
			PaddingBottom(2).
			Render(fullName)

		content = lipgloss.JoinVertical(lipgloss.Center, imagePlaceholder, nameTag)
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#A6E3A1")).
		Width(PreviewW).
		Height(PreviewH).
		Render(content)
}
func renderProgressBar(pct float64, width int) string {
	if pct > 1.0 {
		pct = 1.0
	}
	filledW := int(float64(width) * pct)
	if filledW < 0 {
		filledW = 0
	}
	emptyW := width - filledW

	filled := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Render(strings.Repeat("█", filledW))
	empty := lipgloss.NewStyle().Foreground(lipgloss.Color("#45475A")).Render(strings.Repeat("░", emptyW))

	return fmt.Sprintf("[%s%s] %d%%", filled, empty, int(pct*100))
}
