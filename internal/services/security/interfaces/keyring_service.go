package interfaces

type IKeyService interface {
	StorePassword(password string) error
	RetrievePassword() (string, error)
}
