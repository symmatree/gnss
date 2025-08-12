package main

import (
	"github.com/symmatree/gnss/ubx"
	"log"
	"os"
)

func main() {

	// disable all NMEA on all ports
	for _, msg := range []ubx.Message{
		ubx.CfgMsg2{MsgClass: 0xF0, MsgID: 0x0},         // GGA
		ubx.CfgMsg2{MsgClass: 0xF0, MsgID: 0x1},         // GLL
		ubx.CfgMsg2{MsgClass: 0xF0, MsgID: 0x2},         // GSA
		ubx.CfgMsg2{MsgClass: 0xF0, MsgID: 0x3},         // GSV
		ubx.CfgMsg2{MsgClass: 0xF0, MsgID: 0x4},         // RMC
		ubx.CfgMsg2{MsgClass: 0xF0, MsgID: 0x5},         // VTG
		ubx.CfgMsg2{MsgClass: 0xF0, MsgID: 0x6},         // GRS
		ubx.CfgMsg2{MsgClass: 0xF0, MsgID: 0x7},         // GST
		ubx.CfgMsg2{MsgClass: 0xF0, MsgID: 0x8},         // ZDA
		ubx.CfgMsg2{MsgClass: 0xF0, MsgID: 0x9},         // GBS
		ubx.CfgMsg2{MsgClass: 0xF0, MsgID: 0xD},         // GNS
		ubx.CfgMsg2{MsgClass: 0xF0, MsgID: 0xF},         // VLW
		ubx.CfgRate{MeasRate_ms: 62, NavRate_cycles: 1}, // 62ms (16Hz) measurement
		ubx.CfgMsg1{MsgClass: 0x1, MsgID: 0x7, Rate: 1}, // NAV-PVT
	} {
		b, err := ubx.Encode(msg)
		if err != nil {
			log.Fatal(err)
		}
		os.Stdout.Write(b)
	}
}
