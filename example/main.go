package main

import (
	"context"
	"fmt"
	"github.com/1iza/safeMap"
)

/*
写入跟读取通过channel操作，不需要加锁
创建对象时，存入一个ctx跟一个空map
*/

func main() {
	m, err := safemap.NewSafeMap(context.Background(), map[int]int{})
	if err != nil {
		panic(err)
	}

	if err = m.Set(1, 1); err != nil {
		fmt.Printf("set :%+v \n", err)
	}

	if v, err := m.Get(1); err != nil {
		fmt.Printf("get err :%+v \n", err)
	} else {
		fmt.Printf("get :%+v \n", v.(int))
	}

	if err = m.Del(1); err != nil {
		fmt.Printf("del :%+v \n", err)
	}

	if v, err := m.Get(1); err != nil {
		fmt.Printf("get err :%+v \n", err)
	} else {
		fmt.Printf("get :%+v \n", v.(int))
	}
	m.Close()
}
