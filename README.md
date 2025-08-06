## Fork

This codebase started as a fork of the M8 support in github.com/daedaleanai/ublox which appears to be a
dead project. I forked it "manually" via a tarball after the third time VSCode tried to push to, or open
a PR or Issue against, the upstream repo instead of my own.

# Original README

# ublox

Encoding and decoding of μ-Blox UBX and NMEA messages

[![GoDoc](https://godoc.org/github.com/daedaleanai/ublox?status.svg)](https://godoc.org/github.com/daedaleanai/ublox)

This Go package implements encoders and decoders for the NMEA and UBX messages defined in
[u-blox 8 / u-blox M8 Receiver description](https://www.u-blox.com/en/docs/UBX-13003221) chapters 31 and 32.
