package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"

	// "github.com/caseymrm/go-smc"
	"github.com/fatih/color"
	ag "github.com/guptarohit/asciigraph"
)

var (
	version = "dev"
	helpMsg = ``
	chart   = false
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
	if hasOption, i := argsHaveOption("chart", "c"); hasOption {
		chart = !chart
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
		cool(70)
	}
	temp, err := strconv.Atoi(os.Args[1])
	if err != nil {
		handleErr(err)
		return
	}
	cool(temp)
}

func cool(target int) {
	fmt.Println("Cooling to", color.YellowString("%v C", target))
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		<-c
		setFanSpeed(1200) // default min fan speed
		os.Exit(0)
	}()
	t := getTemp()
	var s int

	var tplot, fplot []float64
	setFanSpeed(1200 + 100*(int(t)-target)) // quickly set it initally
	for ; ; time.Sleep(time.Second * 5) {   // fine tune
		s = getFanSpeed()
		t = getTemp()

		if !chart {
			fmt.Printf("%v %8v RPM\n", color.YellowString("%.1f C", t), s)
		} else {
			ag.Clear()
			fmt.Println("Cooling to", color.YellowString("%v C", target))
			// fmt.Println("\033c") // clear screen

			tplot = append(tplot, t)
			for len(tplot) > 99 { // window width - 1
				tplot = removeKeepOrderFloat(tplot, 0)
			}
			if !(tplot[0] == 50) {
				tplot = append([]float64{50}, tplot...) // get 50 value so plot starts from 50
			}
			fmt.Println(ag.Plot(tplot, ag.Height(8), ag.Caption("Temperature (C)")))

			// fmt.Println()

			fplot = append(fplot, float64(s))
			for len(fplot) > 99 { // window width - 1
				fplot = removeKeepOrderFloat(fplot, 0)
			}
			if !(fplot[0] == 1200) {
				fplot = append([]float64{1200}, fplot...) // get 1200 value so plot scales from 1200
			}
			fmt.Println(ag.Plot(fplot, ag.Height(8), ag.Caption("Fan speed (RPM)")))
			fmt.Printf("Now at %v, %v RPM\n", color.YellowString("%.1f C", t), s)
		}
		setFanSpeed(s + 10*(int(t)-target)) // set current to current + 10 times the difference in temps. This will automatically correct when temp is too low.
	}
}

func setFanSpeed(minSpeed int) {
	if minSpeed > 6500 { // max safe fan speed
		minSpeed = 6500
	}
	if minSpeed < 1200 { // min safe fan speed
		minSpeed = 1200
	}
	//fmt.Println(strconv.FormatInt(int64(minSpeed<<2), 16))
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

func removeKeepOrderFloat(s []float64, i int) []float64 {
	return append(s[:i], s[i+1:]...)
}
