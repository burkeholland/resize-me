//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	hotkeyID = 0x524d

	wmDestroy        = 0x0002
	wmCommand        = 0x0111
	wmHotkey         = 0x0312
	wmUser           = 0x0400
	wmTrayIcon       = wmUser + 1
	wmApplyHotkey    = wmUser + 2 // dispatches hotkey registration to the message-loop thread
	wmShowMenu       = wmUser + 3 // dispatches tray menu display to the message-loop thread
	wmLButtonUp      = 0x0202
	wmRButtonUp      = 0x0205
	wmContextMenu    = 0x007B

	modAlt      = 0x0001
	modControl  = 0x0002
	modShift    = 0x0004
	modWin      = 0x0008
	modNoRepeat = 0x4000

	nimAdd    = 0x00000000
	nimModify = 0x00000001
	nimDelete = 0x00000002

	nifMessage = 0x00000001
	nifIcon    = 0x00000002
	nifTip     = 0x00000004
	nifInfo    = 0x00000010

	niifNone    = 0x00000000
	niifWarning = 0x00000002

	mfString    = 0x00000000
	mfSeparator = 0x00000800
	mfChecked   = 0x00000008
	mfDisabled  = 0x00000002

	tpmRightButton = 0x0002
	tpmReturnCmd   = 0x0100
	tpmNonotify    = 0x0080

	swpNoZOrder   = 0x0004
	swpNoActivate = 0x0010

	swRestore = 9

	monitorDefaultToNearest = 0x00000002

	cmdPresetBase = 1000
	cmdCenter     = 2000
	cmdSettings   = 2002
	cmdQuit       = 2003
)

var (
	user32  = windows.NewLazySystemDLL("user32.dll")
	shell32 = windows.NewLazySystemDLL("shell32.dll")
	kernel  = windows.NewLazySystemDLL("kernel32.dll")

	procRegisterClassEx       = user32.NewProc("RegisterClassExW")
	procCreateWindowEx        = user32.NewProc("CreateWindowExW")
	procDefWindowProc         = user32.NewProc("DefWindowProcW")
	procDestroyWindow         = user32.NewProc("DestroyWindow")
	procGetMessage            = user32.NewProc("GetMessageW")
	procTranslateMessage      = user32.NewProc("TranslateMessage")
	procDispatchMessage       = user32.NewProc("DispatchMessageW")
	procPostMessage           = user32.NewProc("PostMessageW")
	procPostQuitMessage       = user32.NewProc("PostQuitMessage")
	procRegisterHotKey        = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey      = user32.NewProc("UnregisterHotKey")
	procLoadIcon              = user32.NewProc("LoadIconW")
	procExtractIconEx         = shell32.NewProc("ExtractIconExW")
	procShellNotifyIcon       = shell32.NewProc("Shell_NotifyIconW")
	procGetModuleHandle       = kernel.NewProc("GetModuleHandleW")
	procGetCursorPos          = user32.NewProc("GetCursorPos")
	procSetForegroundWindow   = user32.NewProc("SetForegroundWindow")
	procCreatePopupMenu       = user32.NewProc("CreatePopupMenu")
	procAppendMenu            = user32.NewProc("AppendMenuW")
	procTrackPopupMenu        = user32.NewProc("TrackPopupMenu")
	procDestroyMenu           = user32.NewProc("DestroyMenu")
	procGetForegroundWindow   = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcID = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible       = user32.NewProc("IsWindowVisible")
	procGetWindowText         = user32.NewProc("GetWindowTextW")
	procGetClassName          = user32.NewProc("GetClassNameW")
	procGetWindowRect         = user32.NewProc("GetWindowRect")
	procSetWindowPos          = user32.NewProc("SetWindowPos")
	procMonitorFromWindow     = user32.NewProc("MonitorFromWindow")
	procGetMonitorInfo        = user32.NewProc("GetMonitorInfoW")
	procIsIconic              = user32.NewProc("IsIconic")
	procIsZoomed              = user32.NewProc("IsZoomed")
	procShowWindow            = user32.NewProc("ShowWindow")

	trayMu       sync.RWMutex
	activeAgent  *WindowsAgent
	windowProcCB = syscall.NewCallback(windowProc)
)

