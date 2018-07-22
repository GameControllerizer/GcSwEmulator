package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"strconv"
	"io"

	ROBOT "github.com/go-vgo/robotgo"
	SET "github.com/deckarep/golang-set"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	WS "golang.org/x/net/websocket"
)

type GcWordMouse struct {
	Btn []int  `json:"btn"`
	Mov [2]int `json:"mov"`
	Dur int    `json:"dur"`
}

type GcWordKeyboard struct {
	Key []string `json:"key"`
	Mod []int    `json:"mod"`
	Dur int      `json:"dur"`
}

/**
 * CmdReceiver Interface
 */
type CmdReceiver interface {
	Start(chan<- GcWordMouse, chan<- GcWordKeyboard)
	Stop()
}

/**
 * MQTT CmdReceiver
 */
type MqttReceiver struct {
	MqttClient MQTT.Client
	Topic string
}

func NewMqttReceiver(aHost string, aPort int64, aTopic string) *MqttReceiver {
	r := &MqttReceiver{}
	
	tBrokerAddr := fmt.Sprintf("tcp://%s:%d", aHost, aPort)
	tOpts := MQTT.NewClientOptions()
	tOpts.AddBroker(tBrokerAddr)
	tOpts.SetOnConnectHandler(func(_ MQTT.Client) {
		fmt.Println("[*] MQTT client is ONLINE")
	})
	r.MqttClient = MQTT.NewClient(tOpts)
	if tToken := r.MqttClient.Connect(); tToken.Wait() && tToken.Error() != nil {
		panic("MQTT connection failure")
	}
	r.Topic = aTopic
	return r
}

func (r *MqttReceiver) Start(aChMouse chan<- GcWordMouse, aChKeyboard chan<- GcWordKeyboard) {

	tTokenMouse := r.MqttClient.Subscribe(
		r.Topic + "/mouse",
		0,
		func(_ MQTT.Client, aMsg MQTT.Message) {
			tPayload := aMsg.Payload()
			var tGcSentence []GcWordMouse
			if tError := json.Unmarshal(tPayload, &tGcSentence); tError != nil {
				fmt.Println("[!] Illegal mouse command : ", tError)
			}
			for _, c := range tGcSentence {
				aChMouse <- c
			}
		})
	if tTokenMouse.Wait() && tTokenMouse.Error() != nil {
		fmt.Println(tTokenMouse.Error())
	}

	tTokenKeyboard := r.MqttClient.Subscribe(
		r.Topic + "/keyboard",
		0,
		func(_ MQTT.Client, aMsg MQTT.Message) {
			tPayload := aMsg.Payload()
			var tGcSentence []GcWordKeyboard
			if tError := json.Unmarshal(tPayload, &tGcSentence); tError != nil {
				fmt.Println("[!] Illegal keyboard command : ", tError)
			}
			for _, c := range tGcSentence {
				aChKeyboard <- c
			}
		})
	if tTokenKeyboard.Wait() && tTokenKeyboard.Error() != nil {
		fmt.Println(tTokenKeyboard.Error())
	}
}

func (r *MqttReceiver) Stop() {
	r.MqttClient.Disconnect(0)
}

/**
 * Ws Client
 */
type WsReceiver struct {
	WsMouse *WS.Conn
	WsKeyboard *WS.Conn
}

func NewWsReceiver(aHost string, aPort int64, aTopic string) *WsReceiver {
	r := &WsReceiver{}

	// inner function
	newWebsocket := func(aDev string) *WS.Conn {
		tAddr := fmt.Sprintf("%s:%d/%s/%s", aHost, aPort, aTopic, aDev)
		tWsOriginalAddr := fmt.Sprintf("http://%s", tAddr)
		tWsAddr := fmt.Sprintf("ws://%s", tAddr)
	
		tWs, tErr := WS.Dial(tWsAddr, "", tWsOriginalAddr)
		if tErr != nil {
			panic("Websocket connection failure : " + aDev)
		}
		fmt.Println("[*] Websocket is ONLINE : " + aDev)
		return tWs
	}

	r.WsMouse = newWebsocket("mouse")
	r.WsKeyboard = newWebsocket("keyboard")
	return r
}

func (r *WsReceiver) Start(aChMouse chan<- GcWordMouse, aChKeyboard chan<- GcWordKeyboard) {

	go func() {
		for {
			var tGcSentence []GcWordMouse
			if tError := WS.JSON.Receive(r.WsMouse, &tGcSentence); tError != nil {
				if (tError == io.EOF){
					panic("Websocket connection lost")
				}
				fmt.Println(" [!] Illegal mouse command : ", tError)
			}
			for _, c := range tGcSentence {
				aChMouse <- c
			}
		}
	}()

	go func() {
		for {
			var tGcSentence []GcWordKeyboard
			if tError := WS.JSON.Receive(r.WsKeyboard, &tGcSentence); tError != nil {				
				if (tError == io.EOF){
					panic("Websocket connection lost")
				}
				fmt.Println("[!] Illegal keyboard command : ", tError)
			}
			for _, c := range tGcSentence {
				aChKeyboard <- c
			}
		}
	}()
}

