package transport

type Transport interface {
	Connect() error
	Send([]byte) error
	Recv() ([]byte, error)
	Close() error
}
