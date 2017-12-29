package main

import (
    "time"
	"bytes"
	"encoding/json"
	"github.com/marcusolsson/tui-go"
	"io/ioutil"
	"net/http"
	"regexp"
)

type DeviceInfo struct {
	Id           string           `json:"entity_id"`
	State        string           `json:"state"`
	Last_changed string           `json:"last_changed"`
	Last_updated string           `json:"last_updated"`
	Attributes   DeviceAttributes `json:"attributes"`
}

type DeviceAttributes struct {
	Manufacturer      string `json:"device_manufacturer"`
	FriendlyName      string `json:"friendly_name"`
	DeviceModel       string `json:"manufacturer_device_model"`
	ModelName         string `json:"model_name"`
	SupportedFeatures int    `json:"supported_features"`
}

type BaseResult struct {
	errorMessage string
}

type DeviceResult struct {
	BaseResult
	Devices []DeviceInfo
}

func perror(err error) {
	if err != nil {
		panic(err)
	}
}

func (f *DeviceResult) UnmarshalJSON(bs []byte) error {
	return json.Unmarshal(bs, &f.Devices)
}

var lights = []DeviceInfo{}

var deviceTable *tui.Table

func getDeviceState() *DeviceResult {
	url := "http://home-assistant.lan.productionservers.net:8123/api/states"

	res, err := http.Get(url)
	perror(err)

	body, err := ioutil.ReadAll(res.Body)
	perror(err)

	//fmt.Printf("Unparsed: %s\n\n", string(body[:]))

	keys := DeviceResult{}
	//fmt.Printf("Byte Array: %#v\n\n", body[:])
	json.Unmarshal(body[:], &keys)
	//fmt.Printf("Parsed: %#v\n", keys)
	return &keys
}

func toggleLight(d int, t *tui.Table) {
    l := lights[t.Selected()]
	url := ""
	isOn, _ := regexp.MatchString("^on$", l.State)
	if isOn {
		url = "http://home-assistant.lan.productionservers.net:8123/api/services/light/turn_off"
		l.State = "off"
	} else {
		url = "http://home-assistant.lan.productionservers.net:8123/api/services/light/turn_on"
		l.State = "on"
	}

	type SimpleJson struct {
		Id string `json:"entity_id"`
	}

	j := SimpleJson{
		Id: l.Id,
	}

	b, _ := json.Marshal(&j)
    //fmt.Printf("Whatever: %s\n", t[d])
	//fmt.Printf("Json: %s\n", b)
	r := bytes.NewReader(b)

	req, _ := http.NewRequest("POST", url, r)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	perror(err)
	resp.Body.Close()
    //refreshScreen()

	//body, _ := ioutil.ReadAll(resp.Body)
	//fmt.Printf("%s\n", body)
}

func getScreen() *tui.Box {
	deviceTable = tui.NewTable(0, 0)
	deviceTable.SetColumnStretch(0, 0)
	deviceTable.SetColumnStretch(1, 0)

	devices := getDeviceState()
	for devicePos := range devices.Devices {
		device := devices.Devices[devicePos]
		isLight, _ := regexp.MatchString("(^on$|^off$)", device.State)
		if isLight {
			//			fmt.Printf("ID: %s\n", device.Id)
			//			fmt.Printf("State: %s\n", device.State)
			//			fmt.Printf("Friendly Name: %s\n", device.Attributes.FriendlyName)
			//			fmt.Printf("\n")
			lights = append(lights, device)
			deviceTable.AppendRow(
				tui.NewLabel(device.Attributes.FriendlyName),
				tui.NewLabel(device.State),
                tui.NewLabel(device.Attributes.ModelName),
			)
		}
	}

	var (
		devicename   = tui.NewLabel("")
		manufacturer = tui.NewLabel("")
		modelname    = tui.NewLabel("")
	)

	deviceTable.OnSelectionChanged(func(t *tui.Table) {
		l := lights[t.Selected()]
		devicename.SetText(l.Attributes.FriendlyName)
		manufacturer.SetText(l.Attributes.Manufacturer)
		modelname.SetText(l.Attributes.ModelName)
	})

	deviceTable.Select(0)
	deviceTable.OnItemActivated(func(t *tui.Table) {
		//l := lights[t.Selected()]
		toggleLight(t.Selected(), t)
	})

	root := tui.NewVBox(deviceTable, tui.NewLabel(""))
    go startRefreshing()
	return root
}

func refreshScreen() {
    mySelected := deviceTable.Selected()
    deviceTable.RemoveRows()
    devices := getDeviceState()
    lights = []DeviceInfo{}
	for devicePos := range devices.Devices {
		device := devices.Devices[devicePos]
		isLight, _ := regexp.MatchString("(^on$|^off$)", device.State)
		if isLight {
			//			fmt.Printf("ID: %s\n", device.Id)
			//			fmt.Printf("State: %s\n", device.State)
			//			fmt.Printf("Friendly Name: %s\n", device.Attributes.FriendlyName)
			//			fmt.Printf("\n")
			lights = append(lights, device)
			deviceTable.AppendRow(
				tui.NewLabel(device.Attributes.FriendlyName),
				tui.NewLabel(device.State),
                tui.NewLabel(device.Attributes.ModelName),
			)
		}
	}
    deviceTable.SetSelected(mySelected)
}

func startRefreshing() {
    for {
        time.Sleep(10 * time.Second)
        refreshScreen()
    }
}

func main() {
	root := getScreen()
	ui := tui.New(root)
	ui.SetKeybinding("Esc", func() { ui.Quit() })
	ui.SetKeybinding("Shift+Alt+Up", func() { ui.Quit() })
	ui.SetKeybinding("q", func() { ui.Quit() })
    ui.SetKeybinding("r", func () {
        refreshScreen()
    })

	if err := ui.Run(); err != nil {
		panic(err)
	}
}
