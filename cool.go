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
	_ "github.com/guptarohit/asciigraph"
)

var (
	version = "dev"
	helpMsg = ``
)

func main() {
	if len(os.Args) == 1 {
		handleErrStr("too few arguments")
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
	temp, err := strconv.Atoi(os.Args[1])
	if err != nil {
		handleErr(err)
		return
	}
	cool(temp)
}

func cool(target int) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		<-c
		setFanSpeed(1200) // default min fan speed
		os.Exit(0)
	}()
	t := getTemp()
	var s int

	setFanSpeed(1200 + 100*(int(t)-target)) // quickly set it initally
	for ; ; time.Sleep(time.Second * 5) {   // fine tune
		s = getFanSpeed()
		t = getTemp()
		fmt.Printf("%.1f C %8v RPM\n", t, s)
		setFanSpeed(s + 10*(int(t)-target)) // set current to current + 10 times the difference in temps. This will automatically correct when temp is too low.
	}
	//// at this point we've made it cool enough, so reduce noise and try to lower rpm
	//for t := getTemp(); t < float64(target); t = getTemp() {
	//	fmt.Printf("%.1f C %8v RPM\n", t, s)
	//	if s < 6500 {
	//		setFanSpeed(s - 50)
	//	}
	//	s = getFanSpeed()
	//	time.Sleep(time.Second * 5)
	//}
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

// getTemp returns the value of the SMC's CPU PECI sensor (Celsius)
func getTemp() float64 {
	t, _ := strconv.ParseFloat(getKey("TCXC"), 64) // PECI is especially for fan control https://en.wikipedia.org/wiki/Platform_Environment_Control_Interface. However, this seems to report the actual temp.
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
