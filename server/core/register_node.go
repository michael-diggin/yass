package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type location struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

// RegisterServerWithGateway will register the storage server with the api gateway so it can accept requests
func (s server) RegisterServerWithGateway(gateway, nodeAddress string, port int) error {

	addr := location{IP: nodeAddress, Port: fmt.Sprintf("%d", port)}
	payload, err := json.Marshal(addr)
	if err != nil {
		return err
	}
	resp, err := http.Post("http://"+gateway+"/register", "application/json", bytes.NewBuffer(payload))
	if err != nil || resp.StatusCode != http.StatusCreated {
		return err
	}
	return nil
}
