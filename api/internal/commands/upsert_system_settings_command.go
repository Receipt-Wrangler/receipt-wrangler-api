package commands

import (
	"encoding/json"
	"net/http"
	"net/url"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"time"
)

// Bounds on the configurable refresh-token lifetimes. They live here rather than
// alongside the resolver in services/ because internal/commands cannot import
// internal/services, and the validation and the read-side clamp must agree.
const (
	MinRefreshTokenValidForHours = 1
	MaxRefreshTokenValidForHours = 720 // 30 days
)

// Bounds and default for the temp-file retention window, here for the same
// reason as the refresh-token bounds above: the validator below and the
// read-side clamp in wranglerasynq must not drift.
//
// The floor is a day because the window doubles as the grace period a user has
// to rerun or download a failed upload — anything shorter reintroduces the bug
// this setting exists to fix. The ceiling is a year.
const (
	MinTempFileRetentionHours     = 24
	MaxTempFileRetentionHours     = 8760 // 365 days
	DefaultTempFileRetentionHours = 720  // 30 days
)

type UpsertSystemSettingsCommand struct {
	EnableLocalSignUp                   bool                                  `json:"enableLocalSignUp"`
	DebugOcr                            bool                                  `json:"debugOcr"`
	CurrencyDisplay                     string                                `json:"currencyDisplay"`
	CurrencyThousandthsSeparator        models.CurrencySeparator              `json:"currencyThousandthsSeparator"`
	CurrencyDecimalSeparator            models.CurrencySeparator              `json:"currencyDecimalSeparator"`
	CurrencySymbolPosition              models.CurrencySymbolPosition         `json:"currencySymbolPosition"`
	CurrencyHideDecimalPlaces           bool                                  `json:"currencyHideDecimalPlaces"`
	NumWorkers                          int                                   `json:"numWorkers"`
	EmailPollingInterval                int                                   `json:"emailPollingInterval"`
	ReceiptProcessingSettingsId         *uint                                 `json:"receiptProcessingSettingsId"`
	FallbackReceiptProcessingSettingsId *uint                                 `json:"fallbackReceiptProcessingSettingsId"`
	TaskConcurrency                     int                                   `json:"taskConcurrency"`
	PdfDpi                              int                                   `json:"pdfDpi"`
	TaskQueueConfigurations             []UpsertTaskQueueConfigurationCommand `json:"taskQueueConfigurations"`
	McpEnabled                          bool                                  `json:"mcpEnabled"`
	McpPublicUrl                        string                                `json:"mcpPublicUrl"`
	ShowLoginQr                         bool                                  `json:"showLoginQr"`
	MobileServerUrl                     string                                `json:"mobileServerUrl"`
	// Pointers so an omitted key is distinguishable from an explicit zero value.
	// The repository writes every column (Select("*")), so a plain field would
	// persist as its zero value and silently reset a configured setting whenever a
	// client PUTs a body without these keys. Same reasoning as the pointer fields
	// on UpdateGroupReceiptSettingsCommand.
	RefreshTokenValidForHours    *int `json:"refreshTokenValidForHours"`
	McpRefreshTokenValidForHours *int `json:"mcpRefreshTokenValidForHours"`
	TempFileRetentionHours       *int `json:"tempFileRetentionHours"`
	// ServerPublicUrl is a pointer for the same reason, and the stakes are higher
	// than a reset default: it builds the OIDC redirect URI registered at each
	// identity provider, so clearing it makes every provider reject the callback
	// with a redirect_uri mismatch — a failure that surfaces at the IdP rather
	// than in our logs.
	ServerPublicUrl *string `json:"serverPublicUrl"`
	// TimeZone is an IANA zone name. A pointer for the same omitted-key reason
	// as the lifetimes above: a client that does not know the field (an older
	// build) must not reset a configured zone to "".
	TimeZone *string `json:"timeZone"`
}

func (command *UpsertSystemSettingsCommand) LoadDataFromRequest(w http.ResponseWriter, r *http.Request) error {
	bytes, err := utils.GetBodyData(w, r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &command)
	if err != nil {
		return err
	}

	return nil
}

