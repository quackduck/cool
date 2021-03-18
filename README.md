# Cool

Never let the heat slow your Mac down again. 

Note: Cool is still in beta because I've heard of some possible bugs. Cool is still perfectly safe and can cause no damage (the absolute worst case is that you have to reset the SMC).

Cool is a fan control CLI that lets you cool your CPU to any temperature you'd like. Cool also displays a chart of the temperature changes, a chart of fan speed changes and much more. It reduced the CPU temperature from 97 to 75 in just 1 minute and 10 seconds on a MacBook Air 2017. 

[![asciicast](https://asciinema.org/a/400179.svg)](https://asciinema.org/a/400179?speed=2)

## Usage
```text
Usage: sudo cool [-c/--no-chart] [<temperature>]
       cool [-h/--help | -v/--version]
```

Be careful of commands that require sudo! Cool needs sudo to control fan speeds.

You can specify a temperature to cool your Mac down to:
```shell
sudo cool 57
```
or let Cool choose the default (75 C)
```
sudo cool
```

## FAQ

**Isn't fan control bad for your Mac?**  
Only when done incorrectly. Cool only changes the minimum fan speed; macOS can decide the actual fan speed to set it to. This means that your fan speed will never be below the default. Likewise, the maximum fan speed Cool can set (this is hard-coded) is the maximum safe speed: 6500 RPM. This means that your fan speed is always in safe values!

**How does this work?**  
Cool sets fan speeds, reads fan speeds and reads temperatures using the brilliant [smcFanControl CLI](https://github.com/hholtmann/smcFanControl/tree/master/smc-command). The `smc` binary is the compiled executable. If you're curious, Cool changes the value of the SMC key `F0Mn`. It reads the CPU 1 temperature sensor (`TC0E`).

## Thanks

Thanks to [Sam](https://github.com/sampoder) and [Jubril](https://github.com/s1ntaxe770r) for their help with testing Cool.
