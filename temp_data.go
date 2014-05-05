package goio

type TempData map[string]string

func (self *TempData) Get(key string) string {
	str, ok := (*self)[key]
	if !ok {
		return ""
	}

	return str
}

func (self *TempData) Set(key, val string) {
	(*self)[key] = val
}
