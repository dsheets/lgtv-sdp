# lgtv-sdp

<a href="https://goreportcard.com/report/dsheets/lgtv-sdp">
  <img src="https://goreportcard.com/badge/github.com/dsheets/lgtv-sdp" alt="Go Report Card" />
</a>

tl;dr: LG 'smart' TV and projector devices want to phone home to get
the date and time and this daemon fixes these defective devices.

You just got a fancy new display from LG! Congratulations!

Unfortunately, when you decided to make it behave itself by
firewalling it, it stopped automatically updating its time and
date. Maybe your display is the amazing HU-80 CineBeam portable 4K
laster projector and its time and date get thrown away every time you
put it away. Or maybe you have a TV on a smart plug. Or maybe you just
never want to have to worry about date and time on a bloody 'smart'
display device. It's so nice to just have the YouTube app on the
device work properly... so easy. Except Google serves TLS certificates
with NotValidBefore set so the YouTube app is unusable until you set
the date and time manually. :-( And the device doesn't listen to the
NTP server setting that it itself requests via DHCP. :-(

Luckily, no certificate checking is done on the initial phone home
request it makes and the server response contains an HTTP header with
the current time in milliseconds since the UNIX epoch. Whew! Some
lovely LG engineer or team lead has some ethics even if the parent
corporation holds their customers in contempt.

## Installation for the impatient

1. Download correct build for your platform.
2. Copy binary to the file system location on the server that you want
to 'install' it.
3. Run `./lgtv-sdp -s install BIND_ADDRESS` where `BIND_ADDRESS` is
the IP you want the daemon to listen on or `0.0.0.0` or `::` for all
addresses.
4. Redirect DNS for *.lgtvsdp.com to an address where the daemon is listening.
5. Enjoy _your_ LG display!

## Building

1. Ensure go 1.14 is installed.
2. Clone this repo.
3. `go build` or `GOOS=TARGET_OS GOARCH=TARGET_ARCH go build`

## Help wanted

We need:

1. A Windows service tester
2. LG webOS 4 and 5 testers
3. A free world

## Contributing

Patches, documentation, cool features, and more are
welcome. Negativity and lawsuits are not welcome.