package initbundle

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sc-js/core_backend/src/bundles/authbundle"
	"github.com/sc-js/core_backend/src/bundles/cachebundle"
	"github.com/sc-js/core_backend/src/bundles/deepcorebundle"
	"github.com/sc-js/core_backend/src/bundles/localizationbundle"
	"github.com/sc-js/core_backend/src/errs"
	"github.com/sc-js/core_backend/src/mongowrap"
	"github.com/sc-js/core_backend/src/tools"
	"github.com/sc-js/pour"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var wrap *tools.DataWrap
var r *gin.Engine
var gr *gin.RouterGroup
var wsr *gin.RouterGroup
var autoMigrate bool
var SystemConfig Config
var initConf InitConfiguration
var isDocker = false

type Bundle struct {
	Handler  func(*gin.RouterGroup, *tools.DataWrap, bool, map[string]string)
	Settings map[string]string
}

func InitializeCoreWithBundles(bundles []Bundle, conf *InitConfiguration) {
	defer errs.Defer()
	time.Sleep(time.Second)
	dockerFlag := flag.Bool("docker", false, "Running in docker")
	flag.Parse()
	isDocker = *dockerFlag
	if err := godotenv.Load(); err != nil {
		pour.LogColor(false, pour.ColorYellow, "Error loading .env file")
	}

	handleInitConf(conf)
	readConfig()
	tools.Init(SystemConfig.Salt)
	tools.SetDocker(*dockerFlag)

	pour.Setup(isDocker)

	pour.LogColor(false, pour.ColorCyan, "Docker:", *dockerFlag)

	//Connect PostgreSQL DB and optionally Mongo
	wrap = getDataWrap()

	deepcorebundle.Init(wrap.DB, autoMigrate)

	localizationbundle.InitLocales()

	//HTTP Router
	gin.SetMode(initConf.GinMode)
	r = gin.New()
	r.Use(cors.Default())
	r.SetTrustedProxies(nil)
	r.MaxMultipartMemory = initConf.MaxMultipartMemory << 20
	gr = r.Group("")
	gr.Use(timeoutMiddleware())
	gr.Use(authbundle.AuthMiddleware(wrap.DB))
	wsr = r.Group("/ws")

	//Cache Engine
	switch strings.ToLower(SystemConfig.Cache.CacheEngine) {
	case ("aerospike"):
		cachebundle.InitCache(cachebundle.AeroSpike, SystemConfig.Cache.Address, SystemConfig.Cache.PortOverride, SystemConfig.Cache.Username, SystemConfig.Cache.Password, SystemConfig.Cache.Workspace)
	case ("redis"):
		cachebundle.InitCache(cachebundle.Redis, SystemConfig.Cache.Address, SystemConfig.Cache.PortOverride, SystemConfig.Cache.Username, SystemConfig.Cache.Password, SystemConfig.Cache.Workspace)
	}

	pour.LogColor(false, pour.ColorBlue, "Initializing", len(bundles), "external bundle(s)..")
	bundleNames := []string{"auth"}
	for _, element := range bundles {
		bundleName := strings.Split(tools.GetPackageName(element.Handler), "bundle")[0]
		bundleNames = append(bundleNames, bundleName)
		if bundleName == "websocket" {
			_, wrap, migrate, settings := getBundleRequirements(element.Settings)
			element.Handler(wsr, wrap, migrate, settings)
			continue
		}
		element.Handler(getBundleRequirements(element.Settings))
	}
	pour.LogColor(false, pour.ColorBlue, "Bundles initialized:", bundleNames)
}

func handleInitConf(c *InitConfiguration) {
	path := "./config.json"

	if isDocker {
		path = "./config_docker.json"
		pour.LogColor(false, pour.ColorCyan, "Loading docker-configured config:", path)
	} else {
		pour.LogColor(false, pour.ColorCyan, "Loading config:", path)
	}

	if c == nil {
		initConf = InitConfiguration{GinMode: gin.ReleaseMode, ConfigPath: path, MaxMultipartMemory: 8, DBLoggerConfig: logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Warn, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Disable color
		}}
	} else {
		initConf = *c
	}
}

func getBundleRequirements(settings map[string]string) (*gin.RouterGroup, *tools.DataWrap, bool, map[string]string) {
	return gr, wrap, autoMigrate, settings
}

func Run(settings map[string]string) {
	settings = setupSettings(settings)
	authbundle.InitBundle(gr, wrap, true, settings)
	pour.LogColor(false, pour.ColorGreen, "Running non-TLS server at", SystemConfig.Server.Host+":"+fmt.Sprint(SystemConfig.Server.Port))

	srv := &http.Server{
		ReadTimeout:  6 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Addr:         fmt.Sprint(SystemConfig.Server.Host, ":", SystemConfig.Server.Port),
		Handler:      r,
	}

	err := srv.ListenAndServe()
	pour.LogPanicKill(1, err)
}

