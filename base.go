package safemap

type SAFEMAP_OP uint

const (
	SAFEMAP_SET SAFEMAP_OP = iota
	SAFEMAP_DEL
	SAFEMAP_GET
	SAFEMAP_CLEAR
)

type base interface {
	Key() interface{}
	Value() interface{}
	Op() SAFEMAP_OP
}
