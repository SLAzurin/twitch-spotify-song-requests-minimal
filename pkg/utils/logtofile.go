package utils

import (
	"log"
	"os"
	"time"
)

func LogToFile(description string, err error, stack string) {
	s := description + "\n" + err.Error() + "\n" + stack + "\n"

	log.Println("Log to file", s)
	if _, err := os.Stat("./errorlogs"); os.IsNotExist(err) {
		os.Mkdir("./errorlogs", os.ModePerm)
	}
	f, err := os.Create("./errorlogs/" + time.Now().Format("2006-01-02T15:04:05") + ".log")
	if err != nil {
		log.Println("Cannot create error log")
		return
	}
	_, err = f.WriteString(s)
	if err != nil {
		log.Println("Cannot write to logfile")
	}
	defer f.Close()
}
