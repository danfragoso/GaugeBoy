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
	"image/png"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
)

const SCREEN_WIDTH = 640
const SCREEN_HEIGHT = 480

func (app *GaugeBoy) Init() *GaugeBoy {
	log("Initing App...")
	log("SDL init...")
	C.init()
	log("SDL init ok!")

	app.DC = gg.NewContext(SCREEN_WIDTH, SCREEN_HEIGHT)
	app.FB, _ = app.DC.Image().(*image.RGBA)
	app.DC.LoadFontFace("./assets/ui_font.ttf", 25)
	app.DC.SetRGB(0, 0, 0)

	file, err := os.Open("./assets/splash.png")
	if err != nil {
		log(fmt.Sprintf("Failed to load splash: %v", err))
		panic(err)
	}
	defer file.Close()

	app.Splash, err = png.Decode(file)
	if err != nil {
		log(fmt.Sprintf("Failed to decode splash: %v", err))
		panic(err)
	}

	app.DisplayText("MMP ODB2 Tool!")

	app.Config = &Config{}
	app.Running = true

	log("App init OK!")
	return app
}

func (app *GaugeBoy) InitSelectedGauge() {
	newGauge := app.Config.Gauges[app.SelectedGaugeIndex]
	file, err := os.Open(newGauge.Bg)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	bg, err := png.Decode(file)
	if err != nil {
		panic(err)
	}

	newGauge.DC = gg.NewContext(SCREEN_WIDTH, SCREEN_HEIGHT)
	newGauge.DC.DrawImage(bg, 0, 0)
	newGauge.DC.Fill()

	newGauge.FontFace, err = gg.LoadFontFace(newGauge.Font, newGauge.FontSize)
	if err != nil {
		app.Panic(err)
	}

	newGauge.Initialized = true
	app.SelectedGauge = newGauge
}

func (app *GaugeBoy) DrawRPMGauge() {
	rpmMsg, err := app.SendODB2MSG("010C")
	if err != nil {
		app.Panic(err)
	}

	rpmBytes := strings.ReplaceAll(rpmMsg, "410C", "")
	rpmBytes = strings.ReplaceAll(rpmBytes, " ", "")

	rpmBytes = rpmBytes[:4]
	rpmValue, err := strconv.ParseInt(rpmBytes, 16, 64)
	if err != nil {
		app.DC.LoadFontFace("./assets/ui_font.ttf", 25)
		app.DC.SetRGB(0, 0, 0)

		app.DisplayText(rpmMsg)
	}

	app.DC.DrawImage(app.SelectedGauge.DC.Image(), 0, 0)
	app.DC.SetFontFace(app.SelectedGauge.FontFace)
	app.DC.SetHexColor(app.SelectedGauge.TextColor)
	app.DC.DrawStringWrapped(strconv.Itoa(int(rpmValue/4)), app.SelectedGauge.TextX, app.SelectedGauge.TextY, 0.5, 1, SCREEN_WIDTH-40, 1.5, gg.AlignCenter)
	app.DC.Fill()

	app.ShouldRefresh = true
}

func (app *GaugeBoy) DrawSpeedGauge() {
	speedMsg, err := app.SendODB2MSG("010D")
	if err != nil {
		app.Panic(err)
	}

	speedBytes := strings.ReplaceAll(speedMsg, "410D", "")
	speedBytes = strings.ReplaceAll(speedBytes, " ", "")

	speedBytes = speedBytes[:2]
	speedValue, err := strconv.ParseInt(speedBytes, 16, 64)
	if err != nil {
		app.DC.LoadFontFace("./assets/ui_font.ttf", 25)
		app.DC.SetRGB(0, 0, 0)

		app.DisplayText(speedMsg)
	}

	app.DC.DrawImage(app.SelectedGauge.DC.Image(), 0, 0)
	app.DC.SetFontFace(app.SelectedGauge.FontFace)
	app.DC.SetHexColor(app.SelectedGauge.TextColor)
	app.DC.DrawStringWrapped(strconv.Itoa(int(speedValue)), app.SelectedGauge.TextX, app.SelectedGauge.TextY, 0.5, 1, SCREEN_WIDTH-40, 1.5, gg.AlignCenter)
	app.DC.Fill()

	app.ShouldRefresh = true
}

func (app *GaugeBoy) DrawGauges() {
	if app.SelectedGauge != nil && app.SelectedGauge.Initialized {
		switch app.SelectedGauge.Type {
		case "rpm":
			app.DrawRPMGauge()
		case "speed":
			app.DrawSpeedGauge()
		}
	}
}

func (app *GaugeBoy) SelectNextGauge() {
	app.SelectedGauge = nil

	if app.SelectedGaugeIndex+1 >= len(app.Config.Gauges) {
		app.SelectedGaugeIndex = 0
	} else {
		app.SelectedGaugeIndex++
	}

	app.InitSelectedGauge()
}