func (command *UpsertSystemSettingsCommand) Validate() structs.ValidatorError {
	vErr := structs.ValidatorError{}
	errorMap := make(map[string]string)
	vErr.Errors = errorMap

	if command.EmailPollingInterval < 0 {
		errorMap["emailPollingInterval"] = "Email polling interval must be greater than 0"
	}

	if command.ReceiptProcessingSettingsId != nil && *command.ReceiptProcessingSettingsId <= 0 {
		errorMap["receiptProcessingSettingsId"] = "Invalid receipt processing settings ID"
	}

	if command.FallbackReceiptProcessingSettingsId != nil && *command.FallbackReceiptProcessingSettingsId <= 0 {
		errorMap["fallbackReceiptProcessingSettingsId"] = "Invalid fallback receipt processing settings ID"
	}

	if command.ReceiptProcessingSettingsId == nil && command.FallbackReceiptProcessingSettingsId != nil {
		errorMap["fallbackReceiptProcessingSettingsId"] = "Fallback receipt processing settings ID cannot be set without receipt processing settings ID"
	}

	if command.ReceiptProcessingSettingsId != nil &&
		command.FallbackReceiptProcessingSettingsId != nil &&
		*command.ReceiptProcessingSettingsId ==
			*command.FallbackReceiptProcessingSettingsId {
		errorMap["fallbackReceiptProcessingSettingsId"] = "Fallback receipt processing settings ID cannot be the same as receipt processing settings ID"
	}

	if len(command.CurrencySymbolPosition) == 0 {
		errorMap["currencySymbolPosition"] = "Currency symbol position is required"
	}

	if len(command.CurrencyThousandthsSeparator) == 0 {
		errorMap["currencyThousandthsSeparator"] = "Currency thousandths separator is required"
	}

	if len(command.CurrencyDecimalSeparator) == 0 {
		errorMap["currencyDecimalSeparator"] = "Currency decimal separator is required"
	}

	if command.PdfDpi != 0 && (command.PdfDpi < 72 || command.PdfDpi > 1200) {
		errorMap["pdfDpi"] = "PDF DPI must be between 72 and 1200"
	}

	if command.TaskConcurrency < 0 {
		errorMap["taskConcurrency"] = "Task concurrency must be greater than or equal to 0"
	}

	queueNames := models.GetQueueNames()
	if len(command.TaskQueueConfigurations) != len(queueNames) {
		errorMap["taskQueueConfigurations"] = "Task queue configurations must be provided for all queues"
	}

	trimmedMcpPublicUrl := strings.TrimSpace(command.McpPublicUrl)
	if command.McpEnabled && len(trimmedMcpPublicUrl) == 0 {
		errorMap["mcpPublicUrl"] = "A public URL is required to enable the MCP server"
	} else if len(trimmedMcpPublicUrl) > 0 && !isValidAbsoluteUrl(trimmedMcpPublicUrl) {
		errorMap["mcpPublicUrl"] = "MCP public URL must be an absolute origin like https://receipts.example.com"
	}

	trimmedMobileServerUrl := strings.TrimSpace(command.MobileServerUrl)
	if command.ShowLoginQr && len(trimmedMobileServerUrl) == 0 {
		errorMap["mobileServerUrl"] = "A server URL is required to show the login QR code"
	} else if len(trimmedMobileServerUrl) > 0 && !isValidAbsoluteUrl(trimmedMobileServerUrl) {
		errorMap["mobileServerUrl"] = "Mobile server URL must be an absolute URL like https://receipts.example.com/api"
	}

	// A nil pointer means the key was omitted, which leaves the stored value alone
	// and is always valid. An explicit empty string clears it, which is allowed
	// HERE: it is only required once an OIDC provider is enabled, and that is a
	// cross-setting rule a pure validator cannot see.
	//
	// That rule is enforced from BOTH directions, and it needs to be. Enabling a
	// provider without the URL is refused by OidcProviderService; clearing the URL
	// while one is enabled is refused by handlers.serverPublicUrlStillNeeded. Only
	// the first existed at first, so the same broken configuration was reachable by
	// doing the two steps in the other order.
	if command.ServerPublicUrl != nil {
		trimmedServerPublicUrl := strings.TrimSpace(*command.ServerPublicUrl)
		if len(trimmedServerPublicUrl) > 0 && !isValidAbsoluteUrl(trimmedServerPublicUrl) {
			errorMap["serverPublicUrl"] = "Server public URL must be an absolute origin like https://receipts.example.com"
		}
	}

	if msg := validateRefreshTokenValidForHours(command.RefreshTokenValidForHours); len(msg) > 0 {
		errorMap["refreshTokenValidForHours"] = msg
	}

	if msg := validateRefreshTokenValidForHours(command.McpRefreshTokenValidForHours); len(msg) > 0 {
		errorMap["mcpRefreshTokenValidForHours"] = msg
	}

	if msg := validateTempFileRetentionHours(command.TempFileRetentionHours); len(msg) > 0 {
		errorMap["tempFileRetentionHours"] = msg
	}

	if msg := validateTimeZone(command.TimeZone); len(msg) > 0 {
		errorMap["timeZone"] = msg
	}

	return vErr
}

// validateTimeZone requires a sent time zone to be a name the Go runtime can
// load. The embedded time/tzdata (imported by main) makes this independent of
// the host's zoneinfo, so a name accepted here resolves identically everywhere.
// A nil pointer means the key was omitted and is always valid. An explicit ""
// is rejected rather than read as "unset": the column default is "UTC", and a
// blank would only ever reach the resolver's fallback.
func validateTimeZone(timeZone *string) string {
	if timeZone == nil {
		return ""
	}

	// time.LoadLocation reads "" as UTC and "Local" as the process zone; the
	// latter is exactly the implicit dependency this setting exists to remove.
	name := *timeZone
	if name == "" || name == "Local" {
		return invalidTimeZoneMessage
	}

	if _, err := time.LoadLocation(name); err != nil {
		return invalidTimeZoneMessage
	}

	return ""
}

