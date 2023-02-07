# diskache

Lightweight Golang disk cache.

## Get

```Shell
$ go get github.com/pushyzheng/diskache
```

## Use

```Go
import (
    "fmt"
    "github.com/GitbookIO/diskache"
)

// Create an instance
opts := diskache.Opts{
    Directory: "diskache_place",
}
dc, err := diskache.New(&opts)
if err != nil {
    log.Fatalln(err)
}

// Add data to cache
spelling := []byte{'g', 'o', 'l', 'a', 'n', 'g'}
err := dc.Set("spelling", spelling)
if err != nil {
    fmt.Println("Impossible to set data in cache")
}

// Add data to cache with expired time (1s)
data := []byte("Hello World")
err = cache.SetExpired("spelling-expired", data, time.Second.Milliseconds())
if err != nil {
    log.Fatalln(err)
}

// Read from cache
cached, inCache := dc.Get("spelling")
if inCache {
    fmt.Println(string(cached))
}

// Delete from cache
ok := dc.Delete("spelling")
if ok {
    fmt.Println("delete cache succeed")
}

// Read stats
stats := dc.Stats()
reflect.DeepEqual(stats, Stats{
    Directory: "diskache_place",
    Items:     1,
})

// Cleanup
err = dc.Clean()
if err != nil {
    fmt.Println("Impossible to clean cache")
}
```
