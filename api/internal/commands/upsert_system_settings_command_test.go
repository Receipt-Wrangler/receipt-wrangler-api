package commands

import (
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/utils"
	"testing"
)

func validSystemSettingsCommand() UpsertSystemSettingsCommand {
	queueNames := models.GetQueueNames()
	configs := make([]UpsertTaskQueueConfigurationCommand, len(queueNames))

	return UpsertSystemSettingsCommand{
		CurrencySymbolPosition:       models.START,
		CurrencyThousandthsSeparator: models.COMMA,
		CurrencyDecimalSeparator:     models.DOT,
		TaskConcurrency:              1,
		EmailPollingInterval:         60,
		TaskQueueConfigurations:      configs,
	}
}

func TestUpsertSystemSettingsCommand_Validate_ValidInputs(t *testing.T) {
	primaryId := uint(1)
	fallbackId := uint(2)

	tests := map[string]struct {
		command UpsertSystemSettingsCommand
	}{
		"valid minimal": {
			command: validSystemSettingsCommand(),
		},
		"valid with processing settings": {
			command: func() UpsertSystemSettingsCommand {
				cmd := validSystemSettingsCommand()
				cmd.ReceiptProcessingSettingsId = &primaryId
				cmd.FallbackReceiptProcessingSettingsId = &fallbackId
				return cmd
			}(),
		},
		"valid with zero email polling interval": {
			command: func() UpsertSystemSettingsCommand {
				cmd := validSystemSettingsCommand()
				cmd.EmailPollingInterval = 0
				return cmd
			}(),
		},
		"valid with zero task concurrency": {
			command: func() UpsertSystemSettingsCommand {
				cmd := validSystemSettingsCommand()
				cmd.TaskConcurrency = 0
				return cmd
			}(),
		},
		"valid with mcp enabled and a public url": {
			command: func() UpsertSystemSettingsCommand {
				cmd := validSystemSettingsCommand()
				cmd.McpEnabled = true
				cmd.McpPublicUrl = "https://receipts.example.com"
				return cmd
			}(),
		},
		"valid with mcp disabled and no public url": {
			command: func() UpsertSystemSettingsCommand {
				cmd := validSystemSettingsCommand()
				cmd.McpEnabled = false
				cmd.McpPublicUrl = ""
				return cmd
			}(),
		},
		"valid with login qr enabled and a mobile server url": {
			command: func() UpsertSystemSettingsCommand {
				cmd := validSystemSettingsCommand()
				cmd.ShowLoginQr = true
				cmd.MobileServerUrl = "https://receipts.example.com/api"
				return cmd
			}(),
		},
		"valid with login qr disabled and no mobile server url": {
			command: func() UpsertSystemSettingsCommand {
				cmd := validSystemSettingsCommand()
				cmd.ShowLoginQr = false
				cmd.MobileServerUrl = ""
				return cmd
			}(),
		},
		// A nil pointer means the key was absent from the request body. It must
		// validate, because omission leaves the stored value alone.
		"valid with omitted refresh token lifetimes": {
			command: func() UpsertSystemSettingsCommand {
				cmd := validSystemSettingsCommand()
				cmd.RefreshTokenValidForHours = nil
				cmd.McpRefreshTokenValidForHours = nil
				return cmd
			}(),
		},
		// An explicit zero means "unset": the read side falls back to the
		// default instead.
		"valid with unset refresh token lifetimes": {
			command: func() UpsertSystemSettingsCommand {
				cmd := validSystemSettingsCommand()
				cmd.RefreshTokenValidForHours = intPtr(0)
				cmd.McpRefreshTokenValidForHours = intPtr(0)
				return cmd
			}(),
		},
		"valid with the minimum refresh token lifetimes": {
			command: func() UpsertSystemSettingsCommand {
				cmd := validSystemSettingsCommand()
				cmd.RefreshTokenValidForHours = intPtr(1)
				cmd.McpRefreshTokenValidForHours = intPtr(1)
				return cmd
			}(),
		},
		"valid with the maximum refresh token lifetimes": {
			command: func() UpsertSystemSettingsCommand {
				cmd := validSystemSettingsCommand()
				cmd.RefreshTokenValidForHours = intPtr(720)
				cmd.McpRefreshTokenValidForHours = intPtr(720)
				return cmd
			}(),
		},
		// Same two-flavoured "no value" as the lifetimes above: omitted leaves
		// the stored window alone, an explicit zero means "use the default".
		"valid with an omitted temp file retention": {
			command: func() UpsertSystemSettingsCommand {
				cmd := validSystemSettingsCommand()
				cmd.TempFileRetentionHours = nil
				return cmd
			}(),
		},
		"valid with an unset temp file retention": {
			command: func() UpsertSystemSettingsCommand {
				cmd := validSystemSettingsCommand()
				cmd.TempFileRetentionHours = intPtr(0)
				return cmd
			}(),
		},
		"valid with the minimum temp file retention": {
			command: func() UpsertSystemSettingsCommand {
				cmd := validSystemSettingsCommand()
				cmd.TempFileRetentionHours = intPtr(MinTempFileRetentionHours)
				return cmd
			}(),
		},
		"valid with the maximum temp file retention": {
			command: func() UpsertSystemSettingsCommand {
				cmd := validSystemSettingsCommand()
				cmd.TempFileRetentionHours = intPtr(MaxTempFileRetentionHours)
				return cmd
			}(),
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			vErr := test.command.Validate()

			if len(vErr.Errors) > 0 {
				utils.PrintTestError(t, len(vErr.Errors), 0)
			}
		})
	}
}

