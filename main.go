package main

import (
	"fmt"
	"time"

	"github.com/getlantern/systray"
	"github.com/ncruces/zenity"
)

type TempoDiPomo struct {
	workDuration  time.Duration
	breakDuration time.Duration
	running       bool
	paused        bool
	timeLeft      time.Duration
	isWorkSession bool
}

func main() {
	p := &TempoDiPomo{
		workDuration:  25 * time.Minute,
		breakDuration: 5 * time.Minute,
		timeLeft:      25 * time.Minute,
		isWorkSession: true,
	}
	systray.Run(p.onReady, p.onExit)
}

func (p *TempoDiPomo) onReady() {
	systray.SetTitle("🍅 25:00")
	systray.SetTooltip("TempoDiPomo - Pomodoro Timer")

	mStart := systray.AddMenuItem("Start", "Start the timer")
	mPause := systray.AddMenuItem("Pause", "Pause/Resume the timer")
	mReset := systray.AddMenuItem("Reset", "Reset the timer")
	mConfig := systray.AddMenuItem("Configure", "Configure timer durations")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the app")

	mPause.Disable()

	go func() {
		for {
			select {
			case <-mStart.ClickedCh:
				if !p.running {
					p.running = true
					mStart.Disable()
					mPause.Enable()
					go p.runTimer()
				}
			case <-mPause.ClickedCh:
				p.paused = !p.paused
				if p.paused {
					mPause.SetTitle("Resume")
				} else {
					mPause.SetTitle("Pause")
				}
			case <-mReset.ClickedCh:
				p.reset()
				mStart.Enable()
				mPause.Disable()
				mPause.SetTitle("Pause")
			case <-mConfig.ClickedCh:
				p.configureDurations()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
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
	systray.SetTitle(fmt.Sprintf("🍅 %s %02d:%02d", session, minutes, seconds))
}

func (p *TempoDiPomo) sessionComplete() {
	p.running = false
	if p.isWorkSession {
		p.notify("Work session complete! Time for a break!")
		p.timeLeft = p.breakDuration
		p.isWorkSession = false
	} else {
		p.notify("Break session complete! Time to work!")
		p.timeLeft = p.workDuration
		p.isWorkSession = true
	}
	p.updateDisplay()
}

func (p *TempoDiPomo) reset() {
	p.running = false
	p.paused = false
	p.isWorkSession = true
	p.timeLeft = p.workDuration
	p.updateDisplay()
}

func (p *TempoDiPomo) notify(message string) {
	go zenity.Info(message, zenity.Title("TempoDiPomo"))
}

func (p *TempoDiPomo) configureDurations() {
	input, err := zenity.Entry("Enter work and break durations in minutes (e.g., 25,5)",
		zenity.Title("Set Timer Durations"), zenity.EntryText(fmt.Sprintf("%d,%d", int(p.workDuration.Minutes()), int(p.breakDuration.Minutes()))))
	if err != nil {
		return // User canceled
	}

	var workMinutes, breakMinutes int
	_, err = fmt.Sscanf(input, "%d,%d", &workMinutes, &breakMinutes)
	if err != nil || workMinutes < 5 || workMinutes > 120 || breakMinutes < 1 || breakMinutes > 60 {
		zenity.Error("Invalid input. Work: 5-120 min, Break: 1-60 min.", zenity.Title("Invalid Input"))
		return
	}

	p.workDuration = time.Duration(workMinutes) * time.Minute
	p.breakDuration = time.Duration(breakMinutes) * time.Minute

	if p.isWorkSession {
		p.timeLeft = p.workDuration
	} else {
		p.timeLeft = p.breakDuration
	}
	p.updateDisplay()
}

func (p *TempoDiPomo) onExit() {
	fmt.Println("TempoDiPomo app exited")
}
