# lgtv-sdp

[![Go Report Card](https://goreportcard.com/badge/github.com/dsheets/lgtv-sdp)](https://goreportcard.com/report/github.com/dsheets/lgtv-sdp)

**tl;dr**: LG 'smart' TV and projector devices want to phone home to get
the date and time and this daemon fixes these defective devices.

You just got a fancy new display from LG! Congratulations!

Unfortunately, when you decided to make it behave itself by
firewalling it, it stopped automatically updating its time and
date. Maybe your display is the amazing HU-80 CineBeam portable 4K
laster projector and its time and date get thrown away every time you
store it. Or maybe you have a TV on a smart plug. Or maybe your power
is intermittent. Or maybe you just never want to have to worry about
date and time on a bloody 'smart' display device. It's so nice to just
have the YouTube app on the device work properly... so easy. Except
Google serves TLS certificates with *NotValidBefore* set so the YouTube
app is unusable until you set the date and time manually. :-( And the
device doesn't listen to the NTP server setting that it itself
requests via DHCP. :-(

Luckily, no certificate checking is done on the initial phone home
request it makes and the server response contains an HTTP header with
the current time in milliseconds since the UNIX epoch. Whew!

## Installation for the impatient

1. Download correct build for your platform.
2. Unpack archive (it will create a subdirectory for you) where you
want to 'install' the daemon.
3. `cd` into the installation directory and run `sudo ./lgtv-sdp -s
install BIND_ADDRESS` where `BIND_ADDRESS` is the IP you want the
daemon to listen on or `0.0.0.0` or `::` for all addresses.
4. Redirect DNS for `*.lgtvsdp.com` to an address where the daemon is listening.
5. Enjoy _your_ LG display with the correct time!

## Building

1. Ensure go 1.14 is installed.
2. Clone this repo.
3. `go build` or `GOOS=TARGET_OS GOARCH=TARGET_ARCH go build`

## How it works

### TLS

The daemon will check if `key.pem` and `cert.pem` files are present in
its directory and, if they aren't, will create new certificate
authority (CA) certificate and a domain certificate and sign the
domain certificate with the CA key. The daemon will then start an
HTTPS server on the standard port 443.

### API

The HTTPS server will respond with the contents of `initservices.json`
with the headers in the files under `initservices.headers/` to
requests for any paths. The `X-Server-Time` header will always be set
to the number of milliseconds since the UNIX epoch. You may be able to
customize more aspects of your device(s) by sending specific headers
or JSON content but none of that is necessary to set the time. Home
screen widgets and further request behavior are both probably
controllable with specific responses. Your contributions are very welcome.

## Help wanted

We need:

1. A Windows service tester
2. LG webOS 4 and 5 testers
3. A free world

## Contributing

Patches, documentation, cool features, and more are
welcome. Negativity and lawsuits are not welcome.