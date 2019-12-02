package main

import (
	"fmt"
	"sync"
	"time"
)

var t_ = 0

func main() {
	fmt.Println(time.Now().UnixNano())
	t_struct := &T_Struct{
		map_: []int{},
		data: make(chan int),
	}
	go func() {
		for {
			select {
			case data := <-t_struct.data:
				go print_(data)
			default:
				break
			}
		}
	}()
	for i := 0; i < 1000; i++ {
		t_struct.add_(i)
	}

	fmt.Println(time.Now().UnixNano())
	fmt.Println("map数量", len(t_struct.map_))
	fmt.Println("t_ 数值", t_)
	close(t_struct.data)
}

type T_Struct struct {
	sync.Mutex
	map_ []int
	data chan int
}

func (this *T_Struct) add_(data int) {
	this.Lock()
	defer this.Unlock()
	content := 1000 + data
	this.map_ = append(this.map_, content)
	this.data <- content
}

func print_(data int) {
	t_ = data
	fmt.Println(data)
}
