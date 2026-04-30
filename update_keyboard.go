package main

import tea "charm.land/bubbletea/v2"

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
			studentId := m.modal.studentId
			m.modal = modal{kind: modalNone}
			m.isDownloading = true
			return m, m.cacheAndSelectWithBackup(studentId)
		}
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
		return m, func() tea.Msg {
			clearImagesTermimg()
			return nil
		}

	case "tab":
		if m.tab == TabBrowse {
			m.tab = TabInstalled
		} else {
			m.tab = TabBrowse
		}
		m.gridIdx = 0
		m.gridOffset = 0
		return m, nil

	case "?":
		m.modal = modal{kind: modalHelp}
		return m, func() tea.Msg {
			clearImagesTermimg()
			return nil
		}

	case "up", "k":
		if m.gridIdx >= m.gridCols {
			m.gridIdx -= m.gridCols
			m.clampOffset()
			return m, m.scheduleRenderCmd()
		}

	case "down", "j":
		items := m.getCurrentItems()
		if m.gridIdx+m.gridCols < len(items) {
			m.gridIdx += m.gridCols
			m.clampOffset()
			return m, m.scheduleRenderCmd()
		}

	case "left", "h":
		if m.gridIdx > 0 {
			m.gridIdx--
			m.clampOffset()
			return m, m.scheduleRenderCmd()
		}

	case "right", "l":
		items := m.getCurrentItems()
		if m.gridIdx < len(items)-1 {
			m.gridIdx++
			m.clampOffset()
			return m, m.scheduleRenderCmd()
		}

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
			m.isDownloading = true
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
	case "ctrl+c", "q":
		return m, tea.Quit
	case "b", "esc", "i":
		m.screen = screenMain
		m.status = ""
		return m, m.scheduleRenderCmd()
	case "up", "k":
		if m.settingsIdx > 0 {
			m.settingsIdx--
		}
	case "down", "j":
		if m.settingsIdx < 2 {
			m.settingsIdx++
		}
	case "left", "h":
		switch m.settingsIdx {
		case 0:
			if m.logoFormat == "kitty" {
				m.logoFormat = "raw"
			} else {
				m.logoFormat = "kitty"
			}
		case 1:
			if m.cacheSize > 1 {
				m.cacheSize--
				m.cache.SetMaxSize(m.cacheSize)
			}
		case 2:
			m.autoBackup = !m.autoBackup
		}
	case "right", "l":
		switch m.settingsIdx {
		case 0:
			if m.logoFormat == "kitty" {
				m.logoFormat = "raw"
			} else {
				m.logoFormat = "kitty"
			}
		case 1:
			if m.cacheSize < 20 {
				m.cacheSize++
				m.cache.SetMaxSize(m.cacheSize)
			}
		case 2:
			m.autoBackup = !m.autoBackup
		}
	}
	return m, nil
}