func TestUpsertSystemSettingsCommand_Validate_InvalidInputs(t *testing.T) {
	zeroId := uint(0)
	sameId := uint(5)

	tests := map[string]struct {
		modify        func(cmd *UpsertSystemSettingsCommand)
		expectedError string
	}{
		"negative email polling interval": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.EmailPollingInterval = -1 },
			expectedError: "emailPollingInterval",
		},
		"invalid receipt processing settings id": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.ReceiptProcessingSettingsId = &zeroId },
			expectedError: "receiptProcessingSettingsId",
		},
		"invalid fallback receipt processing settings id": {
			modify: func(cmd *UpsertSystemSettingsCommand) {
				id := uint(1)
				cmd.ReceiptProcessingSettingsId = &id
				cmd.FallbackReceiptProcessingSettingsId = &zeroId
			},
			expectedError: "fallbackReceiptProcessingSettingsId",
		},
		"fallback without primary": {
			modify: func(cmd *UpsertSystemSettingsCommand) {
				id := uint(1)
				cmd.FallbackReceiptProcessingSettingsId = &id
			},
			expectedError: "fallbackReceiptProcessingSettingsId",
		},
		"fallback same as primary": {
			modify: func(cmd *UpsertSystemSettingsCommand) {
				cmd.ReceiptProcessingSettingsId = &sameId
				sameIdCopy := sameId
				cmd.FallbackReceiptProcessingSettingsId = &sameIdCopy
			},
			expectedError: "fallbackReceiptProcessingSettingsId",
		},
		"missing currency symbol position": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.CurrencySymbolPosition = "" },
			expectedError: "currencySymbolPosition",
		},
		"missing currency thousandths separator": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.CurrencyThousandthsSeparator = "" },
			expectedError: "currencyThousandthsSeparator",
		},
		"missing currency decimal separator": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.CurrencyDecimalSeparator = "" },
			expectedError: "currencyDecimalSeparator",
		},
		"negative task concurrency": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.TaskConcurrency = -1 },
			expectedError: "taskConcurrency",
		},
		"wrong queue config count": {
			modify: func(cmd *UpsertSystemSettingsCommand) {
				cmd.TaskQueueConfigurations = []UpsertTaskQueueConfigurationCommand{}
			},
			expectedError: "taskQueueConfigurations",
		},
		"mcp enabled without a public url": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.McpEnabled = true; cmd.McpPublicUrl = "" },
			expectedError: "mcpPublicUrl",
		},
		"mcp public url missing scheme": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.McpPublicUrl = "receipts.example.com" },
			expectedError: "mcpPublicUrl",
		},
		"mcp public url with unsupported scheme": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.McpPublicUrl = "ftp://receipts.example.com" },
			expectedError: "mcpPublicUrl",
		},
		"mcp public url with embedded credentials": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.McpPublicUrl = "https://user:token@receipts.example.com" },
			expectedError: "mcpPublicUrl",
		},
		"login qr enabled without a mobile server url": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.ShowLoginQr = true; cmd.MobileServerUrl = "" },
			expectedError: "mobileServerUrl",
		},
		"mobile server url missing scheme": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.MobileServerUrl = "receipts.example.com/api" },
			expectedError: "mobileServerUrl",
		},
		"mobile server url with unsupported scheme": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.MobileServerUrl = "ftp://receipts.example.com" },
			expectedError: "mobileServerUrl",
		},
		// The login QR is served to unauthenticated clients, so credentials in
		// the URL would be published as a scannable code.
		"mobile server url with embedded credentials": {
			modify: func(cmd *UpsertSystemSettingsCommand) {
				cmd.MobileServerUrl = "https://user:token@receipts.example.com/api"
			},
			expectedError: "mobileServerUrl",
		},
		"mobile server url with an embedded username only": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.MobileServerUrl = "https://user@receipts.example.com/api" },
			expectedError: "mobileServerUrl",
		},
		"negative refresh token lifetime": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.RefreshTokenValidForHours = intPtr(-1) },
			expectedError: "refreshTokenValidForHours",
		},
		"refresh token lifetime above the maximum": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.RefreshTokenValidForHours = intPtr(721) },
			expectedError: "refreshTokenValidForHours",
		},
		"negative mcp refresh token lifetime": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.McpRefreshTokenValidForHours = intPtr(-1) },
			expectedError: "mcpRefreshTokenValidForHours",
		},
		"mcp refresh token lifetime above the maximum": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.McpRefreshTokenValidForHours = intPtr(721) },
			expectedError: "mcpRefreshTokenValidForHours",
		},
		"temp file retention below the floor": {
			modify: func(cmd *UpsertSystemSettingsCommand) {
				cmd.TempFileRetentionHours = intPtr(MinTempFileRetentionHours - 1)
			},
			expectedError: "tempFileRetentionHours",
		},
		"temp file retention above the ceiling": {
			modify: func(cmd *UpsertSystemSettingsCommand) {
				cmd.TempFileRetentionHours = intPtr(MaxTempFileRetentionHours + 1)
			},
			expectedError: "tempFileRetentionHours",
		},
		"negative temp file retention": {
			modify:        func(cmd *UpsertSystemSettingsCommand) { cmd.TempFileRetentionHours = intPtr(-1) },
			expectedError: "tempFileRetentionHours",
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			cmd := validSystemSettingsCommand()
			test.modify(&cmd)

			vErr := cmd.Validate()

			if len(vErr.Errors) == 0 {
				utils.PrintTestError(t, len(vErr.Errors), "greater than 0")
			}

			if _, exists := vErr.Errors[test.expectedError]; !exists {
				utils.PrintTestError(t, "error should exist for field", test.expectedError)
			}
		})
	}
}

