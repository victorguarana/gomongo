package mongo

type IMongo interface {
	All(string) (interface{}, error)
	Create(string, interface{}) error
	Delete(string, interface{}) error
	First(string) (interface{}, error)
}
