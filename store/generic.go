package store

import (
	"context"
	"database/sql"
	"fmt"

	. "github.com/fazrithe/siasn-jf-backend-git/errnum"
	"github.com/fazrithe/siasn-jf-backend-git/store/models"
	"github.com/if-itb/siasn-libs-backend/auth"
	"github.com/if-itb/siasn-libs-backend/ec"
	"github.com/if-itb/siasn-libs-backend/metricutil"
	"github.com/lib/pq"
)

// getAsnNipNames retrieves all ASNs that matches the given IDs.
func (c *Client) getAsnNipNames(ctx context.Context, profileDh metricutil.DbHandler, asnIds []string) (asns map[string]*asnNipName, err error) {
	asns = make(map[string]*asnNipName)
	rows, err := profileDh.QueryContext(
		ctx,
		"select distinct on (nip_baru) t.id, nip_baru, orang.nama from orang join (select id, nip_baru from pns where pns.id = ANY($1) union select id, nip_baru from pppk where pppk.id = ANY($1)) t on orang.id = t.id",
		pq.Array(asnIds),
	)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve ASN data: %w", err))
	}
	defer rows.Close()

	for rows.Next() {
		asn := &asnNipName{}
		err = rows.Scan(&asn.AsnId, &asn.Nip, &asn.AsnName)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve ASN data: %w", err))
		}
		asns[asn.AsnId] = asn
	}
	if err = rows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve ASN data: %w", err))
	}

	return asns, nil
}

// getOrganizationUnitNames retrieve unor names from the given unor IDs in one agency.
func (c *Client) getOrganizationUnitNames(ctx context.Context, referenceDh metricutil.DbHandler, unorIds []string) (unor map[string]string, err error) {
	rows, err := referenceDh.QueryContext(ctx, "select id, nama_unor from unor where id = any($1)", pq.Array(unorIds))
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve unor: %w", err))
	}
	defer rows.Close()

	unor = make(map[string]string)
	for rows.Next() {
		id := ""
		name := ""
		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve unor: %w", err))
		}
		unor[id] = name
	}
	if err = rows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve unor: %w", err))
	}

	return unor, nil
}

// getOrganizationUnitNamesByAgencyId retrieve unor names from the given unor IDs in one agency.
// Limited only to the specified agencyId.
func (c *Client) getOrganizationUnitNamesByAgencyId(ctx context.Context, referenceDh metricutil.DbHandler, agencyId string, unorIds []string) (unor map[string]string, err error) {
	rows, err := referenceDh.QueryContext(ctx, "select id, nama_unor from unor where id = any($1) and instansi_id = $2", pq.Array(unorIds), agencyId)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve unor: %w", err))
	}
	defer rows.Close()

	unor = make(map[string]string)
	for rows.Next() {
		id := ""
		name := ""
		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve unor: %w", err))
		}
		unor[id] = name
	}
	if err = rows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve unor: %w", err))
	}

	return unor, nil
}

// getFunctionalPositionNames retrieve JF names from the given JF IDs.
func (c *Client) getFunctionalPositionNames(ctx context.Context, referenceDh metricutil.DbHandler, functionalPositionIds []string) (positions map[string]string, err error) {
	rows, err := referenceDh.QueryContext(ctx, "select id, nama from jabatan_fungsional where id = any($1)", pq.Array(functionalPositionIds))
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve JF names: %w", err))
	}
	defer rows.Close()

	positions = make(map[string]string)
	for rows.Next() {
		id := ""
		name := ""
		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve JF names: %w", err))
		}
		positions[id] = name
	}
	if err = rows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve JF names: %w", err))
	}

	return positions, nil
}

// getAgencyNames retrieve agency names from the given agency IDs.
func (c *Client) getAgencyNames(ctx context.Context, referenceDh metricutil.DbHandler, agencyIds []string) (agency map[string]string, err error) {
	rows, err := referenceDh.QueryContext(ctx, "select id, nama from instansi where id = any($1)", pq.Array(agencyIds))
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve agency names: %w", err))
	}
	defer rows.Close()

	agency = make(map[string]string)
	for rows.Next() {
		id := ""
		name := ""
		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve agency names: %w", err))
		}
		agency[id] = name
	}
	if err = rows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve agency names: %w", err))
	}

	return agency, nil
}

