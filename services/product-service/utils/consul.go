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
	registration.Tags = []string{"product-service"}
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
