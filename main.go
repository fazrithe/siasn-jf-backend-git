package main

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EMCECS/ecs-object-client-go"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/fazrithe/siasn-jf-backend-git/libs/auth"
	"github.com/fazrithe/siasn-jf-backend-git/libs/breaker"
	"github.com/fazrithe/siasn-jf-backend-git/libs/config"
	"github.com/fazrithe/siasn-jf-backend-git/libs/docx"
	"github.com/fazrithe/siasn-jf-backend-git/libs/httputil"
	"github.com/fazrithe/siasn-jf-backend-git/libs/logutil"
	"github.com/fazrithe/siasn-jf-backend-git/libs/metricutil"
	"github.com/fazrithe/siasn-jf-backend-git/store"
	"github.com/fazrithe/siasn-jf-backend-git/store/object"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// version variable is left empty as it is populated in compile time by the linker.
// To populate this variable during compile time, add -X flag to the linker using -ldflags go build command:
//
//	go build -ldflags "-X main.version=0.1.0"
//
// This way, during runtime, this variable will contain a string `0.1.0`. For more information about -X flag, visit
// Go linker documentation https://golang.org/cmd/link/.
//
//goland:noinspection GoUnusedGlobalVariable
var version string

const (
	// OidcSessionBackendMemory initiates the auth handler to store sessions in memory.
	// Sessions will be lost when the service is restarted.
	OidcSessionBackendMemory = "memory"
	// OidcSessionBackendRedis initiates the auth handler to store sessions in Redis.
	// REDIS_ADDRESS config must be configured.
	OidcSessionBackendRedis = "redis"
)

