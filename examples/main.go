package main

import (
	"fmt"
	"github.com/pushyzheng/diskache"
	"log"
	"time"
)

func main() {
	opts := diskache.Opts{
		Directory: "tmp",
	}
	cache, err := diskache.New(&opts)
	if err != nil {
		log.Fatalln(err)
	}
	err = cache.SetStr("foo", "Hello World")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(cache.GetStr("foo"))

	err = cache.SetExpired("foo-expired", []byte("Hello World"), 500)
	if err != nil {
		log.Fatalln(err)
	}
	v, exists := cache.GetStr("foo-expired")
	if exists {
		log.Println("The value of 'foo-expired' is:", v)
	}
	time.Sleep(time.Second)
	_, exists = cache.GetStr("foo-expired")
	if !exists {
		log.Println("The value of 'foo-expired' don't exists", v)
	}
}
