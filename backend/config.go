package main

import (
	"flag"
	"log"
)

type dbConfig struct {
	Host string
	Port int
	User string
	Pass string
	Name string
}

type tracksysURLs struct {
	Client string
	API    string
	Jobs   string
}

type configData struct {
	port        int
	db          dbConfig
	imagesDir   string
	scanDir     string
	finalizeDir string
	iiifURL     string
	serviceURL  string
	tracksys    tracksysURLs
	jwtKey      string
	devAuthUser string
}

func getConfiguration() *configData {
	var config configData
	flag.IntVar(&config.port, "port", 8080, "Port to offer service on (default 8085)")
	flag.StringVar(&config.imagesDir, "images", "/digiserv-production/dpg_imaging", "Images directory")
	flag.StringVar(&config.scanDir, "scan", "/digiserv-production/scan", "Scanning directory")
	flag.StringVar(&config.finalizeDir, "finalize", " /digiserv-production/finalization", "Finalization directory")
	flag.StringVar(&config.iiifURL, "iiif", "", "IIIF server URL")
	flag.StringVar(&config.serviceURL, "url", "", "Base URL for DPG Imaging service")
	flag.StringVar(&config.jwtKey, "jwtkey", "", "JWT signature key")

	// tracksys config
	flag.StringVar(&config.tracksys.API, "tsapiurl", "https://tracksys-api-ws.internal.lib.virginia.edu/api", "URL for TrackSysAPI service")
	flag.StringVar(&config.tracksys.Client, "tsurl", "https://tracksys.lib.virginia.edu", "URL for TrackSys")
	flag.StringVar(&config.tracksys.Jobs, "jobsurl", "", "URL for dpg-jobs processing")

	// DB connection params
	flag.StringVar(&config.db.Host, "dbhost", "", "Database host")
	flag.IntVar(&config.db.Port, "dbport", 3306, "Database port")
	flag.StringVar(&config.db.Name, "dbname", "", "Database name")
	flag.StringVar(&config.db.User, "dbuser", "", "Database user")
	flag.StringVar(&config.db.Pass, "dbpass", "", "Database password")

	// dev setup
	flag.StringVar(&config.devAuthUser, "devuser", "", "Authorized computing id for dev")

	flag.Parse()

	if config.jwtKey == "" {
		log.Fatal("Parameter jwtkey is required")
	}
	if config.imagesDir == "" {
		log.Fatal("images param is required")
	}
	if config.finalizeDir == "" {
		log.Fatal("finalize param is required")
	}
	if config.iiifURL == "" {
		log.Fatal("iiif param is required")
	}
	if config.serviceURL == "" {
		log.Fatal("url param is required")
	}
	if config.db.Host == "" {
		log.Fatal("Parameter dbhost is required")
	}
	if config.db.Name == "" {
		log.Fatal("Parameter dbname is required")
	}
	if config.db.User == "" {
		log.Fatal("Parameter dbuser is required")
	}
	if config.db.Pass == "" {
		log.Fatal("Parameter dbpass is required")
	}

	if config.tracksys.API == "" {
		log.Fatal("Parameter tsapiurl is required")
	}
	if config.tracksys.Client == "" {
		log.Fatal("Parameter tsurl is required")
	}
	if config.tracksys.Jobs == "" {
		log.Fatal("Parameter jobsurl is required")
	}

	log.Printf("[CONFIG] port          = [%d]", config.port)
	log.Printf("[CONFIG] imagesDir     = [%s]", config.imagesDir)
	log.Printf("[CONFIG] scanDir       = [%s]", config.scanDir)
	log.Printf("[CONFIG] finalizeDir   = [%s]", config.finalizeDir)
	log.Printf("[CONFIG] iiifURL       = [%s]", config.iiifURL)
	log.Printf("[CONFIG] serviceURL    = [%s]", config.serviceURL)
	log.Printf("[CONFIG] tracksysAPI   = [%s]", config.tracksys.API)
	log.Printf("[CONFIG] tracksysURL   = [%s]", config.tracksys.Client)
	log.Printf("[CONFIG] jobsURL       = [%s]", config.tracksys.Jobs)
	log.Printf("[CONFIG] dbhost        = [%s]", config.db.Host)
	log.Printf("[CONFIG] dbport        = [%d]", config.db.Port)
	log.Printf("[CONFIG] dbname        = [%s]", config.db.Name)
	log.Printf("[CONFIG] dbuser        = [%s]", config.db.User)

	return &config
}
