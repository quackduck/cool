package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	ag "github.com/guptarohit/asciigraph"
	"github.com/muesli/termenv"
	"github.com/olekukonko/ts"
)

var (
	version = "dev"
	helpMsg = ``
	chart   = true
)

func main() {
	if hasOption, _ := argsHaveOption("help", "h"); hasOption {
		fmt.Println(helpMsg)
		return
	}
	if hasOption, _ := argsHaveOption("version", "v"); hasOption {
		fmt.Println("Cool " + version)
		return
	}
	if hasOption, i := argsHaveOption("no-chart", "c"); hasOption {
		chart = false
		os.Args = removeKeepOrder(os.Args, i)
		main()
		return
	}
	currentUser, err := user.Current()
	if err != nil {
		handleErr(err)
	}
	if !(currentUser.Username == "root") {
		handleErrStr("Setting fan values needs root permissions. Try running with sudo.")
		return
	}
	if len(os.Args) == 1 {
		cool(75)
	}
	temp, err := strconv.ParseFloat(os.Args[1], 32)
	if err != nil {
		handleErr(err)
		return
	}
	cool(temp)
}

func cool(target float64) {
	// Init:
	if !chart {
		fmt.Println("Cooling to", color.YellowString("%v C", target))
	} else {
		termenv.AltScreen()
		termenv.HideCursor()
	}
	setupInterrupt()
	var (
		speed       int
		tplot       []float64
		splot       = []float64{1200} // start at default
		size        ts.Size
		temp        = getTemp()
		printedTime = false
		g           = color.New(color.FgHiGreen)
		start       = time.Now()
	)

	setFanSpeed(1200 + int(math.Round(150*(temp-target)))) // quickly set it at the start
	for ; ; time.Sleep(time.Second * 2) {                  // fine tuning
		speed = getFanSpeed()
		temp = getTemp()

		if chart {
			size, _ = ts.GetSize()
			termenv.ClearScreen()
			fmt.Println("Target:", color.YellowString("%v C", target))
			// fmt.Println()

			tplot = append(tplot, temp)
			fmt.Println(ag.Plot(tplot, ag.Height((size.Row()/2)-2-2), ag.Width(size.Col()-7), ag.Caption("Temperature (C)")))

			splot = append(splot, float64(speed))
			fmt.Println(ag.Plot(splot, ag.Height((size.Row()/2)-2-2), ag.Width(size.Col()-7), ag.Offset(4), ag.Caption("Fan speed (RPM)")))

			fmt.Printf("Now at %v, %v RPM\n", color.YellowString("%.1f C", temp), speed)
			if math.Round(target) == math.Round(temp) { // nolint
				g.Print("At target")
				if !printedTime {
					g.Print(" in " + time.Since(start).Round(time.Second).String())
					printedTime = true
				}
				g.Print("!")
			} else if target > temp {
				color.New(color.FgCyan).Print("Cooler than target!")
			} else {
				color.New(color.FgRed).Print("Hotter than target")
			}
		} else {
			fmt.Printf("%v %8v RPM\n", color.YellowString("%.1f C", temp), speed)
		}
		setFanSpeed(speed + int(math.Round(temp-float64(target)))) // set current to current + the difference in temps. This will automatically correct when temp is too low.
	}
}

func setupInterrupt() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		<-c
		setFanSpeed(1200) // default min fan speed
		if chart {
			termenv.ExitAltScreen()
			termenv.ShowCursor()
		}
		os.Exit(0)
	}()
}

func setFanSpeed(minSpeed int) {
	if minSpeed > 6500 { // max safe fan speed
		minSpeed = 6500
	}
	if minSpeed < 1200 { // min safe fan speed
		minSpeed = 1200
	}
	setKey("F0Mn", strconv.FormatInt(int64(minSpeed<<2), 16)) // https://github.com/hholtmann/smcFanControl/tree/master/smc-command
}

func getFanSpeed() int {
	s, _ := strconv.ParseFloat(getKey("F0Mn"), 64) // get min fan speed (shouldn't change target)
	return int(s)
}

// getTemp returns the value of the SMC's CPU 1 sensor (Celsius)
func getTemp() float64 {
	t, _ := strconv.ParseFloat(getKey("TC0E"), 64)
	return t
}

func getKey(key string) string {
	return strings.Fields(strings.TrimSpace(run("smc -r -k " + key)))[2] // format is "   F0Mn  [fpe2]  2400.00 (bytes 25 80)" so we trim, split and return third
}

func setKey(key string, value string) {
	run("smc -k " + key + " -w " + value)
}

func run(command string) string {
	cmdArr := strings.Fields(strings.TrimSpace(command))
	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		handleErr(err)
		return ""
	}
	return string(b)
}

func argsHaveOption(long string, short string) (hasOption bool, foundAt int) {
	for i, arg := range os.Args {
		if arg == "--"+long || arg == "-"+short {
			return true, i
		}
	}
	return false, 0
}

func handleErr(err error) {
	handleErrStr(err.Error())
}

func handleErrStr(str string) {
	_, _ = fmt.Fprintln(os.Stderr, color.RedString("error: ")+str)
}

func removeKeepOrder(s []string, i int) []string {
	return append(s[:i], s[i+1:]...)
}
