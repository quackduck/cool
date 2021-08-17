package main

import (
	"encoding/binary"
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
	helpMsg = `Cool - Never let the heat slow your Mac down
Usage: cool [-c/--no-chart] [<temperature>]
       cool [-h/--help | -v/--version]`
	chart       = true
	defaultTemp = 75.0

	fanKeyEncodingIsFLT = getFanIsFLT()
)

func main() {
	if hasOption, i := argsHaveOption("no-chart", "c"); hasOption {
		chart = false
		os.Args = removeKeepOrder(os.Args, i)
		main()
		return
	}
	if len(os.Args) > 2 {
		handleErrStr("Too many arguments")
		fmt.Println(helpMsg)
		return
	}
	if hasOption, _ := argsHaveOption("help", "h"); hasOption {
		fmt.Println(helpMsg)
		return
	}
	if hasOption, _ := argsHaveOption("version", "v"); hasOption {
		fmt.Println("Cool " + version)
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
		cool(defaultTemp)
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
		fanSpeed             int
		termsize, _          = ts.GetSize()
		temp                 = getTemp()
		timeTaken            = ""
		alreadyReachedTarget = false
		green                = color.New(color.FgHiGreen)
		greenDot             = green.Sprint("·")
		cyan                 = color.New(color.FgCyan)
		cyanDown             = cyan.Sprint("↓")
		yellow               = color.New(color.FgHiYellow)
		yellowUp             = yellow.Sprint("↑")
		start                = time.Now()
		tplot                []float64
		splot                = []float64{float64(getFanSpeed())} // start at current because we'll change it soon
		lastTemp             = temp
		arrLengthLim         = 1000 // we'll keep an array limit so the values "scroll"
		screen               string
		statusFmt            = "Now at %.1f °C %v RPM Time: %v"
	)

	setFanSpeed(1200 + int(math.Round(150*(temp-target)))) // quickly set it at the start
	for ; ; time.Sleep(time.Second * 2) {                  // fine tuning
		fanSpeed = getFanSpeed()
		lastTemp = temp
		temp = getTemp()

		if chart {
			screen = ""
			screen += "Target " + green.Sprintf("%v °C", target) + timeTaken + "\n"

			tplot = append(tplot, temp)
			if len(tplot) > arrLengthLim {
				tplot = tplot[len(tplot)-arrLengthLim:] // cut off the front so we have max 100 vals
			}
			screen += ag.Plot(tplot, ag.Height((termsize.Row()/2)-4), ag.Width(termsize.Col()-7), ag.Caption("Temperature (C)")) + "\n"

			splot = append(splot, float64(fanSpeed))
			if len(splot) > arrLengthLim {
				splot = splot[len(splot)-arrLengthLim:]
			}
			screen += ag.Plot(splot, ag.Height((termsize.Row()/2)-4), ag.Width(termsize.Col()-7), ag.Offset(4), ag.Caption("Fan speed (RPM)")) + "\n"

			if math.Round(target) == math.Round(temp) { // nolint
				screen += greenDot + " Reached target!\n"
				if !alreadyReachedTarget {
					timeTaken = "reached in " + color.HiGreenString(time.Since(start).Round(time.Second).String())
					alreadyReachedTarget = true
				}
			} else if target > temp {
				screen += cyanDown + " Cooler than target!\n"
			} else {
				screen += yellowUp + " Hotter than target\n"
			}

			if lastTemp == temp { // nolint
				screen += greenDot + " Temperature is stable\n"
			} else if lastTemp > temp {
				screen += cyanDown + " Temperature is decreasing\n"
			} else {
				screen += yellowUp + " Temperature is increasing\n"
			}
			screen += fmt.Sprintf(statusFmt, temp, fanSpeed, time.Since(start).Round(time.Second))
			termsize, _ = ts.GetSize()
			termenv.ClearScreen()
			fmt.Print(screen)
		} else {
			screen += fmt.Sprintf(statusFmt, temp, fanSpeed, time.Since(start).Round(time.Second))
		}
		setFanSpeed(fanSpeed + int(math.Round(temp-target))) // set current to current + the difference in temps. This will automatically correct when temp is too low.
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
	//math.Float32frombits(binary.LittleEndian.Uint32([]byte{0x00, 0x40, 0x9c, 0x44})) // YES IT PRINTS 1250
	//math.Float32frombits(binary.LittleEndian.Uint32([]byte{0x00, 0x40, 0x9c, 0x44}))
	if fanKeyEncodingIsFLT {
		setKey("F0Mn", fmt.Sprintf("%x", float32ToBytes(float32(minSpeed))))
		return
	}
	setKey("F0Mn", strconv.FormatInt(int64(minSpeed<<2), 16)) // https://github.com/hholtmann/smcFanControl/tree/master/smc-command
}

func float32ToBytes(f float32) []byte {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], math.Float32bits(f))
	return buf[:]
}


func getFanSpeed() int {
	s, _ := strconv.ParseFloat(getKey("F0Mn"), 64) // get min fan speed (shouldn't change target)
	return int(s)
}

// getTemp returns the value of the SMC's CPU PECI Die filtered and adjusted temp for fan/power control in Celsius
func getTemp() float64 {
	t, _ := strconv.ParseFloat(getKey("TC0F"), 64)
	return t
}

func getKey(key string) string {
	v := run("smc -r -k " + key)                           // v now has the format: "   F0Mn  [fpe2]  2400.00 (bytes 25 80)"
	v = strings.TrimSpace(v[strings.LastIndex(v, "]")+1:]) // cut it till the last bracket and trim so it's now "2400.00 (bytes 25 80)"
	return strings.Fields(v)[0]                            // split by whitespace to get ["2400.00", "(bytes", "25", "80)"] and then return the first value which is what we want
}

func getFanIsFLT() bool {
	v := run("smc -r -k F0Mn")
	return strings.Contains(v, "[flt ]")
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
