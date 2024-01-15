package errnum

const (
	// ErrCodeActivityAdmissionInsertAsnNotFound - 11401: Some ASNs cannot be found by the given NIP.
	ErrCodeActivityAdmissionInsertAsnNotFound = iota + ServiceErrorCode*10000 + 1*1000 + 401
	// ErrCodeActivityAdmissionInsertNoAttendees - 11402: No attendees NIP supplied.
	ErrCodeActivityAdmissionInsertNoAttendees
	// ErrCodeActivityAdmissionInsertNameEmpty - 11403: Activity name needed.
	ErrCodeActivityAdmissionInsertNameEmpty
	// ErrCodeActivityAdmissionInsertTypeInvalid - 11404: Activity type is invalid.
	ErrCodeActivityAdmissionInsertTypeInvalid
	// ErrCodeActivityAdmissionNoSupportDocs - 11405: No support documents saved, must supply at least one filename.
	ErrCodeActivityAdmissionNoSupportDocs
	// ErrCodeActivityAdmissionPositionGradeEmpty - 11406: jabatan_jenjang is empty.
	ErrCodeActivityAdmissionPositionGradeEmpty
	// ErrCodeActivityAdmissionDatePeriodInvalid - 11407: start/end date is invalid.
	ErrCodeActivityAdmissionDatePeriodInvalid
	// ErrCodeActivityAdmissionTrainingYearInvalid - 11408: training year is invalid
	ErrCodeActivityAdmissionTrainingYearInvalid
	// ErrCodeActivityAdmissionDurationInvalid - 11409: duration is invalid.
	ErrCodeActivityAdmissionDurationInvalid
	// ErrCodeActivityAdmissionNumberInvalid - 11410: admission number is invalid.
	ErrCodeActivityAdmissionNumberInvalid

	// ErrCodeActivityCsrStatusSetNotAccepted - 11411: Cannot set activity status to certificate request, activity is not accepted yet.
	ErrCodeActivityCsrStatusSetNotAccepted
	// ErrCodeActivityCsrStatusSetOngoing - 11412: Certificate request is ongoing.
	ErrCodeActivityCsrStatusSetOngoing
	// ErrCodeActivityCsrStatusNoAttendees - 11413: No attendees supplied.
	ErrCodeActivityCsrStatusNoAttendees

	// ErrCodeActivityVerificationStatusProcessedFurther - 11414: Cannot set activity status to accepted, activity already processed further.
	ErrCodeActivityVerificationStatusProcessedFurther
	// ErrCodeActivityVerificationStatusAlreadyAccepted - 11415: Activity already accepted.
	ErrCodeActivityVerificationStatusAlreadyAccepted
	// ErrCodeActivityVerificationStatusNoAttendees - 11416: No attendees supplied.
	ErrCodeActivityVerificationStatusNoAttendees

	// ErrCodeActivityCertSubmitNotRequested - 11417: Activity admission status is not accepted.
	ErrCodeActivityCertSubmitNotRequested
	// ErrCodeActivityCertSubmitPakUnsupported - 11418: PAK document is supported only for uji kompetensi perpindahan jabatan.
	ErrCodeActivityCertSubmitPakUnsupported
	// ErrCodeActivityCertSubmitNoDocs - 11419: No documents were uploaded.
	ErrCodeActivityCertSubmitNoDocs
	// ErrCodeActivityCertTypeUnsupported - 11420: Certificate type is not supported.
	ErrCodeActivityCertTypeUnsupported

	// ErrCodeFilterStatusInvalid - 11421: admission status supplied contains value outside the valid range (1-5).
	ErrCodeFilterStatusInvalid
	// ErrCodeFilterInvalidDate - 11422: the date format supplied does not conform to the date format required.
	ErrCodeFilterInvalidDate
	// ErrCodeFilterInvalidType - 11423: admission type supplied contains value outside the valid range (1-7).
	ErrCodeFilterInvalidType
	// ErrCodeFilterInvalidAgencyId - 11424: agency ID cannot be empty.
	ErrCodeFilterInvalidAgencyId

	// ErrCodeActivityAdmissionDetailForbidden - 11425: the user is not authorized to see the detail of the activity
	// admission.
	ErrCodeActivityAdmissionDetailForbidden
	// ErrCodeActivityNoRecommendationLetter - 11426: No recommendation letter supplied.
	ErrCodeActivityNoRecommendationLetter
)

