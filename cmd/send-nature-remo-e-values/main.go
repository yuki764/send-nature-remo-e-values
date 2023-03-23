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

	datadogStatsd := os.Getenv("DATADOG_STATSD")
	datadogNs := os.Getenv("DATADOG_NAMESPACE")
	if datadogNs == "" {
		datadogNs = "home"
	}
	statsdClient, err := statsd.New(datadogStatsd+":8125", statsd.WithNamespace(datadogNs))
	if err != nil {
		panic(err)
	}

	pubsubTopic := os.Getenv("PUBSUB_TOPIC")

	for {
		go fetchEnergyValuesFromNatureAPI(applienceId, token, statsdClient, pubsubTopic)

		time.Sleep(60 * time.Second)
	}
}

func fetchEnergyValuesFromNatureAPI(applienceId string, token string, statsdClient *statsd.Client, pubsubTopic string) {
	req, err := http.NewRequest("GET", "https://api.nature.global/1/appliances", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

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

	fmt.Printf("%#v\n", energy)

	if statsdClient != nil {
		go func() {
			// send Instantaneous
			instTs, err := time.Parse(time.RFC3339, energy.InstantaneousTimestamp)
			if err != nil {
				panic(err)
			}
			if err := statsdClient.GaugeWithTimestamp("nature_remo.electric_energy.instantaneous", float64(energy.InstantaneousValue), []string{"home:Home"}, 1, instTs); err != nil {
				panic(err)
			}
			// send (Normal) Cumulative to Datadog
			ncumTs, err := time.Parse(time.RFC3339, energy.NormalCumulativeTimestamp)
			if err != nil {
				panic(err)
			}
			if err := statsdClient.GaugeWithTimestamp("nature_remo.electric_energy.cumulative", float64(energy.NormalCumulativeValue), []string{"home:Home"}, 1, ncumTs); err != nil {
				panic(err)
			}
		}()
	}
}
