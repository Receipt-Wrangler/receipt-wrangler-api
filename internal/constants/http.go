package constants

import "time"

const APPLICATION_JSON = "application/json"
const APPLICATION_ZIP = "application/zip"
const APPLICATION_PDF = "application/pdf"
const IMAGE_HEIC = "image/heic"
const ANY_IMAGE = "image/*"
const TEXT_PLAIN = "text/plain"
const MULTIPART_FORM_MAX_SIZE = 50 << 20

const AI_HTTP_TIMEOUT = 10 * time.Minute
