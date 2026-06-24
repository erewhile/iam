package consts

const APIVersion = "v1"

var (
	APIBase = "/api/" + APIVersion

	// Auth
	AuthLoginPath   = APIBase + "/auth/login"
	AuthRefreshPath = APIBase + "/auth/refresh"
	AuthLogoutPath  = APIBase + "/auth/logout"

	// OAuth
	OAuthAuthorizePath = APIBase + "/oauth/authorize"
	OAuthTokenPath     = APIBase + "/oauth/token"

	// Resources
	UsersPath        = APIBase + "/users"
	RolesPath        = APIBase + "/roles"
	ApplicationsPath = APIBase + "/applications"
	TokensPath       = APIBase + "/tokens"
)
