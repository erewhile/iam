package consts

import "time"

const (
	EnvJwtKid       = "IAM_JWT_KID"
	DefaultJwtKid   = "erewhile-iam-public-key"
	AccessTokenTTL  = 5 * time.Minute
	RefreshTokenTTL = 24 * time.Hour
)
