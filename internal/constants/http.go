package constants

import "time"

const ApplicationJson = "application/json"
const ApplicationZip = "application/zip"
const ApplicationPdf = "application/pdf"
const ImageHeic = "image/heic"
const AnyImage = "image/*"
const TextPlain = "text/plain"
const TextCsv = "text/csv"
const MultipartFormMaxSize = 50 << 20

const AiHttpTimeout = 10 * time.Minute
