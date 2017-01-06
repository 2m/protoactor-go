package consul_cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/AsynkronIT/protoactor-go/cluster"
)

type RegisterAgentService struct {
	ID                string   `json:"ID"`
	Name              string   `json:"Name"`
	Tags              []string `json:"Tags"`
	Address           string   `json:"Address"`
	Port              int      `json:"Port"`
	EnableTagOverride bool     `json:"EnableTagOverride"`
	Check             struct {
		DeregisterCriticalServiceAfter string `json:"DeregisterCriticalServiceAfter,omitempty"`
		Script                         string `json:"Script,omitempty"`
		HTTP                           string `json:"HTTP,omitempty"`
		Interval                       string `json:"Interval,omitempty"`
		TTL                            string `json:"TTL,omitempty"`
	} `json:"Check"`
}

type ConsulProvider struct {
	shutdown bool
	id       string
}

func New() *ConsulProvider {
	p := &ConsulProvider{}
	return p
}

func (p *ConsulProvider) RegisterMember(clusterName string, address string, port int, knownKinds []string) error {
	p.id = fmt.Sprintf("%v@%v:%v", clusterName, address, port)
	s := RegisterAgentService{
		ID:      p.id,
		Name:    clusterName,
		Tags:    knownKinds,
		Address: address,
		Port:    port,
		Check: struct {
			DeregisterCriticalServiceAfter string `json:"DeregisterCriticalServiceAfter,omitempty"`
			Script                         string `json:"Script,omitempty"`
			HTTP                           string `json:"HTTP,omitempty"`
			Interval                       string `json:"Interval,omitempty"`
			TTL                            string `json:"TTL,omitempty"`
		}{
			DeregisterCriticalServiceAfter: "20s",
			TTL: "10s",
		},
	}

	b, err := json.Marshal(s)
	if err != nil {
		return err
	}

	url := "http://127.0.0.1:8500/v1/agent/service/register"
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(b))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)

	if resp.StatusCode != 200 {
		log.Fatal(bodyStr)
		return fmt.Errorf("Expected status 200, got: %v", resp.Status)
	}

	p.UpdateTTL()
	return nil
}

func (p *ConsulProvider) UpdateTTL() {
	refresh := func() error {
		url := "http://127.0.0.1:8500//v1/agent/check/pass/service:" + p.id
		req, err := http.NewRequest("PUT", url, nil)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		bodyStr := string(body)

		if resp.StatusCode != 200 {
			log.Fatal(bodyStr)
			return fmt.Errorf("Expected status 200, got: %v", resp.Status)
		}
		return nil
	}

	go func() {
		for !p.shutdown {
			log.Println("Refreshing service TTL")
			err := refresh()
			if err != nil {
				log.Println("Failure refreshing service TTL")
			}
			time.Sleep(2 * time.Second)
		}
	}()
}

func (p *ConsulProvider) Shutdown() error {
	p.shutdown = true
	///v1/agent/service/deregister
	url := "http://127.0.0.1:8500/v1/agent/service/deregister/" + p.id
	req, err := http.NewRequest("GET", url, nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	bodyStr := string(body)

	if resp.StatusCode != 200 {
		log.Fatal(bodyStr)
		return fmt.Errorf("Expected status 200, got: %v", resp.Status)
	}

	return nil
}

func (p *ConsulProvider) GetStatusChanges() <-chan cluster.MemberStatus {
	c := make(chan cluster.MemberStatus)
	go func() {
		for !p.shutdown {

		}
	}()
	return c
}