func (r *WsReceiver) Stop() {
	r.WsMouse.Close()
	r.WsKeyboard.Close()
}

/**
 * Constant
 */
const FRAME_CYCLE_MS = (1000 / 60.0)

var MAP_BTN = [3]string{"left", "right", "center"}
var MAP_MOD = [3]string{"control", "shift", "alt"}
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

	var tHost = flag.String("host", "127.0.0.1", "MQTT/WS host")
	var tPort = flag.Int64("port", 1883, "MQTT/WS port")
	var tTopic = flag.String("topic", "dev", "MQTT/WS topic")
	var tMode = flag.Int64("mode", 0, "0:MQTT, 1:Websocket")
	flag.Parse()

	fmt.Printf("[*] GcSwEmulator\n")
	fmt.Printf(" - MQTT/WS host  : %v\n", *tHost)
	fmt.Printf(" - MQTT/WS port  : %v\n", *tPort)
	fmt.Printf(" - MQTT/WS topic : '%v/mouse'\n", *tTopic)
	fmt.Printf("                   '%v/keyboard'\n", *tTopic)
	fmt.Printf(" - MQTT/WS mode  : %v (0:MQTT, 1:Websocket)\n", *tMode)

	tChMouse := make(chan GcWordMouse, 32)
	tChKeyboard := make(chan GcWordKeyboard, 32)

	go func() {
		tPrevBtn := SET.NewSet()
		for {
			tGcWord := <-tChMouse
			fmt.Println(" - mouse : ", tGcWord)
			tLatestBtn := SET.NewSet()
			for _, id := range tGcWord.Btn {
				tLatestBtn.Add(MAP_BTN[id])
			}
			// get changed buttons
			tUppedBtn := tPrevBtn.Difference(tLatestBtn)
			for b := range tUppedBtn.Iter() {
				ROBOT.MouseToggle("up", b)
			}
			tDownedBtn := tLatestBtn.Difference(tPrevBtn)
			for b := range tDownedBtn.Iter() {
				ROBOT.MouseToggle("down", b)
			}
			if tGcWord.Mov[0] != 0 || tGcWord.Mov[1] != 0 {
				tX, tY := ROBOT.GetMousePos()
				tNewX := tX + tGcWord.Mov[0]
				tNewY := tY + tGcWord.Mov[1]
				ROBOT.MoveMouse(tNewX, tNewY)
			}
			tPrevBtn = tLatestBtn
			if tGcWord.Dur > 0 {
				tDur := time.Duration(float32(tGcWord.Dur) * FRAME_CYCLE_MS)
				time.Sleep(tDur * time.Millisecond)
				if len(tChMouse) == 0 {
					for b := range tDownedBtn.Iter() {
						ROBOT.MouseToggle("up", b)
					}
					for b := range tPrevBtn.Iter() {
						ROBOT.MouseToggle("up", b)
					}
					tPrevBtn = SET.NewSet()
				}
			}
		}
	}()

	go func() {
		tPrevKey := SET.NewSet()
		for {
			tGcWord := <-tChKeyboard
			fmt.Println(" - keyboard : ", tGcWord)

			tLatestKey := SET.NewSet()
			for _, keyname := range tGcWord.Key {
				if k := MAP_KEY[keyname]; k != "" {
					tLatestKey.Add(k)
				}
			}
			for _, id := range tGcWord.Mod {
				tLatestKey.Add(MAP_MOD[id])
			}

			// get changed key
			tUppedKey := tPrevKey.Difference(tLatestKey)
			for k := range tUppedKey.Iter() {
				ROBOT.KeyToggle(k.(string), "up")
			}
			tDownedKey := tLatestKey.Difference(tPrevKey)
			for k := range tDownedKey.Iter() {
				ROBOT.KeyToggle(k.(string), "down")
			}

			tPrevKey = tLatestKey
			if tGcWord.Dur > 0 {
				tDur := time.Duration(float32(tGcWord.Dur) * FRAME_CYCLE_MS)
				time.Sleep(tDur * time.Millisecond)
				if len(tChKeyboard) == 0 {
					for k := range tDownedKey.Iter() {
						ROBOT.KeyToggle(k.(string), "up")
					}
					for k := range tPrevKey.Iter() {
						ROBOT.KeyToggle(k.(string), "up")
					}
					tPrevKey = SET.NewSet()
				}
			}
		}
	}()

	var tReceiver CmdReceiver
	switch *tMode {
	case 0:
		tReceiver = NewMqttReceiver(*tHost, *tPort, *tTopic)
	case 1:
		tReceiver = NewWsReceiver(*tHost, *tPort, *tTopic)
	default:
		panic("Unknown mode : " + strconv.FormatInt(*tMode, 10))
	}
	tReceiver.Start(tChMouse, tChKeyboard)

	// quit program
	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

	close(tChMouse)
	close(tChKeyboard)
	tReceiver.Stop()
}
