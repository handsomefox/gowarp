package warpgen

import (
	"encoding/json"
	"fmt"
	"gowarp/pkg/config"
	"gowarp/pkg/progressbar"
	"math/rand"
	"net/http"
	"time"
)

func init() {
	go refillStash()
	// background task for refilling as well
	go func() {
		time.Sleep(2 * time.Hour)
		refillStash()
	}()
}

func Generate(w http.ResponseWriter, r *http.Request) error {
	client := createClient()
	handleBrowsers(w, r)

	stashedData := getFromStash()
	if stashedData != nil {
		fmt.Println("Got key from stash")
		str := fmt.Sprintf("\n\nAccount type: %v\nData available: %vGB\nLicense: %v\n",
			stashedData.Type, stashedData.RefCount, stashedData.License)

		fmt.Fprint(w, str)
		return nil
	}

	fmt.Println("Couldn't get key from stash, going the slow path")

	flusher, _ := w.(http.Flusher)
	pb := progressbar.New(w, flusher)
	pb.Update(10)

	acc1, err := registerAccount(client)
	if err != nil {
		return err
	}
	pb.Update(20)

	acc2, err := registerAccount(client)
	if err != nil {
		return err
	}
	pb.Update(30)

	if err := acc1.addReferrer(client, acc2); err != nil {
		return err
	}
	pb.Update(40)

	if err := acc2.removeDevice(client); err != nil {
		return nil
	}
	pb.Update(50)

	if err := acc1.setKey(client, config.Keys[rand.Intn(len(config.Keys))]); err != nil {
		return err
	}
	pb.Update(60)

	if err := acc1.setKey(client, acc1.Account.License); err != nil {
		return err
	}
	pb.Update(70)

	accData, err := acc1.fetchAccountData(client)
	if err != nil {
		return err
	}
	pb.Update(80)

	if err := acc1.removeDevice(client); err != nil {
		return err
	}
	pb.Update(90)

	str := fmt.Sprintf("\n\nAccount type: %v\nData available: %vGB\nLicense: %v\n",
		accData.Type, accData.RefCount, accData.License)

	fmt.Fprint(w, str)
	return nil
}

func registerAccount(client *http.Client) (*account, error) {
	request, err := http.NewRequest("POST", config.BaseURL+"/reg", nil)
	if err != nil {
		return nil, ErrCreateRequest
	}
	setCommonHeaders(request)

	response, err := client.Do(request)
	if err != nil {
		return nil, ErrRegisterAccount
	}
	defer response.Body.Close()

	acc := account{}
	err = json.NewDecoder(response.Body).Decode(&acc)
	if err != nil {
		return nil, ErrDecodeResponse
	}
	return &acc, nil
}
