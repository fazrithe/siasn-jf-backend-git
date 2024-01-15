package models

// Asn (ASN or PNS) represents a single ASN profile.
type Asn struct {
	// The ASN ID is a 36 uppercase hexadecimal characters string.
	AsnId    string `json:"id"`
	NewNip   string `json:"nip_baru"`
	OldNip   string `json:"nip_lama"`
	Name     string `json:"nama"`
	Email    string `json:"email"`
	Username string `json:"username"`
	//TypeId               string `json:"jenis_pegawai_id"`
	ParentAgencyId              string `json:"instansi_induk_id"`
	ParentAgency                string `json:"instansi_induk_nama"`
	WorkAgencyId                string `json:"instansi_kerja_id"`
	WorkAgency                  string `json:"instansi_kerja_nama"`
	FunctionalPositionId        string `json:"jabatan_fungsional_id"`
	GenericFunctionalPositionId string `json:"jabatan_fungsional_umum_id"`
	Position                    string `json:"position"`
}
