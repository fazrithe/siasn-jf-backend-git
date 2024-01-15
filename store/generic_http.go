package store

import (
	"context"
	"net/http"
	"net/url"
	"path"

	"github.com/gorilla/mux"
	. "github.com/if-itb/siasn-jf-backend/errnum"
	"github.com/if-itb/siasn-libs-backend/auth"
	"github.com/if-itb/siasn-libs-backend/ec"
	"github.com/if-itb/siasn-libs-backend/httputil"
)

const (
	userDetailContextKey = "user"

	TimeoutPositionGradesGet         = TimeoutDefault
	TimeoutPositionGradeBezettingGet = TimeoutDefault
	TimeoutListOrganizationUnits     = TimeoutDefault
	TimeoutUploadTemplate            = TimeoutDefault
	TimeoutDownloadTemplate          = TimeoutDefault
)

// HandleProfileGet reads the user detail from ID token and returns it.
func (c *Client) HandleProfileGet(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)
	if user == nil {
		return
	}

	_ = httputil.WriteObj200(writer, user)
}

func (a *Client) HandleRoleGet(writer http.ResponseWriter, request *http.Request) {
	role, err := request.Cookie("role")
	if err != nil {
		return
	}

	_ = httputil.WriteObj200(writer, map[string]interface{}{
		"role": role,
	})
}

// HandlePositionGradesGet handles a request to get position list of a work agency.
// Agency ID will be retrieved from authentication token.
func (c *Client) HandlePositionGradesGet(writer http.ResponseWriter, request *http.Request) {
	user := auth.AssertReqGetUserDetail(request)
	if user == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPositionGradesGet)
	defer cancel()

	positions, err := c.GetPositionGradesCtx(ctx, user.WorkAgencyId)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, positions)
}

// HandlePositionGradeBezettingGet handles a request to get a position grade bezetting.
func (c *Client) HandlePositionGradeBezettingGet(writer http.ResponseWriter, request *http.Request) {
	type schemaRequirementId struct {
		PositionGradeId    string `schema:"jabatan_id"`
		OrganizationUnitId string `schema:"unit_organisasi_id"`
	}

	ctx, cancel := context.WithTimeout(context.Background(), TimeoutPositionGradeBezettingGet)
	defer cancel()

	s := &schemaRequirementId{}
	err := c.decodeRequestSchema(writer, request, s)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	if s.PositionGradeId == "" {
		c.httpError(writer, ec.NewErrorBasic(ErrCodeEntryNotFound, Errs[ErrCodeEntryNotFound]))
		return
	}

	res, err := c.GetRequirementBezettingCtx(ctx, s.PositionGradeId, s.OrganizationUnitId)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, &res)
}

// HandleListOrganizationUnits handles request to list all organization units in the work agency of the logged in user.
func (c *Client) HandleListOrganizationUnits(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutListOrganizationUnits)
	defer cancel()

	user := auth.AssertReqGetUserDetail(request)

	ou, err := c.ListOrganizationUnitsCtx(ctx, user.WorkAgencyId)
	if err != nil {
		c.httpError(writer, err)
		return
	}

	_ = httputil.WriteObj200(writer, ou)
}

// HandleUploadTemplate handles uploading an arbitrary template.
func (c *Client) HandleUploadTemplate(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutUploadTemplate)
	defer cancel()

	v := mux.Vars(request)
	templatePath := v["path"]
	if templatePath == "" {
		http.NotFound(writer, request) // Mimic the behavior of endpoint not found
		return
	}

	templatePath = path.Join("template", templatePath)

	_, activityOk := activityTemplatePaths[templatePath]
	_, requirementOk := requirementTemplatePaths[templatePath]
	_, promotionOk := promotionTemplatePaths[templatePath]
	_, dismissalOk := dismissalTemplatePaths[templatePath]

	if !activityOk && !requirementOk && !promotionOk && !dismissalOk {
		http.NotFound(writer, request) // Mimic the behavior of endpoint not found
		return
	}

	var signFunc func(context.Context, string) (*url.URL, error)

	if activityOk {
		signFunc = c.ActivityStorage.GenerateActivityDocPutSignDirect
	} else if requirementOk {
		signFunc = c.RequirementStorage.GenerateRequirementDocPutSignDirect
	} else if promotionOk {
		signFunc = c.PromotionStorage.GeneratePromotionDocPutSignDirect
	} else if dismissalOk {
		signFunc = c.DismissalStorage.GenerateDismissalDocPutSignDirect
	}

	u, err := signFunc(ctx, templatePath)
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	_ = httputil.WriteObj200HtmlEscape(writer, map[string]string{
		"url": u.String(),
	}, false)
}

// HandleDownloadTemplate handles downloading an arbitrary template.
func (c *Client) HandleDownloadTemplate(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDownloadTemplate)
	defer cancel()

	v := mux.Vars(request)
	templatePath := v["path"]
	if templatePath == "" {
		http.NotFound(writer, request) // Mimic the behavior of endpoint not found
		return
	}

	templatePath = path.Join("template", templatePath)

	_, activityOk := activityTemplatePaths[templatePath]
	_, requirementOk := requirementTemplatePaths[templatePath]
	_, promotionOk := promotionTemplatePaths[templatePath]
	_, dismissalOk := dismissalTemplatePaths[templatePath]

	if !activityOk && !requirementOk && !promotionOk {
		http.NotFound(writer, request) // Mimic the behavior of endpoint not found
		return
	}

	var signFunc func(context.Context, string) (*url.URL, error)

	if activityOk {
		signFunc = c.ActivityStorage.GenerateActivityDocGetSign
	} else if requirementOk {
		signFunc = c.RequirementStorage.GenerateRequirementDocGetSign
	} else if promotionOk {
		signFunc = c.PromotionStorage.GeneratePromotionDocGetSign
	} else if dismissalOk {
		signFunc = c.DismissalStorage.GenerateDismissalDocGetSign
	}

	u, err := signFunc(ctx, templatePath)
	if err != nil {
		c.httpError(writer, ec.NewError(ErrCodeStorageSignFail, Errs[ErrCodeStorageSignFail], err))
		return
	}

	http.Redirect(writer, request, u.String(), http.StatusFound)
}
