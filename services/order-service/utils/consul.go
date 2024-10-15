package utils

import (
	"fmt"
	"log"
	"os"

	consulapi "github.com/hashicorp/consul/api"
)

func RegisterServiceWithConsul(serviceName string, servicePort int, consulAddress string) error {
	config := consulapi.DefaultConfig()
	config.Address = consulAddress

	client, err := consulapi.NewClient(config)
	if err != nil {
		return fmt.Errorf("consul client error: %v", err)
	}

	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = serviceName
	registration.Name = serviceName
	registration.Address = os.Getenv("SERVICE_HOST")
	registration.Port = servicePort
	registration.Check = &consulapi.AgentServiceCheck{
		GRPC:                           fmt.Sprintf("%s:%d/%s", registration.Address, servicePort, serviceName),
		Interval:                       "10s",
		DeregisterCriticalServiceAfter: "1m",
	}

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("service registration error: %v", err)
	}

	log.Printf("Registered service %s with Consul", serviceName)
	return nil
}

func GetServiceAddress(client *consulapi.Client, serviceName string) (string, error) {
	services, _, err := client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "", fmt.Errorf("error retrieving instances from Consul: %v", err)
	}

	if len(services) == 0 {
		return "", fmt.Errorf("no instances of service %s found", serviceName)
	}

	service := services[0].Service
	address := fmt.Sprintf("%s:%d", service.Address, service.Port)
	return address, nil
}
