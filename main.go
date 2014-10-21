package main
 
import (
	"fmt"
	"time"
	"net"
	"bufio"
	"net/url"
	"net/http"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"runtime"
)

var timeout = time.Duration(3 * time.Second)
var debug = false
var urlTarget = "http://www.google.de"
var addProxy = make(chan string)
var proxyList = make(map[string] string)
var useragents = [...]string{
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; .NET CLR 1.1.4322; .NET CLR 2.0.50727; .NET CLR 3.0.04506.30)",
	"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; .NET CLR 1.1.4322)",
	"Googlebot/2.1 (http://www.googlebot.com/bot.html)",
	"Opera/9.20 (Windows NT 6.0; U; en)",
	"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.8.1.1) Gecko/20061205 Iceweasel/2.0.0.1 (Debian-2.0.0.1+dfsg-2)",
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; Trident/4.0; FDM; .NET CLR 2.0.50727; InfoPath.2; .NET CLR 1.1.4322)",
	"Opera/10.00 (X11; Linux i686; U; en) Presto/2.2.0",
	"Mozilla/5.0 (Windows; U; Windows NT 6.0; he-IL) AppleWebKit/528.16 (KHTML, like Gecko) Version/4.0 Safari/528.16",
	"Mozilla/5.0 (compatible; Yahoo! Slurp/3.0; http://help.yahoo.com/help/us/ysearch/slurp)",
	"Mozilla/5.0 (X11; U; Linux x86_64; en-US; rv:1.9.2.13) Gecko/20101209 Firefox/3.6.13",
	"Mozilla/4.0 (compatible; MSIE 9.0; Windows NT 5.1; Trident/5.0)",
	"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 5.1; Trident/4.0; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
	"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 6.0)",
	"Mozilla/4.0 (compatible; MSIE 6.0b; Windows 98)",
	"Mozilla/5.0 (Windows; U; Windows NT 6.1; ru; rv:1.9.2.3) Gecko/20100401 Firefox/4.0 (.NET CLR 3.5.30729)",
	"Mozilla/5.0 (X11; U; Linux x86_64; en-US; rv:1.9.2.8) Gecko/20100804 Gentoo Firefox/3.6.8",
	"Mozilla/5.0 (X11; U; Linux x86_64; en-US; rv:1.9.2.7) Gecko/20100809 Fedora/3.6.7-1.fc14 Firefox/3.6.7",
	"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
	"Mozilla/5.0 (compatible; Yahoo! Slurp; http://help.yahoo.com/help/us/ysearch/slurp)",
	"YahooSeeker/1.2 (compatible; Mozilla 4.0; MSIE 5.5; yahooseeker at yahoo-inc dot com ; http://help.yahoo.com/help/us/shop/merchant/)",
	"Mozilla/5.0 (Linux x86_64; rv:33.0) Firefox/33.0",
}

func dialTimeout() func(net, addr string) (c net.Conn, err error) {
	 return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, timeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(timeout))
        tcp_conn := conn.(*net.TCPConn)
        tcp_conn.SetLinger(0)
        return tcp_conn, err
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("Clicker start (c) by Patrick Othmer")
	proxyCh := make(chan string)
	wg := new(sync.WaitGroup)

	handleNewProxys()

	fmt.Println("Add 50 worker")
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go worker(proxyCh, wg)
	}

	fmt.Println("Read proxy.txt file")
	lines, err := readLines("proxy.txt")
	if err != nil {
		fmt.Printf("readLines: %s", err)
	}

	fmt.Println("Start click process")
	for _, line := range lines {
		proxyCh <- line
	}
	runtime.GC()
	fmt.Println("Use only good proxies")
	for {
		for _, line := range proxyList {
			proxyCh <- line
		}
		runtime.GC()
	}

	close(proxyCh)

	wg.Wait()
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func handleNewProxys() {
	go func() {
		for {
			select {
				case proxy := <-addProxy:
					if _, ok := proxyList[proxy]; !ok {
						fmt.Printf("%s: Good proxy\n", proxy)
						proxyList[proxy] = proxy
					}
			}
		}
	}()
}

func worker(proxyChan chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for proxy := range proxyChan {
		if visit(proxy) {
			addProxy <- proxy
		}
	}
}

func visit(proxy string) bool {
	if proxy == "" {
		return false
	}
	proxyUrl, err := url.Parse("http://" + proxy)

	transport := &http.Transport{
		Dial: dialTimeout(),
		Proxy: http.ProxyURL(proxyUrl),
		DisableKeepAlives: true,
	}
	transport.CloseIdleConnections()

    client := &http.Client{
        Transport: transport,
    }
    req, err := http.NewRequest("GET", urlTarget, nil)
	if err != nil {
		return false
	}
    req.Header.Set("User-Agent", useragents[rand.Intn(len(useragents))])
    resp, err := client.Do(req)
	if err != nil {
		if debug {
			fmt.Printf("%s: Error - %s\n", proxy, err)
		}
		return false
	} else {
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			return false
		} else {
			if string(contents) != "" {
				fmt.Printf("%s: Visited\n", proxy)
				resp.Body.Close()
				return true
			}
			resp.Body.Close()
			return false
		}
	}
	return false
}