func (app *GaugeBoy) DisplayText(text string) {
	app.DC.DrawImage(app.Splash, 0, 0)
	app.DC.DrawStringWrapped(text, SCREEN_WIDTH/2, 360, 0.5, 1, 600, 1.5, gg.AlignCenter)
	app.DC.Fill()

	app.ShouldRefresh = true
}

func (app *GaugeBoy) RunUI() {
	for app.Running {
		if app.ShouldRefresh {
			C.refreshScreenPtr((*C.uchar)(unsafe.Pointer(&app.FB.Pix[0])))
			app.ShouldRefresh = false
		}
	}

	//C.quit()
}

func (app *GaugeBoy) Panic(err error) {
	log(fmt.Sprintf("Panic: %v", err))

	app.SetupFailed = true
	app.Connected = false
	app.DC.LoadFontFace("./assets/ui_font.ttf", 25)
	app.DC.SetRGB(0, 0, 0)
	app.DisplayText(err.Error())
	time.Sleep(3 * time.Second)
	app.Running = false
}

func (app *GaugeBoy) Configure() {
	log("Configuring App...")
	configFile, err := os.Open("./GaugeBoy.json")
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
		app.Panic(fmt.Errorf("Invalid host and port on GaugeBoy.json"))
	}

	app.ConnectODB2()

	app.SelectedGaugeIndex = 0
	app.InitSelectedGauge()
}

func (app *GaugeBoy) ConnectODB2() {
	log("Connecting to ODB2 Device...")
	app.DisplayText("Connecting to " + app.Config.Host + ":" + app.Config.Port + "...\n 10 seconds timeout...")

	var err error
	app.Socket, err = net.DialTimeout("tcp", app.Config.Host+":"+app.Config.Port, 10*time.Second)
	if err != nil {
		app.Panic(err)
	}

	app.SendODB2MSG("AT D")
	app.SendODB2MSG("AT Z")
	app.SendODB2MSG("AT E0")
	app.SendODB2MSG("AT L0")
	app.SendODB2MSG("AT S0")
	app.SendODB2MSG("AT H0")
	app.SendODB2MSG("AT SP 0")

	if app.LastMsg == "" {
		app.Panic(fmt.Errorf("Failed to configure ODB2; Response: '" + app.LastMsg + "'"))
	}

	app.Connected = true
	log("Connected to ODB2 Device!")
	app.DisplayText("Connected to ODB2 Device!")
	time.Sleep(3 * time.Second)
}

func (app *GaugeBoy) SendODB2MSG(msg string) (string, error) {
	log(fmt.Sprintf("Sending ODB2 MSG: %v", msg))
	_, err := app.Socket.Write([]byte(msg + "\r\n"))
	if err != nil {
		log(fmt.Sprintf("Failed to send ODB2 MSG: %v", err))
		return "", err
	}
	time.Sleep(13 * time.Millisecond)

	msgLen, err := app.Socket.Read(app.ODB2ReaderBuffer[:])
	if err != nil {
		log(fmt.Sprintf("Failed to read ODB2 response: %v", err))
		return "", err
	}

	app.LastMsg = strings.TrimSpace(string(app.ODB2ReaderBuffer[:msgLen-3]))
	log(fmt.Sprintf("ODB2 Response: %v", app.LastMsg))
	return app.LastMsg, nil
}

type Config struct {
	Host   string   `json:"host"`
	Port   string   `json:"port"`
	Gauges []*Gauge `json:"gauges"`
}

type Gauge struct {
	Initialized bool
	DC          *gg.Context
	FontFace    font.Face

	Type      string  `json:"type"`
	Bg        string  `json:"bg"`
	TextColor string  `json:"textColor"`
	Font      string  `json:"font"`
	FontSize  float64 `json:"fontSize"`
	TextX     float64 `json:"textX"`
	TextY     float64 `json:"textY"`
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

	Socket net.Conn
	Config *Config

	SelectedGauge      *Gauge
	SelectedGaugeIndex int
}

func createApp() *GaugeBoy {
	return &GaugeBoy{
		Running: true,
		FB:      image.NewRGBA(image.Rect(0, 0, SCREEN_WIDTH, SCREEN_HEIGHT)),
	}
}

func main() {
	log("\n\n\n-----------")
	log("GaugeBoy started!")
	App := createApp()
	App.Init()

	go App.RunUI()
	time.Sleep(2 * time.Second)

	keyPresses := make(chan Key)
	go listenKeyPresses(keyPresses)

	for App.Running {
		select {
		case key := <-keyPresses:
			switch key {
			case MENU:
				App.Panic(fmt.Errorf("Exiting..."))

			case A:
				App.SelectNextGauge()
			}

		default:
			if !App.Connected && !App.SetupFailed {
				if App.Config.Host == "" && App.Config.Port == "" {
					App.Configure()
				}
			} else {
				App.DrawGauges()
			}
		}
	}
}

func C_GetKeyPress() int {
	return int(C.pollEvents())
}