type hotkeyReq struct {
	mods, vk uint32
	result    chan error
}

type WindowsAgent struct {
	app *App

	mu           sync.RWMutex
	config       Config
	presetByCmd  map[uint32]string
	hwnd         windows.Handle
	hIcon        windows.Handle
	currentMods  uint32
	currentVK    uint32
	hotkeyActive bool
	stopped      bool
	hotkeyCh     chan hotkeyReq // dispatches RegisterHotKey to the message-loop OS thread
}

type point struct {
	X int32
	Y int32
}

type rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type message struct {
	Hwnd    windows.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
}

type wndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     windows.Handle
	HIcon         windows.Handle
	HCursor       windows.Handle
	HbrBackground windows.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       windows.Handle
}

type notifyIconData struct {
	CbSize            uint32
	HWnd              windows.Handle
	UID               uint32
	UFlags            uint32
	UCallbackMessage  uint32
	HIcon             windows.Handle
	SzTip             [128]uint16
	DwState           uint32
	DwStateMask       uint32
	SzInfo            [256]uint16
	UTimeoutOrVersion uint32
	SzInfoTitle       [64]uint16
	DwInfoFlags       uint32
	GuidItem          windows.GUID
	HBalloonIcon      windows.Handle
}

type monitorInfo struct {
	CbSize    uint32
	RcMonitor rect
	RcWork    rect
	DwFlags   uint32
}

func NewPlatformAgent(app *App) PlatformAgent {
	app.mu.RLock()
	config := app.config.Clone()
	app.mu.RUnlock()
	return &WindowsAgent{
		app:        app,
		config:     config,
		presetByCmd: map[uint32]string{},
		hotkeyCh:   make(chan hotkeyReq, 1),
	}
}

func (w *WindowsAgent) Start() error {
	started := make(chan error, 1)
	go w.run(started)
	if err := <-started; err != nil {
		return err
	}
	return w.ApplySettings(w.config)
}

func (w *WindowsAgent) Stop() {
	w.mu.Lock()
	if w.stopped {
		w.mu.Unlock()
		return
	}
	w.stopped = true
	hwnd := w.hwnd
	w.mu.Unlock()

	if hwnd != 0 {
		w.deleteTrayIcon()
		_, _, _ = procUnregisterHotKey.Call(uintptr(hwnd), hotkeyID)
		_, _, _ = procDestroyWindow.Call(uintptr(hwnd))
	}
}

func (w *WindowsAgent) ApplySettings(config Config) error {
	mods, vk, err := parseHotkey(config.Hotkey)
	if err != nil {
		return err
	}

	w.mu.Lock()
	hwnd := w.hwnd
	needsHotkey := !w.hotkeyActive || w.currentMods != mods || w.currentVK != vk
	w.mu.Unlock()

	// RegisterHotKey must run on the OS thread that owns the hidden window.
	// We send the request through hotkeyCh and wake the message loop via PostMessage.
	if hwnd != 0 && needsHotkey {
		req := hotkeyReq{mods: mods, vk: vk, result: make(chan error, 1)}
		w.hotkeyCh <- req
		_, _, _ = procPostMessage.Call(uintptr(hwnd), wmApplyHotkey, 0, 0)
		if err := <-req.result; err != nil {
			return err
		}
	}

	// Only write autostart after the hotkey is confirmed working.
	if err := setAutoStart(config.AutoStart); err != nil {
		return err
	}

	w.mu.Lock()
	w.config = config.Clone()
	w.mu.Unlock()

	w.updateTrayIcon()
	return nil
}

