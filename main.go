package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/ak1ra24/dlogview/api"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type primitiveKey struct {
	Primitive tview.Primitive
}

type DockerLogTUI struct {
	leftPanel  *DockerList
	rightPanel *tview.TextView

	layout            *tview.Flex
	app               *tview.Application
	focusPrimitives   []primitiveKey
	currentFocusIndex int
}

type DockerList struct {
	*tview.List
	Containers []types.Container
}

func NewDockerLogTUI() *DockerLogTUI {
	ui := &DockerLogTUI{
		focusPrimitives:   make([]primitiveKey, 0),
		currentFocusIndex: 0,
	}
	ui.app = tview.NewApplication()

	ui.leftPanel = ui.createLeftPanel()
	if err := ui.drawLeftPanel(); err != nil {
		log.Fatal(err)
	}

	ui.rightPanel = ui.createRightPanel()

	// ui.drawRightPanel()

	ui.layout = tview.NewFlex().
		AddItem(ui.leftPanel, 0, 2, false).
		AddItem(ui.rightPanel, 0, 10, false)

	return ui
}

func (ui *DockerLogTUI) createLeftPanel() *DockerList {
	panel := &DockerList{List: tview.NewList()}
	panel.SetBorder(true).SetTitle("Container List")

	return panel
}

func (ui *DockerLogTUI) createRightPanel() *tview.TextView {
	panel := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			ui.app.Draw()
		})
	panel.SetBorder(true).SetTitle("Log View")

	return panel
}

func (ui *DockerLogTUI) drawLeftPanel() error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	c := api.NewContainer(ctx, cli)
	ui.leftPanel.Containers, err = c.ListContainer()
	if len(ui.leftPanel.Containers) == 0 {
		err = errors.New("Sorry, No Docker Container process")
		return err
	}

	for i, container := range ui.leftPanel.Containers {
		fmt.Println(container.Names, container.Image, i)
		shortcutNum := rune(i)
		ui.leftPanel.AddItem(container.Names[0], container.Image, shortcutNum, nil)
	}

	return nil
}

func (ui *DockerLogTUI) drawRightPanel(containerID string) {
	ui.rightPanel.Clear()
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}
	c := api.NewContainer(ctx, cli)
	r, err := c.Cli.ContainerLogs(c.Ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		w := tview.ANSIWriter(ui.rightPanel)
		if _, err := io.Copy(w, r); err != nil {
			panic(err)
		}
	}()
}

func main() {
	ui := NewDockerLogTUI()

	ui.focusPrimitives = append(ui.focusPrimitives, primitiveKey{Primitive: ui.leftPanel})
	ui.focusPrimitives = append(ui.focusPrimitives, primitiveKey{Primitive: ui.rightPanel})

	var currentIndex = 0
	ui.drawRightPanel(ui.leftPanel.Containers[currentIndex].ID)
	ui.leftPanel.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyUp {
			currentIndex = currentIndex - 1
			if currentIndex < 0 {
				currentIndex = len(ui.leftPanel.Containers) - 1
				ui.drawRightPanel(ui.leftPanel.Containers[currentIndex].ID)
			} else {
				ui.drawRightPanel(ui.leftPanel.Containers[currentIndex].ID)
			}
		} else if event.Key() == tcell.KeyDown {
			currentIndex = currentIndex + 1
			if currentIndex > len(ui.leftPanel.Containers)-1 {
				currentIndex = 0
				ui.drawRightPanel(ui.leftPanel.Containers[currentIndex].ID)
			} else {
				ui.drawRightPanel(ui.leftPanel.Containers[currentIndex].ID)
			}
		}

		return event
	})

	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlQ {
			ui.app.Stop()
		} else if event.Key() == tcell.KeyTab {
			nextFocusIndex := ui.currentFocusIndex + 1
			if nextFocusIndex > len(ui.focusPrimitives)-1 {
				nextFocusIndex = 0
			}

			ui.app.SetFocus(ui.focusPrimitives[nextFocusIndex].Primitive)
			ui.currentFocusIndex = nextFocusIndex

			return nil
		}
		return event
	})

	if err := ui.app.SetRoot(ui.layout, true).Run(); err != nil {
		panic(err)
	}
}
