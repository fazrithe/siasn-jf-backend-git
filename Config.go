package main

import (
	"encoding/json"
)

type Config struct {
	// Listen address is an array of IP addresses and port combinations.
	// Listen address is an array so that this service can listen to many interfaces at once.
	// You can use this value for example: []string{"192.168.1.12:80", "25.49.25.73:80"} to listen to
	// listen to interfaces with IP address of 192.168.1.12 and 25.49.25.73, both on port 80.
	ListenAddress []string `config:"LISTEN_ADDRESS"`

	// Set true to enable TLS listener (and its configurations).
	EnableTls bool `config:"ENABLE_TLS"`

	// Listen address is an array of IP addresses and port combinations.
	// This is for TLS connection (HTTPS).
	TlsListenAddress []string `config:"TLS_LISTEN_ADDRESS"`

	// The path to TLS certificate file.
	TlsCertFile string `config:"TLS_CERT_FILE"`

	// The path to TLS private key file.
	TlsKeyFile string `config:"TLS_KEY_FILE"`

	PrometheusListenAddress string `config:"PROMETHEUS_LISTEN_ADDRESS"`

	CorsAllowedHeaders []string `config:"CORS_ALLOWED_HEADERS"`
	CorsAllowedMethods []string `config:"CORS_ALLOWED_METHODS"`
	CorsAllowedOrigins []string `config:"CORS_ALLOWED_ORIGINS"`

	// The full PostgreSQL URL, starting with `postgres://`.
	PostgresUrl string `config:"POSTGRES_URL"`
	// The full PostgreSQL URL, starting with `postgres://`.
	// This config is for read only access of ASN profile database provided by BKN.
	ProfilePostgresUrl string `config:"PROFILE_POSTGRES_URL"`
	// The full PostgreSQL URL, starting with `postgres://`.
	// This config is for read only access to reference database provided by BKN.
	ReferencePostgresUrl string `config:"REFERENCE_POSTGRES_URL"`

	// OidcProviderUrl is the base URL of the identity provider.
	// It is used to retrieve OpenID Connect discovery settings, which is available under <OidcProviderUrl>/.well-known.
	OidcProviderUrl        string `config:"OIDC_PROVIDER_URL"`
	OidcClientId           string `config:"OIDC_CLIENT_ID"`
	OidcClientSecret       string `config:"OIDC_CLIENT_SECRET"`
	OidcEndSessionEndpoint string `config:"OIDC_END_SESSION_ENDPOINT"`
	// OidcRedirectUrl is the URL registered as redirect URL in the identity provider.
	OidcRedirectUrl string `config:"OIDC_REDIRECT_URL"`
	// OidcSuccessRedirectUrl is the URL to which the user is redirected after successful login attempt.
	OidcSuccessRedirectUrl string `config:"OIDC_SUCCESS_REDIRECT_URL"`
	OidcSessionBackend     string `config:"OIDC_SESSION_BACKEND"`

	RedisAddress  string `config:"REDIS_ADDRESS"`
	RedisUsername string `config:"REDIS_USERNAME"`
	RedisPassword string `config:"REDIS_PASSWORD"`
	RedisDbIndex  int    `config:"REDIS_DB_INDEX"`

	EmcEcsEndpoint  string `config:"EMC_ECS_ENDPOINT"`
	EmcEcsAccessKey string `config:"EMC_ECS_ACCESS_KEY"`
	EmcEcsSecretKey string `config:"EMC_ECS_SECRET_KEY"`
	EmcEcsRegion    string `config:"EMC_ECS_REGION"`
	// Bucket name to store temporary files.
	TempBucket string `config:"TEMP_BUCKET"`
	// Directory relative to TEMP_BUCKET without leading/trailing slash to store temporary activity files.
	TempActivityDir string `config:"TEMP_ACTIVITY_DIR"`
	// Bucket name to store activity support doc files.
	ActivityBucket string `config:"ACTIVITY_BUCKET"`
	// Directory relative to ACTIVITY_BUCKET without leading/trailing slash to store activity files.
	ActivityDir string `config:"ACTIVITY_DIR"`
	// Directory relative to TEMP_BUCKET without leading/trailing slash to store temporary requirement files.
	TempRequirementDir string `config:"TEMP_REQUIREMENT_DIR"`
	// Bucket name to store requirement support doc files.
	RequirementBucket string `config:"REQUIREMENT_BUCKET"`
	// Directory relative to REQUIREMENT_BUCKET without leading/trailing slash to store requirement files.
	RequirementDir string `config:"REQUIREMENT_DIR"`
	// Directory for storing template documents.
	RequirementTemplateDir string `config:"REQUIREMENT_TEMPLATE_DIR"`
	// The filename of cover letter template including its extension.
	RequirementTemplateCoverLetterFilename string `config:"REQUIREMENT_TEMPLATE_COVER_LETTER_FILENAME"`
	// Directory relative to TEMP_BUCKET without leading/trailing slash to store temporary dismissal files.
	TempDismissalDir string `config:"TEMP_DISMISSAL_DIR"`
	// Bucket name to store dismissal support doc files.
	DismissalBucket string `config:"DISMISSAL_BUCKET"`
	// Directory relative to DISMISSAL_BUCKET without leading/trailing slash to store dismissal files.
	DismissalDir string `config:"DISMISSAL_DIR"`
	// Directory for storing template documents.
	DismissalTemplateDir string `config:"DISMISSAL_TEMPLATE_DIR"`
	// The filename of acceptance letter template including its extension.
	DismissalTemplateAcceptanceLetterFilename string `config:"DISMISSAL_TEMPLATE_ACCEPTANCE_LETTER_FILENAME"`
	// Bucket name to store promotion support doc files.
	PromotionBucket string `config:"PROMOTION_BUCKET"`
	// Directory relative to TEMP_BUCKET without leading/trailing slash to store temporary promotion files.
	TempPromotionDir string `config:"TEMP_PROMOTION_DIR"`
	// Directory relative to PROMOTION_BUCKET without leading/trailing slash to store promotion files.
	PromotionDir string `config:"PROMOTION_DIR"`
	// Directory for storing template documents.
	PromotionTemplateDir string `config:"PROMOTION_TEMPLATE_DIR"`
	// The filename of PAK letter template including its extension.
	PromotionTemplatePakLetterFilename string `config:"PROMOTION_TEMPLATE_PAK_LETTER_FILENAME"`
	// The filename of recommendation letter template including its extension.
	PromotionTemplateRecommendationLetterFilename string `config:"PROMOTION_TEMPLATE_RECOMMENDATION_LETTER_FILENAME"`
	// Bucket name to store promotion for CPNS support doc files.
	PromotionCpnsBucket string `config:"PROMOTION_CPNS_BUCKET"`
	// Directory relative to TEMP_PROMOTION_CPNS_BUCKET without leading/trailing slash to store temporary promotion files.
	TempPromotionCpnsDir string `config:"TEMP_PROMOTION_CPNS_DIR"`
	// Directory relative to PROMOTION_CPNS_BUCKET without leading/trailing slash to store promotion files.
	PromotionCpnsDir string `config:"PROMOTION_CPNS_DIR"`
	// Bucket name to store assessment team support doc files.
	AssessmentTeamBucket string `config:"ASSESSMENT_TEAM_BUCKET"`
	// Directory relative to TEMP_BUCKET without leading/trailing slash to store temporary assessment team files.
	TempAssessmentTeamDir string `config:"TEMP_ASSESSMENT_TEAM_DIR"`
	// Directory relative to ASSESSMENT_TEAM_BUCKET without leading/trailing slash to store assessment team files.
	AssessmentTeamDir string `config:"ASSESSMENT_TEAM_DIR"`

	// The command for siasn-docx binary.
	// Can be just a command name if the binary exists in PATH.
	SiasnDocxCmd string `config:"SIASN_DOCX_CMD"`
	// The command for soffice binary (libreoffice).
	// Can be just a command name if the binary exists in PATH.
	SofficeCmd string `config:"SOFFICE_CMD"`

	LoggingToStd    bool   `config:"LOGGING_TO_STD"`
	LoggingStdColor bool   `config:"LOGGING_STD_COLOR"`
	LoggingToFile   bool   `config:"LOGGING_TO_FILE"`
	LoggingFilePath string `config:"LOGGING_FILE_PATH"`
}

