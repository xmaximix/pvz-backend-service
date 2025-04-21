package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const base = "http://localhost:8080"

func waitFor() {
	for i := 0; i < 30; i++ {
		resp, err := http.Post(base+"/dummyLogin", "application/json", bytes.NewBufferString(`{"role":"employee"}`))
		if err == nil && resp.StatusCode == http.StatusOK {
			return
		}
		time.Sleep(time.Second)
	}
	panic("service unavailable")
}

func TestE2E(t *testing.T) {
	cmd := exec.Command("docker", "compose", "up", "--build", "--detach")
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, string(out))
	defer exec.Command("docker", "compose", "down", "--volumes").Run()

	waitFor()

	modToken := getToken(t, "moderator")
	cliToken := getToken(t, "employee")

	var pvz struct {
		ID   string `json:"id"`
		City string `json:"city"`
	}
	{
		reqBody := `{"city":"Москва"}`
		req, _ := http.NewRequest("POST", base+"/pvz", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+modToken)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		data, _ := io.ReadAll(resp.Body)
		assert.Equalf(t, http.StatusCreated, resp.StatusCode,
			"CreatePVZ returned %d:\n%s", resp.StatusCode, data,
		)

		err = json.Unmarshal(data, &pvz)
		assert.NoError(t, err)
		assert.Equal(t, "Москва", pvz.City)
	}

	{
		reqBody := fmt.Sprintf(`{"pvzId":"%s"}`, pvz.ID)
		req, _ := http.NewRequest("POST", base+"/receptions", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+cliToken)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		data, _ := io.ReadAll(resp.Body)
		assert.Equalf(t, http.StatusCreated, resp.StatusCode,
			"OpenReception returned %d:\n%s", resp.StatusCode, data,
		)
	}

	for i := 0; i < 50; i++ {
		reqBody := fmt.Sprintf(`{"pvzId":"%s","type":"электроника"}`, pvz.ID)
		req, _ := http.NewRequest("POST", base+"/products", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+cliToken)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		data, _ := io.ReadAll(resp.Body)
		assert.Equalf(t, http.StatusCreated, resp.StatusCode,
			"AddProduct #%d returned %d:\n%s", i+1, resp.StatusCode, data,
		)
	}

	{
		closeURL := fmt.Sprintf("%s/pvz/%s/close_last_reception", base, pvz.ID)
		req, _ := http.NewRequest("POST", closeURL, nil)
		req.Header.Set("Authorization", "Bearer "+cliToken)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		data, _ := io.ReadAll(resp.Body)
		assert.Equalf(t, http.StatusOK, resp.StatusCode,
			"CloseReception returned %d:\n%s", resp.StatusCode, data,
		)
	}
}

func getToken(t *testing.T, role string) string {
	payload := fmt.Sprintf(`{"role":"%s"}`, role)
	resp, err := http.Post(base+"/dummyLogin", "application/json", bytes.NewBufferString(payload))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "/dummyLogin must return 200")
	var out struct{ Token string }
	data, _ := io.ReadAll(resp.Body)
	assert.NoError(t, json.Unmarshal(data, &out), "couldn't parse /dummyLogin JSON: "+string(data))
	return out.Token
}
