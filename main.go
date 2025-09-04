package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"fyne.io/systray"
	"fyne.io/systray/example/icon"
	"github.com/go-co-op/gocron/v2"
)

const configFile = "config.json"

type config struct {
	Url      string `json:"url"`
	Filename string `json:"Filename"`
	Cron     string `json:"cron"`
	Cmd      string `json:"cmd"`
	Args     string `json:"args"`

	filenameRegex *regexp.Regexp
}

var cfg *config

func main() {
	initCfg()
	scheduler, err := gocron.NewScheduler()
	panicIfErr(err)
	setupCronJob(scheduler)

	scheduler.Start()

	systray.Run(onReady, func() {
		err = scheduler.Shutdown()
		logErr(err)
	})
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("WPaper")
	systray.SetTooltip("Pretty awesome超级棒")
	setupNext()
	setupQuit()
}

func setupNext() {
	mNext := systray.AddMenuItem("Next", "Next wallpaper")
	go func() {
		for range mNext.ClickedCh {
			filePath := download(cfg.Url)
			execute(filePath)
		}
	}()
}

func setupQuit() {
	mQuit := systray.AddMenuItem("Quit", "Quit the app")
	go func() {
		for range mQuit.ClickedCh {
			systray.Quit()
		}
	}()
}

func initCfg() {
	wd, err := os.Getwd()
	panicIfErr(err)
	configPath := filepath.Join(wd, configFile)
	cfg = loadConfigFromFile(configPath)

	if cfg != nil && cfg.Filename != "" {
		cfg.filenameRegex, err = regexp.Compile(cfg.Filename)
		panicIfErr(err)
	}
}

func loadConfigFromFile(file string) *config {
	data, err := os.ReadFile(file)
	panicIfErr(err)
	var c config
	err = json.Unmarshal(data, &c)
	panicIfErr(err)
	return &c
}

func filename(url string) string {
	seconds := time.Now().Unix()
	return fmt.Sprintf("%d", seconds)
}

func download(url string) string {
	res, err := http.Get(url)
	logErr(err)
	defer func() {
		err = res.Body.Close()
		logErr(err)
	}()

	name := filename(res.Request.URL.String())
	filePath := wallpaperPath(name)

	file, err := os.Create(filePath)
	panicIfErr(err)
	defer func() {
		err = file.Close()
		logErr(err)
	}()
	_, err = io.Copy(file, res.Body)
	panicIfErr(err)

	return filePath
}

func wallpaperPath(filename string) string {
	wd, err := os.Getwd()
	panicIfErr(err)
	dw := filepath.Join(wd, "download")
	err = os.MkdirAll(dw, 0755)
	panicIfErr(err)
	return filepath.Join(dw, filename)
}

func setupCronJob(s gocron.Scheduler) {
	_, err := s.NewJob(
		gocron.CronJob(cfg.Cron, true),
		gocron.NewTask(
			func() {
				fmt.Println(11)
			},
		),
	)
	panicIfErr(err)
}

func execute(absPath string) {
	args := fmt.Sprintf(cfg.Args, absPath)
	cmd := exec.Command(cfg.Cmd, args)
	stdout, err := cmd.Output()
	logErr(err)
	fmt.Println(string(stdout))
}

func logErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
