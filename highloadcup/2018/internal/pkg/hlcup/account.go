package hlcup

type MaritalStatus string

const (
	StatusAvailable      MaritalStatus = "свободны"
	StatusInRelationship MaritalStatus = "заняты"
	StatusComplicated    MaritalStatus = "всё сложно"
)

type RawAccount struct {
	ID             int32  `json:"id"`
	EMail          string `json:"email"`
	Firstname      string `json:"fname"`
	Surname        string `json:"sname"`
	Phone          string `json:"phone"`
	Sex            string `json:"sex"`
	BirthTimestamp int64  `json:"birth"`
	Country        string `json:"country"`
	City           string `json:"city"`

	JoinedTimestamp int64         `json:"joined"`
	Status          MaritalStatus `json:"status"`
	Interests       []string      `json:"interests"`

	Premium struct {
		StartTimestamp  int64 `json:"start"`
		FinishTimestamp int64 `json:"finish"`
	}

	Likes []struct {
		ID        int32 `json:"id"`
		Timestamp int64 `json:"ts"`
	}
}

type AccountSaver interface {
	Save(account RawAccount) error
}
