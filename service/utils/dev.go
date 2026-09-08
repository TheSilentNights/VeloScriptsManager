package utils

var IsDev bool

func IsDevEnv() bool {
	return IsDev
}

func SetDev() {
	IsDev = true
}
