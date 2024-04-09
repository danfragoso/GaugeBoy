package main

/*
#cgo CFLAGS: -I/opt/miyoomini-toolchain/arm-linux-gnueabihf/libc/usr/include/SDL -O2 -w -D_GNU_SOURCE=1 -D_REENTRANT
#cgo LDFLAGS: -L/opt/miyoomini-toolchain/arm-linux-gnueabihf/libc/usr/lib -lSDL -lpthread
#include "main.c"
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
	"unsafe"

	"github.com/fogleman/gg"
)

func (app *GaugeBoy) Init() *GaugeBoy {
	println("GaugeBoy - If you are reading this give it a star on github!")
	C.init()

	app.DC = gg.NewContext(640, 480)
	app.FB, _ = app.DC.Image().(*image.RGBA)
	app.DC.LoadFontFace("./assets/fonts/Inter.ttf", 25)
	app.DC.SetRGB(0, 0, 0)

	file, err := os.Open("./assets/imgs/splash.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	app.Splash, err = png.Decode(file)
	if err != nil {
		panic(err)
	}

	app.DisplayText("MMP ODB2 Tool!")

	app.Config = &gaugeConfig{}
	app.Running = true

	go app.RunUI()
	time.Sleep(4 * time.Second)
	return app
}

func (app *GaugeBoy) DrawSelectedGauge() {
	if app.SelectedGauge == nil {
		app.SelectedGauge = &Gauge{}
	}

	if !app.SelectedGauge.Initialized {
		file, err := os.Open("./gauges/rpm/bg.png")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		bg, err := png.Decode(file)
		if err != nil {
			panic(err)
		}
		app.SelectedGauge.DC = gg.NewContext(640, 480)

		app.SelectedGauge.DC.DrawImage(bg, 0, 0)

		app.SelectedGauge.DC.LoadFontFace("./assets/fonts/Inter.ttf", 25)
		app.SelectedGauge.DC.SetColor(color.White)

		app.SelectedGauge.DC.DrawStringWrapped("RPM", 640/2, 300, 0.5, 1, 600, 1.5, gg.AlignCenter)

		app.SelectedGauge.DC.Fill()

		app.SelectedGauge.Initialized = true
		app.DC.LoadFontFace("./assets/fonts/Inter.ttf", 150)
		app.DC.SetColor(color.White)
	}

	app.Inc++
	rpmMsg, err := app.SendMSG("01 0C")
	if err != nil {
		app.Panic(err)
	}

	rpmBytes := rpmMsg[3:]

	app.DC.DrawImage(app.SelectedGauge.DC.Image(), 0, 0)
	rpmString := rpmBytes
	app.DC.DrawStringWrapped(rpmString, 640/2, 240, 0.5, 1, 600, 1.5, gg.AlignCenter)
	app.DC.Fill()

	app.ShouldRefresh = true
}

func (app *GaugeBoy) DisplayText(text string) {
	app.DC.DrawImage(app.Splash, 0, 0)
	app.DC.DrawStringWrapped(text, 640/2, 360, 0.5, 1, 600, 1.5, gg.AlignCenter)
	app.DC.Fill()

	app.ShouldRefresh = true
}

func (app *GaugeBoy) RunUI() {
	for app.Running {
		value := C.pollEvents()
		switch value {
		case 0:
			app.Running = false
		case 1:
			for y := 0; y < 480; y++ {
				for x := 0; x < 640; x++ {
					app.FB.SetRGBA(x, y, color.RGBA{uint8(rand.Intn(256)), uint8(rand.Intn(256)), uint8(rand.Intn(256)), 255})
				}
			}
			app.ShouldRefresh = true
		}

		if app.ShouldRefresh {
			C.refreshScreenPtr((*C.uchar)(unsafe.Pointer(&app.FB.Pix[0])))
			app.ShouldRefresh = false
		}
	}

	C.quit()
}

func (app *GaugeBoy) Panic(err error) {
	app.SetupFailed = true
	app.Connected = false
	app.DisplayText(err.Error())
	time.Sleep(3 * time.Second)
	app.Running = false
}

func (app *GaugeBoy) Run() {
	for app.Running {
		if !app.Connected && !app.SetupFailed {
			if app.Config.Host == "" && app.Config.Port == "" {
				configFile, err := os.Open("./assets/gaugeConfig.json")
				if err != nil {
					app.Panic(err)
				}

				jsonData, err := io.ReadAll(configFile)
				if err != nil {
					app.Panic(err)
				}

				if err := json.Unmarshal(jsonData, app.Config); err != nil {
					app.Panic(err)
				}

				if app.Config.Host == "" || app.Config.Port == "" {
					app.Panic(fmt.Errorf("Invalid host and port on gaugeConfig.json"))
				}

				app.DisplayText("Connecting to " + app.Config.Host + ":" + app.Config.Port + "...\n 10 seconds timeout...")
				app.Socket, err = net.DialTimeout("tcp", app.Config.Host+":"+app.Config.Port, 10*time.Second)
				if err != nil {
					app.Panic(err)
				}

				app.SendMSG("AT D")
				app.SendMSG("AT Z")
				app.SendMSG("AT E0")
				app.SendMSG("AT L0")
				app.SendMSG("AT S0")
				app.SendMSG("AT H0")
				app.SendMSG("AT SP 0")

				if app.LastMsg != "OK" {
					app.Panic(fmt.Errorf("Failed to configure ODB2; Response: '" + app.LastMsg + "'"))
				}

				app.Connected = true
				app.DisplayText("Successfully connected to ODB2")
				time.Sleep(3 * time.Second)
			}
		} else {
			app.DrawSelectedGauge()
		}
	}
}

func (app *GaugeBoy) SendMSG(msg string) (string, error) {
	app.Socket.Write([]byte(msg + "\r\n"))
	time.Sleep(33 * time.Millisecond)

	msgLen, err := app.Socket.Read(app.ODB2ReaderBuffer[:])
	if err != nil {
		return "", err
	}

	app.LastMsg = strings.TrimSpace(string(app.ODB2ReaderBuffer[:msgLen-3]))
	return app.LastMsg, nil
}

type gaugeConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type GaugeBoy struct {
	Running       bool
	ShouldRefresh bool

	DC     *gg.Context
	FB     *image.RGBA
	Splash image.Image

	Connected   bool
	SetupFailed bool

	ODB2ReaderBuffer [2048]byte
	LastMsg          string

	Inc int

	Socket net.Conn
	Config *gaugeConfig

	SelectedGauge *Gauge
}

type Gauge struct {
	Initialized bool
	BG          *image.RGBA
	DC          *gg.Context
}

func createApp() *GaugeBoy {
	return &GaugeBoy{
		Running: true,
		FB:      image.NewRGBA(image.Rect(0, 0, 640, 480)),
	}
}

func main() {
	App := createApp()
	App.Init().Run()
}
