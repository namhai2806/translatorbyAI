package main

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// CỘT 1: INPUT
func createInputColumn(models []string) (*fyne.Container, func()) {

	tabs := createInputTabs(models)
	customLay := layout.CustomPaddedLayout{
		TopPadding:   20,
		LeftPadding:  30,
		RightPadding: 30,
	}
	inputComponents := container.New(customLay, tabs)
	refresh := func() {
		inputComponents.Objects = []fyne.CanvasObject{createInputTabs(models)}
		inputComponents.Refresh()
	}

	return inputComponents, refresh
}

func createInputTabs(models []string) *container.AppTabs {

	tabs := container.NewAppTabs()
	var tokenData tokenUsageData
	tokenUsagePath, err := defaultTokenUsagePath()
	if err != nil {
		fmt.Println("Error finding token file:", err)
	} else {
		tokenFile, err := os.ReadFile(tokenUsagePath)
		if err != nil {
			fmt.Println("Error reading token file:", err)
		} else if tokenData, err = decodeTokenUsageData(tokenFile); err != nil {
			fmt.Println("Error parsing token file:", err)
		}
	}

	for _, model := range models {
		tableCost := CostTable(tokenData, model)
		tabs.Append(container.NewTabItemWithIcon(
			model,
			theme.DocumentIcon(),
			container.NewStack(tableCost),
		),
		)
	}

	priceModelColumn := container.NewVBox(widget.NewLabelWithStyle("Mô hình", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	priceInputColumn := container.NewVBox(widget.NewLabelWithStyle("Input (USD)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	priceOutputColumn := container.NewVBox(widget.NewLabelWithStyle("Output (USD)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	priceCachedColumn := container.NewVBox(widget.NewLabelWithStyle("Cached (USD)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	for _, model := range models {
		price := modelPrices[model]
		priceModelColumn.Add(widget.NewLabel(model))
		priceInputColumn.Add(widget.NewLabel(fmt.Sprintf("$%.6f", price.Input*limitToken)))
		priceOutputColumn.Add(widget.NewLabel(fmt.Sprintf("$%.6f", price.Output*limitToken)))
		priceCachedColumn.Add(widget.NewLabel(fmt.Sprintf("$%.6f", price.Cached*limitToken)))
	}

	priceTable := container.NewThemeOverride(container.NewGridWithColumns(4,
		priceModelColumn,
		priceInputColumn,
		priceOutputColumn,
		priceCachedColumn,
	), outputEditorTheme{Theme: theme.DefaultTheme()})
	tabs.Append(container.NewTabItemWithIcon(
		"Đơn giá",
		theme.DocumentIcon(),
		container.NewVBox(
			widget.NewLabelWithStyle("Đơn giá trên 1 triệu token", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			priceTable,
		),
	))

	return tabs
}

func CostTable(entries tokenUsageData, model string) *widget.Table {
	headers := []string{"Ngày", "Giờ", "Input", "Output", "Cached", "Chi phí"}
	sessions := entries[model]
	totalRow := len(sessions) + 1
	if len(sessions) == 0 {
		totalRow++
	}
	rows := totalRow + 1
	price := modelPrices[model]
	sessionCost := func(session tokenUsageSession) float64 {
		inputTokens := session.InputTokens - session.CachedTokens
		if inputTokens < 0 {
			inputTokens = 0
		}
		return float64(inputTokens)*price.Input +
			float64(session.CachedTokens)*price.Cached +
			float64(session.OutputTokens+session.ThoughtTokens)*price.Output
	}
	var totalCost float64
	for _, session := range sessions {
		totalCost += sessionCost(session)
	}
	table := widget.NewTable(
		func() (int, int) { return rows, len(headers) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(cell widget.TableCellID, object fyne.CanvasObject) {
			label := object.(*widget.Label)
			if cell.Row == 0 {
				label.TextStyle = fyne.TextStyle{Bold: true}
				label.SetText(headers[cell.Col])
				return
			}
			if cell.Row == totalRow {
				label.TextStyle = fyne.TextStyle{Bold: true}
				switch cell.Col {
				case 0:
					label.SetText("Tổng")
				case 5:
					label.SetText(fmt.Sprintf("$%.6f", totalCost))
				default:
					label.SetText("")
				}
				return
			}
			label.TextStyle = fyne.TextStyle{}
			if len(sessions) == 0 {
				if cell.Col == 0 {
					label.SetText("Chưa có dữ liệu")
				} else {
					label.SetText("")
				}
				return
			}

			session := sessions[cell.Row-1]
			switch cell.Col {
			case 0:
				label.SetText(session.StartedAt.Format("2006-01-02"))
			case 1:
				label.SetText(session.StartedAt.Format("15:04:05"))
			case 2:
				label.SetText(fmt.Sprintf("%d", session.InputTokens))
			case 3:
				label.SetText(fmt.Sprintf("%d", session.OutputTokens+session.ThoughtTokens))
			case 4:
				label.SetText(fmt.Sprintf("%d", session.CachedTokens))
			case 5:
				label.SetText(fmt.Sprintf("$%.6f", sessionCost(session)))
			}
		},
	)
	for col, width := range []float32{120, 100, 110, 110, 110, 120} {
		table.SetColumnWidth(col, width)
	}
	return table
}
