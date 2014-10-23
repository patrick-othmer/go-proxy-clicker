package main
 
import (
	"bufio"
	"os"
	"runtime"
	"log"
	"flag"
)

var urlTarget = "http://www.google.de"
var disableWeb bool

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.BoolVar(&disableWeb, "disable-web", false, "Disable webserver")
	flag.Parse()
	if disableWeb {
		Message("Clicker start (c) by Patrick Othmer", false)
		voter()
	} else {
		go h.run()
		loadIndex()
		Message("Clicker start (c) by Patrick Othmer", false)
		webServer()
	}
}

func Message(message string, output bool) {
	log.Printf("%s\n", message)
	if !disableWeb && output {
		h.broadcast <- []byte(message)
	}
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
						Message(proxy + "|Good proxy", true)
						proxyList[proxy] = proxy
					}
			}
		}
	}()
}

