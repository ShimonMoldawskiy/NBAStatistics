package common

type Row interface {
	Scan(dest ...interface{}) error
}

type Rows interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...interface{}) error
}
