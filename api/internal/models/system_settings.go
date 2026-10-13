package models

type SystemSettings struct {
	BaseModel
	EnableLocalSignUp                   bool                      `json:"enableLocalSignUp" gorm:"default:false"`
	DebugOcr                            bool                      `json:"debugOcr" gorm:"default:false"`
	NumWorkers                          int                       `json:"numWorkers"`
	EmailPollingInterval                int                       `json:"emailPollingInterval" gorm:"default:1800"`
	CurrencyDisplay                     string                    `json:"currencyDisplay" gorm:"default:$"`
	CurrencyThousandthsSeparator        CurrencySeparator         `json:"currencyThousandthsSeparator" gorm:"default:,"`
	CurrencyDecimalSeparator            CurrencySeparator         `json:"currencyDecimalSeparator" gorm:"default:."`
	CurrencySymbolPosition              CurrencySymbolPosition    `json:"currencySymbolPosition" gorm:"default:START"`
	CurrencyHideDecimalPlaces           bool                      `json:"currencyHideDecimalPlaces" gorm:"default:false"`
	ReceiptProcessingSettings           ReceiptProcessingSettings `json:"-"`
	ReceiptProcessingSettingsId         *uint                     `json:"receiptProcessingSettingsId"`
	FallbackReceiptProcessingSettings   ReceiptProcessingSettings `json:"-"`
	FallbackReceiptProcessingSettingsId *uint                     `json:"fallbackReceiptProcessingSettingsId"`
	TaskConcurrency                     int                       `json:"taskConcurrency" gorm:"default:10"`
	PdfDpi                              int                       `json:"pdfDpi" gorm:"default:300"`
	TaskQueueConfigurations             []TaskQueueConfiguration  `json:"taskQueueConfigurations"`
	// McpEnabled toggles the OAuth 2.1-protected MCP server live, without a
	// restart. The routes are always mounted; this gates them at request time.
	McpEnabled bool `json:"mcpEnabled" gorm:"default:false"`
	// McpPublicUrl is the externally reachable origin (scheme + host) used to
	// build the OAuth issuer/metadata/redirect URLs and the MCP token audience.
	McpPublicUrl string `json:"mcpPublicUrl"`
	// ServerPublicUrl is the externally reachable origin (scheme + host) of this
	// API, used to build the OIDC redirect URI registered at each identity
	// provider: {ServerPublicUrl}/api/oidc/{name}/callback. It is deliberately
	// separate from McpPublicUrl — that value is bound into the MCP token
	// audience, so repurposing it would invalidate live connector tokens.
	ServerPublicUrl string `json:"serverPublicUrl"`
	// ShowLoginQr toggles the self-contained setup QR on the desktop login page.
	ShowLoginQr bool `json:"showLoginQr" gorm:"default:false"`
	// MobileServerUrl is the server/API URL mobile clients connect to. It is
	// encoded (in the login QR's deep link) so scanning it sets up the app.
	MobileServerUrl string `json:"mobileServerUrl"`
	// RefreshTokenValidForHours is how long a refresh token stays valid, i.e. how
	// long a user can be away and still return signed in. Refresh tokens rotate on
	// every use, so this is an inactivity window rather than an absolute session
	// cap. Zero means "unset" and falls back to the built-in default.
	RefreshTokenValidForHours int `json:"refreshTokenValidForHours" gorm:"default:24"`
	// McpRefreshTokenValidForHours is the same for MCP/OAuth connector tokens. It
	// is kept separate from RefreshTokenValidForHours so a long window chosen for
	// human convenience does not silently extend tokens held by third-party
	// clients. Zero means "unset" and falls back to the built-in default.
	McpRefreshTokenValidForHours int `json:"mcpRefreshTokenValidForHours" gorm:"default:24"`
	// TempFileRetentionHours is how long a file in temp/ is kept once nothing can
	// still act on it — how long a user has to rerun, preview or download the
	// source image of an upload that failed. Zero means "unset" and falls back to
	// the built-in default; the read-side clamp, not this column default, is what
	// guarantees a sane value, since AutoMigrate's back-fill of an added column on
	// the pre-existing row is engine-dependent.
	TempFileRetentionHours int `json:"tempFileRetentionHours" gorm:"default:720"`
	// TimeZone is the IANA name (e.g. "America/New_York") the app works out
	// calendar boundaries in: which day an instant such as created_at falls on,
	// "today", "this month", report periods and {{generatedAt}}. A receipt's own
	// Date is a calendar day and never shifts with it. An empty or unknown value
	// resolves to UTC (see repositories.GetAppLocation), which is also the
	// default, so an existing install behaves exactly as before.
	TimeZone string `json:"timeZone" gorm:"default:UTC"`
	// SecureCookies marks every cookie the API sets as Secure, so browsers only
	// send them over HTTPS. Off by default because plain-http self-hosting (a LAN
	// address, a bare IP) is supported, and a browser silently drops a Secure
	// cookie set over http — turning this on there breaks login. See
	// services.UseSecureCookies.
	SecureCookies bool `json:"secureCookies" gorm:"default:false"`
}