func TestUpsertSystemSettingsCommand_Validate_MultipleErrors(t *testing.T) {
	command := UpsertSystemSettingsCommand{
		EmailPollingInterval: -1,
		TaskConcurrency:      -1,
	}

	vErr := command.Validate()

	if len(vErr.Errors) < 5 {
		utils.PrintTestError(t, len(vErr.Errors), "at least 5")
	}
}

func TestUpsertSystemSettingsCommand_Validate_PdfDpi(t *testing.T) {
	// 0 means "unset / use default" and must be allowed; in-range values pass;
	// out-of-range values are rejected with a pdfDpi error.
	validValues := []int{0, 72, 300, 1200}
	for _, v := range validValues {
		cmd := validSystemSettingsCommand()
		cmd.PdfDpi = v
		vErr := cmd.Validate()
		if _, ok := vErr.Errors["pdfDpi"]; ok {
			utils.PrintTestError(t, "pdfDpi error for valid value "+utils.UintToString(uint(v)), "no error")
		}
	}

	invalidValues := []int{71, 1201, 5000}
	for _, v := range invalidValues {
		cmd := validSystemSettingsCommand()
		cmd.PdfDpi = v
		vErr := cmd.Validate()
		if _, ok := vErr.Errors["pdfDpi"]; !ok {
			utils.PrintTestError(t, "no pdfDpi error for invalid value "+utils.UintToString(uint(v)), "pdfDpi error")
		}
	}
}

