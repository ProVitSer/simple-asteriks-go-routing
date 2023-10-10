package main

import (
	"context"
	"encoding/csv"
	"github.com/CyCoreSystems/ari/v6"
	"github.com/CyCoreSystems/ari/v6/client/native"
	"github.com/joho/godotenv"
	"log"
	"os"
)

const (
	Bridge3CX             = "3cx-bridge"
	Base3CXRoute          = "3cx-default-route"
	DefaultRouteExtension = "99911"
	UnknownBranch         = "000"
)

type PhoneNumber string
type BranchCode string

func main() {
	phoneBranchMap, err := readCSVAndCreateMap("route.csv")
	if err != nil {
		log.Fatalf("Error read or create data from csv: %v", err)
	}

	logFile, err := setupLogFile("app.log")
	if err != nil {
		log.Fatal("Error open log:", err)
	}
	defer logFile.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Connecting to ARI")

	cl, err := native.Connect(&native.Options{
		Application:  goDotEnvVariable("STASIS_APPLICATION_NAME"),
		Username:     goDotEnvVariable("ARI_USERNAME"),
		Password:     goDotEnvVariable("ARI_PASSWORD"),
		URL:          goDotEnvVariable("ARI_URL"),
		WebsocketURL: goDotEnvVariable("ARI_WS_URL"),
	})

	if err != nil {
		log.Println("Failed to build ARI client", "error", err)
		return
	}

	log.Println("Starting listener app")
	log.Println("Listening for new calls")

	sub := cl.Bus().Subscribe(nil, "StasisStart")

	for {
		select {
		case e := <-sub.Events():
			v := e.(*ari.StasisStart)
			log.Println("Got stasis start", "channel", v.Channel.ID)
			go continueDialplan(ctx, cl, v, phoneBranchMap)
		case <-ctx.Done():
			return
		}
	}
}

func goDotEnvVariable(key string) string {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func continueDialplan(ctx context.Context, cl ari.Client, v *ari.StasisStart, phoneBranchMap map[PhoneNumber]BranchCode) {
	channel := cl.Channel().Get(v.Key(ari.ChannelKey, v.Channel.ID))

	phone := PhoneNumber(v.Channel.Caller.Number)
	branch, found := phoneBranchMap[phone]
	if found {
		log.Printf("Number %s corresponds branch %s\n", phone, branch)
		if branch == UnknownBranch {
			channel.Continue(Base3CXRoute, DefaultRouteExtension, 1)
		} else {
			channel.Continue(Bridge3CX, string(branch), 1)
		}
	} else {
		log.Printf("Number %s not found\n", phone)
		channel.Continue(Base3CXRoute, DefaultRouteExtension, 1)
	}
}

func readCSVAndCreateMap(filename string) (map[PhoneNumber]BranchCode, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	phoneBranchMap := make(map[PhoneNumber]BranchCode)

	lines, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, line := range lines {
		if len(line) >= 2 {
			phone := PhoneNumber(line[0])
			branch := BranchCode(line[1])
			phoneBranchMap[phone] = branch
		}
	}

	return phoneBranchMap, nil
}

func setupLogFile(logFileName string) (*os.File, error) {
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	log.SetOutput(logFile)

	return logFile, nil
}
