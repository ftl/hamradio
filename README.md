# HAMRADIO Tools
This is a collection of libaries and tools I use to do HF QSOs as DL3NEY:

* calculate distance and azimuth between locations (given as lat/lon or maidenhead locator): [locator](./cmd/locator) and [latlon](./cmd/latlon)
* find DXCC information about radio callsign prefixes: [dxcc](./cmd/dxcc)
* retrieve information about a radio callsign from [HamQTH.com](https://hamqth.com) and [QRZ.com](https://qrz.com): [callbook](./cmd/callbook)
* use the callsign database from [Super Check Partial](http://www.supercheckpartial.com): [supercheck](./cmd/supercheck)
* talk to the [cwdaemon](https://github.com/acerion/cwdaemon) to output CW on your transceiver: [cw](./cmd/cw)
* more to come as I have time and need

The tools are written Go on Linux. They might also work on OsX or Windows, but I did not try that out.

## Disclaimer
I develop these tools for myself and just for fun in my free time. If you find it useful, I'm happy to hear about that. If you have trouble using it, you have all the source code to fix the problem yourself (although pull requests are welcome).

## Links
* [Wiki](https://github.com/ftl/hamradio/wiki)

## License
This tool is published under the [MIT License](https://www.tldrlegal.com/l/mit).

Copyright [Florian Thienel](http://thecodingflow.com/)