func (c *Client) GetUserDetailByNip(ctx context.Context, nip string) (user *auth.Asn, err error) {
	profileMdb := metricutil.NewDB(c.ProfileDb, c.SqlMetrics)
	referenceMdb := metricutil.NewDB(c.ReferenceDb, c.SqlMetrics)

	return c.getUserDetail(ctx, profileMdb, referenceMdb, nip, "")
}

func (c *Client) GetUserDetailByNipWorkAgencyId(ctx context.Context, nip string, workAgencyId string) (user *auth.Asn, err error) {
	if workAgencyId == "" {
		panic("workAgencyId cannot be empty string")
	}

	profileMdb := metricutil.NewDB(c.ProfileDb, c.SqlMetrics)
	referenceMdb := metricutil.NewDB(c.ReferenceDb, c.SqlMetrics)

	return c.getUserDetail(ctx, profileMdb, referenceMdb, nip, workAgencyId)
}

// getUserDetail retrieves user extended detail from the database, based on the user NIP. workAgencyId is optional
// This function does not make distinction about new or old NIP. When the user cannot be found, it will return no error
// but will return nil user instead.
func (c *Client) getUserDetail(ctx context.Context, profileDh metricutil.DbHandler, referenceDh metricutil.DbHandler, nip string, workAgencyId string) (user *auth.Asn, err error) {
	user = &auth.Asn{}
	unorId := ""
	positionTypeId := 0
	birthday := sql.NullString{}

	err = profileDh.QueryRowContext(
		ctx,
		`
select 
       pns.id,
       nip_baru,
       coalesce(nip_lama, ''),
       coalesce(nama, ''),
       coalesce(nomor_id_document, ''),
       coalesce(nomor_hp, ''),
       to_char(tgl_lhr, 'YYYY-MM-DD'),
       instansi_induk_id,
       coalesce(instansi_induk_nama, ''),
       instansi_kerja_id,
       coalesce(instansi_kerja_nama, ''),
       coalesce(jabatan_fungsional_id, ''),
       coalesce(jabatan_fungsional_umum_id, ''),
       jenis_jabatan_id,
       coalesce(unor_id, ''),
       golongan_id
from pns 
    left join orang on pns.id = orang.id
where (nip_baru = $1 or nip_lama = $1) and ($2::text is null or $2 = '' or instansi_kerja_id = $2) order by nip_baru limit 1`,
		nip,
		workAgencyId,
	).Scan(
		&user.AsnId,
		&user.NewNip,
		&user.OldNip,
		&user.Name,
		&user.Nik,
		&user.PhoneNumber,
		&birthday,
		&user.ParentAgencyId,
		&user.ParentAgency,
		&user.WorkAgencyId,
		&user.WorkAgency,
		&user.FunctionalPositionId,
		&user.GenericFunctionalPositionId,
		&positionTypeId,
		&unorId,
		&user.BracketId,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve data from pns: %w", err))
	}

	if err == sql.ErrNoRows {
		err = profileDh.QueryRowContext(
			ctx,
			`
select 
       pppk.id,
       nip_baru,
       coalesce(nama, ''),
       coalesce(nomor_hp, ''),
       to_char(tgl_lhr, 'YYYY-MM-DD'),
       coalesce(nomor_id_document, ''),
       instansi_induk_id,
       coalesce(instansi_induk_nama, ''),
       instansi_kerja_id,
       coalesce(instansi_kerja_nama, ''),
       coalesce(jabatan_fungsional_id, ''),
       coalesce(jabatan_fungsional_umum_id, ''),
       jenis_jabatan_id,
       coalesce(unor_id, ''),
	   golongan_id
from pppk 
    left join orang on pppk.id = orang.id
where nip_baru = $1 and ($2::text is null or $2 = '' or instansi_kerja_id = $2) order by nip_baru limit 1`,
			nip,
			workAgencyId,
		).Scan(
			&user.AsnId,
			&user.NewNip,
			&user.Name,
			&user.Nik,
			&user.PhoneNumber,
			&birthday,
			&user.ParentAgencyId,
			&user.ParentAgency,
			&user.WorkAgencyId,
			&user.WorkAgency,
			&user.FunctionalPositionId,
			&user.GenericFunctionalPositionId,
			&positionTypeId,
			&unorId,
			&user.BracketId,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve data from pppk: %w", err))
		}
	}

	user.Birthday = birthday.String

	unorPosition := ""
	err = referenceDh.QueryRowContext(ctx, `select nama_unor, coalesce(nama_jabatan, '') from unor where id = $1`, unorId).Scan(&user.OrganizationUnit, &unorPosition)
	if err != nil && err != sql.ErrNoRows {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve data from unor for pns: %w", err))
	}

	if positionTypeId == 1 || positionTypeId == 3 {
		user.Position = unorPosition
	} else if positionTypeId == 2 && user.FunctionalPositionId != "" {
		err = referenceDh.QueryRowContext(ctx, `select nama from jabatan_fungsional where id = $1`, user.FunctionalPositionId).Scan(&user.Position)
		if err != nil && err != sql.ErrNoRows {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve data from jabatan_fungsional for pns: %w", err))
		}
	} else if positionTypeId == 4 && user.GenericFunctionalPositionId != "" {
		err = referenceDh.QueryRowContext(ctx, `select nama from jabatan_fungsional_umum where id = $1`, user.GenericFunctionalPositionId).Scan(&user.Position)
		if err != nil && err != sql.ErrNoRows {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve data from jabatan_fungsional_umum for pns: %w", err))
		}
	}

	err = nil

	err = referenceDh.QueryRowContext(ctx, "select nama, nama_pangkat from golongan where id = $1", user.BracketId).Scan(&user.Bracket, &user.Rank)
	if err != nil && err != sql.ErrNoRows {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve data from golongan for pns: %w", err))
	}

	err = nil

	return user, nil
}

