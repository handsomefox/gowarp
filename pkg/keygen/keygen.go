package keygen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gowarp/pkg/progressbar"
	"math/rand"
	"net/http"
)

const (
	cfClientVersion = "a-6.3-2223"
	userAgent       = "okhttp/3.12.1"
	baseURL         = "https://api.cloudflareclient.com/v0a2223"
)

var keys = []string{
	"47d58Hqv-ueR37x50-db3l70n2",
	"bwK01o62-c15C78MH-Z2g5Ji74",
	"c3Dd52l4-K28SzY14-0wKO47c8",
	"Vq765xO8-TR392VI5-z4j6Xp28",
	"3C651Aqt-3PHl50V7-3751AOCb",
	"WE38q17B-25DP3um4-9A7xg48Z",
	"x0yZ894a-3Sh2b90q-718Wrw4A",
	"3z1tW5s7-03L78SFw-6w2T9F4c",
	"bkp5301R-VDB6234t-x954U8Jb",
	"ofQ759g4-9628GDiy-6i58VGa2",
}

func Generate(w http.ResponseWriter, flusher http.Flusher) error {
	pb := progressbar.New(w, flusher)

	kg := keygen{
		client: createClient(),
		writer: w,
	}
	pb.Update(10)

	if err := kg.createAccounts(); err != nil {
		return err
	}
	pb.Update(20)

	if err := kg.addReferrer(); err != nil {
		return err
	}
	pb.Update(30)

	if err := kg.deleteSecondAccount(); err != nil {
		return err
	}
	pb.Update(40)

	if err := kg.setFirstAccountKey(); err != nil {
		return err
	}
	pb.Update(50)

	if err := kg.setFirstAccountLicense(); err != nil {
		return err
	}
	pb.Update(60)

	result, err := kg.getLicenseInformation()
	if err != nil {
		return err
	}
	pb.Update(70)

	if err := kg.deleteAccount(); err != nil {
		return err
	}
	pb.Update(80)

	out := fmt.Sprintf("\n\nAccount type: %s\nData available: %sGB\nLicense: %s\n", result.Type, result.RefCount.String(), result.License)
	pb.Update(90)

	_, _ = fmt.Fprintln(w, out)
	return nil
}

func (kg *keygen) createAccounts() error {
	acc1, err := kg.register()
	if err != nil {
		return err
	}

	acc2, err := kg.register()
	if err != nil {
		return err
	}

	kg.accounts = &createdAccounts{
		First:  acc1,
		Second: acc2,
	}

	return nil
}

func (kg *keygen) register() (*account, error) {
	req, err := http.NewRequest("POST", baseURL+"/reg", nil)
	req.Header.Add("CF-Client-Version", cfClientVersion)
	req.Header.Add("User-Agent", userAgent)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	resp, err := kg.client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var reg account
	err = toJSON(resp, &reg)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &reg, err
}

func (kg *keygen) addReferrer() error {
	payload, _ := json.Marshal(map[string]interface{}{
		"referrer": kg.accounts.Second.Id,
	})
	url := baseURL + fmt.Sprintf("/reg/%s", kg.accounts.First.Id)
	patchRequest, err := http.NewRequest("PATCH", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	patchRequest.Header.Set("Content-Type", "application/json; charset=UTF-8")
	patchRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", kg.accounts.First.Token))

	_, err = kg.client.Do(patchRequest)
	return err
}

func (kg *keygen) deleteSecondAccount() error {
	url := baseURL + fmt.Sprintf("/reg/%s", kg.accounts.Second.Id)
	deleteRequest, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	deleteRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", kg.accounts.Second.Token))
	_, err = kg.client.Do(deleteRequest)
	return err
}

func (kg *keygen) setFirstAccountKey() error {
	key := keys[rand.Intn(len(keys))]
	payload, _ := json.Marshal(map[string]interface{}{
		"license": key,
	})
	url := baseURL + fmt.Sprintf("/reg/%s/account", kg.accounts.First.Id)
	putRequest, err := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	putRequest.Header.Set("Content-Type", "application/json; charset=UTF-8")
	putRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", kg.accounts.First.Token))
	_, err = kg.client.Do(putRequest)
	return err
}

func (kg *keygen) setFirstAccountLicense() error {
	payload, _ := json.Marshal(map[string]interface{}{
		"license": kg.accounts.First.Account.License,
	})
	url := baseURL + fmt.Sprintf("/reg/%s/account", kg.accounts.First.Id)
	putRequest, err := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	putRequest.Header.Set("Content-Type", "application/json; charset=UTF-8")
	putRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", kg.accounts.First.Token))
	_, err = kg.client.Do(putRequest)
	return err
}

func (kg *keygen) getLicenseInformation() (*result, error) {
	url := baseURL + fmt.Sprintf("/reg/%s/account", kg.accounts.First.Id)
	getRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	getRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", kg.accounts.First.Token))

	resp, err := kg.client.Do(getRequest)
	if err != nil {
		return nil, err
	}

	var result result
	if err := toJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (kg *keygen) deleteAccount() error {
	url := baseURL + fmt.Sprintf("/reg/%s", kg.accounts.First.Id)
	deleteRequest, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	deleteRequest.Header.Set("Authorization", fmt.Sprintf("Bearer %s", kg.accounts.First.Token))
	_, err = kg.client.Do(deleteRequest)
	return err
}
