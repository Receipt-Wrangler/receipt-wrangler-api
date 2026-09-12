/// User-facing copy for the receipt-entry affordances (the Scan/Add bottom-nav
/// slot, its long-press menu, the receipts-screen overflow menu, and the
/// "Quick Scan unavailable" banner on the manual form).
///
/// Centralised because the same reason is surfaced in more than one place — the
/// snackbar shown when the Quick Scan sheet is opened directly, and the banner
/// shown when the Scan tap falls through to manual entry — and the two drifting
/// apart would read as two different problems.

// "Receipt Processing Settings" is the product's name for the entity and stays
// plural even for a single record ("Create Receipt Processing Settings", "...
// deleted successfully" on desktop), so the sentence is built around it rather
// than given a singular article.
const quickScanAiDisabledMessage =
    "Receipt Processing Settings must be configured to use Quick Scan. "
    "Contact your administrator for more information.";

const quickScanNoPermissionMessage =
    "You don't have Quick Scan permission here, so this opens manual entry "
    "instead.";

/// [quickScanNoPermissionMessage]'s wording when the group being acted on is
/// known (i.e. the user is inside a single group rather than on group-select).
String quickScanNoPermissionMessageForGroup(String groupName) =>
    "You don't have Quick Scan permission in $groupName, so this opens manual "
    "entry instead.";

const quickScanUnavailableTitle = "Quick Scan unavailable";

const noReceiptEntryPermissionMessage =
    "You don't have permission to add receipts here.";

const cameraDeniedFallbackMessage =
    "Camera access is off — pick a photo instead.";

/// The three picker sources fail for different reasons, so they say different
/// things — a single "gallery" message would misdescribe two of the three.
const photoPickerUnavailableMessage =
    "Couldn't open your photos on this device.";

const filePickerUnavailableMessage =
    "Couldn't open the file picker on this device.";

const scannerUnavailableMessage =
    "Couldn't open the scanner on this device.";

const addManualReceiptLabel = "Add Manual Receipt";
const quickScanLabel = "Quick Scan";

/// The photo library and the document browser are separate entries because
/// they are separate OS pickers with different reach: only [uploadFileLabel]
/// can produce a PDF, and on iOS only [uploadPhotoLabel] can reach the camera
/// roll at all.
const uploadPhotoLabel = "Upload Photo";
const uploadFileLabel = "Upload File";
const uploadFromCameraLabel = "Upload from Camera";

/// `PopupMenuButton`'s default tooltip is the localized "Show menu", which says
/// nothing about what the menu is for.
const uploadSourceTooltip = "Add a photo or file";

const enterDetailsManuallyLabel = "Enter details manually instead";
const quickScanQueuedMessage = "Queued — we'll fill in the details for you.";

/// Extra key carrying a [QuickScanBlockedReason] into `/receipts/add`, so the
/// form can explain why the Scan tap landed there. Only the tap path sets it —
/// a deliberate "Add Manual Receipt" never shows the banner.
const quickScanBlockedReasonExtraKey = "quickScanBlockedReason";

/// Extra key carrying the group name the blocked reason refers to.
const quickScanBlockedGroupExtraKey = "quickScanBlockedGroup";
