package layer2

import (
	"mime"
)

var mediaManager mediaManagerSingleton

type mediaType string

/*
	Allows a Marshaler to implement AdvancedMarshaler interfaces.
*/
type marshalWrapper struct {
	Marshaler
}

func (m marshalWrapper) Marshal(v interface{}) (bt []byte,
	params map[string]string, err error){
	bt, err = m.Marshaler.Marshal(v)
	return
}


type mediaManagerSingleton map[mediaType] AdvancedMarshaler
func (m mediaManagerSingleton) exist() {
	if m == nil {
		m = make(mediaManagerSingleton)
	}
}

func (m mediaManagerSingleton) addMarshaler(mdt mediaType, ma AdvancedMarshaler) error {
	m.exist()
	mt, _, err := mime.ParseMediaType(string(mdt))
	if err != nil {
		return err
	}
	m[mediaType(mt)] = ma
	return nil
}
