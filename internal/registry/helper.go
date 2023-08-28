package registry

func Get[T any](r Registry, key string) (val T, ok bool) {
	var stored interface{}
	stored, ok = r.Get(key)
	if !ok {
		return
	}
	val, ok = stored.(T)
	return
}
