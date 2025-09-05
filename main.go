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

	systray.Run(func() {
		setupCronJob(scheduler)
		scheduler.Start()
		onReady()
	}, func() {
		err = scheduler.Shutdown()
		logErr(err)
	})
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("WPaper")
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
	if cfg.filenameRegex != nil {
		found := cfg.filenameRegex.FindStringSubmatch(url)
		if len(found) > 1 {
			return found[1]
		}
	}
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

	today := time.Now().Format("2006-01-02")
	dw := filepath.Join(wd, "download", today)
	err = os.MkdirAll(dw, 0755)
	panicIfErr(err)
	return filepath.Join(dw, filename)
}

func setupCronJob(s gocron.Scheduler) {
	if cfg.Cron == "" {
		return
	}

	var j gocron.Job
	j, err := s.NewJob(
		gocron.CronJob(cfg.Cron, true),
		gocron.NewTask(
			func() {
				t, err := j.NextRun()
				logErr(err)
				nextFireTime := t.Format("15:04:05")
				systray.SetTooltip(fmt.Sprintf("Next Fire Time: %s", nextFireTime))

				filePath := download(cfg.Url)
				execute(filePath)
			},
		),
	)
	panicIfErr(err)

	t, err := j.NextRun()
	logErr(err)
	nextFireTime := t.Format("15:04:05")
	systray.SetTooltip(fmt.Sprintf("Next Fire Time: %s", nextFireTime))
}

func execute(absPath string) {
	args := fmt.Sprintf(cfg.Args, absPath)
	cmd := exec.Command(cfg.Cmd, args)
	initCmd(cmd)
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