func (w *WindowsAgent) run(started chan<- error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	trayMu.Lock()
	activeAgent = w
	trayMu.Unlock()
	defer func() {
		trayMu.Lock()
		if activeAgent == w {
			activeAgent = nil
		}
		trayMu.Unlock()
	}()

	hwnd, err := createHiddenWindow()
	if err != nil {
		started <- err
		return
	}

	// Extract the custom app icon embedded in this exe at the small tray size.
	// ExtractIconEx is more reliable than LoadIcon+resource-ID lookup since it
	// works regardless of which resource ID Wails assigns to the icon group.
	var hSmallIcon windows.Handle
	exePath, _ := os.Executable()
	lpszFile, _ := windows.UTF16PtrFromString(exePath)
	count, _, _ := procExtractIconEx.Call(
		uintptr(unsafe.Pointer(lpszFile)),
		0,                                   // first icon group, index 0
		0,                                   // skip large icon
		uintptr(unsafe.Pointer(&hSmallIcon)),
		1,
	)
	var hIcon windows.Handle
	if count > 0 && hSmallIcon != 0 {
		hIcon = hSmallIcon
	} else {
		// Fallback: generic Windows application icon
		icon, _, _ := procLoadIcon.Call(0, 32512)
		hIcon = windows.Handle(icon)
	}
	w.mu.Lock()
	w.hwnd = hwnd
	w.hIcon = hIcon
	w.mu.Unlock()

	if err := w.addTrayIcon(); err != nil {
		started <- err
		return
	}

	started <- nil

	var msg message
	for {
		ret, _, _ := procGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) == -1 || ret == 0 {
			return
		}
		_, _, _ = procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		_, _, _ = procDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func createHiddenWindow() (windows.Handle, error) {
	className, err := windows.UTF16PtrFromString("ResizeMeTrayWindow")
	if err != nil {
		return 0, err
	}
	instance, _, _ := procGetModuleHandle.Call(0)
	wc := wndClassEx{
		CbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		LpfnWndProc:   windowProcCB,
		HInstance:     windows.Handle(instance),
		LpszClassName: className,
	}
	atom, _, registerErr := procRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 && registerErr != windows.ERROR_CLASS_ALREADY_EXISTS {
		return 0, fmt.Errorf("register tray window: %w", registerErr)
	}

	hwnd, _, createErr := procCreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(className)),
		0,
		0, 0, 0, 0,
		0,
		0,
		instance,
		0,
	)
	if hwnd == 0 {
		return 0, fmt.Errorf("create tray window: %w", createErr)
	}
	return windows.Handle(hwnd), nil
}

