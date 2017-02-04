package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/VividCortex/godaemon"
	"github.com/fullpipe/screen-time-lapse/icon"
	"github.com/getlantern/systray"
	"github.com/vova616/screenshot"
)

var shotTimeout float64
var savePath string

func main() {
	flag.Parse()

	var err error
	shotTimeout, err = strconv.ParseFloat(flag.Arg(0), 64)
	if err != nil {
		panic(err)
	}

	shotTimeout = shotTimeout * 1000 //convert to milliseconds

	savePath, err = filepath.Abs(flag.Arg(1))
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(savePath, 0777)
	if err != nil {
		panic(err)
	}

	godaemon.MakeDaemon(&godaemon.DaemonAttr{})
	systray.Run(onReady)
}

func onReady() {
	systray.SetIcon(icon.Default)
	systray.SetTooltip(fmt.Sprintf("Screen time lapse (%f sec.)", shotTimeout))

	mCaptures := systray.AddMenuItem("0 captures", "captures")
	mCaptures.Disable()

	pause := false
	quit := false

	go func() {
		mPause := systray.AddMenuItem("Pause", "Pause screen capturing")
		mQuit := systray.AddMenuItem("Stop", "Stop screen capturing")

		for {
			select {
			case <-mQuit.ClickedCh:
				quit = true
				systray.Quit()
				return
			case <-mPause.ClickedCh:
				if pause {
					pause = false
					mPause.SetTitle("Pause")
					mPause.SetTooltip("Pause screen capturing")
				} else {
					pause = true
					mPause.SetTitle("Resume")
					mPause.SetTooltip("Resume screen capturing")
				}
			}
		}
	}()

	go func() {
		counter := 0
		for {
			if quit {
				return
			}

			time.Sleep(time.Millisecond * time.Duration(shotTimeout))

			if pause {
				continue
			}

			go makeScreenshot(fmt.Sprintf("%s/%d.png", savePath, counter))
			go blinkTray()

			counter++
			mCaptures.SetTitle(fmt.Sprintf("%d captures", counter))
		}
	}()
}

func blinkTray() {
	systray.SetIcon(icon.Red)
	time.Sleep(time.Millisecond * 200)

	systray.SetIcon(icon.Default)
}

func makeScreenshot(imagePath string) {
	img, err := screenshot.CaptureScreen()
	if err != nil {
		panic(err)
	}

	println(imagePath)

	f, err := os.Create(imagePath)
	if err != nil {
		panic(err)
	}

	err = png.Encode(f, img)
	if err != nil {
		panic(err)
	}
	f.Close()
}