func TestUpsertSystemSettingsCommand_Validate_RefreshTokenLifetimes(t *testing.T) {
	// 0 means "unset / use default" and must be allowed; 1-720 pass; anything
	// else is rejected. Both fields share one helper, so both are exercised to
	// prove the two error keys are wired to the right field.
	fields := map[string]struct {
		set      func(cmd *UpsertSystemSettingsCommand, hours int)
		errorKey string
	}{
		"refreshTokenValidForHours": {
			set:      func(cmd *UpsertSystemSettingsCommand, hours int) { cmd.RefreshTokenValidForHours = &hours },
			errorKey: "refreshTokenValidForHours",
		},
		"mcpRefreshTokenValidForHours": {
			set:      func(cmd *UpsertSystemSettingsCommand, hours int) { cmd.McpRefreshTokenValidForHours = &hours },
			errorKey: "mcpRefreshTokenValidForHours",
		},
	}

	validValues := []int{0, 1, 24, 168, 720}
	invalidValues := []int{-100, -1, 721, 100000}

	for fieldName, field := range fields {
		for _, hours := range validValues {
			t.Run(fieldName+" accepts "+utils.UintToString(uint(hours)), func(t *testing.T) {
				cmd := validSystemSettingsCommand()
				field.set(&cmd, hours)

				if vErr := cmd.Validate(); len(vErr.Errors) > 0 {
					utils.PrintTestError(t, vErr.Errors, "no errors")
				}
			})
		}

		for _, hours := range invalidValues {
			t.Run(fieldName+" rejects out of range value", func(t *testing.T) {
				cmd := validSystemSettingsCommand()
				field.set(&cmd, hours)

				vErr := cmd.Validate()
				if _, exists := vErr.Errors[field.errorKey]; !exists {
					utils.PrintTestError(t, vErr.Errors, "an error for "+field.errorKey)
				}
			})
		}
	}
}

func intPtr(value int) *int {
	return &value
}

func stringPtr(value string) *string {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

// OmittedColumns is what keeps a partial PUT from resetting a field it
// never mentioned: the repository writes every column (Select("*")), so a value
// that is not excluded here is silently overwritten with zero.
func TestOmittedColumnsNamesEveryUnsentField(t *testing.T) {
	tests := map[string]struct {
		modify   func(cmd *UpsertSystemSettingsCommand)
		expected []string
	}{
		"all omitted": {
			modify:   func(cmd *UpsertSystemSettingsCommand) {},
			expected: []string{"RefreshTokenValidForHours", "McpRefreshTokenValidForHours", "TempFileRetentionHours", "ServerPublicUrl", "TimeZone", "SecureCookies"},
		},
		"retention sent, lifetimes omitted": {
			modify: func(cmd *UpsertSystemSettingsCommand) {
				cmd.TempFileRetentionHours = intPtr(48)
			},
			expected: []string{"RefreshTokenValidForHours", "McpRefreshTokenValidForHours", "ServerPublicUrl", "TimeZone", "SecureCookies"},
		},
		"lifetimes sent, retention omitted": {
			modify: func(cmd *UpsertSystemSettingsCommand) {
				cmd.RefreshTokenValidForHours = intPtr(24)
				cmd.McpRefreshTokenValidForHours = intPtr(24)
			},
			expected: []string{"TempFileRetentionHours", "ServerPublicUrl", "TimeZone", "SecureCookies"},
		},
		"only time zone sent": {
			modify: func(cmd *UpsertSystemSettingsCommand) {
				cmd.TimeZone = stringPtr("America/New_York")
			},
			expected: []string{"RefreshTokenValidForHours", "McpRefreshTokenValidForHours", "TempFileRetentionHours", "ServerPublicUrl", "SecureCookies"},
		},
		"only secure cookies sent, as false": {
			modify: func(cmd *UpsertSystemSettingsCommand) {
				cmd.SecureCookies = boolPtr(false)
			},
			expected: []string{"RefreshTokenValidForHours", "McpRefreshTokenValidForHours", "TempFileRetentionHours", "ServerPublicUrl", "TimeZone"},
		},
		"all sent": {
			modify: func(cmd *UpsertSystemSettingsCommand) {
				cmd.RefreshTokenValidForHours = intPtr(24)
				cmd.McpRefreshTokenValidForHours = intPtr(24)
				cmd.TempFileRetentionHours = intPtr(48)
				cmd.ServerPublicUrl = stringPtr("https://receipts.example.com")
				cmd.TimeZone = stringPtr("UTC")
				cmd.SecureCookies = boolPtr(true)
			},
			expected: []string{},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			command := UpsertSystemSettingsCommand{}
			test.modify(&command)

			got := command.OmittedColumns()
			if len(got) != len(test.expected) {
				t.Fatalf("expected %v, got %v", test.expected, got)
			}
			for i, column := range test.expected {
				if got[i] != column {
					t.Errorf("column %d = %q, want %q", i, got[i], column)
				}
			}
		})
	}
}