func setupSettings(settings map[string]string) map[string]string {
	if settings == nil {
		settings = make(map[string]string)
	}
	settings["jwt_secret"] = SystemConfig.JWTSecret
	settings["jwt_refresh_secret"] = SystemConfig.JWTRefreshSecret

	if len(SystemConfig.JWTSecret) == 0 || len(SystemConfig.JWTRefreshSecret) == 0 {
		pour.LogPanicKill(1, errors.New("JWT secret or JWT refresh secret was empty, please check config"))
	}

	return settings
}

func RunTLS(settings map[string]string, generate bool) {
	settings = setupSettings(settings)
	if generate {
		tools.GenerateTLS()
	}

	certFile := *flag.String("certfile", "cert.pem", "certificate PEM file")
	keyFile := *flag.String("keyfile", "key.pem", "key PEM file")

	authbundle.InitBundle(gr, wrap, true, settings)
	pour.LogColor(false, pour.ColorGreen, "Running TLS server at", SystemConfig.Server.Host+":"+fmt.Sprint(SystemConfig.Server.Port))

	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
		},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}
	srv := &http.Server{
		ReadTimeout:  6 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Addr:         fmt.Sprint(SystemConfig.Server.Host, ":", SystemConfig.Server.Port),
		Handler:      r,
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

	err := srv.ListenAndServeTLS(certFile, keyFile)
	pour.LogPanicKill(1, err)
}

func getDataWrap() *tools.DataWrap {
	db := initialMigration()
	wrap := tools.DataWrap{DB: db}

	if mongo, err := initMongo(); err == nil {
		wrap.Mongo = mongo
	} else {
		pour.Log("Mongo failed:", err)
	}

	return &wrap
}

func initMongo() (*mongowrap.Mongo, error) {
	if SystemConfig.Mongo.Port == 0 || len(SystemConfig.Mongo.Address) == 0 {
		return nil, errors.New("mongo config not specified")
	}

	mngClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(buildMongoUri(SystemConfig)))
	if err != nil {
		return nil, err
	}
	ctx, ctxCancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer ctxCancel()

	mongoClient := &mongowrap.Mongo{Client: mngClient, Database: mngClient.Database("data")}
	err = mongoClient.Client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}
	pour.LogTagged(false, pour.TAG_SUCCESS, "MongoDB connected at", SystemConfig.Mongo.Address+":"+fmt.Sprint(SystemConfig.Mongo.Port))
	return mongoClient, nil
}

func buildMongoUri(config Config) string {
	return "mongodb://" + config.Mongo.Username + ":" + config.Mongo.Password + "@" + config.Mongo.Address + ":" + fmt.Sprint(config.Mongo.Port) + "/?maxPoolSize=10000&w=majority"
}

func initialMigration() *gorm.DB {
	defer errs.Defer()
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		initConf.DBLoggerConfig,
	)

	var err error
	dns := "host=" + SystemConfig.Database.Address + " user=" + SystemConfig.Database.Username + " password=" + SystemConfig.Database.Password + " dbname=" + SystemConfig.Database.Name + " port=" + fmt.Sprint(SystemConfig.Database.Port) + " sslmode=disable TimeZone=Europe/Berlin"
	db, err := gorm.Open(postgres.Open(dns), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		pour.LogPanicKill(1, fmt.Sprint("Cannot connect to DB at", dns))
	} else {
		pour.LogColor(false, pour.ColorPurple, "Connected to PostgreSQL")
	}
	return db
}

func readConfig() {
	content, err := os.ReadFile(initConf.ConfigPath)
	if err != nil {
		pour.LogPanicKill(1, err)
	}
	config := Config{}
	if err := json.Unmarshal(content, &config); err != nil {
		pour.LogPanicKill(1, err)
	}
	config = putDefaultConfigValues(config)
	autoMigrate = config.AutoMigrate
	SystemConfig = config
}

func putDefaultConfigValues(config Config) Config {
	if len(config.Cache.Workspace) == 0 {
		config.Cache.Workspace = cachebundle.AerospikeDefaultWorkspace
	}

	return config
}

func timeoutMiddleware() gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(5000*time.Millisecond),
		timeout.WithHandler(func(c *gin.Context) {
			c.Next()
		}),
		timeout.WithResponse(timeoutResponse),
	)
}

func timeoutResponse(c *gin.Context) {
	tools.RespondError(errors.New("timeout"), http.StatusRequestTimeout, c)
}