// getUserDetail retrieves user extended detail from the database, based on the user NIP. workAgencyId is optional
// This function does not make distinction about new or old NIP. When the user cannot be found, it will return no error
// but will return nil user instead.
func (c *Client) getUserDetailByAsnId(ctx context.Context, profileDh metricutil.DbHandler, referenceDh metricutil.DbHandler, asnId string, workAgencyId string) (user *auth.Asn, err error) {
	user = &auth.Asn{}
	unorId := ""
	positionTypeId := 0
	birthday := sql.NullString{}

	err = profileDh.QueryRowContext(
		ctx,
		`
select 
       pns.id,
       nip_baru,
       coalesce(nip_lama, ''),
       coalesce(nama, ''),
       coalesce(nomor_id_document, ''),
       coalesce(nomor_hp, ''),
       to_char(tgl_lhr, 'YYYY-MM-DD'),
       instansi_induk_id,
       coalesce(instansi_induk_nama, ''),
       instansi_kerja_id,
       coalesce(instansi_kerja_nama, ''),
       coalesce(jabatan_fungsional_id, ''),
       coalesce(jabatan_fungsional_umum_id, ''),
       jenis_jabatan_id,
       coalesce(unor_id, ''),
       golongan_id
from pns 
    left join orang on pns.id = orang.id
where pns.id = $1 and ($2::text is null or $2 = '' or instansi_kerja_id = $2) order by nip_baru limit 1`,
		asnId,
		workAgencyId,
	).Scan(
		&user.AsnId,
		&user.NewNip,
		&user.OldNip,
		&user.Name,
		&user.Nik,
		&user.PhoneNumber,
		&birthday,
		&user.ParentAgencyId,
		&user.ParentAgency,
		&user.WorkAgencyId,
		&user.WorkAgency,
		&user.FunctionalPositionId,
		&user.GenericFunctionalPositionId,
		&positionTypeId,
		&unorId,
		&user.BracketId,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve data from pns: %w", err))
	}

	if err == sql.ErrNoRows {
		err = profileDh.QueryRowContext(
			ctx,
			`
select 
       pppk.id,
       nip_baru,
       coalesce(nama, ''),
       coalesce(nomor_hp, ''),
       to_char(tgl_lhr, 'YYYY-MM-DD'),
       coalesce(nomor_id_document, ''),
       instansi_induk_id,
       coalesce(instansi_induk_nama, ''),
       instansi_kerja_id,
       coalesce(instansi_kerja_nama, ''),
       coalesce(jabatan_fungsional_id, ''),
       coalesce(jabatan_fungsional_umum_id, ''),
       jenis_jabatan_id,
       coalesce(unor_id, ''),
       golongan_id
from pppk 
    left join orang on pppk.id = orang.id
where pppk.id = $1 and ($2::text is null or $2 = '' or instansi_kerja_id = $2) order by nip_baru limit 1`,
			asnId,
			workAgencyId,
		).Scan(
			&user.AsnId,
			&user.NewNip,
			&user.Name,
			&user.Nik,
			&user.PhoneNumber,
			&birthday,
			&user.ParentAgencyId,
			&user.ParentAgency,
			&user.WorkAgencyId,
			&user.WorkAgency,
			&user.FunctionalPositionId,
			&user.GenericFunctionalPositionId,
			&positionTypeId,
			&unorId,
			&user.BracketId,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve data from pppk: %w", err))
		}
	}

	user.Birthday = birthday.String

	unorPosition := ""
	err = referenceDh.QueryRowContext(ctx, `select nama_unor, coalesce(nama_jabatan, '') from unor where id = $1`, unorId).Scan(&user.OrganizationUnit, &unorPosition)
	if err != nil && err != sql.ErrNoRows {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve data from unor for pns: %w", err))
	}

	if positionTypeId == 1 || positionTypeId == 3 {
		user.Position = unorPosition
	} else if positionTypeId == 2 && user.FunctionalPositionId != "" {
		err = referenceDh.QueryRowContext(ctx, `select nama from jabatan_fungsional where id = $1`, user.FunctionalPositionId).Scan(&user.Position)
		if err != nil && err != sql.ErrNoRows {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve data from jabatan_fungsional for pns: %w", err))
		}
	} else if positionTypeId == 4 && user.GenericFunctionalPositionId != "" {
		err = referenceDh.QueryRowContext(ctx, `select nama from jabatan_fungsional_umum where id = $1`, user.GenericFunctionalPositionId).Scan(&user.Position)
		if err != nil && err != sql.ErrNoRows {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve data from jabatan_fungsional_umum for pns: %w", err))
		}
	}

	err = nil

	err = referenceDh.QueryRowContext(ctx, "select nama, nama_pangkat from golongan where id = $1", user.BracketId).Scan(&user.Bracket, &user.Rank)
	if err != nil && err != sql.ErrNoRows {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], fmt.Errorf("cannot retrieve data from golongan for pns: %w", err))
	}

	err = nil

	return user, nil
}

