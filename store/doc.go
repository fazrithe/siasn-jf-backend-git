// Package store provides codes to interact with the database, object storage, and so on.
// The subpackages created in store package have the purposes of helping to provide data for example
// by making an API call or interacting with object storage.
//
// The store package resolves around a single Client. The Client is used to interact with the database,
// object storage, and other APIs. This package specifically provides codes that interact with the database.
// Other subpackages provide codes to interact with other storages/APIs.
//
// This package also support the standard PotatoBeans error code convention. Error code is supported through the use
// of `ec` package from the standard SIASN library, and error code numbers are stored globally in errnum package. This
// standard convention that we have used can easily be extended or customized as you wish.
//
// Use the store package functionalities by creating a Client, by invoking NewClient:
//   client := store.NewClient(db, sqlMetrics, rcb)
package store

import (
	"github.com/if-itb/siasn-jf-backend/store/models"
	"time"
)

const (
	// ActivitySupportDocSubdir is the name of subdirectory for storing support documents uploading for activity
	// submission.
	ActivitySupportDocSubdir = "support"
	// ActivityCertSubdir is the name of subdirectory for storing activity certificates.
	ActivityCertSubdir = "cert"
	// ActivityPakSubdir is the name of subdirectory for storing PAK documents.
	ActivityPakSubdir = "pak"

	ActivityRecommendationLetterSubdir = "recommendation-letter"

	RequirementCoverLetterSubdir          = "cover-letter"
	RequirementEstimationDocSubdir        = "estimation"
	RequirementRecommendationLetterSubdir = "recommendation-letter"

	DismissalSupportDocSubdir       = "dismissal-support"
	DismissalAcceptanceLetterSubdir = "acceptance-letter"
	DismissalDenySupportDocSubdir   = "deny-support"

	PromotionPakLetterSubdir            = "pak"
	PromotionRecommendationLetterSubdir = "recommendation-letter"
	PromotionTestCertificateSubdir      = "test-certificate"
	PromotionPromotionLetterSubdir      = "promotion-letter"

	PromotionCpnsPakLetterSubdir       = "pak"
	PromotionCpnsPromotionLetterSubdir = "promotion-cpns-letter"

	AssessmentTeamSupportDocSubdir           = "support"
	AssessmentTeamRecommendationLetterSubdir = "recommendation"
)

var ActivityCertTypeToSubdir = map[int]string{
	models.ActivityCertTypeCert: ActivityCertSubdir,
	models.ActivityCertTypePak:  ActivityPakSubdir,
}

const TimeoutDefault = 15 * time.Second
