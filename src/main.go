package main

import (
	"flag"
	"os"
	"runtime"

	"github.com/google/gops/agent"
	"github.com/kafkaesque-io/prometheus-pushproxy/src/route"
	"github.com/kafkaesque-io/pulsar-beam/src/util"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

var mode = util.AssignString(os.Getenv("ProcessMode"), *flag.String("mode", "hybrid", "server running mode"))

func main() {
	// runtime.GOMAXPROCS does not the container's CPU quota in Kubernetes
	// therefore, it requires to be set explicitly
	runtime.GOMAXPROCS(util.GetEnvInt("GOMAXPROCS", 1))

	// gops debug instrument
	if err := agent.Listen(agent.Options{}); err != nil {
		log.Panicf("gops instrument error %v", err)
	}

	exit := make(chan bool) // future use to exit the main program if in broker only mode
	util.Init()

	flag.Parse()
	log.Warnf("start server mode %s", mode)
	if !util.IsValidMode(&mode) {
		log.Panic("Unsupported server mode")
	}

	route.Init()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8085", "http://localhost:8080"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "PulsarTopicUrl"},
	})

	router := route.NewRouter(&mode)

	handler := c.Handler(router)
	config := util.GetConfig()
	port := util.AssignString(config.PORT, "8085")
	certFile := util.GetConfig().CertFile
	keyFile := util.GetConfig().KeyFile
	log.Fatal(util.ListenAndServeTLS(":"+port, certFile, keyFile, handler))

}