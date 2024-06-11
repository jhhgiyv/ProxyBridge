package proxy

import (
	"ProxyBridge/config"
	"ProxyBridge/router_struct"
	"errors"
	"fmt"
	"github.com/elazarl/goproxy"
	"golang.org/x/net/proxy"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type proxyData struct {
	protocol string
	address  string
	username string
	password string
}

func (p proxyData) string() string {
	if p.username != "" {
		return p.protocol + "://" + p.username + ":" + p.password + "@" + p.address
	}
	return p.protocol + "://" + p.address
}

func newProxyData(requests *router_struct.NewProxyRequests) *proxyData {
	split := strings.Split(requests.Proxy, "://")
	var protocol string
	var withoutProtocol string
	var username string
	var password string
	var address string
	if len(split) == 1 {
		protocol = "http"
		withoutProtocol = split[0]
	} else {
		protocol = split[0]
		withoutProtocol = split[1]
	}
	split = strings.Split(withoutProtocol, "@")
	if len(split) == 1 {
		address = split[0]
	} else {
		auth := split[0]
		address = split[1]
		split = strings.Split(auth, ":")
		username = split[0]
		if len(split) == 2 {
			password = split[1]
		}
	}
	return &proxyData{
		protocol: protocol,
		address:  address,
		username: username,
		password: password,
	}
}

func setOnRequest(server *goproxy.ProxyHttpServer, proxyData *proxyData, response chan router_struct.Response) error {
	proxyUrl, err := url.Parse(proxyData.string())
	if err != nil {
		response <- router_struct.Response{Status: 400, Message: "Invalid proxy address"}
		return err
	}
	var transport *http.Transport

	switch proxyData.protocol {
	case "http", "https":
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
	case "socks4":
		response <- router_struct.Response{Status: 501, Message: "Not implemented"}
		return errors.New("not implemented")
	case "socks5":
		if proxyData.username != "" || proxyData.password != "" {
			dialer, err := proxy.SOCKS5("tcp", proxyData.address, &proxy.Auth{User: proxyData.username, Password: proxyData.password}, proxy.Direct)
			if err != nil {
				response <- router_struct.Response{Status: 500, Message: "Failed to create proxy dialer"}
				return err
			}
			transport = &http.Transport{
				Dial: dialer.Dial,
			}
		} else {
			dialer, err := proxy.SOCKS5("tcp", proxyData.address, nil, proxy.Direct)
			if err != nil {
				response <- router_struct.Response{Status: 500, Message: "Failed to create proxy dialer"}
				return err
			}
			transport = &http.Transport{
				Dial: dialer.Dial,
			}
		}
	default:
		response <- router_struct.Response{Status: 400, Message: "Invalid protocol"}
		return errors.New("not implemented")
	}
	server.Tr = transport
	server.ConnectDial = transport.Dial
	return nil
}

func CreateProxy(requests *router_struct.NewProxyRequests, response chan router_struct.Response) {
	p := newProxyData(requests)
	proxyServer := goproxy.NewProxyHttpServer()
	err := setOnRequest(proxyServer, p, response)
	if err != nil {
		log.Println("Failed to create proxy: ", err)
		return
	}

	config.C.Ports.Mu.Lock()
	defer config.C.Ports.Mu.Unlock()
	port, _ := config.C.Ports.Queue.Dequeue()
	if port == nil {
		response <- router_struct.Response{Status: 500, Message: "No available ports"}
		return
	}
	listenAddress := fmt.Sprintf("%s:%d", config.C.ProxyListenIP, port)
	server := &http.Server{
		Addr:    listenAddress,
		Handler: proxyServer,
	}
	isErr := false
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			isErr = true
			response <- router_struct.Response{Status: 500, Message: "Failed to start proxy server"}
			log.Println("Failed to start proxy server: ", err)
		}
	}()
	time.AfterFunc(time.Duration(100), func() {
		if isErr {
			return
		}
		response <- router_struct.Response{Status: 200, Message: "Proxy server started", Data: map[string]any{
			"port": port,
		}}
	})

	time.AfterFunc(time.Duration(requests.ProxyLifetimeSec)*time.Second, func() {
		log.Println("Closing proxy server " + listenAddress)
		if isErr {
			return
		}
		err := server.Close()
		if err != nil {
			log.Println("Failed to close proxy server: ", err)
			return
		}
		config.C.Ports.Mu.Lock()
		defer config.C.Ports.Mu.Unlock()
		config.C.Ports.Queue.Enqueue(port)
	})
}
