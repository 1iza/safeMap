package safemap

import (
	"context"
	"errors"
	"reflect"
)

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

type SafeMap struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	ch         chan base
	reflectmap reflect.Value
	mapType    reflect.Type
}

func NewSafeMap(c context.Context, target interface{}) (*SafeMap, error) {
	ctx, cancelfunc := context.WithCancel(c)
	mapType := reflect.TypeOf(target)
	if mapType.Kind() != reflect.Map {
		return nil, errors.New("not a map")
	}
	ch := make(chan base)
	reflectmap := reflect.MakeMap(mapType)
	go func() {
		for {
		loop:
			select {
			case <-ctx.Done():
				close(ch)
				return
			case b, ok := <-ch:
				if !ok {
					return
				}
				k, v, op := b.Key(), b.Value(), b.Op()
				keyVal := reflect.ValueOf(k)
				//get value
				switch op {
				case SAFEMAP_GET:
					var empty interface{}
					//map的key类型，跟get的key类型判断
					if mapType.Key() != reflect.TypeOf(k) {
						v.(chan interface{}) <- empty
						goto loop
					}
					val := reflectmap.MapIndex(keyVal)
					//是否为空
					if !val.IsValid() {
						v.(chan interface{}) <- empty
						goto loop
					}
					v.(chan interface{}) <- val.Interface()
					goto loop
				case SAFEMAP_SET:
					valueType := reflect.TypeOf(v)
					if mapType.Key() != reflect.TypeOf(k) || mapType.Elem() != valueType {
						goto loop
					}
					valueVal := reflect.ValueOf(v)
					reflectmap.SetMapIndex(keyVal, valueVal)
				case SAFEMAP_DEL:
					if mapType.Key() != reflect.TypeOf(k) {
						goto loop
					}
					valueVal := reflect.ValueOf(v)
					reflectmap.SetMapIndex(keyVal, valueVal)
				case SAFEMAP_CLEAR:
					for _, key := range reflectmap.MapKeys() {
						reflectmap.SetMapIndex(key, reflect.Value{})
					}
				default:
					//
				}

			}
		}
	}()
	return &SafeMap{
		ch:         ch,
		ctx:        ctx,
		cancelFunc: cancelfunc,
		reflectmap: reflectmap,
		mapType:    mapType,
	}, nil
}

func (m *SafeMap) Close() {
	m.cancelFunc()
}

func (m *SafeMap) Set(key interface{}, value interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
			return
		}
	}()
	m.ch <- &opStruct{key, value, SAFEMAP_SET}
	return
}

func (m *SafeMap) Del(key interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
			return
		}
	}()
	m.ch <- &opStruct{key, nil, SAFEMAP_DEL}
	return
}

func (m *SafeMap) Clear() (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
			return
		}
	}()
	m.ch <- &opStruct{nil, nil, SAFEMAP_CLEAR}
	return
}

func (m *SafeMap) Get(key interface{}) (val interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
			return
		}
	}()
	o := &getStruct{key, make(chan interface{}), SAFEMAP_GET}
	m.ch <- o
	val = <-(o.value)
	if val == nil {
		err = errors.New("not found")
		return
	}
	return

}