func main() {
	logutil.SetDefaultLogger(logutil.NewStdLogger(logutil.IsSupportColor(), "main"))
	globalConfig := NewConfigDefault()

	sc := &config.ServiceConfig{
		Prefix:         "SIASN_JF",
		ArraySeparator: " ",
	}
	err := sc.ParseTo(globalConfig)
	if err != nil {
		logutil.Errorf("cannot parse configuration: %v", err)
		os.Exit(1)
	}

	mustLoadConfig(globalConfig, sc)
	logutil.SetDefaultLogger(createLogger(globalConfig, "main"))
	logutil.Debugf("using configuration: %s", globalConfig.AsString())
	logutil.Infof("version: %s", version)

	rcb := &breaker.RateCircuitBreaker{
		Limit:    3,
		Cooldown: 10 * time.Second,
		Logger:   createLogger(globalConfig, "breaker"),
	}

	db, err := sql.Open("pgx", globalConfig.PostgresUrl)
	if err != nil {
		logutil.Errorf("cannot initiate connection to database: %v", err)
		os.Exit(1)
		return
	}
	defer db.Close()

	logutil.Tracef("pinging database for a maximum of 30 seconds")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Attempt to contact the database.
	err = db.PingContext(ctx)
	if err != nil {
		logutil.Errorf("cannot ping database: %v", err)
		os.Exit(1)
		return
	}
	logutil.Tracef("primary database can be pinged successfully")

	profileDb, err := sql.Open("pgx", globalConfig.ProfilePostgresUrl)
	if err != nil {
		logutil.Errorf("cannot initiate connection to profile database: %v", err)
		os.Exit(1)
		return
	}
	defer profileDb.Close()

	// Attempt to contact the database.
	err = profileDb.PingContext(ctx)
	if err != nil {
		logutil.Errorf("cannot ping profile database: %v", err)
		os.Exit(1)
		return
	}
	logutil.Tracef("profile database can be pinged successfully")

	referenceDb, err := sql.Open("pgx", globalConfig.ReferencePostgresUrl)
	if err != nil {
		logutil.Errorf("cannot initiate connection to profile database: %v", err)
		os.Exit(1)
		return
	}
	defer referenceDb.Close()

	// Attempt to contact the database.
	err = referenceDb.PingContext(ctx)
	if err != nil {
		logutil.Errorf("cannot ping profile database: %v", err)
		os.Exit(1)
		return
	}
	logutil.Tracef("reference database can be pinged successfully")

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(globalConfig.EmcEcsAccessKey, globalConfig.EmcEcsSecretKey, ""),
		Endpoint:         aws.String(globalConfig.EmcEcsEndpoint),
		Region:           aws.String(globalConfig.EmcEcsRegion),
		S3ForcePathStyle: aws.Bool(true),
	}

	sess, err := session.NewSession(s3Config)
	if err != nil {
		logutil.Errorf("cannot connect to object storage: %v", err)
		os.Exit(1)
		return
	}
	svc := s3.New(sess)

	storeClient := store.NewClient(db, profileDb, referenceDb, &object.EmcEcsStorage{
		Client:                                 ecs.New(svc),
		Logger:                                 createLogger(globalConfig, "emcecs"),
		TempBucket:                             globalConfig.TempBucket,
		TempActivityDir:                        globalConfig.TempActivityDir,
		ActivityBucket:                         globalConfig.ActivityBucket,
		ActivityDir:                            globalConfig.ActivityDir,
		TempRequirementDir:                     globalConfig.TempRequirementDir,
		RequirementBucket:                      globalConfig.RequirementBucket,
		RequirementDir:                         globalConfig.RequirementDir,
		RequirementTemplateDir:                 globalConfig.RequirementTemplateDir,
		RequirementTemplateCoverLetterFilename: globalConfig.RequirementTemplateCoverLetterFilename,
		TempDismissalDir:                       globalConfig.TempDismissalDir,
		DismissalBucket:                        globalConfig.DismissalBucket,
		DismissalDir:                           globalConfig.DismissalDir,
		DismissalTemplateDir:                   globalConfig.DismissalTemplateDir,
		DismissalTemplateAcceptanceLetterFilename: globalConfig.DismissalTemplateAcceptanceLetterFilename,
		PromotionBucket:                    globalConfig.PromotionBucket,
		TempPromotionDir:                   globalConfig.TempPromotionDir,
		PromotionDir:                       globalConfig.PromotionDir,
		PromotionTemplateDir:               globalConfig.PromotionTemplateDir,
		PromotionTemplatePakLetterFilename: globalConfig.PromotionTemplatePakLetterFilename,
		PromotionCpnsBucket:                globalConfig.PromotionCpnsBucket,
		TempPromotionCpnsDir:               globalConfig.TempPromotionCpnsDir,
		PromotionCpnsDir:                   globalConfig.PromotionCpnsDir,
		AssessmentTeamBucket:               globalConfig.AssessmentTeamBucket,
		TempAssessmentTeamDir:              globalConfig.TempAssessmentTeamDir,
		AssessmentTeamDir:                  globalConfig.AssessmentTeamDir,
		SignUrlExpire:                      1 * time.Hour,
	}, &docx.SiasnRenderer{
		DocxCmd:    globalConfig.SiasnDocxCmd,
		SofficeCmd: globalConfig.SofficeCmd,
	}, sqlMetrics, rcb)
	storeClient.Logger = createLogger(globalConfig, "store")

	authHandler, err := auth.NewAuth(
		globalConfig.OidcProviderUrl,
		profileDb,
		referenceDb,
		globalConfig.OidcClientId,
		globalConfig.OidcClientSecret,
		globalConfig.OidcEndSessionEndpoint,
		globalConfig.OidcRedirectUrl,
		globalConfig.OidcSuccessRedirectUrl,
	)
	if err != nil {
		logutil.Errorf("cannot initialize OpenID Connect configurations: %v", err)
		os.Exit(1)
		return
	}
	switch globalConfig.OidcSessionBackend {
	case OidcSessionBackendMemory:
		authHandler.AccessTokenCache = auth.NewMemoryAccessTokenCache()
	case OidcSessionBackendRedis:
		sessionCache := cache.New(&cache.Options{
			Redis: redis.NewClient(&redis.Options{
				Addr:     globalConfig.RedisAddress,
				Username: globalConfig.RedisUsername,
				Password: globalConfig.RedisPassword,
				DB:       globalConfig.RedisDbIndex,
			}),
		})
		redisTokenCache := auth.NewRedisAccessTokenCache(sessionCache)
		redisTokenCache.Prefix = "jf"
		authHandler.AccessTokenCache = redisTokenCache
	default:
		logutil.Errorf("cannot initialize session cache, unknown OIDC_SESSION_BACKEND: %s, can only be \"memory\" or \"redis\"", globalConfig.OidcSessionBackend)
		os.Exit(1)
		return
	}

	// Spawn the Prometheus metric server
	prometheusServer := metricutil.NewPrometheusServer(globalConfig.PrometheusListenAddress)
	prometheusServer.Logger = createLogger(globalConfig, "prometheus")
	go prometheusServer.Start()

	serverManager := httputil.NewServerManager()
	serverManager.Logger = createLogger(globalConfig, "serverManager")

	router := createRouter(
		globalConfig.CorsAllowedHeaders,
		globalConfig.CorsAllowedMethods,
		globalConfig.CorsAllowedOrigins,
		authHandler,
		storeClient,
		apiMetrics,
		&logutil.AccessLoggerWriter{Logger: createLogger(globalConfig, "apiRouter")},
	)
	serverManager.BatchSpawnHttp(globalConfig.ListenAddress, router)
	if globalConfig.EnableTls {
		serverManager.BatchSpawnHttps(globalConfig.TlsListenAddress, router, globalConfig.TlsCertFile, globalConfig.TlsKeyFile)
	}

	// Catch SIGTERM signal and shut down the system gracefully.
	// If the CB is tripped, also throw exit code 1 and add a 30 second timer to shut down the system forcefully.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-c:
			logutil.Warn("shutting down")
			serverManager.Shutdown()
		case <-rcb.Tripped:
			logutil.Warn("shutting down")
			go func() {
				logutil.Warn("forcing the system to shut down in 30 seconds")
				time.Sleep(30 * time.Second)
				os.Exit(1)
			}()
			serverManager.Shutdown()
			os.Exit(1)
		}
	}()

	time.Sleep(2 * time.Second)
	if serverManager.Err() != nil {
		c <- os.Interrupt
	}

	serverManager.Wait()
}

// mustLoadConfig loads the configuration, and terminates the process if configuration cannot be loaded.
func mustLoadConfig(cfg *Config, sc *config.ServiceConfig) {
	err := sc.ParseTo(cfg)
	if err != nil {
		logutil.Errorf("cannot parse configuration: %s", err)
		os.Exit(1)
	}

	logutil.Infof("configuration loaded: %v", *cfg)
	return
}

// createLogger creates an instance of logutil.Logger based on the given configuration and prefix.
// Depending on the given config, if the config demands log to be written also to a file, then this function
// will create a logger that outputs log to a file. If it demands logs only to be written to stdout, it will
// create an logutil.StdLogger. Either way the returned logger is a type of loguitl.MultiLogger which contains either
// an StdLogger, FileLogger, or both which will output logs to both outputs.
func createLogger(loggingConfig *Config, prefix string) logutil.Logger {
	var loggers []logutil.Logger

	if loggingConfig.LoggingToStd {
		loggers = append(loggers, logutil.NewStdLogger(loggingConfig.LoggingStdColor, prefix))
	}

	if loggingConfig.LoggingToFile {
		loggers = append(loggers, logutil.NewFileLogger(loggingConfig.LoggingFilePath, prefix))
	}

	return logutil.NewMultiLogger(loggers...)
}
