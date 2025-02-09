package ivy

type KV struct {
	m map[any]any
}

// Set sets a key into the request level Key-Value store
func (kv *KV) Set(k any, v any) {
	if kv.m == nil {
		kv.m = make(map[any]any, 1)
	}
	kv.m[k] = v
}

// Get fetches the value of key in request level KV store
// in case, key is not present default value is returned
func (kv *KV) Get(k any) any {
	return kv.m[k]
}

func (kv *KV) All() map[any]any {
	return kv.m
}

// Lookup fetches the value of key in request level KV store
// in case default value is not present, ok will be false
func (kv *KV) Lookup(k any) (any, bool) {
	v, ok := kv.m[k]
	return v, ok
}
