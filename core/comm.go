package core

import "syscall"

type FDComm struct {
	FD int
}

func (f FDComm) Write(b []byte) (int, error) {
	return syscall.Write(f.FD, b)
}

func (f FDComm) Read(b []byte) (int, error) {
	return syscall.Read(f.FD, b)
}
