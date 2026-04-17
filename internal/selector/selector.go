package selector

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	minItemsPerPage = 7
	maxItemsPerPage = 9
	keyEsc          = '\x1b'
	keyArrowUp      = -1001
	keyArrowDown    = -1002
	keyArrowRight   = -1003
	keyArrowLeft    = -1004
)

var ErrExitRequested = errors.New("exit requested")

var preselectedItems = []string{"mods", "Новый мод", "Новые моды", "Новый моды"}
var topLevelItemGroups = [][]string{
	{"mods"},
	{"shaderpacks"},
	{"Новый мод", "Новые моды", "Новый моды"},
}

type item struct {
	path     string
	name     string
	isDir    bool
	parent   int
	children []int
	expanded bool
	checked  bool
}

type pageAction struct {
	label string
	item  int
	kind  actionKind
}

type actionKind int

const (
	actionToggle actionKind = iota
	actionPrevPage
	actionNextPage
	actionConfirm
)

func SelectFilesForMods(root string) ([]string, error) {
	items, roots, err := buildTree(root)
	if err != nil {
		return nil, err
	}

	terminal, err := enableRawMode()
	if err != nil {
		return nil, err
	}
	defer terminal.restore()

	page := 0
	cursor := 0
	confirmExit := false

	for {
		clearScreen()

		visible := visibleItems(items, roots)
		pages := buildPages(len(visible))
		if page >= len(pages) {
			page = len(pages) - 1
		}
		if page < 0 {
			page = 0
		}
		current := pages[page]
		pageItemsCount := current[1] - current[0]
		if cursor >= pageItemsCount {
			cursor = pageItemsCount - 1
		}
		if cursor < 0 {
			cursor = 0
		}

		actions := printSelection(items, visible, pages, page, cursor, confirmExit)

		fmt.Print("Ваш выбор: ")
		key, err := readKey()
		fmt.Println()
		if err != nil {
			return nil, err
		}

		if key != keyEsc {
			confirmExit = false
		}

		switch key {
		case keyArrowUp:
			if cursor > 0 {
				cursor--
			} else if page > 0 {
				page--
				cursor = maxItemsPerPage
			}
			continue
		case keyArrowDown:
			if cursor < pageItemsCount-1 {
				cursor++
			} else if page < len(pages)-1 {
				page++
				cursor = 0
			}
			continue
		case keyArrowRight:
			if pageItemsCount > 0 {
				index := visible[current[0]+cursor]
				if canExpand(items[index]) {
					items[index].expanded = true
					page = 0
					cursor = 0
				}
			}
			continue
		case keyArrowLeft:
			if collapseExpandedRoot(items, roots) {
				page = 0
				cursor = 0
			}
			continue
		case keyEsc:
			if collapseExpandedRoot(items, roots) {
				page = 0
				cursor = 0
				continue
			}
			if confirmExit {
				return nil, ErrExitRequested
			}
			confirmExit = true
			continue
		case '\r', '\n':
			return selectedFiles(items), nil
		case ' ':
			if pageItemsCount > 0 {
				index := visible[current[0]+cursor]
				setCheckedRecursive(items, index, state(items, index) != 1)
			}
			continue
		}

		action, ok := actions[key]
		if !ok {
			fmt.Println("Не понял выбор. Нажмите 1-9, Shift+1-9 для раскрытия верхней папки или 0 для продолжения.")
			continue
		}

		switch action.kind {
		case actionConfirm:
			return selectedFiles(items), nil
		case actionPrevPage:
			page--
			cursor = 0
		case actionNextPage:
			page++
			cursor = 0
		case actionToggle:
			if isShiftKey(key) {
				if !canExpand(items[action.item]) {
					fmt.Println("Раскрывать можно только папки верхнего уровня.")
					continue
				}
				items[action.item].expanded = !items[action.item].expanded
				page = 0
				cursor = 0
				continue
			}

			setCheckedRecursive(items, action.item, state(items, action.item) != 1)
		}
	}
}

