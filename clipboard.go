// Copyright (c) 2013 Ato Araki. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//   - Redistributions of source code must retain the above copyright
//
// notice, this list of conditions and the following disclaimer.
//   - Redistributions in binary form must reproduce the above
//
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//   - Neither the name of @atotto. nor the names of its
//
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
package main

import (
	"context"
	"fmt"
	"runtime"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"
)

const (
	textFormat   = 13
	gmemMoveable = 0x0002
)

var (
	user32                     = syscall.MustLoadDLL("user32")
	getClipboardSequenceNumber = user32.MustFindProc("GetClipboardSequenceNumber")
	isClipboardFormatAvailable = user32.MustFindProc("IsClipboardFormatAvailable")
	openClipboard              = user32.MustFindProc("OpenClipboard")
	closeClipboard             = user32.MustFindProc("CloseClipboard")
	getClipboardData           = user32.MustFindProc("GetClipboardData")
	emptyClipboard             = user32.MustFindProc("EmptyClipboard")
	setClipboardData           = user32.MustFindProc("SetClipboardData")

	kernel32     = syscall.NewLazyDLL("kernel32")
	globalAlloc  = kernel32.NewProc("GlobalAlloc")
	globalFree   = kernel32.NewProc("GlobalFree")
	globalLock   = kernel32.NewProc("GlobalLock")
	globalUnlock = kernel32.NewProc("GlobalUnlock")
	lstrcpy      = kernel32.NewProc("lstrcpyW")
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

func writeClipboardString(data string) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for {
		ok, _, _ := openClipboard.Call()
		if ok == 0 {
			time.Sleep(time.Millisecond)
			continue
		}
		break
	}
	defer closeClipboard.Call()

	r, _, err := emptyClipboard.Call(0)
	if r == 0 {
		err = fmt.Errorf("failed to empty clipboard: %w", err)
		return
	}

	rawData, err := syscall.UTF16FromString(data)
	if err != nil {
		err = fmt.Errorf("failed to encode clipboard data: %w", err)
		return
	}

	hMem, _, err := globalAlloc.Call(gmemMoveable, uintptr(len(rawData)*int(unsafe.Sizeof(rawData[0]))))
	if hMem == 0 {
		err = fmt.Errorf("failed to allocate clipboard memory: %w", err)
		return
	}

	l, _, err := globalLock.Call(hMem)
	if l == 0 {
		err = fmt.Errorf("failed to lock clipboard memory: %w", err)
		globalFree.Call(hMem)
		return
	}

	r, _, err = lstrcpy.Call(l, uintptr(unsafe.Pointer(&rawData[0])))
	if r == 0 {
		err = fmt.Errorf("failed to allocate clipboard memory: %w", err)
		defer globalFree.Call(hMem)
		return
	}

	r, _, err = globalUnlock.Call(hMem)
	if r == 0 {
		if err.(syscall.Errno) != 0 {
			err = fmt.Errorf("failed to unlock clipboard memory: %w", err)
			globalFree.Call(hMem)
			return
		}
	}

	r, _, err = setClipboardData.Call(textFormat, hMem)
	if r == 0 {
		err = fmt.Errorf("failed to allocate clipboard memory: %w", err)
		defer globalFree.Call(hMem)
		return
	} else {
		err = nil
	}
	return
}

func writeAll(text string) error {
	// LockOSThread ensure that the whole method will keep executing on the same thread from begin to end (it actually locks the goroutine thread attribution).
	// Otherwise if the goroutine switch thread during execution (which is a common practice), the OpenClipboard and CloseClipboard will happen on two different threads, and it will result in a clipboard deadlock.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for {
		ok, _, _ := openClipboard.Call()
		if ok == 0 {
			time.Sleep(time.Millisecond)
			continue
		}
		break
	}

	r, _, err := emptyClipboard.Call(0)
	if r == 0 {
		_, _, _ = closeClipboard.Call()
		return err
	}

	data := syscall.StringToUTF16(text)

	// "If the hMem parameter identifies a memory object, the object must have
	// been allocated using the function with the GMEM_MOVEABLE flag."
	h, _, err := globalAlloc.Call(gmemMoveable, uintptr(len(data)*int(unsafe.Sizeof(data[0]))))
	if h == 0 {
		_, _, _ = closeClipboard.Call()
		return err
	}
	defer func() {
		if h != 0 {
			globalFree.Call(h)
		}
	}()

	l, _, err := globalLock.Call(h)
	if l == 0 {
		_, _, _ = closeClipboard.Call()
		return err
	}

	r, _, err = lstrcpy.Call(l, uintptr(unsafe.Pointer(&data[0])))
	if r == 0 {
		_, _, _ = closeClipboard.Call()
		return err
	}

	r, _, err = globalUnlock.Call(h)
	if r == 0 {
		if err.(syscall.Errno) != 0 {
			_, _, _ = closeClipboard.Call()
			return err
		}
	}

	r, _, err = setClipboardData.Call(textFormat, h)
	if r == 0 {
		_, _, _ = closeClipboard.Call()
		return err
	}
	h = 0 // suppress deferred cleanup
	closed, _, err := closeClipboard.Call()
	if closed == 0 {
		return err
	}
	return nil
}
