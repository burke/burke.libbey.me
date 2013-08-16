Conserving File Descriptors in Go
---
Go makes it very easy to write highly-concurrent applications -- so easy, in
fact, that it exposes you to Operating System limits fairly quickly.

Imagine building a web crawler. You have a long list of URLs, and you want to
fetch the contents of each page. In many languages, you might spawn up a number
of worker threads, each looping, pulling a URL off the queue and fetching it,
until the queue is empty. This is a reasonable model, and it's the most obvious
choice in ruby and languages like it.

```ruby
def crawl(queue)
  1.upto(100).each do
    Thread.new do
      loop { fetch(queue.pop) }
    end
  end
end
```

In Go, on the other hand, we have goroutines. Goroutines are cheaper to create
and to schedule than real threads, and more of them can run at the same time
without losing performance to thread overhead. It's not uncommon to have many
hundreds or thousands of goroutines running concurrently, without incurring much
performance overhead.

In fact, Go is quite capable of handling thousands of HTTP requests at the same
time. However, most operating systems won't let you do that by default. OS X has
a default maximum of 256 open files per process (an HTTP request uses a File
Descriptor). If you're coming to Go from a slow scripting language, this isn't
something you've had to think about on a daily basis, but it will be with Go.

One obvious approach to building this web crawler in Go would be to just spawn a
goroutine for each URL in the queue and let the scheduler figure it out. For
example:

```go
func crawl(urlProducer chan string) {
  for url := range(urlProducer)
    go fetch(url)
  }
}
```

This works really well if the channel never hands off urls above a certain rate.
If it produces too fast -- if you already have 256 open files and you try to
spawn a new HTTP request, one of two things will happen, depending in part on
which version of Go you're using:

1. In Go 1.0, DNS lookup will fail cryptically, claiming "no such host". This is
   because DNS lookup consumes a file descriptor temporarily, and the failed
   allocation is interpreted as NXDOMAIN.

2. In Go 1.1, or if you didn't trigger the case above in 1.0, `Dial` will simply
   return an error rather than a connection.

The first case in particular is rather misleading, but it's almost always an
indicator that you are out of File Descriptors.

So how can you moderate your FD consumption in the above example? There are two
simple ways:

1. Worker Pool

2. Semaphore

The Worker Pool strategy is reminiscent of the ruby example above:

```go
const nWorkers = 100
func crawl(urlProducer chan string) {
  for i := 0; i < nWorkers; i++ {
    go func() {
      for {
        fetch(<-urlProducer)
      }
    }()
  }
}
```

The outer loop runs `nWorkers` times, and spawns `nWorkers` goroutines. Each
loops indefinitely, calling fetch synchronously with a URL from the input
channel. This limits the concurrency to `nWorkers`, meaning that you will never
have more than 100 HTTP requests pending, and never be consuming more than 100
FDs as a result of this function.

Using a semaphore is a simliar strategy, but I feel it's cleaner in certain
circumstances:

```go
const nTokens = 100
func crawl(urlProducer chan string) {
  sem := make(chan bool, nTokens)
  for i := 0, i < nTokens; i++ {
    sem <- true
  }
  for url := range(urlProducer) {
    go func() {
      <- sem
      defer sem <- true

      fetch(<-urlProducer)
    }
  }()
}
```

You first pre-fill the semaphore with 100 tokens. Each time you want to start
an HTTP request, you must first withdraw a token. If all 100 tokens are
currently out, you must wait until one has been returned.

The semaphore strategy accomplishes largely the same performance
characteristics as the worker pool strategy, but one or the other can feel more
appropriate depending on the problem. It's good to be familiar with both. Both
will prevent you from exceeding the file descriptor limit (which, incidentally,
you can adjust using `ulimit`).


