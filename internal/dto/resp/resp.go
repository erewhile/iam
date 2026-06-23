package resp

type List struct {
	Content any `json:"content"`
	Count   int `json:"count"`
}
