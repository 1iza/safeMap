package safemap

import (
	"context"
	"errors"
	"reflect"
)

type safeMap struct {
	noCopy     noCopy
	ctx        context.Context
	cancelFunc context.CancelFunc
	ch         chan base
	reflectmap reflect.Value
	mapType    reflect.Type
}

func NewSafeMap(c context.Context, target interface{}) (*safeMap, error) {
	ctx, cancelfunc := context.WithCancel(c)
	mapType := reflect.TypeOf(target)
	if mapType.Kind() != reflect.Map {
		cancelfunc()
		return nil, errors.New("not a map")
	}
	m := &safeMap{
		ch:         make(chan base),
		ctx:        ctx,
		cancelFunc: cancelfunc,
		reflectmap: reflect.MakeMap(mapType),
		mapType:    mapType,
	}
	go m.eventHandler()
	return m, nil
}

func (m *safeMap) eventHandler() {
	for {
	loop:
		select {
		case <-m.ctx.Done():
			close(m.ch)
			return
		case b, ok := <-m.ch:
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
				if m.mapType.Key() != reflect.TypeOf(k) {
					v.(chan interface{}) <- empty
					goto loop
				}
				val := m.reflectmap.MapIndex(keyVal)
				//是否为空
				if !val.IsValid() {
					v.(chan interface{}) <- empty
					goto loop
				}
				v.(chan interface{}) <- val.Interface()
				goto loop
			case SAFEMAP_SET:
				valueType := reflect.TypeOf(v)
				if m.mapType.Key() != reflect.TypeOf(k) || m.mapType.Elem() != valueType {
					goto loop
				}
				valueVal := reflect.ValueOf(v)
				m.reflectmap.SetMapIndex(keyVal, valueVal)
			case SAFEMAP_DEL:
				if m.mapType.Key() != reflect.TypeOf(k) {
					goto loop
				}
				valueVal := reflect.ValueOf(v)
				m.reflectmap.SetMapIndex(keyVal, valueVal)
			case SAFEMAP_CLEAR:
				for _, key := range m.reflectmap.MapKeys() {
					m.reflectmap.SetMapIndex(key, reflect.Value{})
				}
			default:
				//
			}
		}
	}
}

func (m *safeMap) Close() {
	m.cancelFunc()
}

func (m *safeMap) Set(key interface{}, value interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
			return
		}
	}()
	m.ch <- &opStruct{key, value, SAFEMAP_SET}
	return
}

func (m *safeMap) Del(key interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
			return
		}
	}()
	m.ch <- &opStruct{key, nil, SAFEMAP_DEL}
	return
}

func (m *safeMap) Clear() (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
			return
		}
	}()
	m.ch <- &opStruct{nil, nil, SAFEMAP_CLEAR}
	return
}

func (m *safeMap) Get(key interface{}) (val interface{}, err error) {
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
