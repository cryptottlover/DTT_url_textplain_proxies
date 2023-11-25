package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	proxies           []string
	proxiesMutex      sync.Mutex
	minProxyCount     int
	proxyURL          string
	userWebServerPort string
	protocol          string
)

func fetchProxies() ([]string, error) {
	response, err := http.Get(proxyURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	proxies := strings.Split(string(body), "\n")

	var cleanedProxies []string
	for _, proxy := range proxies {
		proxy = strings.TrimSpace(proxy)
		if proxy != "" {
			cleanedProxies = append(cleanedProxies, protocol+proxy)
		}
	}

	return cleanedProxies, nil
}

func startWebServer() {
	http.HandleFunc("/getproxy", func(w http.ResponseWriter, r *http.Request) {
		proxiesMutex.Lock()
		if len(proxies) > 0 {
			proxy := proxies[0]
			proxies = proxies[1:]
			proxiesMutex.Unlock()
			fmt.Fprint(w, proxy)
		} else {
			proxiesMutex.Unlock()
			fmt.Fprint(w, "")
		}
	})
	http.ListenAndServe(userWebServerPort, nil)
}

func updateProxiesIfNeeded() {
	for {
		time.Sleep(1 * time.Second)
		if len(proxies) < minProxyCount {
			proxiesMutex.Lock()
			newProxies, err := fetchProxies()
			if err == nil && len(newProxies) > 0 {
				proxies = append(proxies, newProxies...)
				fmt.Println("Updated proxies. Total:", len(proxies))
			} else {
				fmt.Println("Failed to update proxies:", err)
			}
			proxiesMutex.Unlock()
		}
	}
}

func main() {
	var port, protocolarg string

	fmt.Print("Enter the proxy URL: ")
	fmt.Scan(&proxyURL)

	fmt.Print("Enter the minimum number of proxies to update: ")
	fmt.Scan(&minProxyCount)

	fmt.Print("Enter the user web server port (default: 1337): ")
	fmt.Scan(&port)

	fmt.Print("Enter the protocol number:\n1 - http://\n2 - socks4://\n3 - socks5://\n")
	fmt.Scan(&protocolarg)

	switch protocolarg {
	case "1":
		protocol = "http://"
	case "2":
		protocol = "socks4://"
	case "3":
		protocol = "socks5://"
	default:
		fmt.Println("1 or 2 or 3!")
		os.Exit(1)
	}

	userWebServerPort = ":" + port

	go startWebServer()

	newProxies, err := fetchProxies()
	if err != nil {
		fmt.Println("Failed to parse proxies:", err)
		return
	}

	proxiesMutex.Lock()
	proxies = append(proxies, newProxies...)
	proxiesMutex.Unlock()

	go updateProxiesIfNeeded()

	select {}
}
