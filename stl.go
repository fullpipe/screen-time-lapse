package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"log"
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
var gifFramerate = flag.Int("gif", 0, "convert images to gif with framerate")

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Printf("  screen-time-lapse [-gif 30] TIMEOUT SAVEPATH\n")
		flag.PrintDefaults()
	}

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

	err = os.MkdirAll(savePath, 0755)
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
				time.Sleep(time.Millisecond * 100) //wait makeScreenshot completition
				if *gifFramerate > 0 {
					go generateGif(counter)
				}
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
	defer f.Close()
	if err != nil {
		panic(err)
	}

	err = png.Encode(f, img)
	if err != nil {
		panic(err)
	}
}

func generateGif(num int) {
	var frames []*image.Paletted

	for i := 0; i < num; i++ {
		pngPath := fmt.Sprintf("%s/%d.png", savePath, i)
		f, _ := os.Open(pngPath)
		img, _ := png.Decode(f)

		buf := bytes.Buffer{}
		if err := gif.Encode(&buf, img, nil); err != nil {
			log.Printf("Skipping file %s due to error in gif encoding:%s", pngPath, err)
			continue
		}

		tmpimg, err := gif.Decode(&buf)
		if err != nil {
			log.Printf("Skipping file %s due to weird error reading the temporary gif :%s", pngPath, err)
			continue
		}

		frames = append(frames, tmpimg.(*image.Paletted))
	}

	delays := make([]int, len(frames))
	for j := range delays {
		delays[j] = 1 / *gifFramerate * 100
	}

	opfile, err := os.Create(fmt.Sprintf("%s/out.gif", savePath))
	if err != nil {
		log.Fatalf("Error creating the destination file %s : %s", savePath, err)
	}

	if err := gif.EncodeAll(opfile, &gif.GIF{Image: frames, Delay: delays, LoopCount: 0}); err != nil {
		log.Printf("Error encoding output into animated gif: %s", err)
	}

	opfile.Close()
}
