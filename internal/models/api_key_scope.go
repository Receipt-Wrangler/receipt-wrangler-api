package models

type ApiKeyScope string

const (
	API_KEY_SCOPE_READ       ApiKeyScope = "r"
	API_KEY_SCOPE_WRITE      ApiKeyScope = "w"
	API_KEY_SCOPE_READ_WRITE ApiKeyScope = "rw"
)

func (self *ApiKeyScope) Scan(value string) error {
	*self = ApiKeyScope(value)
	return nil
}

func (self ApiKeyScope) IsValid() bool {
	return self == API_KEY_SCOPE_READ || self == API_KEY_SCOPE_WRITE || self == API_KEY_SCOPE_READ_WRITE
}
