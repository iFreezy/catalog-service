package respondent

type Manifest struct {
	Status       int
	Error        string
	ErrorID      string
	ErrorCode    int
	ErrorDetail  string
	ErrorDetails []string
}
