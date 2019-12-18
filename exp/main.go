package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	darksky "github.com/sophiaehlen/darksky-client"
)

func main() {
	// c := darksky.NewClient(apiKey)

	keybytes, err := ioutil.ReadFile("../dev.txt")
	if err != nil {
		log.Fatal(err)
	}
	key := string(keybytes)
	key = strings.TrimSpace(key)

	c := darksky.Client{
		Key: key,
	}
	lat := 32.589720
	long := -116.466988
	fc, err := c.Forecast(lat, long)
	if err != nil {
		panic(err)
	}
	fmt.Println(fc.CurrentTemperature())

	lt, err := fc.LocalTime()
	if err != nil {
		panic(err)
	}
	fmt.Println(lt)

}
