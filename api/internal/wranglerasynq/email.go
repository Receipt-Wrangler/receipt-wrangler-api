package wranglerasynq

import (
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"receipt-wrangler/api/internal/logging"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/structs"
	"receipt-wrangler/api/internal/utils"
	"slices"
	"strings"
)

func StartEmailPolling() error {
	systemSettingsRepository := repositories.NewSystemSettingsRepository(nil)
	systemSettings, err := systemSettingsRepository.GetSystemSettings()
	if err != nil {
		return err
	}

	inspector, err := GetAsynqInspector()
	if err != nil {
		return nil
	}
	defer inspector.Close()

	_, err = inspector.DeleteAllScheduledTasks(string(models.EmailPollingQueue))
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_INFO, err.Error())
	}

	// The temp-file sweep used to be scheduled from here; it now lives in
	// StartSystemCleanUpTasks so it runs regardless of email configuration. Drain
	// whatever the old cron left behind on its queue.
	retireEmailReceiptImageCleanupQueue(inspector)

	payload := EmailPollTaskPayload{
		PollAllGroups: true,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	pollTask := asynq.NewTask(EmailPoll, payloadBytes)
	_, err = RegisterTask(GetPollTimeString(systemSettings.EmailPollingInterval), pollTask, models.EmailPollingQueue, 0)
	if err != nil {
		return err
	}

	return nil
}

func GetPollTimeString(pollingInterval int) string {
	return fmt.Sprintf("@every %ds", pollingInterval)
}

// PollEmail reads the mailboxes of the requested groups (all groups with email
// integration enabled when pollAllGroups is set) and queues their receipts.
func PollEmail(pollAllGroups bool, groupIds []string) error {
	groupSettingsRepository := repositories.NewGroupSettingsRepository(nil)
	enabled, err := groupSettingsRepository.GetAllGroupSettings("email_integration_enabled = ?", true)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		return err
	}

	requested := enabled
	if !pollAllGroups {
		requested = make([]models.GroupSettings, 0)
		for _, settings := range enabled {
			if slices.Contains(groupIds, utils.UintToString(settings.GroupId)) {
				requested = append(requested, settings)
			}
		}
	}

	if len(requested) == 0 {
		logging.LogStd(logging.LOG_LEVEL_INFO, "No group settings enabled for email polling")
		return nil
	}

	err = newEmailPoller().poll(requested, enabled)
	if err != nil {
		logging.LogStd(logging.LOG_LEVEL_ERROR, err.Error())
		return err
	}
	return nil
}

// shouldRenderEmailBodyPdfForGroup returns true when the email has an HTML
// body with meaningful text content AND the named group has email body
// processing enabled. Extracted so the gating logic is testable without
// launching chromium and so a group missing from the lookup defaults to
// "do not render" (defensive).
//
// We also require the stripped-text Body to be non-whitespace: many email
// clients (Gmail's web composer in particular) auto-wrap every message in
// a boilerplate HTML shell (<html><body><div style="..."></div></body>),
// so BodyHtml alone is not a good proxy for "there's a body worth
// rendering." Rendering those wrappers produces a near-blank PDF that
// just adds a second, useless image to the LLM call.
func shouldRenderEmailBodyPdfForGroup(
	metadata structs.EmailMetadata,
	groupSettingsId uint,
	groupSettingsLookup map[uint]models.GroupSettings,
) bool {
	if len(metadata.BodyHtml) == 0 {
		return false
	}
	if len(strings.TrimSpace(metadata.Body)) == 0 {
		return false
	}
	gs, ok := groupSettingsLookup[groupSettingsId]
	if !ok {
		return false
	}
	return gs.EmailBodyProcessingEnabled
}