func buildTree(root string) ([]item, []int, error) {
	items := make([]item, 0)
	roots := make([]int, 0)
	preselected := preselectedSet()

	var addItems func(string, int, bool) ([]int, error)
	addItems = func(dir string, parent int, checked bool) ([]int, error) {
		entries, err := sortedDirEntries(dir)
		if err != nil {
			return nil, err
		}

		indexes := make([]int, 0, len(entries))
		for _, entry := range entries {
			name := entry.Name()
			path := filepath.Join(dir, name)
			index := len(items)
			item := item{
				path:    path,
				name:    name,
				isDir:   entry.IsDir(),
				parent:  parent,
				checked: checked,
			}

			items = append(items, item)
			indexes = append(indexes, index)

			if entry.IsDir() {
				children, err := addItems(path, index, checked)
				if err != nil {
					return nil, err
				}
				items[index].children = children
			}
		}

		return indexes, nil
	}

	entries, err := sortedTopLevelEntries(root)
	if err != nil {
		return nil, nil, err
	}

	for _, entry := range entries {
		name := entry.Name()
		path := filepath.Join(root, name)
		checked := preselected[name]
		index := len(items)
		item := item{
			path:     path,
			name:     name,
			isDir:    entry.IsDir(),
			parent:   -1,
			checked:  checked,
			expanded: false,
		}

		items = append(items, item)
		roots = append(roots, index)

		if entry.IsDir() {
			children, err := addItems(path, index, checked)
			if err != nil {
				return nil, nil, err
			}
			items[index].children = children
		}
	}

	return items, roots, nil
}

func preselectedSet() map[string]bool {
	result := make(map[string]bool, len(preselectedItems))
	for _, name := range preselectedItems {
		result[name] = true
	}
	return result
}

func sortedDirEntries(dir string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})

	return entries, nil
}

func sortedTopLevelEntries(root string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	byName := make(map[string]os.DirEntry, len(entries))
	for _, entry := range entries {
		byName[entry.Name()] = entry
	}

	result := make([]os.DirEntry, 0, len(topLevelItemGroups))
	for _, names := range topLevelItemGroups {
		for _, name := range names {
			entry, ok := byName[name]
			if ok {
				result = append(result, entry)
				break
			}
		}
	}

	return result, nil
}

func state(items []item, index int) int {
	item := items[index]
	if !item.isDir || len(item.children) == 0 {
		if item.checked {
			return 1
		}
		return 0
	}

	checkedCount := 0
	partialCount := 0
	for _, child := range item.children {
		switch state(items, child) {
		case 1:
			checkedCount++
		case 2:
			partialCount++
		}
	}

	if checkedCount == len(item.children) {
		return 1
	}
	if checkedCount > 0 || partialCount > 0 {
		return 2
	}
	return 0
}

func setCheckedRecursive(items []item, index int, checked bool) {
	items[index].checked = checked
	for _, child := range items[index].children {
		setCheckedRecursive(items, child, checked)
	}
}

func visibleItems(items []item, roots []int) []int {
	result := make([]int, 0, len(items))

	for _, root := range roots {
		result = append(result, root)
		if !items[root].isDir || !items[root].expanded {
			continue
		}
		for _, child := range items[root].children {
			result = append(result, child)
		}
	}

	return result
}

func collapseExpandedRoot(items []item, roots []int) bool {
	for _, root := range roots {
		if items[root].expanded {
			items[root].expanded = false
			return true
		}
	}
	return false
}

func buildPages(total int) [][2]int {
	if total <= 0 {
		return [][2]int{{0, 0}}
	}

	pages := make([][2]int, 0)
	for start := 0; start < total; {
		hasPrev := len(pages) > 0
		remaining := total - start

		lastPageSize := maxItemsPerPage
		if hasPrev {
			lastPageSize--
		}

		size := remaining
		if remaining > lastPageSize {
			size = lastPageSize - 1
		}
		if size < minItemsPerPage && remaining > minItemsPerPage {
			size = minItemsPerPage
		}
		if size > remaining {
			size = remaining
		}

		pages = append(pages, [2]int{start, start + size})
		start += size
	}

	return pages
}