func init() {
	Errs[ErrCodeActivityAdmissionInsertAsnNotFound] = "some ASNs cannot be found by the given NIP"
	Errs[ErrCodeActivityAdmissionInsertNoAttendees] = "attendees NIP(s) are needed, either with old or new version NIP"
	Errs[ErrCodeActivityAdmissionInsertNameEmpty] = "activity name is needed"
	Errs[ErrCodeActivityAdmissionInsertTypeInvalid] = "activity type is invalid"
	Errs[ErrCodeActivityAdmissionNoSupportDocs] = "no support documents saved, must supply at least one filename"
	Errs[ErrCodeActivityAdmissionPositionGradeEmpty] = "jabatan_jenjang is empty"
	Errs[ErrCodeActivityAdmissionDatePeriodInvalid] = "start/end date is invalid"
	Errs[ErrCodeActivityAdmissionTrainingYearInvalid] = "admission training year is invalid"
	Errs[ErrCodeActivityAdmissionDurationInvalid] = "admission duration is invalid"
	Errs[ErrCodeActivityAdmissionNumberInvalid] = "admission number is invalid"

	Errs[ErrCodeActivityCsrStatusSetNotAccepted] = "cannot set activity status to certificate request, activity is not accepted yet"
	Errs[ErrCodeActivityCsrStatusSetOngoing] = "certificate request is already ongoing for this activity"
	Errs[ErrCodeActivityCsrStatusNoAttendees] = "no attendees supplied"

	Errs[ErrCodeActivityVerificationStatusProcessedFurther] = "cannot set activity status to accepted, activity already processed further"
	Errs[ErrCodeActivityVerificationStatusAlreadyAccepted] = "this activity is already accepted"
	Errs[ErrCodeActivityVerificationStatusNoAttendees] = "no attendees supplied"

	Errs[ErrCodeActivityCertSubmitNotRequested] = "activity admission status is not accepted"
	Errs[ErrCodeActivityCertSubmitPakUnsupported] = "PAK document is supported only for uji kompetensi perpindahan jabatan"
	Errs[ErrCodeActivityCertSubmitNoDocs] = "no documents were uploaded"
	Errs[ErrCodeActivityCertTypeUnsupported] = "certificate type is not supported"

	Errs[ErrCodeFilterStatusInvalid] = "admission status must be >= 1 and <= 5"
	Errs[ErrCodeFilterInvalidDate] = "date must be in the format of YYYY-MM-DD (e.g. 2006-12-31)"
	Errs[ErrCodeFilterInvalidType] = "type (jenis_kegiatan) must be >= 1 and <= 7"
	Errs[ErrCodeFilterInvalidAgencyId] = "agency id (instansi_id) cannot be empty"
	Errs[ErrCodeActivityAdmissionDetailForbidden] = "not authorized to see the the detail"

	Errs[ErrCodeActivityNoRecommendationLetter] = "no recommendation letter supplied"

	ErrsToHttp[ErrCodeActivityAdmissionInsertAsnNotFound] = 400
	ErrsToHttp[ErrCodeActivityAdmissionInsertNoAttendees] = 400
	ErrsToHttp[ErrCodeActivityAdmissionInsertNameEmpty] = 400
	ErrsToHttp[ErrCodeActivityAdmissionInsertTypeInvalid] = 400
	ErrsToHttp[ErrCodeActivityAdmissionNoSupportDocs] = 400
	ErrsToHttp[ErrCodeActivityAdmissionPositionGradeEmpty] = 400
	ErrsToHttp[ErrCodeActivityAdmissionDatePeriodInvalid] = 400
	ErrsToHttp[ErrCodeActivityAdmissionTrainingYearInvalid] = 400
	ErrsToHttp[ErrCodeActivityAdmissionDurationInvalid] = 400
	ErrsToHttp[ErrCodeActivityAdmissionNumberInvalid] = 400

	ErrsToHttp[ErrCodeActivityCsrStatusSetNotAccepted] = 400
	ErrsToHttp[ErrCodeActivityCsrStatusSetOngoing] = 400
	ErrsToHttp[ErrCodeActivityCsrStatusNoAttendees] = 400

	ErrsToHttp[ErrCodeActivityVerificationStatusProcessedFurther] = 400
	ErrsToHttp[ErrCodeActivityVerificationStatusAlreadyAccepted] = 400
	ErrsToHttp[ErrCodeActivityVerificationStatusNoAttendees] = 400

	ErrsToHttp[ErrCodeActivityCertSubmitNotRequested] = 400
	ErrsToHttp[ErrCodeActivityCertSubmitPakUnsupported] = 400
	ErrsToHttp[ErrCodeActivityCertSubmitNoDocs] = 400
	ErrsToHttp[ErrCodeActivityCertTypeUnsupported] = 400

	ErrsToHttp[ErrCodeFilterStatusInvalid] = 400
	ErrsToHttp[ErrCodeFilterInvalidDate] = 400
	ErrsToHttp[ErrCodeFilterInvalidType] = 400
	ErrsToHttp[ErrCodeFilterInvalidAgencyId] = 400
	ErrsToHttp[ErrCodeActivityAdmissionDetailForbidden] = 403

	ErrsToHttp[ErrCodeActivityNoRecommendationLetter] = 400
}
