package types

// Entity types we handle
const (
	TypeListenerName     = "listener"
	TypeRouteName        = "route"
	TypeClusterName      = "cluster"
	TypeDeveloperName    = "developer"
	TypeDeveloperAppName = "developerapp"
	TypeAPIProductName   = "apiproduct"
	TypeKeyName          = "key"
	TypeOAuthName        = "oauth"
	TypeUserName         = "user"
	TypeRoleName         = "role"
)

// NameOf returns the name of an object
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

	case Key:
		return TypeKeyName
	case *Key:
		return TypeKeyName

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

// IDOf returns the id of an object (e.g. developer.Email in case of Developer)
func IDOf(entity interface{}) string {

	switch v := entity.(type) {
	case Listener:
		return v.Name
	case *Listener:
		return v.Name

	case Route:
		return v.Name
	case *Route:
		return v.Name

	case Cluster:
		return v.Name
	case *Cluster:
		return v.Name

	case Developer:
		return v.Email
	case *Developer:
		return v.Email

	case DeveloperApp:
		return v.Name
	case *DeveloperApp:
		return v.Name

	case Key:
		return v.ConsumerKey
	case *Key:
		return v.ConsumerKey

	case APIProduct:
		return v.Name
	case *APIProduct:
		return v.Name

	case User:
		return v.Name
	case *User:
		return v.Name

	case Role:
		return v.Name
	case *Role:
		return v.Name

	default:
		return "unknown"
	}
}
