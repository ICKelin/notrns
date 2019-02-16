package ddns

import (
	"fmt"
	"log"
)

func LogErr(format string, v ...interface{}) {
	log.Printf(fmt.Sprintf("[ERROR] %s\n", format), v...)
}

func LogInfo(format string, v ...interface{}) {
	log.Printf(fmt.Sprintf("[INF] %s\n", format), v...)
}

func LogWarn(format string, v ...interface{}) {
	log.Printf(fmt.Sprintf("[WARN] %s\n", format), v...)
}
