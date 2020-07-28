package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// Bank is a collection of basic properties from one bank
type Bank struct {
	Blz  string
	Name string
	Bic  string
}

var rawFilename string = fmt.Sprintf(getBankdataFilename())
var banks map[string]Bank

var countAll int32 = 0
var countValid int32 = 0

func main() {
	file, err := os.Open(rawFilename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	banks = readBanksFromTxtData(file)
	log.Printf("Loaded %d from %d lines.", countValid, countAll)

	r := gin.Default()

	r.GET("/bic/:iban", func(c *gin.Context) {
		blz := blzFromIban(c.Param("iban"))
		bank := bankFromBlz(blz)
		c.JSON(200, gin.H{
			"iban": c.Param("iban"),
			"bic":  bank.Bic,
			"name": bank.Name,
			"blz":  bank.Blz,
		})
	})

	r.GET("/blz/:iban", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"iban": c.Param("iban"),
			"blz":  blzFromIban(c.Param("iban")),
		})
	})

	r.GET("/banks", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"banks": banks,
		})
	})

	r.Run(getListenAddress())
}

func blzFromIban(iban string) string {
	return iban[4:12]
}

func bankFromBlz(blz string) Bank {
	return banks[blz]
}

func readBanksFromTxtData(file io.Reader) map[string]Bank {
	scanner := bufio.NewScanner(file)
	banks := make(map[string]Bank)

	for scanner.Scan() {
		bank := getBankFromRawLine(toUtf8(scanner.Bytes()))
		countAll++

		if bank.Blz == "" {
			continue
		}

		banks[bank.Blz] = bank
		countValid++
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return banks
}

func getBankFromRawLine(rawLine string) Bank {
	parsed := parseRawLine(rawLine)

	if (*parsed)[10] == "D" && (*parsed)[1] == "1" {
		return Bank{}
	}

	return Bank{strings.TrimSpace((*parsed)[0]), strings.TrimSpace((*parsed)[2]), strings.TrimSpace((*parsed)[7])}
}

func parseRawLine(rawLine string) *[]string {
	parsed := make([]string, 13)
	parsed[0] = rawLine[0:8]
	parsed[1] = rawLine[8:9]
	parsed[2] = rawLine[9:67]
	parsed[3] = rawLine[67:72]
	parsed[4] = rawLine[72:107]
	parsed[5] = rawLine[107:134]
	parsed[6] = rawLine[134:139]
	parsed[7] = rawLine[139:150]
	parsed[8] = rawLine[150:152]
	parsed[9] = rawLine[152:158]
	parsed[10] = rawLine[158:159]
	parsed[11] = rawLine[159:160]
	parsed[12] = rawLine[160:168]

	return &parsed
}

func toUtf8(iso8859_1Buf []byte) string {
	buf := make([]rune, len(iso8859_1Buf))
	for i, b := range iso8859_1Buf {
		buf[i] = rune(b)
	}
	return string(buf)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getListenAddress() string {
	host := getEnv("HOST", "")
	port := getEnv("PORT", "8000")
	return fmt.Sprintf("%s:%s", host, port)
}

func getBankdataFilename() string {
	dataDirectory := getEnv("DATADIRECTORY", "./data")
	dataFile := getEnv("DATAFILE", "BLZ.txt")

	return fmt.Sprintf("%s/%s", dataDirectory, dataFile)
}
