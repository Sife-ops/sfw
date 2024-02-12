// Code generated from Pkl module `sfw`. DO NOT EDIT.
package config

type Postgres interface {
	Server

	GetDatabase() string

	GetUsername() string

	GetPassword() string
}

var _ Postgres = (*PostgresImpl)(nil)

type PostgresImpl struct {
	Database string `pkl:"database"`

	Username string `pkl:"username"`

	Password string `pkl:"password"`

	Host string `pkl:"host"`
}

func (rcv *PostgresImpl) GetDatabase() string {
	return rcv.Database
}

func (rcv *PostgresImpl) GetUsername() string {
	return rcv.Username
}

func (rcv *PostgresImpl) GetPassword() string {
	return rcv.Password
}

func (rcv *PostgresImpl) GetHost() string {
	return rcv.Host
}
