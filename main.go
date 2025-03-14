package main

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type TempoDiPomo struct {
	workDuration  time.Duration
	breakDuration time.Duration
	running       bool
	paused        bool
	timeLeft      time.Duration
	isWorkSession bool
	app           fyne.App
	window        fyne.Window
	timeLabel     *widget.Label
	startButton   *widget.Button // Declare buttons as struct fields
	pauseButton   *widget.Button
}

func main() {
	fyneApp := app.New()
	p := &TempoDiPomo{
		workDuration:  25 * time.Minute,
		breakDuration: 5 * time.Minute,
		timeLeft:      25 * time.Minute,
		isWorkSession: true,
		app:           fyneApp,
	}
	p.window = fyneApp.NewWindow("TempoDiPomo")
	p.run()
}

func (p *TempoDiPomo) run() {
	p.timeLabel = widget.NewLabel("🍅 Work 25:00")

	p.startButton = widget.NewButton("Start", func() {
		if !p.running {
			p.running = true
			p.startButton.Disable()
			p.pauseButton.Enable()
			go p.runTimer()
		}
	})
	p.pauseButton = widget.NewButton("Pause", func() {
		p.paused = !p.paused
		if p.paused {
			p.pauseButton.SetText("Resume")
		} else {
			p.pauseButton.SetText("Pause")
		}
	})
	p.pauseButton.Disable()
	resetButton := widget.NewButton("Reset", func() {
		p.reset()
		p.startButton.Enable()
		p.pauseButton.Disable()
		p.pauseButton.SetText("Pause")
	})
	configButton := widget.NewButton("Configure", func() {
		p.showConfigDialog()
	})
	quitButton := widget.NewButton("Quit", func() {
		p.app.Quit()
	})

	content := container.NewVBox(
		p.timeLabel,
		container.NewHBox(p.startButton, p.pauseButton, resetButton),
		configButton,
		quitButton,
	)

	p.window.SetContent(content)
	p.window.ShowAndRun()
}

func (p *TempoDiPomo) runTimer() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for p.running {
		if !p.paused {
			p.timeLeft -= time.Second
			p.updateDisplay()
			if p.timeLeft <= 0 {
				p.sessionComplete()
				break
			}
		}
		<-ticker.C
	}
}

func (p *TempoDiPomo) updateDisplay() {
	minutes := int(p.timeLeft.Minutes())
	seconds := int(p.timeLeft.Seconds()) % 60
	session := "Work"
	if !p.isWorkSession {
		session = "Break"
	}
	p.timeLabel.SetText(fmt.Sprintf("🍅 %s %02d:%02d", session, minutes, seconds))
}

func (p *TempoDiPomo) sessionComplete() {
	p.running = false
	if p.isWorkSession {
		p.notify("Work session complete!", "Time for a break!")
		p.timeLeft = p.breakDuration
		p.isWorkSession = false
	} else {
		p.notify("Break session complete!", "Time to work!")
		p.timeLeft = p.workDuration
		p.isWorkSession = true
	}
	p.updateDisplay()
	p.startButton.Enable()
	p.pauseButton.Disable()
	p.pauseButton.SetText("Pause")
}

func (p *TempoDiPomo) reset() {
	p.running = false
	p.paused = false
	p.isWorkSession = true
	p.timeLeft = p.workDuration
	p.updateDisplay()
}

func (p *TempoDiPomo) notify(title, message string) {
	dialog.NewInformation(title, message, p.window).Show()
}

func (p *TempoDiPomo) showConfigDialog() {
	workEntry := widget.NewEntry()
	workEntry.SetText(strconv.Itoa(int(p.workDuration.Minutes())))
	breakEntry := widget.NewEntry()
	breakEntry.SetText(strconv.Itoa(int(p.breakDuration.Minutes())))

	form := dialog.NewForm("Configure Timer", "Save", "Cancel", []*widget.FormItem{
		{Text: "Work Duration (5-120 min):", Widget: workEntry},
		{Text: "Break Duration (1-60 min):", Widget: breakEntry},
	}, func(confirmed bool) {
		if confirmed {
			workMinutes, workErr := strconv.Atoi(workEntry.Text)
			breakMinutes, breakErr := strconv.Atoi(breakEntry.Text)
			if workErr == nil && breakErr == nil && workMinutes >= 5 && workMinutes <= 120 && breakMinutes >= 1 && breakMinutes <= 60 {
				p.workDuration = time.Duration(workMinutes) * time.Minute
				p.breakDuration = time.Duration(breakMinutes) * time.Minute
				if p.isWorkSession {
					p.timeLeft = p.workDuration
				} else {
					p.timeLeft = p.breakDuration
				}
				p.updateDisplay()
			} else {
				p.notify("Invalid Input", "Work: 5-120 min, Break: 1-60 min")
			}
		}
	}, p.window)

	form.Show()
}