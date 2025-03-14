package main

import (
	"fmt"
	"time"

	"github.com/getlantern/systray"
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
	systray.Run(onReady, onExit)
}

func onReady() {
	p := &TempoDiPomo{
		workDuration:  25 * time.Minute,
		breakDuration: 5 * time.Minute,
		timeLeft:      25 * time.Minute,
		isWorkSession: true,
	}

	systray.SetTitle("🍅 25:00")
	systray.SetTooltip("TempoDiPomo - Pomodoro Timer")

	mStart := systray.AddMenuItem("Start", "Start the timer")
	mPause := systray.AddMenuItem("Pause", "Pause/Resume the timer")
	mReset := systray.AddMenuItem("Reset", "Reset the timer")
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
		notify("Work session complete!", "Time for a break!")
		p.timeLeft = p.breakDuration
		p.isWorkSession = false
	} else {
		notify("Break session complete!", "Time to work!")
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

func notify(title, message string) {
	fmt.Printf("Notification: %s - %s\n", title, message)
}

func onExit() {
	fmt.Println("TempoDiPomo app exited")
}