package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-vgo/robotgo"
)

type GcimlMouse struct {
	Btn []int  `json:"btn"`
	Mov [2]int `json:"mov"`
	Dur int    `json:"dur"`
}

type GcimlKeyboard struct {
	Key []string `json:"key"`
	Mod []int    `json:"mod"`
	Dur int      `json:"dur"`
}

/**
 * MQTT Client
 */
type MqttSub struct {
	mMqttClient MQTT.Client
}

func NewMqttClient(aHost string, aPort int64, aId string) *MqttSub {
	s := &MqttSub{}

	tBrokerAddr := fmt.Sprintf("tcp://%s:%d", aHost, aPort)
	tOpts := MQTT.NewClientOptions()
	tOpts.AddBroker(tBrokerAddr)
	tOpts.SetClientID(aId)
	tOpts.SetOnConnectHandler(func(_ MQTT.Client) {
		fmt.Println("[*] MQTT client is ONLINE")
	})

	s.mMqttClient = MQTT.NewClient(tOpts)
	if tToken := s.mMqttClient.Connect(); tToken.Wait() && tToken.Error() != nil {
		panic(tToken.Error())
	}
	return s
}

func (s *MqttSub) Start(
	aTopic string,
	aChMouse chan<- GcimlMouse, aChKeyboard chan<- GcimlKeyboard) {

	tToken := s.mMqttClient.Subscribe(
		aTopic,
		0,
		func(_ MQTT.Client, aMsg MQTT.Message) {
			tTopicHierarchy := strings.Split(aMsg.Topic(), "/")
			tDevice := tTopicHierarchy[len(tTopicHierarchy)-1]
			tPayload := aMsg.Payload()

			switch tDevice {
			case "mouse":
				var tGcimlList []GcimlMouse
				if tError := json.Unmarshal(tPayload, &tGcimlList); tError != nil {
					fmt.Println(" [!]", tError)
				}
				for _, c := range tGcimlList {
					aChMouse <- c
				}
			case "keyboard":
				var tGcimlList []GcimlKeyboard
				if tError := json.Unmarshal(tPayload, &tGcimlList); tError != nil {
					fmt.Println(" [!]", tError)
				}
				for _, c := range tGcimlList {
					aChKeyboard <- c
				}
			}
		})
	if tToken.Wait() && tToken.Error() != nil {
		panic(tToken.Error())
	}
}

func (s *MqttSub) Stop() {
	s.mMqttClient.Disconnect(0)
}

/**
 * Constant
 */
const FRAME_CYCLE_MS = (1000 / 60.0)

var MAP_KEY = map[string]string{
	"a":          "a",
	"b":          "b",
	"c":          "c",
	"d":          "d",
	"e":          "e",
	"f":          "f",
	"g":          "g",
	"h":          "h",
	"i":          "i",
	"j":          "j",
	"k":          "k",
	"l":          "l",
	"m":          "m",
	"n":          "n",
	"o":          "o",
	"p":          "p",
	"q":          "q",
	"r":          "r",
	"s":          "s",
	"t":          "t",
	"u":          "u",
	"v":          "v",
	"w":          "w",
	"x":          "x",
	"y":          "y",
	"z":          "z",
	"0":          "0",
	"1":          "1",
	"2":          "2",
	"3":          "3",
	"4":          "4",
	"5":          "5",
	"6":          "6",
	"7":          "7",
	"8":          "8",
	"9":          "9",
	"F1":         "f1",
	"F2":         "f2",
	"F3":         "f3",
	"F4":         "f4",
	"F5":         "f5",
	"F6":         "f6",
	"F7":         "f7",
	"F8":         "f8",
	"F9":         "f9",
	"F10":        "f10",
	"F11":        "f11",
	"F12":        "f12",
	"Escape":     "escape",
	"Space":      "space",
	"Tab":        "tab",
	"Enter":      "enter",
	"Backspace":  "backspace",
	"Delete":     "delete",
	"ArrowUp":    "up",
	"ArrowDown":  "down",
	"ArrowRight": "right",
	"ArrowLeft":  "left",
}

/**
 * Main
 */
func main() {

	var tMqttHost = flag.String("mh", "127.0.0.1", "MQTT host")
	var tMqttPort = flag.Int64("mp", 1883, "MQTT port")
	var tMqttTopic = flag.String("mt", "#", "MQTT port")
	var tMqttId = flag.String("mi", "GcSwEmulator", "MQTT client id")
	flag.Parse()

	fmt.Printf("[*] GcSwEmulator\n")
	fmt.Printf(" - MQTT host : %v\n", *tMqttHost)
	fmt.Printf(" - MQTT port : %v\n", *tMqttPort)
	fmt.Printf(" - MQTT topic : '%v/mouse'\n", *tMqttTopic)
	fmt.Printf("                '%v/keyboard'\n", *tMqttTopic)

	tChMouse := make(chan GcimlMouse, 32)
	tChKeyboard := make(chan GcimlKeyboard, 32)

	go func() {
		var tPrevBtn int = 0
		tBtnToggle := [2]string{"up", "down"}
		tBtnName := [3]string{"left", "right", "center"}
		for {
			tGciml := <-tChMouse
			fmt.Println(" - mouse : ", tGciml)
			var tLatestBtn = 0x00
			for _, id := range tGciml.Btn {
				tLatestBtn |= 0x01 << uint(id)
			}
			tBtnChange := tPrevBtn ^ tLatestBtn
			for i := uint(0); i < 3; i++ {
				tBtnMask := 0x01 << i
				if tBtnChange&tBtnMask > 0 {
					robotgo.MouseToggle(tBtnToggle[tLatestBtn&tBtnMask], tBtnName[i])
				}
			}
			if tGciml.Mov[0] != 0 || tGciml.Mov[1] != 0 {
				tX, tY := robotgo.GetMousePos()
				tNewX := tX + tGciml.Mov[0]
				tNewY := tY + tGciml.Mov[1]
				robotgo.MoveMouse(tNewX, tNewY)
			}
			tPrevBtn = tLatestBtn
			if tGciml.Dur > 0 {
				tDur := time.Duration(float32(tGciml.Dur) * FRAME_CYCLE_MS)
				time.Sleep(tDur * time.Millisecond)
			}
		}
	}()

	go func() {
		tModName := []string{"control", "shift", "alt"}
		for {
			tGciml := <-tChKeyboard
			fmt.Println(" - keyboard : ", tGciml)

			tMod := []string{}
			for _, id := range tGciml.Mod {
				tMod = append(tMod, tModName[id])
			}
			for _, keyname := range tGciml.Key {
				if k := MAP_KEY[keyname]; k != "" {
					if len(tMod) > 0 {
						robotgo.KeyTap(k, tMod)
					} else {
						robotgo.KeyTap(k)
					}
				}
			}
			if tGciml.Dur > 0 {
				tDur := time.Duration(float32(tGciml.Dur) * FRAME_CYCLE_MS)
				time.Sleep(tDur * time.Millisecond)
			}
		}
	}()

	tMqttSub := NewMqttClient(*tMqttHost, *tMqttPort, *tMqttId)
	tMqttSub.Start(*tMqttTopic, tChMouse, tChKeyboard)

	// quit program
	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

	close(tChMouse)
	close(tChKeyboard)
	tMqttSub.Stop()
}