func NewConfigDefault() *Config {
	defaultConfig := &Config{
		ListenAddress:    []string{"127.0.0.1:8080"},
		TlsListenAddress: []string{"127.0.0.1:8443"},
		TlsCertFile:      "certs/localhost.crt",
		TlsKeyFile:       "certs/localhost.key",

		PrometheusListenAddress: "0.0.0.0:9080",

		CorsAllowedHeaders: []string{"Accept", "Accept-Language", "Content-Type", "Content-Language", "Content-Disposition", "Origin", "X-Requested-With", "X-Forwarded-For"},
		CorsAllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "PUT"},

		OidcProviderUrl:        "https://iam-siasn.bkn.go.id/auth/realms/public-siasn",
		OidcClientId:           "manajemen-jf",
		OidcEndSessionEndpoint: "https://iam-siasn.bkn.go.id/auth/realms/public-siasn/protocol/openid-connect/logout",
		OidcRedirectUrl:        "http://training-manajemen-jf.bkn.go.id/api/oauth",
		OidcSuccessRedirectUrl: "http://training-manajemen-jf.bkn.go.id",
		OidcSessionBackend:     "memory",

		RedisAddress: "127.0.0.1:6379",

		PostgresUrl:          "postgres://postgres:D3v45n@10.100.8.164/siasn_jf ",
		ProfilePostgresUrl:   "postgres://itb:kepitingberanakpaus@10.100.8.42:5432/db_profileasn",
		ReferencePostgresUrl: "postgres://itb:kepitingberanakpaus@10.100.8.42:5432/db_referensi",

		TempActivityDir:                               "activity",
		ActivityDir:                                   "activity",
		ActivityBucket:                                "dev-siasn21",
		TempRequirementDir:                            "requirement",
		RequirementDir:                                "requirement",
		RequirementTemplateDir:                        "requirement-template",
		RequirementTemplateCoverLetterFilename:        "surat-pengantar-template.docx",
		TempDismissalDir:                              "dismissal",
		DismissalDir:                                  "dismissal",
		DismissalTemplateDir:                          "dismissal-template",
		DismissalTemplateAcceptanceLetterFilename:     "surat-pemberhentian-template.docx",
		TempPromotionDir:                              "promotion",
		PromotionDir:                                  "promotion",
		PromotionTemplateDir:                          "promotion-template",
		PromotionTemplatePakLetterFilename:            "surat-pak-template.docx",
		PromotionTemplateRecommendationLetterFilename: "surat-rekomendasi-template.docx",
		PromotionCpnsDir:                              "promotion-cpns",
		TempPromotionCpnsDir:                          "promotion-cpns",
		AssessmentTeamDir:                             "assessment-team",
		TempAssessmentTeamDir:                         "assessment-team",

		SiasnDocxCmd: "siasn-docx",
		SofficeCmd:   "soffice",

		LoggingToStd:    true,
		LoggingStdColor: true,
		LoggingToFile:   true,
		LoggingFilePath: "logs/siasn-jf-backend.log",
	}

	return defaultConfig
}

func (c *Config) AsString() string {
	data, _ := json.Marshal(c)
	return string(data)
}
