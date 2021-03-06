// +build js,wasm

package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dennwc/dom/js"
	"github.com/zorchenhimer/MovieNight/common"
)

var (
	timestamp bool
	color     string
	auth      common.CommandLevel
)

func getElement(s string) js.Value {
	return js.Get("document").Call("getElementById", s)
}

func join(v []js.Value) {
	color := js.Call("getCookie", "color").String()
	if color == "" {
		// If a color is not set, do a random color
		color = common.RandomColor()
	} else if !common.IsValidColor(color) {
		// Don't show the user the error, just clear the cookie
		common.LogInfof("%#v is not a valid color, clearing cookie", color)
		js.Call("deleteCookie", "color")
	}

	joinData, err := json.Marshal(common.JoinData{
		Name:  getElement("name").Get("value").String(),
		Color: color,
	})
	if err != nil {
		notify("Error prepping data for join")
		common.LogErrorf("Could not prep data: %#v\n", err)
	}

	data, err := json.Marshal(common.ClientData{
		Type:    common.CdJoin,
		Message: string(joinData),
	})
	if err != nil {
		common.LogErrorf("Could not marshal data: %v", err)
	}

	js.Call("websocketSend", string(data))
}

func recieve(v []js.Value) {
	if len(v) == 0 {
		fmt.Println("No data received")
		return
	}

	chatJSON, err := common.DecodeData(v[0].String())
	if err != nil {
		fmt.Printf("Error decoding data: %s\n", err)
		js.Call("appendMessages", fmt.Sprintf("<div>%v</div>", v))
		return
	}

	chat, err := chatJSON.ToData()
	if err != nil {
		fmt.Printf("Error converting ChatDataJSON to ChatData of type %d: %v", chatJSON.Type, err)
	}

	switch chat.Type {
	case common.DTHidden:
		h := chat.Data.(common.HiddenMessage)
		switch h.Type {
		case common.CdUsers:
			names = nil
			for _, i := range h.Data.([]interface{}) {
				names = append(names, i.(string))
			}
			sort.Strings(names)
		case common.CdAuth:
			auth = h.Data.(common.CommandLevel)
		case common.CdColor:
			color = h.Data.(string)
			js.Get("document").Set("cookie", fmt.Sprintf("color=%s; expires=Fri, 31 Dec 9999 23:59:59 GMT", color))
		case common.CdEmote:
			data := h.Data.(map[string]interface{})
			emoteNames = make([]string, 0, len(data))
			emotes = make(map[string]string)
			for k, v := range data {
				emoteNames = append(emoteNames, k)
				emotes[k] = v.(string)
			}
			sort.Strings(emoteNames)
		case common.CdJoin:
			notify("")
			js.Call("openChat")
		case common.CdNotify:
			notify(h.Data.(string))
		}
	case common.DTEvent:
		d := chat.Data.(common.DataEvent)
		// A server message is the only event that doesn't deal with names.
		if d.Event != common.EvServerMessage {
			websocketSend("", common.CdUsers)
		}
		// on join or leave, update list of possible user names
		fallthrough
	case common.DTChat:
		msg := chat.Data.HTML()
		if d, ok := chat.Data.(common.DataMessage); ok {
			if timestamp && (d.Type == common.MsgChat || d.Type == common.MsgAction) {
				h, m, _ := time.Now().Clock()
				msg = fmt.Sprintf(`<span class="time">%02d:%02d</span> %s`, h, m, msg)
			}
		}

		appendMessage(msg)
	case common.DTCommand:
		d := chat.Data.(common.DataCommand)

		switch d.Command {
		case common.CmdPlaying:
			if d.Arguments == nil || len(d.Arguments) == 0 {
				js.Call("setPlaying", "", "")

			} else if len(d.Arguments) == 1 {
				js.Call("setPlaying", d.Arguments[0], "")

			} else if len(d.Arguments) == 2 {
				js.Call("setPlaying", d.Arguments[0], d.Arguments[1])
			}
		case common.CmdRefreshPlayer:
			js.Call("initPlayer", nil)
		case common.CmdPurgeChat:
			js.Call("purgeChat", nil)
			appendMessage(d.HTML())
		case common.CmdHelp:
			url := "/help"
			if d.Arguments != nil && len(d.Arguments) > 0 {
				url = d.Arguments[0]
			}
			appendMessage(d.HTML())
			js.Get("window").Call("open", url, "_blank", "menubar=0,status=0,toolbar=0,width=300,height=600")
		}
	}
}

func appendMessage(msg string) {
	js.Call("appendMessages", "<div>"+msg+"</div>")
}

func websocketSend(msg string, dataType common.ClientDataType) error {
	if strings.TrimSpace(msg) == "" && dataType == common.CdMessage {
		return nil
	}

	data, err := json.Marshal(common.ClientData{
		Type:    dataType,
		Message: msg,
	})
	if err != nil {
		return fmt.Errorf("could not marshal data: %v", err)
	}

	js.Call("websocketSend", string(data))
	return nil
}

func send(this js.Value, v []js.Value) interface{} {
	if len(v) != 1 {
		showChatError(fmt.Errorf("expected 1 parameter, got %d", len(v)))
		return false
	}

	err := websocketSend(v[0].String(), common.CdMessage)
	if err != nil {
		showChatError(err)
		return false
	}
	return true
}

func showChatError(err error) {
	if err != nil {
		fmt.Printf("Could not send: %v\n", err)
		js.Call("appendMessages", `<div><span style="color: red;">Could not send message</span></div>`)
	}
}

func notify(msg string) {
	js.Call("setNotifyBox", msg)
}

func showTimestamp(v []js.Value) {
	if len(v) != 1 {
		// Don't bother with returning a value
		return
	}
	timestamp = v[0].Bool()
}

func isValidColor(this js.Value, v []js.Value) interface{} {
	if len(v) != 1 {
		return false
	}
	return common.IsValidColor(v[0].String())
}

func debugValues(v []js.Value) {
	for k, v := range map[string]interface{}{
		"timestamp":               timestamp,
		"auth":                    auth,
		"color":                   color,
		"current suggestion":      currentSug,
		"current suggestion type": currentSugType,
		"filtered suggestions":    filteredSug,
		"user names":              names,
		"emote names":             emoteNames,
	} {
		fmt.Printf("%s: %#v\n", k, v)
	}
}

func main() {
	common.SetupLogging(common.LLDebug, "")

	js.Set("processMessageKey", js.FuncOf(processMessageKey))
	js.Set("sendMessage", js.FuncOf(send))
	js.Set("isValidColor", js.FuncOf(isValidColor))

	js.Set("recieveMessage", js.CallbackOf(recieve))
	js.Set("processMessage", js.CallbackOf(processMessage))
	js.Set("debugValues", js.CallbackOf(debugValues))
	js.Set("showTimestamp", js.CallbackOf(showTimestamp))
	js.Set("join", js.CallbackOf(join))

	go func() {
		time.Sleep(time.Second * 1)
		inner := `<option value=""></option>`
		for _, c := range common.Colors {
			inner += fmt.Sprintf(`<option value="%s">%s</option>\n`, c, c)
		}

		js.Get("colorSelect").Set("innerHTML", inner)
	}()

	// This is needed so the goroutine does not end
	for {
		// heatbeat to keep connection alive to deal with nginx
		if js.Get("inChat").Bool() {
			websocketSend("", common.CdPing)
		}
		time.Sleep(time.Second * 10)
	}
}
