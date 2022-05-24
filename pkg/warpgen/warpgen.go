package warpgen

import (
	"encoding/json"
	"fmt"
	"gowarp/pkg/config"
	"gowarp/pkg/progressbar"
	"math/rand"
	"net/http"
)

var (
	client = createClient()
)

func Generate(w http.ResponseWriter, r *http.Request) error {
	handleBrowsers(w, r)

	flusher, _ := w.(http.Flusher)
	pb := progressbar.New(w, flusher)
	pb.Update(10)

	acc1, err := registerAccount(w, r)
	if err != nil {
		return err
	}
	pb.Update(20)

	acc2, err := registerAccount(w, r)
	if err != nil {
		return err
	}
	pb.Update(30)

	if err := acc1.addReferrer(acc2); err != nil {
		return err
	}
	pb.Update(40)

	if err := acc2.removeDevice(); err != nil {
		return nil
	}
	pb.Update(50)

	if err := acc1.setKey(config.Keys[rand.Intn(len(config.Keys))]); err != nil {
		return err
	}
	pb.Update(60)

	if err := acc1.setKey(acc1.Account.License); err != nil {
		return err
	}
	pb.Update(70)

	accData, err := acc1.fetchAccountData()
	if err != nil {
		return err
	}
	pb.Update(80)

	if err := acc1.removeDevice(); err != nil {
		return err
	}
	pb.Update(90)

	str := fmt.Sprintf("\n\nAccount type: %v\nData available: %vGB\nLicense: %v\n",
		accData.Type, accData.RefCount, accData.License)

	fmt.Fprint(w, str)
	return nil
}

func registerAccount(w http.ResponseWriter, r *http.Request) (*account, error) {
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