func windowProc(hwnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	trayMu.RLock()
	agent := activeAgent
	trayMu.RUnlock()

	if agent != nil {
		switch msg {
		case wmApplyHotkey:
			// Drain one pending hotkey request and register it on this OS thread.
			select {
			case req := <-agent.hotkeyCh:
				req.result <- agent.registerHotkey(req.mods, req.vk)
			default:
			}
			return 0
		case wmShowMenu:
			// TrackPopupMenu must run on the thread that owns the window.
			agent.showMenu()
			return 0
		case wmTrayIcon:
			if lParam == wmLButtonUp || lParam == wmRButtonUp || lParam == wmContextMenu {
				_, _, _ = procPostMessage.Call(uintptr(hwnd), wmShowMenu, 0, 0)
				return 0
			}
		case wmHotkey:
			if wParam == hotkeyID {
				go func() {
					if err := agent.app.ResizeNow(); err != nil {
						agent.Notify("ResizeMe", err.Error(), true)
					}
				}()
				return 0
			}
		case wmCommand:
			go agent.handleCommand(uint32(wParam & 0xffff))
			return 0
		case wmDestroy:
			agent.deleteTrayIcon()
			_, _, _ = procUnregisterHotKey.Call(hwnd, hotkeyID)
			_, _, _ = procPostQuitMessage.Call(0)
			return 0
		}
	}

	ret, _, _ := procDefWindowProc.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func (w *WindowsAgent) registerHotkey(mods uint32, vk uint32) error {
	w.mu.RLock()
	hwnd := w.hwnd
	prevMods := w.currentMods
	prevVK := w.currentVK
	wasActive := w.hotkeyActive
	w.mu.RUnlock()
	if hwnd == 0 {
		return nil
	}

	// Temporarily unregister to free the ID so we can re-register.
	// We'll restore the old hotkey if the new one fails.
	if wasActive {
		_, _, _ = procUnregisterHotKey.Call(uintptr(hwnd), hotkeyID)
	}

	ret, _, err := procRegisterHotKey.Call(uintptr(hwnd), hotkeyID, uintptr(mods|modNoRepeat), uintptr(vk))
	if ret == 0 {
		// New hotkey failed — restore the previous one so the user isn't left
		// without a working hotkey.
		if wasActive {
			_, _, _ = procRegisterHotKey.Call(uintptr(hwnd), hotkeyID, uintptr(prevMods|modNoRepeat), uintptr(prevVK))
		}
		return fmt.Errorf("register hotkey %s: %w", w.config.Hotkey, err)
	}

	w.mu.Lock()
	w.currentMods = mods
	w.currentVK = vk
	w.hotkeyActive = true
	w.mu.Unlock()
	return nil
}

func parseHotkey(value string) (uint32, uint32, error) {
	parts := strings.Split(normalizeHotkeyText(value), "+")
	var mods uint32
	var key string
	for _, part := range parts {
		switch part {
		case "Ctrl":
			mods |= modControl
		case "Alt":
			mods |= modAlt
		case "Shift":
			mods |= modShift
		case "Win":
			mods |= modWin
		default:
			key = part
		}
	}
	if mods == 0 || key == "" {
		return 0, 0, fmt.Errorf("hotkey must include at least one modifier and one key")
	}
	if len(key) == 1 {
		ch := key[0]
		if (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			return mods, uint32(ch), nil
		}
	}
	if strings.HasPrefix(key, "F") {
		number, err := strconv.Atoi(strings.TrimPrefix(key, "F"))
		if err == nil && number >= 1 && number <= 24 {
			return mods, uint32(0x70 + number - 1), nil
		}
	}
	return 0, 0, fmt.Errorf("unsupported hotkey key %q", key)
}

func (w *WindowsAgent) addTrayIcon() error {
	w.mu.RLock()
	hwnd := w.hwnd
	hIcon := w.hIcon
	w.mu.RUnlock()

	nid := notifyIconData{
		CbSize:           uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:             hwnd,
		UID:              1,
		UFlags:           nifMessage | nifIcon | nifTip,
		UCallbackMessage: wmTrayIcon,
		HIcon:            hIcon,
	}
	copyUTF16(nid.SzTip[:], w.tooltip())
	ret, _, err := procShellNotifyIcon.Call(nimAdd, uintptr(unsafe.Pointer(&nid)))
	if ret == 0 {
		return fmt.Errorf("add tray icon: %w", err)
	}
	return nil
}

func (w *WindowsAgent) updateTrayIcon() {
	w.mu.RLock()
	hwnd := w.hwnd
	hIcon := w.hIcon
	w.mu.RUnlock()
	if hwnd == 0 {
		return
	}

	nid := notifyIconData{
		CbSize:           uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:             hwnd,
		UID:              1,
		UFlags:           nifIcon | nifTip,
		UCallbackMessage: wmTrayIcon,
		HIcon:            hIcon,
	}
	copyUTF16(nid.SzTip[:], w.tooltip())
	_, _, _ = procShellNotifyIcon.Call(nimModify, uintptr(unsafe.Pointer(&nid)))
}

func (w *WindowsAgent) deleteTrayIcon() {
	w.mu.RLock()
	hwnd := w.hwnd
	w.mu.RUnlock()
	if hwnd == 0 {
		return
	}

	nid := notifyIconData{
		CbSize: uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:   hwnd,
		UID:    1,
	}
	_, _, _ = procShellNotifyIcon.Call(nimDelete, uintptr(unsafe.Pointer(&nid)))
}

func (w *WindowsAgent) tooltip() string {
	w.mu.RLock()
	config := w.config.Clone()
	w.mu.RUnlock()
	preset, ok := config.ActivePreset()
	if !ok {
		return "ResizeMe"
	}
	return fmt.Sprintf("ResizeMe - %s (%dx%d)", preset.Name, preset.Width, preset.Height)
}

func (w *WindowsAgent) Notify(title string, body string, warning bool) {
	w.mu.RLock()
	hwnd := w.hwnd
	hIcon := w.hIcon
	w.mu.RUnlock()
	if hwnd == 0 {
		return
	}

	flags := uint32(niifNone)
	if warning {
		flags = niifWarning
	}
	nid := notifyIconData{
		CbSize:      uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:        hwnd,
		UID:         1,
		UFlags:      nifInfo,
		HIcon:       hIcon,
		DwInfoFlags: flags,
	}
	copyUTF16(nid.SzInfoTitle[:], title)
	copyUTF16(nid.SzInfo[:], body)
	_, _, _ = procShellNotifyIcon.Call(nimModify, uintptr(unsafe.Pointer(&nid)))
}

func (w *WindowsAgent) showMenu() {
	w.mu.RLock()
	config := w.config.Clone()
	hwnd := w.hwnd
	w.mu.RUnlock()
	if hwnd == 0 {
		return
	}

	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer procDestroyMenu.Call(menu)

	presetByCmd := make(map[uint32]string, len(config.Presets))
	nextPresetCommand := uint32(cmdPresetBase)
	appendPreset := func(preset Preset) {
		flags := uint32(mfString)
		if preset.ID == config.ActivePresetID {
			flags |= mfChecked
		}
		command := nextPresetCommand
		nextPresetCommand++
		presetByCmd[command] = preset.ID
		appendMenu(menu, flags, command, fmt.Sprintf("%s  %dx%d", preset.Name, preset.Width, preset.Height))
	}

	favoriteSet := make(map[string]bool, len(config.FavoritePresetIDs))
	for _, id := range config.FavoritePresetIDs {
		favoriteSet[id] = true
	}

	hasFavorites := false
	for _, id := range config.FavoritePresetIDs {
		preset, ok := config.FindPreset(id)
		if !ok {
			continue
		}
		if !hasFavorites {
			appendMenu(menu, mfString|mfDisabled, 0, "Favorites")
			hasFavorites = true
		}
		appendPreset(preset)
	}

	otherPresets := make([]Preset, 0, len(config.Presets))
	for _, preset := range config.Presets {
		if favoriteSet[preset.ID] {
			continue
		}
		otherPresets = append(otherPresets, preset)
	}

	if hasFavorites && len(otherPresets) > 0 {
		appendMenu(menu, mfSeparator, 0, "")
		appendMenu(menu, mfString|mfDisabled, 0, "All Presets")
	} else if !hasFavorites {
		appendMenu(menu, mfString|mfDisabled, 0, "Presets")
	}
	for _, preset := range otherPresets {
		appendPreset(preset)
	}

	w.mu.Lock()
	w.presetByCmd = presetByCmd
	w.mu.Unlock()

	appendMenu(menu, mfSeparator, 0, "")
	appendMenu(menu, mfString|mfDisabled, 0, fmt.Sprintf("Hotkey: %s", config.Hotkey))
	appendMenu(menu, mfSeparator, 0, "")

	centerFlags := uint32(mfString)
	if config.CenterAfterResize {
		centerFlags |= mfChecked
	}
	appendMenu(menu, centerFlags, cmdCenter, "Center after resize")
	appendMenu(menu, mfSeparator, 0, "")
	appendMenu(menu, mfString, cmdSettings, "Settings...")
	appendMenu(menu, mfString, cmdQuit, "Quit ResizeMe")

	var pt point
	_, _, _ = procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	_, _, _ = procSetForegroundWindow.Call(uintptr(hwnd))
	cmd, _, _ := procTrackPopupMenu.Call(menu, tpmRightButton|tpmReturnCmd|tpmNonotify, uintptr(pt.X), uintptr(pt.Y), 0, uintptr(hwnd), 0)
	if cmd != 0 {
		// handleCommand calls Go/Wails methods — run off the message-loop thread.
		go w.handleCommand(uint32(cmd))
	}
}

func appendMenu(menu uintptr, flags uint32, command uint32, label string) {
	var labelPtr uintptr
	if label != "" {
		labelPtr = uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(label)))
	}
	_, _, _ = procAppendMenu.Call(menu, uintptr(flags), uintptr(command), labelPtr)
}

