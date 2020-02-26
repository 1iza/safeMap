package safemap

type opStruct struct {
	key   interface{}
	value interface{}
	op    SAFEMAP_OP
}

func (x *opStruct) Key() interface{}   { return x.key }
func (x *opStruct) Value() interface{} { return x.value }
func (x *opStruct) Op() SAFEMAP_OP     { return x.op }

type getStruct struct {
	key   interface{}
	value chan interface{}
	op    SAFEMAP_OP
}

func (x *getStruct) Key() interface{}   { return x.key }
func (x *getStruct) Value() interface{} { return x.value }
func (x *getStruct) Op() SAFEMAP_OP     { return x.op }
