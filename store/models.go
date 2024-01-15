package store

type schemaFilename struct {
	Filename   string `schema:"filename"`
	NoRedirect bool   `schema:"no_redirect"`
}

// organizationPositionId is a local model, containing position + unor ID.
type organizationPositionId struct {
	PositionId         string
	OrganizationUnitId string
}

// bezettingResult represents a single entry of bezetting for position + unor ID.
type bezettingResult struct {
	PositionId         string
	OrganizationUnitId string
	BezettingCount     int
}

type asnNipName struct {
	AsnId   string
	AsnName string
	Nip     string
}