// GetPositionGradesCtx searches for the list of position grades of a particular work agency ID (instansi kerja).
// It will return empty slice if no position are found.
func (c *Client) GetPositionGradesCtx(ctx context.Context, workAgencyId string) (positions []*models.PositionGrade, err error) {
	mdb := metricutil.NewDB(c.ReferenceDb, c.SqlMetrics)

	positionRows, err := mdb.QueryContext(ctx, "select jf.id, jf.nama from jabatan_fungsional jf join kel_jabatan kj on jf.kel_jabatan_id = kj.id join instansi i on kj.pembina_id = i.id where i.id = $1", workAgencyId)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer positionRows.Close()
	positions = []*models.PositionGrade{}
	for positionRows.Next() {
		position := &models.PositionGrade{}
		err = positionRows.Scan(
			&position.PositionGradeId,
			&position.PositionGradeName,
		)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
		positions = append(positions, position)
	}
	if err = positionRows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	return
}

// GetRequirementBezettingCtx counts the bezetting (employee count) of the position grade, specifically the functional
// position from a requirement admission.
func (c *Client) GetRequirementBezettingCtx(ctx context.Context, positionGradeId string, organizationUnitId string) (res *models.Bezetting, err error) {
	res = &models.Bezetting{
		PositionGradeId:    positionGradeId,
		OrganizationUnitId: organizationUnitId,
	}

	mdbProfile := metricutil.NewDB(c.ProfileDb, c.SqlMetrics)

	count := 0
	err = mdbProfile.QueryRowContext(ctx, "select count(*) from pns where jabatan_fungsional_id = $1 and unor_id = $2", positionGradeId, organizationUnitId).Scan(&count)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	res.Count = count

	return res, nil
}

// ListOrganizationUnitsCtx list organization units in an agency.
func (c *Client) ListOrganizationUnitsCtx(ctx context.Context, agencyId string) (ou []*models.OrganizationUnit, err error) {
	mdbReference := metricutil.NewDB(c.ReferenceDb, c.SqlMetrics)

	rows, err := mdbReference.QueryContext(ctx, "select id, nama_unor from unor where instansi_id = $1", agencyId)
	if err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}
	defer rows.Close()

	for rows.Next() {
		o := &models.OrganizationUnit{}
		err = rows.Scan(&o.OrganizationUnitId, &o.OrganizationUnitName)
		if err != nil {
			return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
		}
		ou = append(ou, o)
	}
	if err = rows.Err(); err != nil {
		return nil, ec.NewError(ErrCodeQueryFail, Errs[ErrCodeQueryFail], err)
	}

	return ou, nil
}
