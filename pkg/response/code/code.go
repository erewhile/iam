package code

type Code int

const (
	Success             Code = 200
	Unauthorized        Code = 401
	Forbidden           Code = 403
	NotFound            Code = 404
	InternalServerError Code = 500
	Custom              Code = -621
)

const (
	Parameter Code = 10000 + iota
)

var statusMessages = map[Code]string{
	Success:             "OK",
	Unauthorized:        "Unauthorized",
	Forbidden:           "Forbidden",
	NotFound:            "Not Found",
	InternalServerError: "Internal Server Error",

	Parameter: "Parameter error",
}

func (c Code) Value() int {
	return int(c)
}

func (c Code) Message() string {
	if msg, exists := statusMessages[c]; exists {
		return msg
	}
	return "Unknown Error"
}
