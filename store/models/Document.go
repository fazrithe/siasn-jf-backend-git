package models

// Document is the metadata of document such PDF stored in the object storage.
type Document struct {
	// Filename is the document's unique id.
	Filename string `json:"nama_file,omitempty"`

	// DocumentName is the human-readable name of the document.
	DocumentName string `json:"nama_dokumen,omitempty"`

	// DocumentNumber is the document's registered number.
	DocumentNumber string `json:"no_dokumen,omitempty"`

	// DocumentDate is when the document published.
	DocumentDate Iso8601Date `json:"tgl_dokumen,omitempty"`

	// A freetext custom note.
	Note string `json:"catatan,omitempty"`

	// SignerId is the signer user ID.
	SignerId string `json:"ttd_user_id,omitempty"`

	// Indicates whether this document has been signed.
	// Ignore if the document does not need signing.
	IsSigned bool `json:"sudah_ttd,omitempty"`

	// The timestamp of when this document was created.
	CreatedAt EpochTime `json:"created_ts,omitempty"`

	// The timestamp of when this document was signed.
	SignedAt EpochTime `json:"signed_at,omitempty"`
}
