package types

// StatusThing is a "thing" that can have a status
type StatusThing struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      Status `json:"status"`
}
