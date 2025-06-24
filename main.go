package main

import (
	"strings"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sergi/go-diff/diffmatchpatch"
	"fmt"
)

func countLines(s string) int {
	if s == "" {
		return 0
	}
	return len(strings.Split(s, "\n"))
}

func colorDiff(orig, updated string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(orig, updated, false)
	var b strings.Builder
	for _, d := range diffs {
		switch d.Type {
		case diffmatchpatch.DiffInsert:
			b.WriteString(`[green]` + tview.Escape(d.Text) + `[-]`)
		case diffmatchpatch.DiffDelete:
			b.WriteString(`[red]` + tview.Escape(d.Text) + `[-]`)
		default:
			b.WriteString(tview.Escape(d.Text))
		}
	}
	return b.String()
}

func main() {
	app := tview.NewApplication().EnablePaste(true)

	leftArea := tview.NewTextArea()
	leftArea.SetBorder(true)
	leftArea.SetTitle(" Original (Lines: 0) ")

	rightArea := tview.NewTextArea()
	rightArea.SetBorder(true)
	rightArea.SetTitle(" Updated (Lines: 0) ")

	// Short, colored legend in the bottom panel title
	bottom := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() { app.Draw() })
	bottom.SetBorder(true).SetTitle(` [red]Removed Text[-] |  [green]Added Text[-]  | Delete to Clear `)

	updatePanels := func() {
		origContent := leftArea.GetText()
		updatedContent := rightArea.GetText()
		leftArea.SetTitle(fmt.Sprintf(" Original (Lines: %d) ", countLines(origContent)))
		rightArea.SetTitle(fmt.Sprintf(" Updated (Lines: %d) ", countLines(updatedContent)))
		bottom.SetText(colorDiff(origContent, updatedContent))
	}

	leftArea.SetChangedFunc(updatePanels)
	rightArea.SetChangedFunc(updatePanels)

	flexTop := tview.NewFlex().
		AddItem(leftArea, 0, 1, true).
		AddItem(rightArea, 0, 1, false)
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(flexTop, 0, 3, true).
		AddItem(bottom, 0, 2, false)

	currentPane := 0 // 0 = left, 1 = right

	leftArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if currentPane != 0 {
			return nil // Ignore input if not focused
		}
		switch event.Key() {
		case tcell.KeyTAB:
			currentPane = 1
			app.SetFocus(rightArea)
			return nil
		case tcell.KeyCtrlC, tcell.KeyEscape:
			app.Stop()
			return nil
		case tcell.KeyDelete:
			leftArea.SetText("", true)
			rightArea.SetText("", true)
			updatePanels()
			return nil
		}
		return event
	})
	rightArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if currentPane != 1 {
			return nil // Ignore input if not focused
		}
		switch event.Key() {
		case tcell.KeyTAB:
			currentPane = 0
			app.SetFocus(leftArea)
			return nil
		case tcell.KeyCtrlC, tcell.KeyEscape:
			app.Stop()
			return nil
		case tcell.KeyDelete:
			leftArea.SetText("", true)
			rightArea.SetText("", true)
			updatePanels()
			return nil
		}
		return event
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:
			if currentPane == 0 {
				currentPane = 1
				app.SetFocus(rightArea)
			} else {
				currentPane = 0
				app.SetFocus(leftArea)
			}
			return nil
		case tcell.KeyCtrlC, tcell.KeyEscape:
			app.Stop()
			return nil
		case tcell.KeyDelete:
			leftArea.SetText("", true)
			rightArea.SetText("", true)
			updatePanels()
			return nil
		}
		return event
	})

	app.SetFocus(leftArea)
	updatePanels()
	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
