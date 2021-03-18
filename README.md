# Cool

Never let the heat slow your Mac down again. 

Cool is a fan control CLI that lets you cool your CPU to any temperature you'd like. Cool also displays a chart of the temperature changes, a chart of fan speed changes and much more. It reduced the CPU temperature from 9 

[![asciicast](https://asciinema.org/a/n6Ewg4TYht3E3JHdjJSKoJY7G.svg)](https://asciinema.org/a/n6Ewg4TYht3E3JHdjJSKoJY7G)

## FAQ

**Isn't fan control bad for your Mac?**  
Only when done incorrectly. Cool only changes the minimum fan speed; macOS can decide the actual fan speed to set it to. This means that your fan speed will never be below the default. Likewise, the maximum fan speed Cool can set (this is hard-coded) is the maximum safe speed: 6500 RPM. This means that your fan speed is always in safe values!

**How does this work?**  
Cool sets fan speeds, reads fan speeds and reads temperatures using the brilliant [smcFanControl CLI](https://github.com/hholtmann/smcFanControl/tree/master/smc-command). The `smc` binary is the compiled executable.

## Thanks

Thanks to [Sam](https://github.com/sampoder) and [Jubril](https://github.com/s1ntaxe770r) for their help with testing Cool.
