package config

import (
	"errors"
	aq "github.com/emirpasic/gods/queues/arrayqueue"
	"github.com/spf13/viper"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type PortsQueue struct {
	Queue *aq.Queue
	Mu    sync.Mutex
}

type config struct {
	PortRange           string `json:"port_range"`
	MaxProxyLifetimeSec int    `json:"max_proxy_lifetime_sec"`
	ApiKey              string `json:"api_key"`
	Ports               *PortsQueue
	GinListen           string `json:"gin_listen"`
	ProxyListenIP       string `json:"proxy_listen_ip"`
}

var C *config

func InitConfig() {
	viper.SetDefault("PortRange", "8000-9000")
	viper.SetDefault("MaxProxyLifetimeSec", 3600)
	viper.SetDefault("ApiKey", "your_api_key")
	viper.SetDefault("GinListen", "localhost:8080")
	viper.SetDefault("ProxyListenIP", "127.0.0.1")
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			err := viper.SafeWriteConfig()
			if err != nil {
				log.Fatal("Failed to write config file: ", err)
			}
			log.Println("Config file not found, created a new one. Please edit it and restart the server.")
			os.Exit(0)
		} else {
			log.Fatal("Failed to read config file: ", err)
		}
	}
	viper.GetString("PortRange")
	err := viper.Unmarshal(&C)
	if err != nil {
		log.Fatal("Failed to unmarshal config: ", err)
	}
	portRange := strings.Split(C.PortRange, "-")
	start, err := strconv.Atoi(portRange[0])
	end, err1 := strconv.Atoi(portRange[1])
	if err != nil || err1 != nil {
		log.Fatal("Invalid port range")
	}
	C.Ports = &PortsQueue{Queue: aq.New()}
	C.Ports.Mu.Lock()
	defer C.Ports.Mu.Unlock()
	for i := start; i <= end; i++ {
		C.Ports.Queue.Enqueue(i)
	}
}