const invalidTimeZoneMessage = "Time zone must be an IANA time zone name like America/New_York"

// validateTempFileRetentionHours bounds the temp-file retention window, returning
// an empty string when the value is acceptable.
//
// Same two-flavoured "no value" as the refresh-token lifetimes: a nil pointer
// means the key was omitted and leaves the stored value alone, while an explicit
// 0 means "unset" and lets the read-side clamp fall back to the built-in default.
func validateTempFileRetentionHours(hours *int) string {
	if hours == nil || *hours == 0 {
		return ""
	}

	if *hours < MinTempFileRetentionHours || *hours > MaxTempFileRetentionHours {
		return "Temporary file retention must be between 24 and 8760 hours (1 year)"
	}

	return ""
}

// validateRefreshTokenValidForHours bounds a refresh-token lifetime, returning an
// empty string when the value is acceptable.
//
// A nil pointer means the key was omitted, which leaves the stored value alone
// (see ApplyOmittedValues) and is always valid. An explicit 0 means "unset":
// the read side falls back to the built-in default. Shared by the app and MCP
// settings so the two cannot drift.
func validateRefreshTokenValidForHours(hours *int) string {
	if hours == nil || *hours == 0 {
		return ""
	}

	if *hours < MinRefreshTokenValidForHours || *hours > MaxRefreshTokenValidForHours {
		return "Refresh token lifetime must be between 1 and 720 hours (30 days)"
	}

	return ""
}

// isValidAbsoluteUrl reports whether the value is an absolute http(s) URL.
// A scheme and host are required; paths/queries/fragments are tolerated.
//
// Embedded credentials (https://user:token@host) are rejected: both settings
// this guards are published verbatim to unauthenticated clients — the mobile
// server URL is encoded into the login QR served by the public /featureConfig,
// and the MCP public URL is echoed in the OAuth discovery metadata.
func isValidAbsoluteUrl(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || parsed.User != nil {
		return false
	}

	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func (command *UpsertSystemSettingsCommand) ToSystemSettings(id uint) (models.SystemSettings, error) {
	var systemSettings models.SystemSettings

	bytes, err := json.Marshal(command)
	if err != nil {
		return systemSettings, err
	}

	err = json.Unmarshal(bytes, &systemSettings)
	if err != nil {
		return systemSettings, err
	}

	systemSettings.ID = id
	for _, config := range systemSettings.TaskQueueConfigurations {
		config.SystemSettingsId = id
	}

	return systemSettings, nil
}

// OmittedColumns names the pointer-backed fields the request did not send, so
// the repository can leave those columns out of the UPDATE entirely. It covers
// every such field -- the two token lifetimes, the retention window, the public
// URL and the time zone -- which is why it is not named for the lifetimes it
// started out holding.
//
// Skipping the column is what makes a concurrent update safe. Copying the stored
// value onto the row instead (see ApplyOmittedValues) would still write it, so
// two requests that each set one field and omit the other would clobber each
// other with the values they read before the write. A column that is never
// written cannot be clobbered, and unlike a row lock this works identically on
// SQLite, MySQL and Postgres.
func (command *UpsertSystemSettingsCommand) OmittedColumns() []string {
	columns := make([]string, 0, 5)

	if command.RefreshTokenValidForHours == nil {
		columns = append(columns, "RefreshTokenValidForHours")
	}

	if command.McpRefreshTokenValidForHours == nil {
		columns = append(columns, "McpRefreshTokenValidForHours")
	}

	if command.TempFileRetentionHours == nil {
		columns = append(columns, "TempFileRetentionHours")
	}

	if command.ServerPublicUrl == nil {
		columns = append(columns, "ServerPublicUrl")
	}

	if command.TimeZone == nil {
		columns = append(columns, "TimeZone")
	}

	return columns
}

// ApplyOmittedValues carries the stored value of every omitted pointer field onto
// the settings a PUT is about to write.
//
// ToSystemSettings round-trips the command through JSON, so a nil pointer lands
// as the zero value on the model. The columns themselves are excluded from the
// UPDATE by OmittedColumns, so this exists purely so the object echoed back in
// the response carries the stored value rather than a misleading zero.
func (command *UpsertSystemSettingsCommand) ApplyOmittedValues(existing models.SystemSettings, updated *models.SystemSettings) {
	if command.RefreshTokenValidForHours == nil {
		updated.RefreshTokenValidForHours = existing.RefreshTokenValidForHours
	}

	if command.McpRefreshTokenValidForHours == nil {
		updated.McpRefreshTokenValidForHours = existing.McpRefreshTokenValidForHours
	}

	if command.TempFileRetentionHours == nil {
		updated.TempFileRetentionHours = existing.TempFileRetentionHours
	}

	if command.ServerPublicUrl == nil {
		updated.ServerPublicUrl = existing.ServerPublicUrl
	}

	if command.TimeZone == nil {
		updated.TimeZone = existing.TimeZone
	}
}