func canExpand(item item) bool {
	return item.isDir && item.parent == -1
}

func depth(items []item, index int) int {
	if items[index].parent == -1 {
		return 0
	}
	return 1
}

func checkbox(state int) string {
	switch state {
	case 1:
		return "[x]"
	case 2:
		return "[-]"
	default:
		return "[ ]"
	}
}

func printSelection(items []item, visible []int, pages [][2]int, page, cursor int, confirmExit bool) map[rune]pageAction {
	actions := map[rune]pageAction{
		'0': {label: "Продолжить", kind: actionConfirm},
	}
	current := pages[page]
	pageItems := visible[current[0]:current[1]]
	hasPrev := page > 0
	hasNext := page < len(pages)-1
	itemKeys := availableItemKeys(hasPrev, hasNext)

	fmt.Println()
	fmt.Println("Выберите файлы и папки, которые попадут в новую папку mods.")
	fmt.Println("1-9/Space - отметить/снять, Shift+1-9 - раскрыть/свернуть верхнюю папку, стрелки - навигация, Enter/0 - продолжить.")
	if len(pages) > 1 {
		fmt.Printf("Страница %d из %d.\n", page+1, len(pages))
	}
	if confirmExit {
		fmt.Println("Нажмите Esc еще раз, чтобы закрыть приложение.")
	}
	fmt.Println()

	for number, index := range pageItems {
		key := itemKeys[number]
		item := items[index]
		indent := strings.Repeat("   ", depth(items, index))
		pointer := " "
		if number == cursor {
			pointer = ">"
		}
		marker := " "
		if canExpand(item) {
			if item.expanded {
				marker = "v"
			} else {
				marker = ">"
			}
			for _, shiftKey := range shiftedKeys(key) {
				actions[shiftKey] = pageAction{item: index, kind: actionToggle}
			}
		}

		note := ""
		if item.parent == -1 && item.isDir {
			note = " (все выбранные файлы из папки)"
		}

		fmt.Printf("%s %c. %s%s %s %s%s\n", pointer, key, indent, checkbox(state(items, index)), marker, item.name, note)
		actions[key] = pageAction{item: index, kind: actionToggle}
	}

	if hasPrev {
		fmt.Println("  8. Назад")
		actions['8'] = pageAction{label: "Назад", kind: actionPrevPage}
	}
	if hasNext {
		fmt.Println("  9. Далее")
		actions['9'] = pageAction{label: "Далее", kind: actionNextPage}
	}
	fmt.Println("  Enter/0. Продолжить")
	fmt.Println()

	return actions
}

func availableItemKeys(hasPrev, hasNext bool) []rune {
	keys := make([]rune, 0, maxItemsPerPage)
	for _, key := range []rune{'1', '2', '3', '4', '5', '6', '7', '8', '9'} {
		if hasPrev && key == '8' {
			continue
		}
		if hasNext && key == '9' {
			continue
		}
		keys = append(keys, key)
	}
	return keys
}

func shiftedKeys(key rune) []rune {
	switch key {
	case '1':
		return []rune{'!'}
	case '2':
		return []rune{'@', '"'}
	case '3':
		return []rune{'#', '№'}
	case '4':
		return []rune{'$', ';'}
	case '5':
		return []rune{'%'}
	case '6':
		return []rune{'^', ':'}
	case '7':
		return []rune{'&', '?'}
	case '8':
		return []rune{'*'}
	case '9':
		return []rune{'('}
	default:
		return nil
	}
}

func isShiftKey(key rune) bool {
	return strings.ContainsRune("!@#$%^&*(\"№;:?", key)
}

func selectedFiles(items []item) []string {
	files := make([]string, 0)
	for index, item := range items {
		if !item.isDir && state(items, index) == 1 {
			files = append(files, item.path)
		}
	}
	return files
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
