package layer2

type (
	Marshaler interface {
		Marshal(v interface{}) ([]byte, error)
	}

	MarshalFunc func(v interface{}) ([]byte, error)

	AdvancedMarshaler interface {
		Marshal(v interface{}) (bt []byte, params map[string]string, err error)
	}

	AdvancedMarshalFunc func(v interface{}) (bt []byte, params map[string]string, err error)
)

func RegisterMarshaler(mediatype string, m Marshaler) error {
	return mediaManager.addMarshaler(mediaType(mediatype), marshalWrapper{m})
}

func (m MarshalFunc) Marshal(v interface{}) ([]byte, error) {
	return m(v)
}

func (m AdvancedMarshalFunc) Marshal(v interface{}) ([]byte, map[string]string, error) {
	return m(v)
}

func RegisterAdvancedMarshaler(mediatype string, m AdvancedMarshaler) error {
	return mediaManager.addMarshaler(mediaType(mediatype), m)
}
