package zoox

type HTTPError interface {
	Status() int
	Code() int
	Message() string
	Error() string
	Raw() error
}
