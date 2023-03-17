package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"

	"neigepluie.net/send-nature-remo-e-values/pkg/natureRemoE"
)

func main() {
	applienceId := os.Getenv("APPLIENCE_ID")
	token := os.Getenv("NATURE_API_TOKEN")

	for {
		go fetchEnergyValuesFromNatureAPI(applienceId, token)

		time.Sleep(60 * time.Second)
	}
}

func fetchEnergyValuesFromNatureAPI(applienceId string, token string) {
	req, err := http.NewRequest("GET", "https://api.nature.global/1/appliances", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	//println(resp.Status)

	decoder := json.NewDecoder(resp.Body)

	var allAppliences []natureRemoE.Applience
	if err := decoder.Decode(&allAppliences); err != nil {
		panic(err)
	}

	var energy natureRemoE.Energy

	for _, a := range allAppliences {
		if a.Id == applienceId {
			energy, err = natureRemoE.ParseEnergy(a)
			if err != nil {
				panic(err)
			}
		}
	}

	statsdClient, err := statsd.New("127.0.0.1:8125", statsd.WithNamespace("demo_home_smartmeter"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", energy)
	if err := statsdClient.Gauge("demo_home_smartmeter.energy.instantaneous", float64(energy.Instantaneous), []string{"environment:demo_home"}, 1); err != nil {
		panic(err)
	}
}
