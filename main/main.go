package main

//import "fmt"
import (
	//"library_api/handlers"
	"net/http"
	"fmt"
	"library_api/handlers"
	//"go.uber.org/zap"
	//"github.com/docker/docker/daemon/logger"
	//"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/log"
	//"io"
	"os"
)
//
//func init() {
//	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
//	handlers.Config.Logger = logger
//}
func main() {


	router := handlers.NewMux()

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	handlers.Config.Logger = logger


	//level.Error(logger).Log("err", errors.New("bad data"))
	//level.Info(logger).Log("event", "data saved")
	//level.Debug(logger).Log("next item", 17) // filtered

	err := http.ListenAndServe(":2020",  router)

	if err != nil {
		fmt.Print(err)
	}

	//fmt.Print("hello, golang!")
}
