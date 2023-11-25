package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/rclancey/sensors/netscan"
)

func main() {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	run, err := netscan.ParseNMAP(data)
	if err != nil {
		log.Fatal(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(run)
}
