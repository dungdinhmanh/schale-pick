package main

import "charm.land/bubbletea/v2"

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
			return m, tea.Batch(m.renderKittyImage(), m.getVisibleIconBatch(), m.renderVisibleIconsCmd())
		}

	case "right", "l":
		if m.gridIdx < len(m.getCurrentItems())-1 {
			m.gridIdx++
			m.clampOffset()
			return m, tea.Batch(m.renderKittyImage(), m.getVisibleIconBatch(), m.renderVisibleIconsCmd())
		}

	case "up", "k":
		m.gridIdx -= m.gridCols
		if m.gridIdx < 0 {
			m.gridIdx = 0
		}
		m.clampOffset()
		return m, tea.Batch(m.renderKittyImage(), m.getVisibleIconBatch(), m.renderVisibleIconsCmd())

	case "down", "j":
		m.gridIdx += m.gridCols
		maxIdx := len(m.getCurrentItems()) - 1
		if m.gridIdx > maxIdx {
			m.gridIdx = maxIdx
		}
		m.clampOffset()
		return m, tea.Batch(m.renderKittyImage(), m.getVisibleIconBatch(), m.renderVisibleIconsCmd())

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
	case "ctrl+c", "q":
		return m, tea.Quit
	case "b", "esc":
		m.screen = screenMain
		m.status = ""
		return m, nil
	}
	return m, nil
}