// ApplyOmittedValues exists so the response echoes the stored value rather
// than the misleading zero that ToSystemSettings' JSON round trip produces for a
// nil pointer.
func TestApplyOmittedValuesEchoesStoredValues(t *testing.T) {
	existing := models.SystemSettings{
		RefreshTokenValidForHours:    100,
		McpRefreshTokenValidForHours: 200,
		TempFileRetentionHours:       300,
		ServerPublicUrl:              "https://receipts.example.com",
	}

	command := UpsertSystemSettingsCommand{TempFileRetentionHours: intPtr(48)}
	updated := models.SystemSettings{TempFileRetentionHours: 48}

	command.ApplyOmittedValues(existing, &updated)

	if updated.RefreshTokenValidForHours != 100 || updated.McpRefreshTokenValidForHours != 200 {
		t.Errorf("expected the omitted lifetimes to be carried forward, got %+v", updated)
	}
	if updated.TempFileRetentionHours != 48 {
		t.Errorf("expected the sent retention to survive, got %d", updated.TempFileRetentionHours)
	}
	if updated.ServerPublicUrl != "https://receipts.example.com" {
		t.Errorf("expected the omitted server public url to be carried forward, got %q", updated.ServerPublicUrl)
	}
}

// An omitted time zone echoes the stored one; a sent one survives.
func TestApplyOmittedValuesEchoesTheStoredTimeZone(t *testing.T) {
	existing := models.SystemSettings{TimeZone: "America/New_York"}

	omitted := UpsertSystemSettingsCommand{}
	updated := models.SystemSettings{}
	omitted.ApplyOmittedValues(existing, &updated)
	if updated.TimeZone != "America/New_York" {
		t.Errorf("omitted time zone = %q, want the stored America/New_York", updated.TimeZone)
	}

	sent := UpsertSystemSettingsCommand{TimeZone: stringPtr("Europe/Berlin")}
	updated = models.SystemSettings{TimeZone: "Europe/Berlin"}
	sent.ApplyOmittedValues(existing, &updated)
	if updated.TimeZone != "Europe/Berlin" {
		t.Errorf("sent time zone = %q, want Europe/Berlin", updated.TimeZone)
	}
}

// An omitted secureCookies echoes the stored value; a sent one -- false
// included -- survives. A false that is sent must not be mistaken for omitted.
func TestApplyOmittedValuesEchoesTheStoredSecureCookies(t *testing.T) {
	existing := models.SystemSettings{SecureCookies: true}

	omitted := UpsertSystemSettingsCommand{}
	updated := models.SystemSettings{}
	omitted.ApplyOmittedValues(existing, &updated)
	if !updated.SecureCookies {
		t.Error("omitted secureCookies should echo the stored true")
	}

	sent := UpsertSystemSettingsCommand{SecureCookies: boolPtr(false)}
	updated = models.SystemSettings{SecureCookies: false}
	sent.ApplyOmittedValues(existing, &updated)
	if updated.SecureCookies {
		t.Error("a sent false should survive, not be replaced by the stored true")
	}
}

// A sent time zone must be a name the runtime can load. Omitting it is always
// valid; "Local" and "" are refused because they would read as the server
// process's zone and as UTC respectively, never as what the admin meant.
func TestUpsertSystemSettingsCommand_Validate_TimeZone(t *testing.T) {
	valid := []string{"UTC", "America/New_York", "Europe/Berlin", "Asia/Kolkata", "Australia/Sydney"}
	invalid := []string{"", "Local", "Not/AZone", "America/New York", " UTC", "../../etc/passwd", "EST5EDT,M3.2.0,M11.1.0"}

	t.Run("omitted", func(t *testing.T) {
		cmd := validSystemSettingsCommand()
		if vErr := cmd.Validate(); len(vErr.Errors) > 0 {
			t.Errorf("omitted time zone: unexpected errors %v", vErr.Errors)
		}
	})

	for _, name := range valid {
		t.Run("accepts "+name, func(t *testing.T) {
			cmd := validSystemSettingsCommand()
			cmd.TimeZone = stringPtr(name)
			if vErr := cmd.Validate(); len(vErr.Errors) > 0 {
				t.Errorf("unexpected errors %v", vErr.Errors)
			}
		})
	}

	for _, name := range invalid {
		t.Run("rejects "+name, func(t *testing.T) {
			cmd := validSystemSettingsCommand()
			cmd.TimeZone = stringPtr(name)
			vErr := cmd.Validate()
			if _, exists := vErr.Errors["timeZone"]; !exists {
				t.Errorf("expected a timeZone error for %q, got %v", name, vErr.Errors)
			}
		})
	}
}
