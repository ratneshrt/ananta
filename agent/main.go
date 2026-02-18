package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strconv"
)

type DeployRequest struct {
	AppName   string `json:"app_name"`
	Repo      string `json:"repo"`
	Runtime   string `json:"runtime"`
	Subdomain string `json:"subdomain"`
	Port      int    `json:"port"`
}

type DeployResponse struct {
	URL   string `json:"url"`
	Error string `json:"error,omitempty"`
}

func deployHandler(w http.ResponseWriter, r *http.Request) {
	var req DeployRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	cmd := exec.Command(
		"bash",
		"/srv/deploy/deploy_app.sh",
		req.AppName,
		req.Repo,
		req.Runtime,
		req.Subdomain,
		strconv.Itoa(req.Port),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		json.NewEncoder(w).Encode(DeployResponse{
			Error: string(output),
		})
		return
	}

	json.NewEncoder(w).Encode(DeployResponse{
		URL: string(output),
	})
}

func main() {
	http.HandleFunc("/deploy", deployHandler)
	log.Println("deploy agent listening on :9000")
	log.Fatal(http.ListenAndServe(":9000", nil))
}