func (w *WindowsAgent) handleCommand(command uint32) {
	w.mu.RLock()
	config := w.config.Clone()
	presetID, isPreset := w.presetByCmd[command]
	w.mu.RUnlock()

	switch {
	case isPreset:
		if _, err := w.app.SetActivePreset(presetID); err != nil {
			w.Notify("ResizeMe", err.Error(), true)
		}
	case command == cmdCenter:
		if _, err := w.app.SetCenterAfterResize(!config.CenterAfterResize); err != nil {
			w.Notify("ResizeMe", err.Error(), true)
		}
	case command == cmdSettings:
		w.app.ShowSettings()
	case command == cmdQuit:
		w.app.Quit()
	}
}

func (w *WindowsAgent) ResizeActiveWindow(preset Preset, center bool) error {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return fmt.Errorf("no active window to resize")
	}

	w.mu.RLock()
	agentHwnd := w.hwnd
	w.mu.RUnlock()
	if windows.Handle(hwnd) == agentHwnd {
		return fmt.Errorf("ResizeMe cannot resize its own tray window")
	}

	var pid uint32
	_, _, _ = procGetWindowThreadProcID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == uint32(os.Getpid()) {
		return fmt.Errorf("ResizeMe settings cannot be resized")
	}

	visible, _, _ := procIsWindowVisible.Call(hwnd)
	if visible == 0 {
		return fmt.Errorf("the active window is not resizable")
	}

	className := getWindowClass(windows.Handle(hwnd))
	switch className {
	case "Progman", "WorkerW", "Shell_TrayWnd":
		return fmt.Errorf("the Windows desktop or taskbar cannot be resized")
	}

	if ret, _, _ := procIsIconic.Call(hwnd); ret != 0 {
		_, _, _ = procShowWindow.Call(hwnd, swRestore)
	}
	if ret, _, _ := procIsZoomed.Call(hwnd); ret != 0 {
		_, _, _ = procShowWindow.Call(hwnd, swRestore)
	}

	var current rect
	if ret, _, err := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&current))); ret == 0 {
		return fmt.Errorf("read active window bounds: %w", err)
	}

	x := current.Left
	y := current.Top
	if center {
		workArea, err := monitorWorkArea(windows.Handle(hwnd))
		if err != nil {
			return err
		}
		x = workArea.Left + ((workArea.Right-workArea.Left)-int32(preset.Width))/2
		y = workArea.Top + ((workArea.Bottom-workArea.Top)-int32(preset.Height))/2
	}

	ret, _, err := procSetWindowPos.Call(
		hwnd,
		0,
		uintptr(int32ToUintptr(x)),
		uintptr(int32ToUintptr(y)),
		uintptr(preset.Width),
		uintptr(preset.Height),
		swpNoZOrder|swpNoActivate,
	)
	if ret == 0 {
		title := getWindowTitle(windows.Handle(hwnd))
		if title == "" {
			title = "the active window"
		}
		return fmt.Errorf("could not resize %s: %w", title, err)
	}
	return nil
}

