package utils

func ToPointer[T any](s T) *T {
	return &s
}
