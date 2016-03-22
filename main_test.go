package issue14790_test

import (
	"syscall"
	"testing"
)

var (
	modwinmm    = syscall.NewLazyDLL("winmm.dll")
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")

	proctimeBeginPeriod = modwinmm.NewProc("timeBeginPeriod")
	proctimeEndPeriod   = modwinmm.NewProc("timeEndPeriod")

	procCreateEvent = modkernel32.NewProc("CreateEventW")
	procSetEvent    = modkernel32.NewProc("SetEvent")
)

func timeBeginPeriod(period uint32) {
	syscall.Syscall(proctimeBeginPeriod.Addr(), 1, uintptr(period), 0, 0)
}

func timeEndPeriod(period uint32) {
	syscall.Syscall(proctimeEndPeriod.Addr(), 1, uintptr(period), 0, 0)
}

func createEvent() (syscall.Handle, error) {
	r0, _, e0 := syscall.Syscall6(procCreateEvent.Addr(), 4, 0, 0, 0, 0, 0, 0)
	if r0 == 0 {
		return 0, syscall.Errno(e0)
	}
	return syscall.Handle(r0), nil
}

func setEvent(h syscall.Handle) {
	r0, _, e0 := syscall.Syscall(procSetEvent.Addr(), 1, uintptr(h), 0, 0)
	if r0 == 0 {
		panic(syscall.Errno(e0))
	}
}

func bench(b *testing.B) {
	ch := make(chan int)
	event, err := createEvent()
	if err != nil {
		b.Fatal(err)
	}
	go func() {
		for {
			syscall.WaitForSingleObject(event, syscall.INFINITE)
			ch <- 1
		}
	}()
	for i := 0; i < b.N; i++ {
		setEvent(event)
		<-ch
	}
}

func BenchmarkDefaultResolution(b *testing.B) {
	bench(b)
}

func Benchmark1ms(b *testing.B) {
	timeBeginPeriod(1)
	bench(b)
	timeEndPeriod(1)
}

func benchNoChannel(b *testing.B) {
	event1, err := createEvent()
	if err != nil {
		b.Fatal(err)
	}
	event2, err := createEvent()
	if err != nil {
		b.Fatal(err)
	}
	go func() {
		for {
			syscall.WaitForSingleObject(event1, syscall.INFINITE)
			setEvent(event2)
		}
	}()
	for i := 0; i < b.N; i++ {
		setEvent(event1)
		syscall.WaitForSingleObject(event2, syscall.INFINITE)
	}
}

func BenchmarkNoChannelDefaultResolution(b *testing.B) {
	benchNoChannel(b)
}

func BenchmarkNoChannel1ms(b *testing.B) {
	timeBeginPeriod(1)
	benchNoChannel(b)
	timeEndPeriod(1)
}

func benchOnlyChannel(b *testing.B) {
	ch1 := make(chan int)
	ch2 := make(chan int)
	go func() {
		for {
			<-ch1
			ch2 <- 1
		}
	}()
	for i := 0; i < b.N; i++ {
		ch1 <- 1
		<-ch2
	}
}

func BenchmarkOnlyChannelDefaultResolution(b *testing.B) {
	benchOnlyChannel(b)
}

func BenchmarkOnlyChannel1ms(b *testing.B) {
	timeBeginPeriod(1)
	benchOnlyChannel(b)
	timeEndPeriod(1)
}
