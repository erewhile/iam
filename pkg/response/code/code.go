package code

type Code int

const (
	Success             Code = 200
	BadRequest          Code = 400
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
	BadRequest:          "bad request",
	Unauthorized:        "unauthorized",
	Forbidden:           "forbidden",
	NotFound:            "not found",
	InternalServerError: "internal server error",

	Parameter: "parameter error",
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
