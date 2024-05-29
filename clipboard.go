package main

import (
	"context"
	"runtime"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"
)

const (
	textFormat = 13
)

var (
	user32                     = syscall.MustLoadDLL("user32")
	getClipboardSequenceNumber = user32.MustFindProc("GetClipboardSequenceNumber")
	isClipboardFormatAvailable = user32.MustFindProc("IsClipboardFormatAvailable")
	openClipboard              = user32.MustFindProc("OpenClipboard")
	closeClipboard             = user32.MustFindProc("CloseClipboard")
	getClipboardData           = user32.MustFindProc("GetClipboardData")

	kernel32     = syscall.NewLazyDLL("kernel32")
	globalLock   = kernel32.NewProc("GlobalLock")
	globalUnlock = kernel32.NewProc("GlobalUnlock")
)

// watchClipboard scans the clipboard for new data every interval and notifies the channel when new data is received
func watchClipboard(ctx context.Context, interval time.Duration) (events chan string) {
	events = make(chan string)
	go func() {
		ticker := time.NewTicker(interval)
		oldVersion, _, _ := getClipboardSequenceNumber.Call()
		for {
			<-ticker.C
			err := ctx.Err()
			if err != nil {
				close(events)
				return
			}
			currentVersion, _, _ := getClipboardSequenceNumber.Call()
			if oldVersion != currentVersion {
				oldVersion = currentVersion
				clipboardString := readClipboardString()
				if clipboardString != "" {
					events <- clipboardString
				}
			}
		}
	}()
	return
}

// readClipboardString reads the Windows clipboard data into program memory,
// assuming that the clipboard contains raw text
func readClipboardString() (data string) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ok, _, _ := isClipboardFormatAvailable.Call(textFormat)
	if ok == 0 {
		// Not ok, no data
		return
	}

	for {
		ok, _, _ = openClipboard.Call()
		if ok == 0 {
			time.Sleep(time.Millisecond)
			continue
		}
		break
	}
	defer closeClipboard.Call()

	hMem, _, _ := getClipboardData.Call(textFormat)
	if hMem == 0 {
		return
	}
	p, _, _ := globalLock.Call(hMem)
	if p == 0 {
		return
	}
	defer globalUnlock.Call(hMem)

	rawData := make([]uint16, 0)
	for i := 0; true; i++ {
		char := *(*uint16)(unsafe.Pointer(p + uintptr(i)*unsafe.Sizeof(uint16(0))))
		if char == 0 {
			break
		}
		rawData = append(rawData, char)
	}

	return string(utf16.Decode(rawData))
}