func monitorWorkArea(hwnd windows.Handle) (rect, error) {
	monitor, _, _ := procMonitorFromWindow.Call(uintptr(hwnd), monitorDefaultToNearest)
	if monitor == 0 {
		return rect{}, fmt.Errorf("could not find the active window's monitor")
	}
	info := monitorInfo{CbSize: uint32(unsafe.Sizeof(monitorInfo{}))}
	if ret, _, err := procGetMonitorInfo.Call(monitor, uintptr(unsafe.Pointer(&info))); ret == 0 {
		return rect{}, fmt.Errorf("read monitor bounds: %w", err)
	}
	return info.RcWork, nil
}

func getWindowTitle(hwnd windows.Handle) string {
	buffer := make([]uint16, 256)
	ret, _, _ := procGetWindowText.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	if ret == 0 {
		return ""
	}
	return windows.UTF16ToString(buffer[:ret])
}

func getWindowClass(hwnd windows.Handle) string {
	buffer := make([]uint16, 256)
	ret, _, _ := procGetClassName.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	if ret == 0 {
		return ""
	}
	return windows.UTF16ToString(buffer[:ret])
}

func setAutoStart(enabled bool) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open startup registration: %w", err)
	}
	defer key.Close()

	if !enabled {
		if err := key.DeleteValue("ResizeMe"); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove startup registration: %w", err)
		}
		return nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("read executable path: %w", err)
	}
	if err := key.SetStringValue("ResizeMe", fmt.Sprintf("%q", exePath)); err != nil {
		return fmt.Errorf("write startup registration: %w", err)
	}
	return nil
}

func copyUTF16(target []uint16, value string) {
	encoded := windows.StringToUTF16(value)
	if len(encoded) > len(target) {
		encoded = encoded[:len(target)]
		encoded[len(encoded)-1] = 0
	}
	copy(target, encoded)
}

func int32ToUintptr(value int32) uintptr {
	return uintptr(uint32(value))
}
