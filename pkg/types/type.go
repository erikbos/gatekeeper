package types

// Entity types we handle
const (
	TypeListenerName     = "listener"
	TypeRouteName        = "route"
	TypeClusterName      = "cluster"
	TypeDeveloperName    = "developer"
	TypeDeveloperAppName = "developerapp"
	TypeAPIProductName   = "apiproduct"
	TypeCredentialName   = "credential"
	TypeOAuthName        = "oauth"
	TypeUserName         = "user"
	TypeRoleName         = "role"
)

// NameOf returns the type name of an object
func NameOf(entity interface{}) string {

	switch entity.(type) {
	case Listener:
		return TypeListenerName
	case *Listener:
		return TypeListenerName

	case Route:
		return TypeRouteName
	case *Route:
		return TypeRouteName

	case Cluster:
		return TypeClusterName
	case *Cluster:
		return TypeClusterName

	case Developer:
		return TypeDeveloperName
	case *Developer:
		return TypeDeveloperName

	case DeveloperApp:
		return TypeDeveloperAppName
	case *DeveloperApp:
		return TypeDeveloperAppName

	case DeveloperAppKey:
		return TypeCredentialName
	case *DeveloperAppKey:
		return TypeCredentialName

	case APIProduct:
		return TypeAPIProductName
	case *APIProduct:
		return TypeAPIProductName

	case User:
		return TypeUserName
	case *User:
		return TypeUserName

	case Role:
		return TypeRoleName
	case *Role:
		return TypeRoleName

	default:
		return "unknown"
	}
}
