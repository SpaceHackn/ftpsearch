package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

// takes extentions file.
func readExtensionsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var extensions []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		extensions = append(extensions, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return extensions, nil
}

// WRITES TO LIST
func handleFTP(c *ftp.ServerConn, err error, finAddress string) {
	if err != nil {
		log.Fatal(err)
		return
	}

	extensions, err := readExtensionsFromFile("extensions.txt")
	if err != nil {
		log.Fatal(err)
		return
	}

	var filePaths []string

	w := c.Walk("/")
	for w.Next() {
		if w.Err() != nil {
			continue
		}
		for _, extension := range extensions {
			if strings.HasSuffix(w.Stat().Name, extension) {
				filePaths = append(filePaths, w.Path())
				break
			}
		}
	}

	fmt.Println("File paths:")
	for _, filePath := range filePaths {
		fmt.Println(filePath)
	}
	printFilePaths(filePaths, finAddress)
}

func printFilePaths(filePaths []string, ipAddress string) {
	file, err := os.OpenFile("loot.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	for _, filePath := range filePaths {
		_, err := file.WriteString(ipAddress + " - " + filePath + "\n")
		if err != nil {
			log.Fatal(err)
		}
	}
}

// function to increment IP address
func inc(ip net.IP) net.IP {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
	return ip
}

func main() {
	fmt.Println("golang practice FTP searcher")

	// take command line argument for the path to the .csv file
	filePath := flag.String("file", "ips.csv", "usage: -file ips.csv")
	// parse flag options
	flag.Parse()

	// Read the .csv file
	file, err := os.Open(*filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	ips, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	port := strconv.Itoa(21)

	for _, row := range ips {
		ip := row[0]
		FinAddress := ip + ":" + port
		fmt.Println("Connecting to:", FinAddress)
		c, err := ftp.Dial(FinAddress, ftp.DialWithShutTimeout(2*time.Second))
		if err != nil {
			continue
		}
		err = c.Login("anonymous", "anonymous")
		if err != nil {
			continue
		}
		handleFTP(c, err, FinAddress)
		if err := c.Quit(); err != nil {
			log.Fatal(err)
		}
	}
}
