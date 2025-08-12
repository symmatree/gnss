# geotalker

## Building and Connecting

* To build for the host OS, run `./build.sh`
* To build a Docker image, run `./build_image.sh`

### To run under Windows via WSL

* Run Powershell as admin
* `usbipd list`
* Pick the "USB Serial Device" (should be only one?) and note the first column, `BUSID`
* `usbipd bind --busid 1-2`
* `usbipd attach --wsl --busid 1-2`

From there it's connected and available at `/dev/ttyACM0`

When done, `usbipd wsl detach --busid <busid>`

## Basic design

This utility is intended to talk directly to a u-blox F9R or F9P.

Configuration (read from YAML file):

* Exports: list of dicts
  * Message string id (must be known so we have a parser for it). Note that the messages we export here should agree with what's enabled on the interface in the u-blox unit!
  * Message numeric id (allows unknown messages as long as we don't parse it)
  * Bool: Forward raw binary message
  * Bool: Forward parsed version as json
  * Bool: Retain last message in topic
  * Override binary MQTT topic
  * Override json MQTT topic
  * Bool: Log parsed version (maybe a prettier format than json)
* Incoming GNSS real-time corrections:
  * binary RTCM MQTT topic
  * or JSON RTCM MQTT topic
  * or connection to NTRIP server

Startup sequence:

* Connect to an MQTT server (configured by flags)
* Connect to a u-blox device (configured by flags)
* Register our publish topics as needed
* Subscribe to a flag-configured topic for binary RTCM corrections for the rover (us).

Ongoing listeners:

* When a message is received from u-blox
  * determine its type. If it is one of the configured types:
  * Forward to binary topic (if needed)
  * parse and reformat to JSON (if needed)
  * log and/or send to json topic
* When an RTCM message is received on the incoming RTCM topic, send it to the u-blox device.
* When an RTCM message is received from the NTRIP server, send it to the u-blox device.

## u-blox config

Always a sore spot; I need to find/make some tools to capture a config (from live) at the sort of
semantic level used in teh u-center UI and to diff two states. Basically I want to be able to get
back to stock, and I want to have a nice-enough text format to capture and iterate on config without
needing to connect to a Windows machine just to run u-blox.

For reference if you export settings from u-center it looks like [spark-f9p-config-2024-09-01.txt](https://github.com/symmatree/amateur/blob/main/gnss/ublox/configs/spark-f9p-config-2024-09-01.txt) - row after row of opaque `CFG-VALGET 0A 0B 0C...`

I also talked about this a bunch at https://github.com/symmatree/amateur/blob/main/gnss/ublox/README.md#configuration

For right now I will just document my changes (intended):

| `CFG-USBOUTPROT-NMEA` -> 0 | Simplest way to turn off default nav settings without a good config-capture plan. |
| `CFG-MSGOUT-UBX_MON_COMMS_USB (0x20910352)` -> 60  | num epochs so probably 1m |
| `CFG-MSGOUT-UBX_ESF_ALG_USB (0x20910112)` -> 60 | External Sensor Fusion Auto-alignment status  |
| `CFG-MSGOUT-UBX_ESF_STATUS_USB (0x20910108)` -> 60 | External Sensor Fusion Status: Status for a list of sensors and some Fusion modes |
| `CFG-MSGOUT-UBX_MON_HW_USB (0x209101b7)` -> 60 | Noise level, AGC Monitor, CW Jamming Indicator (plus pin values which we can skip)|
| `CFG-MSGOUT-UBX_MON_IO_USB (0x209101a8)` -> 60 | Send/recv bytes and error counts for each interface |
| `CFG-MSGOUT-UBX_MON_MSGPP_USB (0x20910199)` | Interface X protocol processing counts |
| `CFG-MSGOUT-UBX_MON_SYS_USB (0x209106a0)` | cpu and memory load kinds of things |
| `CFG-MSGOUT-UBX_MON_RXBUF_USB (0x209101a3)` | curr and peak RX buffer by interface |
| `CFG-MSGOUT-UBX_MON_TXBUF_USB (0x2091019e)` | curr and peak RX buffer by interface |
| `CFG-MSGOUT-UBX_NAV_DOP_USB (0x2091003b)` | not every epoch at least |
| Soon `CFG-MSGOUT-UBX_NAV_COV_USB (0x20910086)` | We will end up needing it |
| `CFG-MSGOUT-UBX_NAV_PVAT_USB (0x2091062d)` -> 1 | Primary positioning, also almost everything else |
| `CFG-MSGOUT-UBX_NAV_STATUS_USB (0x2091001d)` -> 60 | just diagnostics |
| `CFG-MSGOUT-UBX_NAV_STATUS_USB (0x2091001d)` -> 120 | I guess |
| `CFG-MSGOUT-UBX_NAV_SIG_USB (0x20910348)` -> 120 | also |
| `CFG-INFMSG-UBX_USB (0x20920004)` -> 0x7 | That's the default for NMEA, so I guess a bitfield  (ERROR=1, WARNING=2, NOTICE=4)|
| `CFG-NAVSPG-DYNMODEL` | Leave it at AUTOMOTIVE actually but this was on my list to change for the mower |
| `CFG-RATE-MEAS (0x30210001)` | Leave it at 1000 for nominal 1 Hz but think about it...|
| `CFG-SFCORE-USE_SF (0x10080001)` -> False | Just to get a baseline |

Assuming I'm somehow getting RTCM corrections onboard and delivered to the device. Plan is to set these to BBR and RAM.
