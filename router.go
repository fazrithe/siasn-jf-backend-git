package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	. "github.com/if-itb/siasn-jf-backend/errnum"
	"github.com/if-itb/siasn-jf-backend/store"
	"github.com/if-itb/siasn-libs-backend/auth"
	"github.com/if-itb/siasn-libs-backend/ec"
	"github.com/if-itb/siasn-libs-backend/httputil"
	"github.com/if-itb/siasn-libs-backend/metricutil"
)

func createRouter(
	corsAllowedHeaders []string,
	corsAllowedMethods []string,
	corsAllowedOrigins []string,
	authHandler *auth.Auth,
	storeClient *store.Client,
	metrics metricutil.GenericApiMetricsPerUrl,
	logWriter io.Writer,
) (router *mux.Router) {
	router = mux.NewRouter()
	router.Use(
		// Create a logger wrapper
		func(handler http.Handler) http.Handler {
			return handlers.CombinedLoggingHandler(logWriter, handler)
		},
		handlers.CORS(
			handlers.AllowCredentials(),
			handlers.AllowedHeaders(corsAllowedHeaders),
			handlers.AllowedMethods(corsAllowedMethods),
			handlers.AllowedOrigins(corsAllowedOrigins),
		),
		handlers.RecoveryHandler(handlers.PrintRecoveryStack(true)),
	)
	router.NotFoundHandler = handlers.CombinedLoggingHandler(logWriter, http.NotFoundHandler())
	router.HandleFunc("/api/version", handleVersion).Methods("GET")
	// router.HandleFunc("/api/v2/sign/submit", storeClient.HandleSignSubmit).Methods("post")
	// router.HandleFunc("/api/v2/sign/req", storeClient.HandleSignReq).Methods("post")
	// router.HandleFunc("/api/v2/sign/tte", storeClient.HandleSignTte).Methods("post")
	router.HandleFunc("/api/login", authHandler.LoginHandler)
	router.HandleFunc("/api/oauth", authHandler.OidcHandler)
	router.HandleFunc("/api/logout", authHandler.LogoutHandler)

	apiV1 := router.PathPrefix("/api/v1").Subrouter()
	apiV1.Use(
		// Create generic API wrapper
		func(handler http.Handler) http.Handler {
			return metricutil.GenericApiMetricsPerUrlWrapper(handler, metrics)
		},
		authHandler.UserExtendedAuthHandler,
		authHandler.UserDetailAuthHandler,
	)

	activityV1 := apiV1.PathPrefix("/activity").Subrouter()

	activityV1.HandleFunc("/statistic/status/get", storeClient.HandleGetActivityStatusStatistic).Methods("GET")

	activityV1.HandleFunc("/admission/submit", storeClient.HandleActivityAdmissionSubmit).Methods("POST")
	activityV1.HandleFunc("/admission/upload", storeClient.HandleActivityAdmissionSupportDocUpload).Methods("POST")
	activityV1.HandleFunc("/admission/preview", storeClient.HandleActivityAdmissionSupportDocPreview).Methods("GET")
	activityV1.HandleFunc("/admission/download", storeClient.HandleActivityAdmissionSupportDocDownload).Methods("GET")
	activityV1.HandleFunc("/admission/search-asn", storeClient.HandleActivityAdmissionAsnGet).Methods("GET")
	activityV1.HandleFunc("/admission/verify", storeClient.HandleActivityVerificationSet).Methods("POST")
	activityV1.HandleFunc("/admission/search", storeClient.HandleActivityAdmissionSearch).Methods("GET")
	activityV1.HandleFunc("/admission/search/paginated", storeClient.HandleActivityAdmissionSearchPaginated).Methods("GET")
	activityV1.HandleFunc("/admission/search-pembina", storeClient.HandleActivityAdmissionSearchPembina).Methods("GET")
	activityV1.HandleFunc("/admission/detail", storeClient.HandleActivityAdmissionDetail).Methods("GET")
	activityV1.HandleFunc("/admission/get", storeClient.HandleActivityAdmissionDetail).Methods("GET")

	activityV1.HandleFunc("/admission/upload/recommendation-letter", storeClient.HandleActivityRecommendationLetterUpload).Methods("POST")
	activityV1.HandleFunc("/admission/preview/recommendation-letter", storeClient.HandleActivityRecommendationLetterPreview).Methods("GET")
	activityV1.HandleFunc("/admission/submit/recommendation-letter", storeClient.HandleActivityRecommendationLetterSubmit).Methods("POST")
	activityV1.HandleFunc("/admission/download/recommendation-letter", storeClient.HandleActivityRecommendationLetterDownload).Methods("GET")

	activityV1.HandleFunc("/csr/submit", storeClient.HandleActivityStatusCsrSubmit).Methods("POST")

	activityV1.HandleFunc("/certgen/upload", storeClient.HandleActivityCertGenDocUpload).Methods("POST")
	activityV1.HandleFunc("/certgen/preview", storeClient.HandleActivityCertGenDocPreview).Methods("GET")
	activityV1.HandleFunc("/certgen/download", storeClient.HandleActivityCertGenDocDownload).Methods("GET")
	activityV1.HandleFunc("/certgen/submit", storeClient.HandleActivityCertGenSubmit).Methods("POST")

	requirementV1 := apiV1.PathPrefix("/requirement").Subrouter()

	requirementV1.HandleFunc("/statistic/status/get", storeClient.HandleGetRequirementStatusStatistic).Methods("GET")

	requirementAdmissionV1 := requirementV1.PathPrefix("/admission").Subrouter()
	requirementAdmissionV1.HandleFunc("/submit", storeClient.HandleRequirementAdmissionSubmit).Methods("POST")
	requirementAdmissionV1.HandleFunc("/edit", storeClient.HandleRequirementAdmissionEdit).Methods("PUT")
	requirementAdmissionV1.HandleFunc("/upload/cover-letter", storeClient.HandleRequirementAdmissionCoverLetterUpload).Methods("POST")
	requirementAdmissionV1.HandleFunc("/upload/estimation", storeClient.HandleRequirementAdmissionEstimationDocUpload).Methods("POST")
	requirementAdmissionV1.HandleFunc("/preview/cover-letter", storeClient.HandleRequirementAdmissionCoverLetterPreview).Methods("GET")
	requirementAdmissionV1.HandleFunc("/preview/estimation", storeClient.HandleRequirementAdmissionEstimationDocPreview).Methods("GET")
	requirementAdmissionV1.HandleFunc("/download/cover-letter", storeClient.HandleRequirementAdmissionCoverLetterDownload).Methods("GET")
	requirementAdmissionV1.HandleFunc("/template/cover-letter", storeClient.HandleRequirementAdmissionCoverLetterTemplateDownload).Methods("GET")
	requirementAdmissionV1.HandleFunc("/download/estimation", storeClient.HandleRequirementAdmissionEstimationDocDownload).Methods("GET")
	requirementAdmissionV1.HandleFunc("/search", storeClient.HandleRequirementAdmissionSearch).Methods("GET")
	requirementAdmissionV1.HandleFunc("/search/paginated", storeClient.HandleRequirementAdmissionSearchPaginated).Methods("GET")
	requirementAdmissionV1.HandleFunc("/detail", storeClient.HandleRequirementAdmissionDetailGet).Methods("GET")
	requirementAdmissionV1.HandleFunc("/get", storeClient.HandleRequirementAdmissionDetailGet).Methods("GET")

	requirementVerifyV1 := requirementV1.PathPrefix("/verify").Subrouter()
	requirementVerifyV1.Handle("/upload", endpointRemovedHandler("/api/v1/requirement/verify/upload/recommendation-letter")).Methods("PUT")
	requirementVerifyV1.Handle("/download", endpointRemovedHandler("/api/v1/requirement/verify/download/recommendation-letter")).Methods("GET")
	requirementVerifyV1.Handle("/preview", endpointRemovedHandler("/api/v1/requirement/verify/preview/recommendation-letter")).Methods("GET")
	requirementVerifyV1.HandleFunc("/upload/recommendation-letter", storeClient.HandleRequirementVerificationRecommendationLetterUpload).Methods("POST")
	requirementVerifyV1.HandleFunc("/download/recommendation-letter", storeClient.HandleRequirementVerificationRecommendationLetterDownload).Methods("GET")
	requirementVerifyV1.HandleFunc("/preview/recommendation-letter", storeClient.HandleRequirementVerificationRecommendationLetterPreview).Methods("GET")
	requirementVerifyV1.HandleFunc("/download/recommendation-letter/unsigned", storeClient.HandleRequirementVerificationRecommendationLetterDownloadUnsigned).Methods("GET")
	requirementVerifyV1.HandleFunc("/download/recommendation-letter/signed", storeClient.HandleRequirementVerificationRecommendationLetterDownloadSigned).Methods("GET")
	requirementVerifyV1.HandleFunc("/bulk-submit/recommendation-letter", storeClient.HandleRequirementVerificationBulkSubmitRecommendationLetter).Methods("POST")
	requirementVerifyV1.HandleFunc("/sign/recommendation-letter", storeClient.HandleRequirementVerificationSignRecommendationLetter).Methods("POST")

	requirementVerifyV1.HandleFunc("/submit", storeClient.HandleRequirementVerificationSet).Methods("POST")
	requirementVerifyV1.HandleFunc("/deny", storeClient.HandleRequirementDenySet).Methods("POST")

	requirementV1.HandleFunc("/verifier/get", storeClient.HandleRequirementVerifiersGet).Methods("GET")

	dismissalV1 := apiV1.PathPrefix("/dismissal").Subrouter()

	dismissalV1.HandleFunc("/statistic/status/get", storeClient.HandleGetDismissalStatusStatistic).Methods("GET")

	dismissalAdmissionV1 := dismissalV1.PathPrefix("/admission").Subrouter()
	// HandleActivityAdmissionAsnGet is reused here.
	dismissalAdmissionV1.HandleFunc("/search-asn", storeClient.HandleActivityAdmissionAsnGet).Methods("GET")
	dismissalAdmissionV1.HandleFunc("/submit", storeClient.HandleDismissalAdmissionSubmit).Methods("POST")
	dismissalAdmissionV1.HandleFunc("/upload", storeClient.HandleDismissalAdmissionSupportDocUpload).Methods("POST")
	dismissalAdmissionV1.HandleFunc("/preview", storeClient.HandleDismissalAdmissionSupportDocPreview).Methods("GET")
	dismissalAdmissionV1.HandleFunc("/download", storeClient.HandleDismissalAdmissionSupportDocDownload).Methods("GET")
	dismissalAdmissionV1.HandleFunc("/get", storeClient.HandleDismissalAdmissionGet).Methods("GET")
	dismissalAdmissionV1.HandleFunc("/search", storeClient.HandleDismissalAdmissionsSearch).Methods("GET")
	dismissalAdmissionV1.HandleFunc("/search/paginated", storeClient.HandleDismissalAdmissionsSearchPaginated).Methods("GET")

	dismissalAcceptV1 := dismissalV1.PathPrefix("/accept").Subrouter()
	dismissalAcceptV1.HandleFunc("/submit", storeClient.HandleDismissalAcceptSet).Methods("POST")
	dismissalAcceptV1.HandleFunc("/download", storeClient.HandleDismissalAcceptanceLetterDownload).Methods("GET")

	dismissalDenyV1 := dismissalV1.PathPrefix("/deny").Subrouter()
	dismissalDenyV1.HandleFunc("/submit", storeClient.HandleDismissalDenySet).Methods("POST")
	dismissalDenyV1.HandleFunc("/upload", storeClient.HandleDismissalDenySupportDocUpload).Methods("POST")
	dismissalDenyV1.HandleFunc("/preview", storeClient.HandleDismissalDenySupportDocPreview).Methods("GET")
	dismissalDenyV1.HandleFunc("/download", storeClient.HandleDismissalDenySupportDocDownload).Methods("GET")

	promotionV1 := apiV1.PathPrefix("/promotion").Subrouter()

	promotionV1.HandleFunc("/statistic/status/get", storeClient.HandleGetPromotionStatusStatistic).Methods("GET")

	promotionAdmissionV1 := promotionV1.PathPrefix("/admission").Subrouter()
	promotionAdmissionV1.HandleFunc("/search-asn", storeClient.HandleActivityAdmissionAsnGet).Methods("GET")
	promotionAdmissionV1.HandleFunc("/submit", storeClient.HandlePromotionAdmissionSubmit).Methods("POST")
	promotionAdmissionV1.HandleFunc("/upload/pak", storeClient.HandlePromotionAdmissionPakLetterUpload).Methods("POST")
	promotionAdmissionV1.HandleFunc("/preview/pak", storeClient.HandlePromotionAdmissionPakLetterPreview).Methods("GET")
	promotionAdmissionV1.HandleFunc("/download/pak", storeClient.HandlePromotionPakLetterDownload).Methods("GET")
	promotionAdmissionV1.HandleFunc("/template/pak", storeClient.HandlePromotionAdmissionPakLetterTemplateDownload).Methods("GET")
	promotionAdmissionV1.HandleFunc("/upload/recommendation-letter", storeClient.HandlePromotionAdmissionRecommendationLetterUpload).Methods("POST")
	promotionAdmissionV1.HandleFunc("/preview/recommendation-letter", storeClient.HandlePromotionAdmissionRecommendationLetterPreview).Methods("GET")
	promotionAdmissionV1.HandleFunc("/download/recommendation-letter", storeClient.HandlePromotionRecommendationLetterDownload).Methods("GET")
	promotionAdmissionV1.HandleFunc("/template/recommendation-letter", storeClient.HandlePromotionAdmissionRecommendationLetterTemplateDownload).Methods("GET")
	promotionAdmissionV1.HandleFunc("/upload/promotion-letter", storeClient.HandlePromotionAdmissionPromotionLetterUpload).Methods("POST")
	promotionAdmissionV1.HandleFunc("/preview/promotion-letter", storeClient.HandlePromotionAdmissionPromotionLetterPreview).Methods("GET")
	promotionAdmissionV1.HandleFunc("/download/promotion-letter", storeClient.HandlePromotionAdmissionPromotionLetterDownload).Methods("GET")
	promotionAdmissionV1.HandleFunc("/upload/test-certificate", storeClient.HandlePromotionAdmissionTestCertificateUpload).Methods("POST")
	promotionAdmissionV1.HandleFunc("/preview/test-certificate", storeClient.HandlePromotionAdmissionTestCertificatePreview).Methods("GET")
	promotionAdmissionV1.HandleFunc("/download/test-certificate", storeClient.HandlePromotionAdmissionTestCertificateDownload).Methods("GET")
	promotionAdmissionV1.HandleFunc("/accept", storeClient.HandlePromotionAdmissionAccept).Methods("POST")
	promotionAdmissionV1.HandleFunc("/reject", storeClient.HandlePromotionAdmissionReject).Methods("POST")
	promotionAdmissionV1.HandleFunc("/search/paginated", storeClient.HandlePromotionAdmissionSearchPaginated).Methods("GET")

	promotionCpnsV1 := apiV1.PathPrefix("/promotion-cpns").Subrouter()

	promotionCpnsV1.HandleFunc("/statistic/status/get", storeClient.HandleGetPromotionCpnsStatusStatistic).Methods("GET")

	promotionCpnsAdmissionV1 := promotionCpnsV1.PathPrefix("/admission").Subrouter()
	promotionCpnsAdmissionV1.HandleFunc("/get", storeClient.HandlePromotionCpnsAdmissionGet).Methods("GET")
	promotionCpnsAdmissionV1.HandleFunc("/search/paginated", storeClient.HandlePromotionCpnsAdmissionSearchPaginated).Methods("GET")
	promotionCpnsAdmissionV1.HandleFunc("/upload/pak", storeClient.HandlePromotionCpnsAdmissionPakLetterUpload).Methods("POST")
	promotionCpnsAdmissionV1.HandleFunc("/preview/pak", storeClient.HandlePromotionCpnsAdmissionPakLetterPreview).Methods("GET")
	promotionCpnsAdmissionV1.HandleFunc("/download/pak", storeClient.HandlePromotionCpnsAdmissionPakLetterDownload).Methods("GET")
	promotionCpnsAdmissionV1.HandleFunc("/upload/promotion-letter", storeClient.HandlePromotionCpnsAdmissionPromotionLetterUpload).Methods("POST")
	promotionCpnsAdmissionV1.HandleFunc("/preview/promotion-letter", storeClient.HandlePromotionCpnsAdmissionPromotionLetterPreview).Methods("GET")
	promotionCpnsAdmissionV1.HandleFunc("/download/promotion-letter", storeClient.HandlePromotionCpnsAdmissionPromotionLetterDownload).Methods("GET")
	promotionCpnsAdmissionV1.HandleFunc("/submit", storeClient.HandlePromotionCpnsAdmissionSubmit).Methods("POST")

	assessmentTeamV1 := apiV1.PathPrefix("/assessment-team").Subrouter()
	assessmentTeamAdmissionV1 := assessmentTeamV1.PathPrefix("/admission").Subrouter()
	assessmentTeamAdmissionV1.HandleFunc("/upload", storeClient.HandleAssessmentTeamAdmissionSupportDocUpload).Methods("POST")
	assessmentTeamAdmissionV1.HandleFunc("/preview", storeClient.HandleAssessmentTeamAdmissionSupportDocPreview).Methods("GET")
	assessmentTeamAdmissionV1.HandleFunc("/download", storeClient.HandleAssessmentTeamAdmissionSupportDocDownload).Methods("GET")
	assessmentTeamAdmissionV1.HandleFunc("/submit", storeClient.HandleAssessmentTeamAdmissionSubmit).Methods("POST")
	assessmentTeamAdmissionV1.HandleFunc("/get", storeClient.HandleAssessmentTeamGet).Methods("GET")
	assessmentTeamAdmissionV1.HandleFunc("/search", storeClient.HandleAssessmentTeamSearch).Methods("GET")

	assessmentTeamVerificationV1 := assessmentTeamV1.PathPrefix("/verification").Subrouter()
	assessmentTeamVerificationV1.HandleFunc("/upload", storeClient.HandleAssessmentTeamVerificationRecommendationLetterUpload).Methods("POST")
	assessmentTeamVerificationV1.HandleFunc("/preview", storeClient.HandleAssessmentTeamVerificationRecommendationLetterPreview).Methods("GET")
	assessmentTeamVerificationV1.HandleFunc("/download", storeClient.HandleAssessmentTeamVerificationRecommendationLetterDownload).Methods("GET")
	assessmentTeamVerificationV1.HandleFunc("/submit", storeClient.HandleAssessmentTeamVerificationSubmit).Methods("POST")

	genericV1 := apiV1.PathPrefix("/generic").Subrouter()
	genericV1.HandleFunc("/profile/get", storeClient.HandleProfileGet).Methods("GET")
	genericV1.HandleFunc("/role/get", storeClient.HandleRoleGet).Methods("GET")
	genericV1.HandleFunc("/position/get", storeClient.HandlePositionGradesGet).Methods("GET")
	genericV1.HandleFunc("/unit/list", storeClient.HandleListOrganizationUnits).Methods("GET")
	genericV1.HandleFunc("/bezetting", storeClient.HandlePositionGradeBezettingGet).Methods("GET")
	genericV1.HandleFunc("/template/upload/{path:.+}", storeClient.HandleUploadTemplate).Methods("POST")
	genericV1.HandleFunc("/template/download/{path:.+}", storeClient.HandleDownloadTemplate).Methods("GET")

	documentV1 := apiV1.PathPrefix("/document").Subrouter()
	documentV1.HandleFunc("/submit", storeClient.HandleActivityDocumentSubmit).Methods("POST")
	documentV1.HandleFunc("/get", storeClient.HandleActivityDocumentGet).Methods("GET")
	documentV1.HandleFunc("/download", storeClient.HandleActivityDocumentDownload).Methods("GET")
	documentV1.HandleFunc("/delete", storeClient.HandleActivityDocumentDelete).Methods("GET")

	moduleV1 := apiV1.PathPrefix("/module-type").Subrouter()
	moduleV1.HandleFunc("/submit", storeClient.HandleModuleTypeSubmit).Methods("POST")
	moduleV1.HandleFunc("/get", storeClient.HandleModuleTypeGet).Methods("GET")

	signerV1 := apiV1.PathPrefix("/type-signer").Subrouter()
	signerV1.HandleFunc("/get", storeClient.HandleTypeSignerGet).Methods("GET")

	signtte := apiV1.PathPrefix("/sign").Subrouter()
	signtte.HandleFunc("/submit", storeClient.HandleSignSubmit).Methods("post")
	return
}

// A simple HTTP handler to output version data as JSON.
func handleVersion(writer http.ResponseWriter, _ *http.Request) {
	_ = httputil.WriteObj200(writer, map[string]string{
		"version": version,
	})
}

// Deprecated: will be deleted once all old endpoints are deleted.
// Tell the requester that the endpoint has been replaced with a new one.
func endpointRemovedHandler(newPath string) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_ = httputil.WriteObj(writer, ec.NewError(ErrCodeEndpointGone, Errs[ErrCodeEndpointGone], fmt.Errorf("use newer one: %s", newPath)), ErrsToHttp[ErrCodeEndpointGone])
	})
}

// Deprecated: will be deleted once all old endpoints are deleted.
// Tell the requester that the endpoint has been replaced with a new one.
func endpointRemovedRedirectHandler(newPath string) http.Handler {
	return http.RedirectHandler(newPath, http.StatusSeeOther)
}
